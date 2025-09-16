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
	modePayload := getPayloadSelectMode(CONTROLLER_DEVICE, PRESET_SELECT_CONTROLLER...)
	if err := c.PublishConfig(modePayload); err != nil {
		return Controller{}, fmt.Errorf("failed to publish controller mode config: %w", err)
	}
	durationPayload := getPayloadSelectDuration(CONTROLLER_DEVICE)
	if err := c.PublishConfig(durationPayload); err != nil {
		return Controller{}, fmt.Errorf("failed to publish controller duration config: %w", err)
	}

	return Controller{
		ReadTopics: ControllerReadTopics{
			Mode:     modePayload.CommandTopic,
			Duration: durationPayload.CommandTopic,
		},
		WriteTopics: ControllerWriteTopics{
			Mode:     modePayload.StateTopic,
			Duration: durationPayload.StateTopic,
		},
	}, nil

}

type ControllerReadTopics struct {
	Mode     ReadTopic
	Duration ReadTopic
}
type ControllerWriteTopics struct {
	Mode     WriteTopic
	Duration WriteTopic
}

type Controller struct {
	ReadTopics  ControllerReadTopics
	WriteTopics ControllerWriteTopics
}
