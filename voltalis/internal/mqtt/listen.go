package mqtt

import (
	"log/slog"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ReadTopic string

func (c *Client) ListenState(topic ReadTopic, f func(data string)) {
	c.Client.Subscribe(string(topic), 0, func(client mqtt.Client, msg mqtt.Message) {
		childlog := slog.With("topic", msg.Topic(), "data", string(msg.Payload()))
		childlog.Debug("MQTT message received")
		data := string(msg.Payload())
		if c.stateMap[topic] == data {
			childlog.Debug("MQTT message ignored, same as last state")
			return
		}

		// MAJ état global
		c.stateMutex.Lock()
		c.stateMap[topic] = data
		c.stateMutex.Unlock()

		f(data)
		relatedWriteTopic := strings.Replace(msg.Topic(), "/get", "/set", 1)
		c.PublishState(WriteTopic(relatedWriteTopic), data)
	})
}
