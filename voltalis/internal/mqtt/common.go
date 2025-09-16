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

func getPayloadSelectMode[T ~string](device DeviceInfo, options ...T) *SelectConfigPayload[T] {
	identifier := device.Identifiers[0] + "_mode"
	return &SelectConfigPayload[T]{
		UniqueID:     identifier,
		Name:         "Select mode",
		CommandTopic: newTopicName[ReadTopic](identifier),
		StateTopic:   newTopicName[WriteTopic](identifier),
		Options:      options,
		Device:       device,
	}
}

func getPayloadSelectDuration[T selectDuration](device DeviceInfo) *SelectConfigPayload[selectDuration] {
	identifier := device.Identifiers[0] + "_duration"
	return &SelectConfigPayload[selectDuration]{
		UniqueID:     identifier,
		Name:         "Select duration",
		CommandTopic: newTopicName[ReadTopic](identifier),
		StateTopic:   newTopicName[WriteTopic](identifier),
		Options:      []selectDuration{selectDurationOneHour, selectDurationTwoHour},
		Device:       device,
	}
}
