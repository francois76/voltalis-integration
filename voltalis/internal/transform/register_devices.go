package transform

import (
	"fmt"
	"log/slog"

	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func RegisterDevices(client *mqtt.Client) map[int64]mqtt.HeaterWriteTopics {
	registerController(client)
	result := map[int64]mqtt.HeaterWriteTopics{}
	result[12345678901234] = registerHeater(client, 12345678901234, "Salon")
	result[23456789012345] = registerHeater(client, 23456789012345, "Chambre")
	return result
}

var VOLTALIS_MODES = []string{"Confort", "Eco", "Hors-Gel", "Manuel", "Arret"}

func registerController(client *mqtt.Client) {
	configPayload := mqtt.InstanciateVoltalisControllerSelectConfig("mode", VOLTALIS_MODES...)
	err := client.PublishConfig(mqtt.ComponentSelect, "voltalis_controller", configPayload)
	if err != nil {
		panic(err)
	}
}
func registerHeater(client *mqtt.Client, deviceID int64, name string) mqtt.HeaterWriteTopics {
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

	go client.ListenState(heaterTopics.Read.PresetMode, func(data string) {
		slog.Info("Target temperature command received", "value", data)
		if data == "none" {
			client.PublishState(heaterTopics.Write.Mode, mqtt.HeaterModeHeat)
			client.PublishState(heaterTopics.Write.Temperature, 18.0)
		} else {
			client.PublishState(heaterTopics.Write.Mode, mqtt.HeaterModeAuto)
			client.PublishState(heaterTopics.Write.Temperature, "None")
		}
		switch mqtt.HeaterPresetMode(data) {
		case mqtt.HeaterPresetEco:
			client.PublishState(heaterTopics.Write.Action, mqtt.HeaterActionIdle)
		case mqtt.HeaterPresetAway:
			client.PublishState(heaterTopics.Write.Action, mqtt.HeaterActionCooling)
		case mqtt.HeaterPresetHome:
			client.PublishState(heaterTopics.Write.Action, mqtt.HeaterActionHeating)
		default:
			slog.Warn("Unknown preset mode received", "value", data)
		}
	})

	return heaterTopics.Write
}
