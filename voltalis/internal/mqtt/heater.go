package mqtt

import (
	"fmt"
)

// InstanciateVoltalisHeaterClimate crée une nouvelle configuration d'entité climate pour un radiateur Voltalis
func (c *Client) InstanciateHeater(id int64, name string) (Heater, error) {
	payload := HeaterConfigPayload{
		ActionTopic:      newHeaterTopic[WriteTopic](id, "action"),
		UniqueID:         fmt.Sprintf("voltalis_heater_%d", id),
		Name:             "Temperature",
		CommandTopic:     newHeaterTopic[WriteTopic](id, "set"),
		ModeStateTopic:   newHeaterTopic[WriteTopic](id, "mode"),
		ModeCommandTopic: newHeaterTopic[ReadTopic](id, "mode"),
		PresetModes: []HeaterPresetMode{HeaterPresetModeHorsGel,
			HeaterPresetModeEco, HeaterPresetModeConfort},
		PresetModeCommandTopic:  newHeaterTopic[ReadTopic](id, "preset_mode"),
		PresetModeStateTopic:    newHeaterTopic[WriteTopic](id, "preset_mode"),
		TemperatureStateTopic:   newHeaterTopic[WriteTopic](id, "temp"),
		TemperatureCommandTopic: newHeaterTopic[ReadTopic](id, "temp"),
		MinTemp:                 15,
		MaxTemp:                 25,
		TempStep:                0.5,
		Modes:                   []HeaterMode{HeaterModeOff, HeaterModeAuto, HeaterModeHeat},
		CurrentTemperatureTopic: newHeaterTopic[WriteTopic](id, "current_temp"),
		Device: DeviceInfo{
			Identifiers:  []string{"voltalis_heater_" + fmt.Sprint(id)},
			Manufacturer: "Voltalis",
			Name:         "Radiateur " + name,
			Model:        "Radiateur voltalis",
			SwVersion:    "0.1.0",
		},
	}

	if err := c.PublishConfig(ComponentClimate, fmt.Sprintf("voltalis_heater_%d", id), payload); err != nil {
		return Heater{}, fmt.Errorf("failed to publish heater config: %w", err)
	}

	if err := c.PublishConfig(ComponentSelect, fmt.Sprintf("voltalis_select_%d", id), SelectConfigPayload{
		UniqueID:     fmt.Sprintf("voltalis_controller_select_%s", name),
		Name:         fmt.Sprintf("Controller Select %s", name),
		CommandTopic: payload.PresetModeCommandTopic,
		StateTopic:   payload.PresetModeStateTopic,
		Options: []string{
			string(HeaterPresetModeHorsGel),
			string(HeaterPresetModeEco),
			string(HeaterPresetModeConfort),
		},
		Device: payload.Device,
	}); err != nil {
		return Heater{}, fmt.Errorf("failed to publish heater select config: %w", err)
	}
	return Heater{
		heaterConfigPayload: payload,
		ReadTopics: HeaterReadTopics{
			Mode:        payload.ModeCommandTopic,
			PresetMode:  payload.PresetModeCommandTopic,
			Temperature: payload.TemperatureCommandTopic,
		},
		WriteTopics: HeaterWriteTopics{
			Action:             payload.ActionTopic,
			Mode:               payload.ModeStateTopic,
			PresetMode:         payload.PresetModeStateTopic,
			Temperature:        payload.TemperatureStateTopic,
			CurrentTemperature: payload.CurrentTemperatureTopic,
		},
	}, nil
}

type HeaterReadTopics struct {
	Mode        ReadTopic
	PresetMode  ReadTopic
	Temperature ReadTopic
}
type HeaterWriteTopics struct {
	Action             WriteTopic
	Mode               WriteTopic
	PresetMode         WriteTopic
	Temperature        WriteTopic
	CurrentTemperature WriteTopic
}

type Heater struct {
	heaterConfigPayload HeaterConfigPayload
	ReadTopics          HeaterReadTopics
	WriteTopics         HeaterWriteTopics
}

func newHeaterTopic[T Topic](id int64, suffix string) T {
	return T(fmt.Sprintf("voltalis/heater/%d/%s", id, suffix))
}
