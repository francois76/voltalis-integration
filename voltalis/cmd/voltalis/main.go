package main

import (
	"fmt"
	"time"

	"github.com/tonuser/voltalis-integration/voltalis/internal/mqtt"
)

func main() {

	client, err := mqtt.InitClient("tcp://localhost:1883", "voltalis-addon")
	if err != nil {
		panic(err)
	}

	configPayload := mqtt.InstanciateVoltalisHeaterBaseConfig(123).WithName("Salon")

	err = client.Publish(mqtt.HomeAssistantClimateConfig, configPayload)
	if err != nil {
		panic(err)
	}
	fmt.Println("Discovery config published")

	// Exemple : publier périodiquement une température et l’état
	i := 0
	for {
		// Température actuelle
		temp := 19.5 + float64(i%3)
		err = client.Publish(configPayload.CurrentTemperatureTopic, fmt.Sprintf("%.1f", temp))
		if err != nil {
			fmt.Println("Failed to publish temperature:", err)
		}
		// Mode (heat/off)
		mode := "heat"
		err = client.Publish(configPayload.ModeCommandTopic, mode)
		if err != nil {
			fmt.Println("Failed to publish mode:", err)
		}

		// Température consigne
		err = client.Publish(configPayload.TemperatureCommandTopic, "21.0")
		if err != nil {
			fmt.Println("Failed to publish target temperature:", err)
		}

		fmt.Println("Published climate state", temp, mode)
		time.Sleep(15 * time.Second)
		i++
	}
}
