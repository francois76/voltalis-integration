package mqtt

type payload interface {
	getIdentifier() string
	getComponent() component
}

type HeaterCommandPayload struct {
	ModeCommandTopic        SetTopic `json:"mode_command_topic"`
	PresetModeCommandTopic  SetTopic `json:"preset_mode_command_topic,omitempty"`
	TemperatureCommandTopic SetTopic `json:"temperature_command_topic"`
}

type HeaterConfigPayload struct {
	HeaterCommandPayload
	ActionTopic              GetTopic           `json:"action_topic,omitempty"`
	Name                     string             `json:"name"`
	UniqueID                 string             `json:"unique_id"`
	CommandTopic             GetTopic           `json:"command_topic"`
	ModeStateTopic           GetTopic           `json:"mode_state_topic"`
	PresetModes              []HeaterPresetMode `json:"preset_modes,omitempty"`
	PresetModeStateTopic     GetTopic           `json:"preset_mode_state_topic,omitempty"`
	TemperatureStateTopic    GetTopic           `json:"temperature_state_topic"`
	MinTemp                  float64            `json:"min_temp"`
	MaxTemp                  float64            `json:"max_temp"`
	TempStep                 float64            `json:"temp_step"`
	Modes                    []HeaterMode       `json:"modes"`
	CurrentTemperatureTopic  GetTopic           `json:"current_temperature_topic"`
	Device                   DeviceInfo         `json:"device"`
	TemperatureStateTemplate string             `json:"temperature_state_template,omitempty"`
}

func (p *HeaterConfigPayload) getIdentifier() string {
	return p.UniqueID
}

func (p *HeaterConfigPayload) getComponent() component {
	return ComponentClimate
}

type SelectConfigPayload[T ~string | ~int64] struct {
	Name         string     `json:"name"`
	UniqueID     string     `json:"unique_id"`
	CommandTopic SetTopic   `json:"command_topic"`
	StateTopic   GetTopic   `json:"state_topic"`
	Options      []T        `json:"options"`
	Device       DeviceInfo `json:"device"`
}

func (p *SelectConfigPayload[T]) getIdentifier() string {
	return p.UniqueID
}

func (p *SelectConfigPayload[T]) getComponent() component {
	return ComponentSelect
}

type SensorConfigPayload struct {
	Name       string     `json:"name"`
	UniqueID   string     `json:"unique_id"`
	StateTopic GetTopic   `json:"state_topic"`
	Device     DeviceInfo `json:"device"`
}

func (p *SensorConfigPayload) getIdentifier() string {
	return p.UniqueID
}

func (p *SensorConfigPayload) getComponent() component {
	return ComponentSensor
}

// DeviceInfo représente les informations du périphérique pour Home Assistant
type DeviceInfo struct {
	Identifiers  []string `json:"identifiers"`
	Name         string   `json:"name"`
	Manufacturer string   `json:"manufacturer"`
	Model        string   `json:"model"`
	SwVersion    string   `json:"sw_version"`
}
