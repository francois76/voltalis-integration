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
			// Radiateur éteint
			heaterState.Mode = state.HeaterModeOff
			heaterState.PresetMode = state.HeaterPresetModeAucunMode
		} else if appliance.Programming.ProgType == "MANUAL" {
			// ManualSetting actif - le Mode indique ECO/CONFORT/HORS_GEL/TEMPERATURE
			mapPreset(appliance, heaterState)
			mapEndDate(appliance, heaterState)
		} else if appliance.Programming.ProgType == "USER" {
			// Programme utilisateur (hebdomadaire)
			heaterState.Mode = state.HeaterModeAuto
			mapPreset(appliance, heaterState)
			mapEndDate(appliance, heaterState)
			states.ControllerState.Program = appliance.Programming.ProgName
		} else if appliance.Programming.ProgType == "QUICK" {
			// QuickSetting (absence courte, etc.)
			mapPreset(appliance, heaterState)
			mapEndDate(appliance, heaterState)
			quickSettingsMappings := map[string]state.HeaterPresetMode{
				"quicksettings.shortleave": state.HeaterPresetModeEco,
				"quicksettings.athome":     state.HeaterPresetModeConfort,
				"quicksettings.longleave":  state.HeaterPresetModeHorsGel,
			}
			states.ControllerState.Mode = quickSettingsMappings[appliance.Programming.ProgName]
		} else if appliance.Programming.ProgType == "DEFAULT" {
			// Mode par défaut (pas de programme actif, pas de manualSetting)
			heaterState.Mode = state.HeaterModeAuto
			mapPreset(appliance, heaterState)
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
		mqttClient.PublishCommand(heaterCommands.SingleDuration, heaterState.Duration)
		// Toujours publier le mode
		if heaterState.Mode != "" {
			mqttClient.PublishCommand(heaterCommands.Mode, string(heaterState.Mode))
		}
		// Toujours publier le preset (même si mode est heat, pour garder l'état à jour)
		if heaterState.PresetMode != "" {
			mqttClient.PublishCommand(heaterCommands.PresetMode, string(heaterState.PresetMode))
		}
		if heaterState.Temperature != 0 {
			mqttClient.PublishCommand(heaterCommands.Temperature, heaterState.Temperature)
		}
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
		heaterState.PresetMode = state.HeaterPresetModeConfort
	case "ECO":
		heaterState.PresetMode = state.HeaterPresetModeEco
	case "HORS_GEL":
		heaterState.PresetMode = state.HeaterPresetModeHorsGel
	case "TEMPERATURE":
		// Mode température personnalisée = mode "heat" dans HA
		heaterState.Mode = state.HeaterModeHeat
		heaterState.Temperature = appliance.Programming.TemperatureTarget
	default:
		heaterState.PresetMode = state.HeaterPresetModeAucunMode
	}
}
