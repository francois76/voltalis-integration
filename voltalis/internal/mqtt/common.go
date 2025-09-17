package mqtt

import (
	"fmt"
	"log/slog"
	"maps"
	"slices"
)

type Topic interface{ WriteTopic | ReadTopic }

func newTopicName[T Topic](base string) T {
	mode := "set"
	// si c'est un writeTopic on suffixe par get, sinon set
	if _, ok := any(*new(T)).(WriteTopic); ok {
		mode = "get"
	}
	result := T(fmt.Sprintf("voltalis/%s/%s", base, mode))
	slog.With("result", result).Debug("instanciating ")
	return result
}

func getPayloadSelectMode[T ~string](device DeviceInfo, options ...T) *SelectConfigPayload[T] {
	identifier := device.Identifiers[0] + "_mode"
	return &SelectConfigPayload[T]{
		UniqueID:     identifier,
		Name:         "Sélectionner le mode",
		CommandTopic: newTopicName[ReadTopic](identifier),
		StateTopic:   newTopicName[WriteTopic](identifier),
		Options:      options,
		Device:       device,
	}
}

func getPayloadSelectDuration(device DeviceInfo) *SelectConfigPayload[string] {
	identifier := device.Identifiers[0] + "_duration"
	return &SelectConfigPayload[string]{
		UniqueID:     identifier,
		Name:         "Sélectionner la durée",
		CommandTopic: newTopicName[ReadTopic](identifier),
		StateTopic:   newTopicName[WriteTopic](identifier),
		Options:      slices.Collect(maps.Keys(DURATION_NAMES_TO_VALUES)),
		Device:       device,
	}
}
