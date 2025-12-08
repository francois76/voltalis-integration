package mqtt

import (
	"log/slog"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/francois76/voltalis-integration/voltalis/internal/state"
)

type SetTopic string

func (c *Client) ListenState(topic SetTopic, publishState func(currentState *state.ResourceState, data string)) {
	c.ListenStateWithPreHook(topic, nil, publishState)
}

func (c *Client) ListenStateWithPreHook(topic SetTopic, preHook func(data string), publishState func(currentState *state.ResourceState, data string)) {
	if topic == "" {
		panic("tentative d'écouter un topic vide, verifier que les composant ayant généré ce topic est bien instancié")
	}

	// S'abonner au topic
	go c.Client.Subscribe(string(topic), 0, func(client mqtt.Client, msg mqtt.Message) {
		data := string(msg.Payload())
		childlog := slog.With("topic", msg.Topic(), "data", data)
		childlog.Debug("MQTT message received")

		// MAJ état global
		c.stateMutex.Lock()
		c.stateTopicMap[topic] = data
		c.stateMutex.Unlock()

		if preHook != nil {
			preHook(data)
		}
		currentState := c.StateManager.GetCurrentState()
		publishState(&currentState, data)
		c.StateManager.UpdateState(currentState)
		relatedGetTopic := strings.Replace(msg.Topic(), "/set", "/get", 1)
		c.PublishState(GetTopic(relatedGetTopic), data)
	})
}
