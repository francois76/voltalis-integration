package mqtt

import "fmt"

type Topic interface{ WriteTopic | ReadTopic }

type DeviceInfo struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SwVersion    string   `json:"sw_version"`
}

type HeaterConfigPayload struct {
	Name                    string     `json:"name"`
	UniqueID                string     `json:"unique_id"`
	CommandTopic            WriteTopic `json:"command_topic"`
	ModeStateTopic          WriteTopic `json:"mode_state_topic"`
	ModeCommandTopic        ReadTopic  `json:"mode_command_topic"`
	TemperatureStateTopic   WriteTopic `json:"temperature_state_topic"`
	TemperatureCommandTopic ReadTopic  `json:"temperature_command_topic"`
	MinTemp                 float64    `json:"min_temp"`
	MaxTemp                 float64    `json:"max_temp"`
	TempStep                float64    `json:"temp_step"`
	Modes                   []string   `json:"modes"`
	CurrentTemperatureTopic WriteTopic `json:"current_temperature_topic"`
	Device                  DeviceInfo `json:"device"`
}

func InstanciateVoltalisHeaterBaseConfig(id int64) *HeaterConfigPayload {
	return &HeaterConfigPayload{
		UniqueID:                fmt.Sprintf("voltalis_heater_%d", id),
		CommandTopic:            newTopic[WriteTopic](id, "set"),
		ModeStateTopic:          newTopic[WriteTopic](id, "mode"),
		ModeCommandTopic:        newTopic[ReadTopic](id, "mode/set"),
		TemperatureStateTopic:   newTopic[WriteTopic](id, "temp"),
		TemperatureCommandTopic: newTopic[ReadTopic](id, "temp/set"),
		MinTemp:                 15,
		MaxTemp:                 25,
		TempStep:                0.5,
		Modes:                   []string{"off", "heat"},
		CurrentTemperatureTopic: newTopic[WriteTopic](id, "current_temp"),
		Device: DeviceInfo{
			Identifiers:  []string{"voltalis_heater_" + fmt.Sprint(id)},
			Manufacturer: "Voltalis",
			Name:         "Radiateur",
			Model:        "Radiateur voltalis",
			SwVersion:    "0.1.0",
		},
	}
}

func (c *HeaterConfigPayload) WithName(name string) *HeaterConfigPayload {
	c.Name = name
	return c
}

func newTopic[T Topic](id int64, suffix string) T {
	return T(fmt.Sprintf("voltalis/heater/%d/%s", id, suffix))
}
