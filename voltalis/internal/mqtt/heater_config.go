package mqtt

import (
	"fmt"
)

// InstanciateVoltalisHeaterClimate crée une nouvelle configuration d'entité climate pour un radiateur Voltalis
func InstanciateVoltalisHeaterClimate(id int64, name string) *HeaterConfigPayload {
	return &HeaterConfigPayload{
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
		Device:                  getHeaterDevice(id, name),
	}
}

func getHeaterDevice(id int64, name string) DeviceInfo {
	return DeviceInfo{
		Identifiers:  []string{"voltalis_heater_" + fmt.Sprint(id)},
		Manufacturer: "Voltalis",
		Name:         "Radiateur " + name,
		Model:        "Radiateur voltalis",
		SwVersion:    "0.1.0",
	}
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
type HeaterTopics struct {
	Read  HeaterReadTopics
	Write HeaterWriteTopics
}

func (p *HeaterConfigPayload) GetTopics() HeaterTopics {
	return HeaterTopics{
		Read: HeaterReadTopics{
			Mode:        p.ModeCommandTopic,
			PresetMode:  p.PresetModeCommandTopic,
			Temperature: p.TemperatureCommandTopic,
		},
		Write: HeaterWriteTopics{
			Mode:               p.ModeStateTopic,
			PresetMode:         p.PresetModeStateTopic,
			Temperature:        p.TemperatureStateTopic,
			CurrentTemperature: p.CurrentTemperatureTopic,
			Action:             p.ActionTopic,
		},
	}
}

func newHeaterTopic[T Topic](id int64, suffix string) T {
	return T(fmt.Sprintf("voltalis/heater/%d/%s", id, suffix))
}
