package mqtt

import (
	"log/slog"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ReadTopic string

func (c *Client) ListenState(topic ReadTopic, f func(data string)) {
	c.Client.Subscribe(string(topic), 0, func(client mqtt.Client, msg mqtt.Message) {
		slog.Debug("MQTT message received", "topic", msg.Topic(), "payload", string(msg.Payload()))
		f(string(msg.Payload()))
	})
}
