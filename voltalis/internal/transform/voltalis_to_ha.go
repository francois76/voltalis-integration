package transform

import (
	"fmt"
	"log/slog"

	"github.com/francois76/voltalis-integration/voltalis/internal/api"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
	"github.com/francois76/voltalis-integration/voltalis/internal/state"
)

func SyncVoltalisHeatersToHA(mqttClient *mqtt.Client, apiClient *api.Client) error {

	// initialisation de l'état global
	states := state.ResourceState{
		HeaterState: make(map[int64]state.HeaterState),
		ControllerState: state.ControllerState{
			Mode:    state.HeaterPresetModeAucunMode,
			Program: "Aucun programme",
		},
	}
	appliances, err := apiClient.GetAppliances()
	if err != nil {
		return err
	}

	for _, appliance := range appliances {
		heaterState := &state.HeaterState{}
		if !appliance.Programming.IsOn {
			heaterState.Mode = state.HeaterModeOff
		} else if appliance.Programming.ProgType == "MANUAL" {
			heaterState.Mode = state.HeaterModeManual
			heaterState.Temperature = appliance.Programming.TemperatureTarget
		} else if appliance.Programming.ProgType == "USER" {
			mapPreset(appliance, heaterState)
			mapEndDate(appliance, heaterState)
			states.ControllerState.Program = appliance.Programming.ProgName
		} else if appliance.Programming.ProgType == "QUICK" {
			mapPreset(appliance, heaterState)
			mapEndDate(appliance, heaterState)
			quickSettingsMappings := map[string]state.HeaterPresetMode{
				"quicksettings.shortleave": state.HeaterPresetModeEco,
				"quicksettings.athome":     state.HeaterPresetModeConfort,
				"quicksettings.longleave":  state.HeaterPresetModeHorsGel,
			}
			states.ControllerState.Mode = quickSettingsMappings[appliance.Programming.ProgName]
		} else {
			slog.Error("unknown prog type", "progType", appliance.Programming.ProgType)
		}
		states.HeaterState[int64(appliance.ID)] = *heaterState
	}
	slog.With("state", states).Debug("state after voltalis fetch")
	controllerCommands := mqttClient.BuildControllerCommandTopic()

	mqttClient.PublishCommand(controllerCommands.Duration, states.ControllerState.Duration)
	mqttClient.PublishCommand(controllerCommands.Mode, string(states.ControllerState.Mode))
	mqttClient.PublishCommand(controllerCommands.Program, states.ControllerState.Program)
	for id, heaterState := range states.HeaterState {
		heaterCommands := mqttClient.BuildHeaterCommandTopic(id)
		if heaterState.Temperature != 0 {
			mqttClient.PublishCommand(heaterCommands.Temperature, fmt.Sprintf("%.1f", heaterState.Temperature))
		}
		mqttClient.PublishCommand(heaterCommands.SingleDuration, heaterState.Duration)
		mqttClient.PublishCommand(heaterCommands.PresetMode, string(heaterState.Mode))
	}
	return nil
}

func mapEndDate(appliance api.Appliance, heaterState *state.HeaterState) {
	if appliance.Programming.EndDate != nil {
		heaterState.Duration = *appliance.Programming.EndDate
	} else {
		heaterState.Duration = "Jusqu'à ce que je change d'avis"
	}
	fmt.Println(heaterState)
}

func mapPreset(appliance api.Appliance, heaterState *state.HeaterState) {
	switch appliance.Programming.Mode {
	case "CONFORT":
		heaterState.Mode = state.HeaterModeConfort
	case "ECO":
		heaterState.Mode = state.HeaterModeEco
	case "HORS_GEL":
		heaterState.Mode = state.HeaterModeHorsGel
	default:
		heaterState.Mode = state.HeaterModeOff
	}
}
