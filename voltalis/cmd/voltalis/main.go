package main

import (
	"encoding/json"
	"fmt"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	// Connexion MQTT
	opts := mqtt.NewClientOptions().
		AddBroker("tcp://localhost:1883").
		SetClientID("voltalis-addon")

	client := mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		panic(token.Error())
	}
	fmt.Println("MQTT connected")

	// Déclaration du climate (discovery)
	configTopic := "homeassistant/climate/voltalis_heater/config"
	configPayload := map[string]any{
		"name":      "Voltalis Heater",
		"unique_id": "voltalis_heater_1",

		"command_topic":             "voltalis/heater/set",
		"mode_state_topic":          "voltalis/heater/mode",
		"mode_command_topic":        "voltalis/heater/mode/set",
		"temperature_state_topic":   "voltalis/heater/temp",
		"temperature_command_topic": "voltalis/heater/temp/set",
		"min_temp":                  15,
		"max_temp":                  25,
		"temp_step":                 0.5,
		"modes":                     []string{"off", "heat"},
		"current_temperature_topic": "voltalis/heater/current_temp",

		// === Device info ===
		"device": map[string]any{
			"identifiers":  []string{"voltalis_hub_123"},
			"connections":  [][]string{{"mac", "AA:BB:CC:DD:EE:FF"}},
			"name":         "Voltalis Hub",
			"manufacturer": "Voltalis",
			"model":        "Virtual Heater v1",
			"sw_version":   "0.1.0",
		},
	}

	confJSON, _ := json.Marshal(configPayload)
	client.Publish(configTopic, 0, true, confJSON)
	fmt.Println("Discovery config published")

	// Exemple : publier périodiquement une température et l’état
	i := 0
	for {
		// Température actuelle
		temp := 19.5 + float64(i%3)
		client.Publish("voltalis/heater/current_temp", 0, false, fmt.Sprintf("%.1f", temp))

		// Mode (heat/off)
		mode := "heat"
		client.Publish("voltalis/heater/mode", 0, false, mode)

		// Température consigne
		client.Publish("voltalis/heater/temp", 0, false, "21.0")

		fmt.Println("Published climate state", temp, mode)
		time.Sleep(15 * time.Second)
		i++
	}
}
