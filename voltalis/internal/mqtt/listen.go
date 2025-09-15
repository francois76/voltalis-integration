package mqtt

import (
	"log/slog"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type ReadTopic string

func (c *Client) ListenState(topic ReadTopic, f func(data string)) {
	lastState := ""
	c.Client.Subscribe(string(topic), 0, func(client mqtt.Client, msg mqtt.Message) {
		slog.Debug("MQTT message received", "topic", msg.Topic(), "payload", string(msg.Payload()))
		data := string(msg.Payload())
		if lastState == data {
			slog.Debug("MQTT message ignored, same as last state", "data", data)
			return
		}
		lastState = data
		f(data)
		slog.With(slog.Any("data", msg)).Debug("MQTT message")
		relatedWriteTopic := strings.Replace(msg.Topic(), "/get", "/set", 1)
		c.PublishState(WriteTopic(relatedWriteTopic), data)
	})
}
