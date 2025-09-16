package transform

import (
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
	heater, err := client.InstanciateHeater(deviceID, name)
	if err != nil {
		panic(err)
	}
	go client.ListenState(heater.ReadTopics.Temperature, func(data string) {
	})

	go client.ListenState(heater.ReadTopics.PresetMode, func(data string) {
		recomputeState(client, heater, data)
	})

	go client.ListenState(heater.ReadTopics.Mode, func(data string) {
		switch mqtt.HeaterMode(data) {
		case mqtt.HeaterModeOff:
			recomputeState(client, heater, string(mqtt.HeaterPresetModeNone))
			client.PublishState(heater.WriteTopics.PresetMode, mqtt.HeaterPresetModeNone)
		case mqtt.HeaterModeAuto:
			lastPreset := client.GetState(heater.ReadTopics.PresetMode)
			if lastPreset == string(mqtt.HeaterPresetModeManuel) || lastPreset == string(mqtt.HeaterPresetModeNone) {
				client.PublishState(heater.WriteTopics.PresetMode, mqtt.HeaterPresetModeConfort)
			}
		case mqtt.HeaterModeHeat:
			recomputeState(client, heater, string(mqtt.HeaterPresetModeNone))
			client.PublishState(heater.WriteTopics.PresetMode, mqtt.HeaterPresetModeNone)
		default:
			slog.Warn("Unknown mode received", "value", data)
		}
	})

	return heater.WriteTopics
}

func recomputeState(client *mqtt.Client, heater mqtt.Heater, data string) {
	slog.Info("Target preset mode received", "value", data)
	targetHeaterMode := mqtt.HeaterModeAuto
	targetTemperature := mqtt.TEMPERATURE_NONE
	targetAction := mqtt.HeaterActionIdle

	switch mqtt.HeaterPresetMode(data) {
	case mqtt.HeaterPresetModeNone:
		// On cherche ici à distinguer 2 cas: soit on a manuellement retiré le preset, dans ce cas on bascule en mode manuel
		// soit on a mis le mode en off, et dans ce cas on ne fait rien
		lastMode := client.GetState(heater.ReadTopics.Mode)
		slog.Debug("Last mode read", "value", lastMode)
		if lastMode == string(mqtt.HeaterModeOff) {
			targetAction = mqtt.HeaterActionOff
			targetHeaterMode = mqtt.HeaterModeOff
		} else {
			targetAction = mqtt.HeaterActionHeating
			targetTemperature = "18"
			targetHeaterMode = mqtt.HeaterModeHeat
		}
	case mqtt.HeaterPresetModeHorsGel:
		targetAction = mqtt.HeaterActionIdle
	case mqtt.HeaterPresetModeEco:
		targetAction = mqtt.HeaterActionCooling
	case mqtt.HeaterPresetModeConfort:
		targetAction = mqtt.HeaterActionHeating
	default:
		slog.Warn("Unknown preset mode received", "value", data)
	}
	client.PublishState(heater.WriteTopics.Action, targetAction)
	client.PublishState(heater.WriteTopics.Mode, targetHeaterMode)
	client.PublishState(heater.WriteTopics.Temperature, targetTemperature)
}
