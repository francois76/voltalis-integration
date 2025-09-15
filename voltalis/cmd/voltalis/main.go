package main

import (
	"time"

	"log/slog"

	"github.com/francois76/voltalis-integration/voltalis/internal/logger"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
	"github.com/francois76/voltalis-integration/voltalis/internal/scheduler"
	"github.com/francois76/voltalis-integration/voltalis/internal/transform"
)

func main() {
	logger.InitLogs()
	client, err := mqtt.InitClient("tcp://localhost:1883", "voltalis-addon")
	if err != nil {
		panic(err)
	}

	configPayload := mqtt.InstanciateVoltalisHeaterBaseConfig(123).WithName("Salon")

	err = client.PublishConfig(mqtt.ComponentClimate, "voltalis_heater", configPayload)
	if err != nil {
		panic(err)
	}
	slog.Info("Discovery config published")
	heaterTopics := configPayload.GetTopics()
	go client.ListenState(heaterTopics.Read.Temperature, func(data string) {
		slog.Info("Target temperature command received", "value", data)
	})
	scheduler.Run(15*time.Second, func() {
		transform.SyncVoltalisHeaterToHA(client, heaterTopics.Write)
	})

}
