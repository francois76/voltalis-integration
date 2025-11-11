package mqtt

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/francois76/voltalis-integration/voltalis/internal/state"
)

func (c *Client) RegisterHeater(id int64, name string) error {
	heater := Heater{
		Client:    c,
		GetTopics: HeaterGetTopics{},
		SetTopics: HeaterSetTopics{},
	}
	c.StateManager.currentState.HeaterState[id] = state.HeaterState{}
	payload, err := heater.addClimate(id, name)
	if err != nil {
		return err
	}

	if err := heater.addSelectMode(payload); err != nil {
		return err
	}

	if err := heater.addSelectDuration(payload); err != nil {
		return err
	}

	if err := heater.addDurationState(payload, heater.GetTopics.SingleDuration); err != nil {
		return err
	}

	updateHeater := func(currentState *state.ResourceState, data string, heaterTreatment func(heaterState *state.HeaterState, data string)) {
		heaterState := currentState.HeaterState[id]
		heaterTreatment(&heaterState, data)
		currentState.HeaterState[id] = heaterState
	}

	heater.ListenState(heater.SetTopics.Temperature, func(currentState *state.ResourceState, data string) {
		updateHeater(currentState, data, func(heaterState *state.HeaterState, data string) {
			dataFloat, err := strconv.ParseFloat(data, 64)
			if err != nil {
				heaterState.Temperature = -1
			} else {
				heaterState.Temperature = dataFloat
			}
		})
	})
	heater.ListenState(heater.SetTopics.SingleDuration, func(currentState *state.ResourceState, data string) {
		updateHeater(currentState, data, func(heaterState *state.HeaterState, data string) {
			heaterState.Duration = data
		})
	})

	heater.ListenStateWithPreHook(heater.SetTopics.PresetMode, func(data string) {
		heater.recomputeState(data)
	}, func(currentState *state.ResourceState, data string) {
		fmt.Println("test")
		updateHeater(currentState, data, func(heaterState *state.HeaterState, data string) {
			heaterState.Mode = state.HeaterMode(data)
		})
	})

	heater.ListenStateWithPreHook(heater.SetTopics.Mode, func(data string) {
		switch HeaterMode(data) {
		case HeaterModeOff:
			heater.recomputeState(string(HeaterPresetModeNone))
			heater.PublishState(heater.GetTopics.PresetMode, HeaterPresetModeNone)
		case HeaterModeAuto:
			heater.recomputeState(string(HeaterPresetModeConfort))
			heater.PublishState(heater.GetTopics.PresetMode, HeaterPresetModeConfort)
		case HeaterModeHeat:
			heater.PublishState(heater.GetTopics.PresetMode, HeaterPresetModeNone)
			heater.recomputeState(string(HeaterPresetModeNone))
		default:
			slog.Warn("Unknown mode received", "value", data)
		}
	}, func(currentState *state.ResourceState, data string) {
		updateHeater(currentState, data, func(heaterState *state.HeaterState, data string) {
			if data == "auto" {
				heaterState.Mode = state.HeaterMode(heater.GetTopicState(heater.SetTopics.PresetMode))
			} else {
				heaterState.Mode = state.HeaterMode(data)
			}
		})
	})

	return nil
}

func (h *Heater) addDurationState(payload *ClimateConfigPayload, topic GetTopic) error {
	statePayload := getPayloadDureeMode(payload.Device, topic)
	if err := h.PublishConfig(statePayload); err != nil {
		return fmt.Errorf("failed to publish heater state config: %w", err)
	}
	h.PublishState(statePayload.StateTopic, "Initialisation de l'intégration voltalis...")
	return nil
}

func (h *Heater) addSelectDuration(payload *ClimateConfigPayload) error {
	durationPayload := getPayloadSelectDuration(payload.Device)
	if err := h.PublishConfig(durationPayload); err != nil {
		return fmt.Errorf("failed to publish heater duration config: %w", err)
	}
	h.GetTopics.SingleDuration = durationPayload.StateTopic
	h.SetTopics.SingleDuration = durationPayload.CommandTopic
	return nil
}

func (h *Heater) addSelectMode(payload *ClimateConfigPayload) error {
	selectPresetPayload := getPayloadSelectMode(payload.Device, PRESET_SELECT_ONE_HEATER...)
	// le select de preset est juste un remapping sur le climate. Donc on ne déclare pas de topic dédiés
	// (on écrase ceux qui sont créés par la méthode au dessus)
	selectPresetPayload.CommandTopic = payload.PresetModeCommandTopic
	selectPresetPayload.StateTopic = payload.PresetModeStateTopic
	if err := h.PublishConfig(selectPresetPayload); err != nil {
		return fmt.Errorf("failed to publish heater select config: %w", err)
	}
	return nil
}

func (h *Heater) addClimate(id int64, name string) (*ClimateConfigPayload, error) {
	payload := &ClimateConfigPayload{
		ClimateCommandPayload: h.buildClimateCommands(id),
		ClimateStatePayload:   h.buildClimateStates(id),
		ActionTopic:           NewHeaterTopic[GetTopic](id, "action"),
		UniqueID:              fmt.Sprintf("voltalis_heater_%d", id),
		Name:                  "Temperature",
		PresetModes: []HeaterPresetMode{HeaterPresetModeHorsGel,
			HeaterPresetModeEco, HeaterPresetModeConfort},
		MinTemp:  15,
		MaxTemp:  25,
		TempStep: 0.5,
		Modes:    []HeaterMode{HeaterModeOff, HeaterModeAuto, HeaterModeHeat},
		Device:   buildDeviceInfo(id, name),
	}
	if err := h.PublishConfig(payload); err != nil {
		return nil, fmt.Errorf("failed to publish heater config: %w", err)
	}
	h.GetTopics.Action = payload.ActionTopic
	h.GetTopics.Mode = payload.ModeStateTopic
	h.SetTopics.Mode = payload.ModeCommandTopic
	h.GetTopics.Temperature = payload.TemperatureStateTopic
	h.SetTopics.Temperature = payload.TemperatureCommandTopic
	h.SetTopics.PresetMode = payload.PresetModeCommandTopic
	h.GetTopics.PresetMode = payload.PresetModeStateTopic
	h.GetTopics.CurrentTemperature = payload.CurrentTemperatureTopic
	return payload, nil
}

func buildDeviceInfo(id int64, name string) DeviceInfo {
	return DeviceInfo{
		Identifiers:  []string{"voltalis_heater_" + fmt.Sprint(id)},
		Manufacturer: "Voltalis",
		Name:         "Radiateur " + name,
		Model:        "Radiateur voltalis",
		SwVersion:    "0.1.0",
	}
}

func (h *Heater) recomputeState(data string) {
	slog.Info("Target preset mode received", "value", data)
	targetHeaterMode := HeaterModeAuto
	targetTemperature := TEMPERATURE_NONE
	targetAction := HeaterActionIdle

	switch HeaterPresetMode(data) {
	case HeaterPresetModeNone:
		// On cherche ici à distinguer 2 cas: soit on a manuellement retiré le preset, dans ce cas on bascule en mode manuel
		// soit on a mis le mode en off, et dans ce cas on ne fait rien
		lastMode := h.GetTopicState(h.SetTopics.Mode)
		slog.Debug("Last mode read", "value", lastMode)
		if lastMode == string(HeaterModeOff) {
			targetAction = HeaterActionOff
			targetHeaterMode = HeaterModeOff
		} else {
			targetAction = HeaterActionHeating
			targetTemperature = "18"
			targetHeaterMode = HeaterModeHeat
		}
	case HeaterPresetModeHorsGel:
		targetAction = HeaterActionIdle
	case HeaterPresetModeEco:
		targetAction = HeaterActionCooling
	case HeaterPresetModeConfort:
		targetAction = HeaterActionHeating
	default:
		slog.Warn("Unknown preset mode received", "value", data)
	}
	h.PublishState(h.GetTopics.Action, targetAction)
	h.PublishState(h.GetTopics.Mode, targetHeaterMode)
	h.PublishState(h.GetTopics.Temperature, targetTemperature)
}

type HeaterSetTopics struct {
	Mode           SetTopic
	PresetMode     SetTopic
	Temperature    SetTopic
	SingleDuration SetTopic
}
type HeaterGetTopics struct {
	Action             GetTopic
	Mode               GetTopic
	PresetMode         GetTopic
	Temperature        GetTopic
	CurrentTemperature GetTopic
	SingleDuration     GetTopic
}

type Heater struct {
	*Client
	SetTopics HeaterSetTopics
	GetTopics HeaterGetTopics
}

func NewHeaterTopic[T Topic](id int64, suffix string) T {
	return newTopicName[T](fmt.Sprintf("heater/%d/%s", id, suffix))
}
