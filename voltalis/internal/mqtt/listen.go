package mqtt

import (
	"log/slog"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ReadTopic string

func (c *Client) ListenState(topic ReadTopic, f func(data string)) {
	c.Client.Subscribe(string(topic), 0, func(client mqtt.Client, msg mqtt.Message) {
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

		f(data)
		relatedWriteTopic := strings.Replace(msg.Topic(), "/set", "/get", 1)
		c.PublishState(WriteTopic(relatedWriteTopic), data)
	})
}
