package mqtt

import (
	"fmt"
	"log/slog"
)

var CONTROLLER_DEVICE = DeviceInfo{
	Identifiers:  []string{"voltalis_controller"},
	Manufacturer: "Voltalis",
	Name:         "Controleur",
	Model:        "Voltalis software Controller",
	SwVersion:    "0.1.0",
}

func (c *Client) RegisterController() error {
	modePayload := getPayloadSelectMode(CONTROLLER_DEVICE, PRESET_SELECT_CONTROLLER...)
	if err := c.PublishConfig(modePayload); err != nil {
		return fmt.Errorf("failed to publish controller mode config: %w", err)
	}
	durationPayload := getPayloadSelectDuration(CONTROLLER_DEVICE)
	if err := c.PublishConfig(durationPayload); err != nil {
		return fmt.Errorf("failed to publish controller duration config: %w", err)
	}

	programPayload := getPayloadSelectProgram()
	if err := c.PublishConfig(programPayload); err != nil {
		return fmt.Errorf("failed to publish controller program config: %w", err)
	}
	statePayload := getPayloadDureeMode(CONTROLLER_DEVICE)
	if err := c.PublishConfig(statePayload); err != nil {
		return fmt.Errorf("failed to publish controller state config: %w", err)
	}
	c.PublishState(statePayload.StateTopic, "Initialisation de l'intégration voltalis...")
	c.ListenState(controller.SetTopics.Mode, func(data string) {
		slog.Debug("received value:", "value", data)
		// Handle controller command state changes
	})
	c.ListenState(controller.SetTopics.Duration, func(data string) {
		slog.Debug("received value:", "value", data)
	})
	return nil
}

type ControllerSetTopics struct {
	Mode     SetTopic
	Duration SetTopic
	Program  SetTopic
}
type ControllerGetTopics struct {
	Mode     GetTopic
	Duration GetTopic
	Program  GetTopic
	State    GetTopic
}

type Controller struct {
	SetTopics ControllerSetTopics
	GetTopics ControllerGetTopics
}

func getPayloadSelectProgram(options ...string) *SelectConfigPayload[string] {
	identifier := CONTROLLER_DEVICE.Identifiers[0] + "_program"
	return &SelectConfigPayload[string]{
		UniqueID:     identifier,
		Name:         "Sélectionner le programme",
		CommandTopic: newTopicName[SetTopic](identifier),
		StateTopic:   newTopicName[GetTopic](identifier),
		Options:      append([]string{"Aucun programme"}, options...),
		Device:       CONTROLLER_DEVICE,
	}
}
