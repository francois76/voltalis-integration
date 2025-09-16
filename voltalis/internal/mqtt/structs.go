package mqtt

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

type SelectConfigPayload struct {
	Name         string     `json:"name"`
	UniqueID     string     `json:"unique_id"`
	CommandTopic ReadTopic  `json:"command_topic"`
	StateTopic   WriteTopic `json:"state_topic"`
	Options      []string   `json:"options"`
	Device       DeviceInfo `json:"device"`
}

// DeviceInfo représente les informations du périphérique pour Home Assistant
type DeviceInfo struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SwVersion    string   `json:"sw_version"`
}
