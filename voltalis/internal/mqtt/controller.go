package mqtt

import "fmt"

var CONTROLLER_DEVICE = DeviceInfo{
	Identifiers:  []string{"voltalis_controller"},
	Manufacturer: "Voltalis",
	Name:         "Controleur",
	Model:        "Voltalis software Controller",
	SwVersion:    "0.1.0",
}

var VOLTALIS_MODES = []string{"Confort", "Eco", "Hors-Gel", "Manuel", "Arret"}

func (c *Client) InstanciateController() (Controller, error) {
	configPayload := getPayloadSelectMode(CONTROLLER_DEVICE, "controller", "mode", VOLTALIS_MODES...)
	err := c.PublishConfig(configPayload)
	if err != nil {
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
