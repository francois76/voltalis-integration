package mqtt

import (
	"fmt"
	"log/slog"
)

func (c *Client) RegisterHeater(id int64, name string) error {
	payload := &HeaterConfigPayload{
		ActionTopic:      newHeaterTopic[GetTopic](id, "action"),
		UniqueID:         fmt.Sprintf("voltalis_heater_%d", id),
		Name:             "Temperature",
		CommandTopic:     newHeaterTopic[GetTopic](id, "set"),
		ModeStateTopic:   newHeaterTopic[GetTopic](id, "mode"),
		ModeCommandTopic: newHeaterTopic[SetTopic](id, "mode"),
		PresetModes: []HeaterPresetMode{HeaterPresetModeHorsGel,
			HeaterPresetModeEco, HeaterPresetModeConfort},
		PresetModeCommandTopic:  newHeaterTopic[SetTopic](id, "preset_mode"),
		PresetModeStateTopic:    newHeaterTopic[GetTopic](id, "preset_mode"),
		TemperatureStateTopic:   newHeaterTopic[GetTopic](id, "temp"),
		TemperatureCommandTopic: newHeaterTopic[SetTopic](id, "temp"),
		MinTemp:                 15,
		MaxTemp:                 25,
		TempStep:                0.5,
		Modes:                   []HeaterMode{HeaterModeOff, HeaterModeAuto, HeaterModeHeat},
		CurrentTemperatureTopic: newHeaterTopic[GetTopic](id, "current_temp"),
		Device: DeviceInfo{
			Identifiers:  []string{"voltalis_heater_" + fmt.Sprint(id)},
			Manufacturer: "Voltalis",
			Name:         "Radiateur " + name,
			Model:        "Radiateur voltalis",
			SwVersion:    "0.1.0",
		},
	}

	if err := c.PublishConfig(payload); err != nil {
		return fmt.Errorf("failed to publish heater config: %w", err)
	}
	selectPresetPayload := getPayloadSelectMode(payload.Device, PRESET_SELECT_ONE_HEATER...)
	// le select de preset est juste un remapping sur le climate. Donc on ne déclare pas de topic dédiés
	// (on écrase ceux qui sont créés par la méthode au dessus)
	selectPresetPayload.CommandTopic = payload.PresetModeCommandTopic
	selectPresetPayload.StateTopic = payload.PresetModeStateTopic
	if err := c.PublishConfig(selectPresetPayload); err != nil {
		return fmt.Errorf("failed to publish heater select config: %w", err)
	}

	durationPayload := getPayloadSelectDuration(payload.Device)
	if err := c.PublishConfig(durationPayload); err != nil {
		return fmt.Errorf("failed to publish heater duration config: %w", err)
	}

	statePayload := getPayloadDureeMode(payload.Device)
	if err := c.PublishConfig(statePayload); err != nil {
		return fmt.Errorf("failed to publish heater state config: %w", err)
	}
	c.PublishState(statePayload.StateTopic, "Initialisation de l'intégration voltalis...")
	c.ListenState(heater.SetTopics.Temperature, func(data string) {
	})

	c.ListenState(heater.SetTopics.PresetMode, func(data string) {
		c.recomputeState(heater, data)
	})

	c.ListenState(heater.SetTopics.Mode, func(data string) {
		switch HeaterMode(data) {
		case HeaterModeOff:
			c.recomputeState(heater, string(HeaterPresetModeNone))
			c.PublishState(heater.GetTopics.PresetMode, HeaterPresetModeNone)
		case HeaterModeAuto:
			lastPreset := c.GetState(heater.SetTopics.PresetMode)
			if lastPreset == string(HeaterPresetModeManuel) || lastPreset == string(HeaterPresetModeNone) {
				c.PublishState(heater.GetTopics.PresetMode, HeaterPresetModeConfort)
			}
		case HeaterModeHeat:
			c.recomputeState(heater, string(HeaterPresetModeNone))
			c.PublishState(heater.GetTopics.PresetMode, HeaterPresetModeNone)
		default:
			slog.Warn("Unknown mode received", "value", data)
		}
	})

	return nil
}

func (c *Client) recomputeState(data string) {
	slog.Info("Target preset mode received", "value", data)
	targetHeaterMode := HeaterModeAuto
	targetTemperature := TEMPERATURE_NONE
	targetAction := HeaterActionIdle

	switch HeaterPresetMode(data) {
	case HeaterPresetModeNone:
		// On cherche ici à distinguer 2 cas: soit on a manuellement retiré le preset, dans ce cas on bascule en mode manuel
		// soit on a mis le mode en off, et dans ce cas on ne fait rien
		lastMode := c.GetState(heater.SetTopics.Mode)
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
	c.PublishState(heater.GetTopics.Action, targetAction)
	c.PublishState(heater.GetTopics.Mode, targetHeaterMode)
	c.PublishState(heater.GetTopics.Temperature, targetTemperature)
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
	State              GetTopic
}

type Heater struct {
	SetTopics HeaterSetTopics
	GetTopics HeaterGetTopics
}

func newHeaterTopic[T Topic](id int64, suffix string) T {
	return newTopicName[T](fmt.Sprintf("heater/%d/%s", id, suffix))
}
