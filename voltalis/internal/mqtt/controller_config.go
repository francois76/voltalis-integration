package mqtt

import "fmt"

var CONTROLLER_DEVICE = DeviceInfo{
	Identifiers:  []string{"voltalis_controller"},
	Manufacturer: "Voltalis",
	Name:         "Controller de gestion voltalis",
	Model:        "Voltalis software Controller",
	SwVersion:    "0.1.0",
}

func InstanciateVoltalisControllerSelectConfig(name string, options ...string) *SelectConfigPayload {
	return &SelectConfigPayload{
		UniqueID:     fmt.Sprintf("voltalis_controller_select_%s", name),
		Name:         fmt.Sprintf("Controller Select %s", name),
		CommandTopic: newTopicName[ReadTopic](name),
		StateTopic:   newTopicName[WriteTopic](name),
		Options:      options,
		Device:       CONTROLLER_DEVICE,
	}
}
