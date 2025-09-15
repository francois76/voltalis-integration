package mqtt

import "fmt"

type DeviceInfo struct {
	Identifiers  []string `json:"identifiers"`
	name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SwVersion    string   `json:"sw_version"`
}

type HeaterConfigPayload struct {
	name                    string     `json:"name"`
	UniqueID                string     `json:"unique_id"`
	CommandTopic            Topic      `json:"command_topic"`
	ModeStateTopic          Topic      `json:"mode_state_topic"`
	ModeCommandTopic        Topic      `json:"mode_command_topic"`
	TemperatureStateTopic   Topic      `json:"temperature_state_topic"`
	TemperatureCommandTopic Topic      `json:"temperature_command_topic"`
	MinTemp                 float64    `json:"min_temp"`
	MaxTemp                 float64    `json:"max_temp"`
	TempStep                float64    `json:"temp_step"`
	Modes                   []string   `json:"modes"`
	CurrentTemperatureTopic Topic      `json:"current_temperature_topic"`
	Device                  DeviceInfo `json:"device"`
}

func InstanciateVoltalisHeaterBaseConfig(id int64) *HeaterConfigPayload {
	newTopic := func(suffix string) Topic {
		return Topic(fmt.Sprintf("voltalis/heater/%d/%s", id, suffix))
	}
	return &HeaterConfigPayload{
		UniqueID:                fmt.Sprintf("voltalis_heater_%d", id),
		CommandTopic:            newTopic("set"),
		ModeStateTopic:          newTopic("mode"),
		ModeCommandTopic:        newTopic("mode/set"),
		TemperatureStateTopic:   newTopic("temp"),
		TemperatureCommandTopic: newTopic("temp/set"),
		MinTemp:                 15,
		MaxTemp:                 25,
		TempStep:                0.5,
		Modes:                   []string{"off", "heat"},
		CurrentTemperatureTopic: newTopic("current_temp"),
		Device: DeviceInfo{
			Identifiers:  []string{"voltalis_heater_" + fmt.Sprint(id)},
			Manufacturer: "Voltalis",
			Model:        "Radiateur voltalis",
			SwVersion:    "0.1.0",
		},
	}
}

func (c *HeaterConfigPayload) WithName(name string) *HeaterConfigPayload {
	c.name = name
	c.Device.name = name + " Hub"
	return c
}
