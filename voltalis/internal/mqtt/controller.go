package mqtt

import (
	"fmt"

	"github.com/francois76/voltalis-integration/voltalis/internal/state"
)

var CONTROLLER_DEVICE = DeviceInfo{
	Identifiers:  []string{"voltalis_controller"},
	Manufacturer: "Voltalis",
	Name:         "Controleur",
	Model:        "Voltalis software Controller",
	SwVersion:    "0.1.0",
}

func (c *Client) RegisterController() error {
	controller := Controller{
		Client:    c,
		GetTopics: ControllerGetTopics{},
		SetTopics: ControllerSetTopics{},
	}

	if err := controller.addSelectMode(); err != nil {
		return err
	}
	if err := controller.addSelectDuration(); err != nil {
		return err
	}

	if err := controller.addSelectProgram(); err != nil {
		return err
	}
	err := controller.addDurationState()
	if err != nil {
		return err
	}
	controller.ListenState(controller.SetTopics.Mode, func(currentState *state.ResourceState, data string) {
		currentState.ControllerState.Mode = state.HeaterPresetMode(data)
	})
	controller.ListenState(controller.SetTopics.Duration, func(currentState *state.ResourceState, data string) {
		currentState.ControllerState.Duration = data
	})
	controller.ListenState(controller.SetTopics.Program, func(currentState *state.ResourceState, data string) {
		currentState.ControllerState.Program = data
	})
	return nil
}

func (controller *Controller) addDurationState() error {
	statePayload := getPayloadDureeMode(CONTROLLER_DEVICE)
	if err := controller.PublishConfig(statePayload); err != nil {
		return fmt.Errorf("failed to publish controller state config: %w", err)
	}
	controller.GetTopics.State = statePayload.StateTopic
	controller.PublishState(statePayload.StateTopic, "Initialisation de l'intégration voltalis...")
	return nil
}

func (controller *Controller) addSelectProgram() error {
	programPayload := getPayloadSelectProgram()
	if err := controller.PublishConfig(programPayload); err != nil {
		return fmt.Errorf("failed to publish controller program config: %w", err)
	}
	controller.GetTopics.Program = programPayload.StateTopic
	controller.SetTopics.Program = programPayload.CommandTopic
	return nil
}

func (c *Controller) addSelectDuration() error {
	durationPayload := getPayloadSelectDuration(CONTROLLER_DEVICE)
	if err := c.PublishConfig(durationPayload); err != nil {
		return fmt.Errorf("failed to publish controller duration config: %w", err)
	}
	c.GetTopics.Duration = durationPayload.StateTopic
	c.SetTopics.Duration = durationPayload.CommandTopic
	return nil
}

func (c *Controller) addSelectMode() error {
	modePayload := getPayloadSelectMode(CONTROLLER_DEVICE, PRESET_SELECT_CONTROLLER...)
	if err := c.PublishConfig(modePayload); err != nil {
		return fmt.Errorf("failed to publish controller mode config: %w", err)
	}
	c.GetTopics.Mode = modePayload.StateTopic
	c.SetTopics.Mode = modePayload.CommandTopic
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
	*Client
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
