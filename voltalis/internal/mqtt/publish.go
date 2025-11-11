package mqtt

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
)

type GetTopic string

// PublishConfig publie une configuration Home Assistant (retained=true)
func (c *Client) PublishConfig(payload payload) error {
	return c.publish(fmt.Sprintf("homeassistant/%s/%s/config", payload.getComponent(), payload.getIdentifier()), true, payload)
}

// PublishState publie une mise à jour d'état (retained=false)
// Si une mise a jour d'état tombe en erreur, on ne fait pas tomber le processus complet
func (c *Client) PublishState(topic GetTopic, payload any) {
	err := c.publish(string(topic), false, payload)
	if err != nil {
		slog.Error("Failed to publish state", "topic", topic, "error", err)
	}
}

// PublishState publie une mise à jour de valeur
// Si une mise a jour d'état tombe en erreur, on ne fait pas tomber le processus complet
func (c *Client) PublishCommand(topic SetTopic, payload any) {
	err := c.publish(string(topic), false, payload)
	if err != nil {
		slog.Error("Failed to publish command", "topic", topic, "error", err)
	}
}

// publish publie sur MQTT :
// - Si T est une struct/pointeur de struct, on fait un json.Marshal
// - Si T est un type primitif (string, []byte, int, float, bool), on l'envoie directement
func (c *Client) publish(topic string, retained bool, payload any) error {
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

	token := c.Client.Publish(string(topic), 0, retained, data)
	token.Wait()
	return token.Error()
}
