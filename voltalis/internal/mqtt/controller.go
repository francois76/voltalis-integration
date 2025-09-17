package mqtt

import (
	"fmt"
)

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

	programPayload := getPayloadSelectProgram()
	if err := c.PublishConfig(programPayload); err != nil {
		return Controller{}, fmt.Errorf("failed to publish controller program config: %w", err)
	}

	return Controller{
		ReadTopics: ControllerReadTopics{
			Mode:     modePayload.CommandTopic,
			Duration: durationPayload.CommandTopic,
			Program:  programPayload.CommandTopic,
		},
		WriteTopics: ControllerWriteTopics{
			Mode:     modePayload.StateTopic,
			Duration: durationPayload.StateTopic,
			Program:  programPayload.StateTopic,
		},
	}, nil

}

type ControllerReadTopics struct {
	Mode     ReadTopic
	Duration ReadTopic
	Program  ReadTopic
}
type ControllerWriteTopics struct {
	Mode     WriteTopic
	Duration WriteTopic
	Program  WriteTopic
}

type Controller struct {
	ReadTopics  ControllerReadTopics
	WriteTopics ControllerWriteTopics
}

func getPayloadSelectProgram(options ...string) *SelectConfigPayload[string] {
	identifier := CONTROLLER_DEVICE.Identifiers[0] + "_program"
	return &SelectConfigPayload[string]{
		UniqueID:     identifier,
		Name:         "Sélectionner le programme",
		CommandTopic: newTopicName[ReadTopic](identifier),
		StateTopic:   newTopicName[WriteTopic](identifier),
		Options:      append([]string{"Aucun programme"}, options...),
		Device:       CONTROLLER_DEVICE,
	}
}
