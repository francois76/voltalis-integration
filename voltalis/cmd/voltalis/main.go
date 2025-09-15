package main

import (
	"fmt"
	"time"

	"log/slog"
	"os"

	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func main() {
	// Configuration du niveau de log via une variable d'environnement
	logLevel := slog.LevelInfo
	if os.Getenv("DEBUG") == "1" {
		logLevel = slog.LevelDebug
	}
	h := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})
	slog.SetDefault(slog.New(h))

	client, err := mqtt.InitClient("tcp://localhost:1883", "voltalis-addon")
	if err != nil {
		panic(err)
	}

	configPayload := mqtt.InstanciateVoltalisHeaterBaseConfig(123).WithName("Salon")

	err = client.Publish(mqtt.HomeAssistantClimateConfig, configPayload)
	if err != nil {
		panic(err)
	}
	slog.Info("Discovery config published")

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

		time.Sleep(15 * time.Second)
		i++
	}
}
