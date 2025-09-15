package transform

import (
	"fmt"
	"log/slog"

	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func RegisterDevices(client *mqtt.Client) map[int64]mqtt.HeaterWriteTopics {
	result := map[int64]mqtt.HeaterWriteTopics{}
	result[12345678901234] = registerDevice(client, 12345678901234, "Salon")
	result[23456789012345] = registerDevice(client, 23456789012345, "Chambre")
	return result
}

func registerDevice(client *mqtt.Client, deviceID int64, name string) mqtt.HeaterWriteTopics {
	configPayload := mqtt.InstanciateVoltalisHeaterBaseConfig(deviceID).WithName(name)
	err := client.PublishConfig(mqtt.ComponentClimate, fmt.Sprintf("voltalis_heater_%d", deviceID), configPayload)
	if err != nil {
		panic(err)
	}
	slog.Info("Discovery config published")
	heaterTopics := configPayload.GetTopics()
	go client.ListenState(heaterTopics.Read.Temperature, func(data string) {
		slog.Info("Target temperature command received", "value", data)
	})

	return heaterTopics.Write
}
