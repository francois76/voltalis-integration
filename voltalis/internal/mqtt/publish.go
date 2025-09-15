package mqtt

import (
	"encoding/json"
	"fmt"
	"reflect"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type Topic string

const (
	HomeAssistantClimateConfig Topic = "homeassistant/climate/voltalis_heater/config"
)

type Client struct {
	mqtt.Client
}

func InitClient(broker string, clientID string) (*Client, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(broker).
		SetClientID(clientID)

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, token.Error()
	}
	return &Client{Client: client}, nil
}

func (c *Client) Publish(topic Topic, payload any) error {
	// Si c'est un pointeur, on récupère la valeur
	if reflect.TypeOf(payload).Kind() == reflect.Ptr {
		payload = reflect.ValueOf(payload).Elem().Interface()
	}
	bytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	token := c.Client.Publish(string(topic), 0, false, bytes)
	token.Wait()
	return token.Error()
}
