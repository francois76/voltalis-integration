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
