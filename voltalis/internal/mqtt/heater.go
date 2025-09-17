package mqtt

import (
	"fmt"
)

// InstanciateVoltalisHeaterClimate crée une nouvelle configuration d'entité climate pour un radiateur Voltalis
func (c *Client) InstanciateHeater(id int64, name string) (Heater, error) {
	payload := &HeaterConfigPayload{
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

	if err := c.PublishConfig(payload); err != nil {
		return Heater{}, fmt.Errorf("failed to publish heater config: %w", err)
	}
	selectPresetPayload := getPayloadSelectMode(payload.Device, PRESET_SELECT_ONE_HEATER...)
	// le select de preset est juste un remapping sur le climate. Donc on ne déclare pas de topic dédiés
	// (on écrase ceux qui sont créés par la méthode au dessus)
	selectPresetPayload.CommandTopic = payload.PresetModeCommandTopic
	selectPresetPayload.StateTopic = payload.PresetModeStateTopic
	if err := c.PublishConfig(selectPresetPayload); err != nil {
		return Heater{}, fmt.Errorf("failed to publish heater select config: %w", err)
	}

	durationPayload := getPayloadSelectDuration(payload.Device)
	if err := c.PublishConfig(durationPayload); err != nil {
		return Heater{}, fmt.Errorf("failed to publish heater duration config: %w", err)
	}

	return Heater{
		ReadTopics: HeaterReadTopics{
			Mode:           payload.ModeCommandTopic,
			PresetMode:     payload.PresetModeCommandTopic,
			Temperature:    payload.TemperatureCommandTopic,
			SingleDuration: durationPayload.CommandTopic,
		},
		WriteTopics: HeaterWriteTopics{
			Action:             payload.ActionTopic,
			Mode:               payload.ModeStateTopic,
			PresetMode:         payload.PresetModeStateTopic,
			Temperature:        payload.TemperatureStateTopic,
			CurrentTemperature: payload.CurrentTemperatureTopic,
			SingleDuration:     durationPayload.StateTopic,
		},
	}, nil
}

type HeaterReadTopics struct {
	Mode           ReadTopic
	PresetMode     ReadTopic
	Temperature    ReadTopic
	SingleDuration ReadTopic
}
type HeaterWriteTopics struct {
	Action             WriteTopic
	Mode               WriteTopic
	PresetMode         WriteTopic
	Temperature        WriteTopic
	CurrentTemperature WriteTopic
	SingleDuration     WriteTopic
}

type Heater struct {
	ReadTopics  HeaterReadTopics
	WriteTopics HeaterWriteTopics
}

func newHeaterTopic[T Topic](id int64, suffix string) T {
	return newTopicName[T](fmt.Sprintf("heater/%d/%s", id, suffix))
}
