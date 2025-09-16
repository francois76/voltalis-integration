package mqtt

import "fmt"

type Topic interface{ WriteTopic | ReadTopic }

func newTopicName[T Topic](base string) T {
	mode := "set"
	// si c'est un writeTopic on suffixe par get, sinon set
	if _, ok := any(new(T)).(WriteTopic); ok {
		mode = "get"
	}
	return T(fmt.Sprintf("voltalis/%s/%s", base, mode))
}

func getPayloadSelectMode(device DeviceInfo, uniqueIdSuffix, name string, options ...string) *SelectConfigPayload {
	identifier := device.Identifiers[0] + "_" + uniqueIdSuffix
	return &SelectConfigPayload{
		UniqueID:     identifier,
		Name:         fmt.Sprintf("Select %s", name),
		CommandTopic: newTopicName[ReadTopic](identifier),
		StateTopic:   newTopicName[WriteTopic](identifier),
		Options:      options,
		Device:       device,
	}
}
