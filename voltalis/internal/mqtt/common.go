package mqtt

import (
	"fmt"
	"maps"
	"slices"
)

type Topic interface{ GetTopic | SetTopic }

func newTopicName[T Topic](base string) T {
	mode := "set"
	// si c'est un writeTopic on suffixe par get, sinon set
	if _, ok := any(*new(T)).(GetTopic); ok {
		mode = "get"
	}
	result := T(fmt.Sprintf("voltalis/%s/%s", base, mode))
	return result
}

func getPayloadSelectMode[T ~string](device DeviceInfo, options ...T) *SelectConfigPayload[T] {
	identifier := device.Identifiers[0] + "_mode"
	return &SelectConfigPayload[T]{
		UniqueID:     identifier,
		Name:         "Sélectionner le mode",
		CommandTopic: newTopicName[SetTopic](identifier),
		StateTopic:   newTopicName[GetTopic](identifier),
		Options:      options,
		Device:       device,
	}
}

func getPayloadSelectDuration(device DeviceInfo) *SelectConfigPayload[string] {
	identifier := device.Identifiers[0] + "_duration"
	return &SelectConfigPayload[string]{
		UniqueID:     identifier,
		Name:         "Sélectionner la durée",
		CommandTopic: newTopicName[SetTopic](identifier),
		StateTopic:   newTopicName[GetTopic](identifier),
		Options:      slices.Sorted(maps.Keys(DURATION_NAMES_TO_VALUES)),
		Device:       device,
	}
}

func getPayloadDureeMode(device DeviceInfo, topic GetTopic) *SensorConfigPayload {
	identifier := device.Identifiers[0] + "_state"
	return &SensorConfigPayload{
		UniqueID:   identifier,
		Name:       "Durée mode",
		StateTopic: topic,
		Device:     device,
	}
}

func getPayloadRefreshButton(device DeviceInfo) *ButtonConfigPayload {
	identifier := device.Identifiers[0] + "_refresh"
	return &ButtonConfigPayload{
		UniqueID:     identifier,
		Name:         "Recharger la configuration",
		CommandTopic: newTopicName[SetTopic](identifier),
		Device:       device,
	}
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
