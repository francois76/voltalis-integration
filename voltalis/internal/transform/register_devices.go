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
	configPayload := mqtt.InstanciateVoltalisHeaterClimate(deviceID, name)
	err := client.PublishConfig(mqtt.ComponentClimate, fmt.Sprintf("voltalis_heater_%d", deviceID), configPayload)
	if err != nil {
		panic(err)
	}
	slog.Info("Discovery config published")
	heaterTopics := configPayload.GetTopics()

	err = client.PublishConfig(mqtt.ComponentSelect, fmt.Sprintf("voltalis_select_%d", deviceID), mqtt.SelectConfigPayload{
		UniqueID:     fmt.Sprintf("voltalis_controller_select_%s", name),
		Name:         fmt.Sprintf("Controller Select %s", name),
		CommandTopic: configPayload.PresetModeCommandTopic,
		StateTopic:   configPayload.PresetModeStateTopic,
		Options: []string{
			string(mqtt.HeaterPresetModeHorsGel),
			string(mqtt.HeaterPresetModeEco),
			string(mqtt.HeaterPresetModeConfort),
		},
		Device: configPayload.Device,
	})
	if err != nil {
		panic(err)
	}
	go client.ListenState(heaterTopics.Read.Temperature, func(data string) {
	})

	go client.ListenState(heaterTopics.Read.PresetMode, func(data string) {
		recomputeState(client, heaterTopics, data)
	})

	go client.ListenState(heaterTopics.Read.Mode, func(data string) {
		switch mqtt.HeaterMode(data) {
		case mqtt.HeaterModeOff:
			recomputeState(client, heaterTopics, string(mqtt.HeaterPresetModeNone))
			client.PublishState(heaterTopics.Write.PresetMode, mqtt.HeaterPresetModeNone)
		case mqtt.HeaterModeAuto:
			client.PublishState(heaterTopics.Write.PresetMode, mqtt.HeaterPresetModeConfort)
		case mqtt.HeaterModeHeat:
			recomputeState(client, heaterTopics, string(mqtt.HeaterPresetModeNone))
			client.PublishState(heaterTopics.Write.PresetMode, mqtt.HeaterPresetModeNone)
		default:
			slog.Warn("Unknown mode received", "value", data)
		}
	})

	return heaterTopics.Write
}

func recomputeState(client *mqtt.Client, heaterTopics mqtt.HeaterTopics, data string) {
	slog.Info("Target preset mode received", "value", data)
	targetHeaterMode := mqtt.HeaterModeAuto
	targetTemperature := mqtt.TEMPERATURE_NONE
	targetAction := mqtt.HeaterActionIdle

	switch mqtt.HeaterPresetMode(data) {
	case mqtt.HeaterPresetModeNone:
		// On cherche ici à distinguer 2 cas: soit on a manuellement retiré le preset, dans ce cas on bascule en mode manuel
		// soit on a mis le mode en off, et dans ce cas on ne fait rien
		lastMode := client.GetState(heaterTopics.Read.Mode)
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
	client.PublishState(heaterTopics.Write.Action, targetAction)
	client.PublishState(heaterTopics.Write.Mode, targetHeaterMode)
	client.PublishState(heaterTopics.Write.Temperature, targetTemperature)
}
