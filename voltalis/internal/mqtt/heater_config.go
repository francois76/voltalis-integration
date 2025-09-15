package mqtt

import (
	"fmt"
)

// DeviceInfo représente les informations du périphérique pour Home Assistant
type DeviceInfo struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SwVersion    string   `json:"sw_version"`
}

type HeaterConfigPayload struct {
	ActionTopic              WriteTopic         `json:"action_topic,omitempty"`
	Name                     string             `json:"name"`
	UniqueID                 string             `json:"unique_id"`
	CommandTopic             WriteTopic         `json:"command_topic"`
	ModeStateTopic           WriteTopic         `json:"mode_state_topic"`
	ModeCommandTopic         ReadTopic          `json:"mode_command_topic"`
	PresetModes              []HeaterPresetMode `json:"preset_modes,omitempty"`
	PresetModeCommandTopic   ReadTopic          `json:"preset_mode_command_topic,omitempty"`
	PresetModeStateTopic     WriteTopic         `json:"preset_mode_state_topic,omitempty"`
	TemperatureStateTopic    WriteTopic         `json:"temperature_state_topic"`
	TemperatureCommandTopic  ReadTopic          `json:"temperature_command_topic"`
	MinTemp                  float64            `json:"min_temp"`
	MaxTemp                  float64            `json:"max_temp"`
	TempStep                 float64            `json:"temp_step"`
	Modes                    []HeaterMode       `json:"modes"`
	CurrentTemperatureTopic  WriteTopic         `json:"current_temperature_topic"`
	Device                   DeviceInfo         `json:"device"`
	TemperatureStateTemplate string             `json:"temperature_state_template,omitempty"`
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

func InstanciateVoltalisHeaterBaseConfig(id int64) *HeaterConfigPayload {
	return &HeaterConfigPayload{
		ActionTopic:             newHeaterTopic[WriteTopic](id, "action"),
		UniqueID:                fmt.Sprintf("voltalis_heater_%d", id),
		Name:                    "Temperature",
		CommandTopic:            newHeaterTopic[WriteTopic](id, "set"),
		ModeStateTopic:          newHeaterTopic[WriteTopic](id, "mode"),
		ModeCommandTopic:        newHeaterTopic[ReadTopic](id, "mode"),
		PresetModes:             []HeaterPresetMode{HeaterPresetEco, HeaterPresetAway, HeaterPresetHome},
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
			Name:         "Radiateur",
			Model:        "Radiateur voltalis",
			SwVersion:    "0.1.0",
		},
	}
}

func (p *HeaterConfigPayload) WithName(name string) *HeaterConfigPayload {
	p.Device.Name = "Radiateur " + name
	return p
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
