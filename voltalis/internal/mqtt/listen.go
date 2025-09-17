package mqtt

import (
	"log/slog"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type SetTopic string

func (c *Client) ListenState(topic SetTopic) {
	c.ListenStateWithPreHook(topic, nil)
}

func (c *Client) ListenStateWithPreHook(topic SetTopic, f func(data string)) {
	if topic == "" {
		panic("tentative d'écouter un topic vide, verifier que les composant ayant généré ce topic est bien instancié")
	}
	go c.Client.Subscribe(string(topic), 0, func(client mqtt.Client, msg mqtt.Message) {
		data := string(msg.Payload())
		if c.stateMap[topic] == data {
			return
		}
		childlog := slog.With("topic", msg.Topic(), "data", data)
		childlog.Debug("MQTT message received")

		// MAJ état global
		c.stateMutex.Lock()
		c.stateMap[topic] = data
		c.stateMutex.Unlock()

		if f != nil {
			f(data)
		}
		currentState := c.StateManager.GetCurrentState()
		currentState.ID++
		c.StateManager.UpdateState(*currentState)
		relatedGetTopic := strings.Replace(msg.Topic(), "/set", "/get", 1)
		c.PublishState(GetTopic(relatedGetTopic), data)
	})
}
