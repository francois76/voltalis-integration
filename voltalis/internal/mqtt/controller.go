package mqtt

import "fmt"

var CONTROLLER_DEVICE = DeviceInfo{
	Identifiers:  []string{"voltalis_controller"},
	Manufacturer: "Voltalis",
	Name:         "Controleur",
	Model:        "Voltalis software Controller",
	SwVersion:    "0.1.0",
}

func (c *Client) InstanciateController() (Controller, error) {
	configPayload := getPayloadSelectMode(CONTROLLER_DEVICE, "controller", "mode", PRESET_SELECT_CONTROLLER...)
	if err := c.PublishConfig(configPayload); err != nil {
		return Controller{}, fmt.Errorf("failed to publish controller config: %w", err)
	}
	return Controller{
		ReadTopics:  ControllerReadTopics{Command: configPayload.CommandTopic},
		WriteTopics: ControllerWriteTopics{State: configPayload.StateTopic},
	}, nil

}

type ControllerReadTopics struct {
	Command ReadTopic
}
type ControllerWriteTopics struct {
	State WriteTopic
}

type Controller struct {
	ReadTopics  ControllerReadTopics
	WriteTopics ControllerWriteTopics
}
