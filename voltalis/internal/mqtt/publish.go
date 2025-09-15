package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
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

// Publish publie sur MQTT :
// - Si T est une struct/pointeur de struct, on fait un json.Marshal
// - Si T est un type primitif (string, []byte, int, float, bool), on l'envoie directement
func (c *Client) Publish(topic Topic, payload any) error {
	var data []byte
	var err error

	val := reflect.ValueOf(payload)
	kind := val.Kind()
	if kind == reflect.Ptr {
		val = val.Elem()
		kind = val.Kind()
	}

	switch kind {
	case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array:
		// Pour struct, map, slice, array : json.Marshal
		data, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
	case reflect.String:
		data = []byte(val.String())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		data = []byte(fmt.Sprintf("%d", val.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		data = []byte(fmt.Sprintf("%d", val.Uint()))
	case reflect.Float32, reflect.Float64:
		data = []byte(fmt.Sprintf("%f", val.Float()))
	case reflect.Bool:
		data = []byte(fmt.Sprintf("%t", val.Bool()))
	case reflect.Invalid:
		return fmt.Errorf("payload is nil")
	default:
		// Pour []byte, on vérifie explicitement
		if b, ok := any(payload).([]byte); ok {
			data = b
		} else {
			return fmt.Errorf("unsupported payload type: %s", kind)
		}
	}

	slog.Debug("MQTT publish", "topic", topic, "payload", string(data))
	token := c.Client.Publish(string(topic), 0, false, data)
	token.Wait()
	return token.Error()
}
