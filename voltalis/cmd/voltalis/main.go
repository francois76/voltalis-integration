package main

import (
	"fmt"

	"log/slog"

	"github.com/francois76/voltalis-integration/voltalis/internal/logger"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
	"github.com/francois76/voltalis-integration/voltalis/internal/scheduler"
)

func main() {
	logger.InitLogs()
	client, err := mqtt.InitClient("tcp://localhost:1883", "voltalis-addon")
	if err != nil {
		panic(err)
	}

	configPayload := mqtt.InstanciateVoltalisHeaterBaseConfig(123).WithName("Salon")

	err = client.PublishConfig(configPayload)
	if err != nil {
		panic(err)
	}
	slog.Info("Discovery config published")
	go client.ListenState(configPayload.TemperatureCommandTopic, func(data string) {
		slog.Info("Target temperature command received", "value", data)
	})
	scheduler.Run(func() {
		client.PublishState(configPayload.CurrentTemperatureTopic, fmt.Sprintf("%.1f", 19.5))
		client.PublishState(configPayload.ModeStateTopic, "heat")
		client.PublishState(configPayload.TemperatureStateTopic, "21")
	})

}
