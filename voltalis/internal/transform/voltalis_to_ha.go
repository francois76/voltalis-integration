package transform

import (
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
			heaterState.Mode = state.HeaterModeAuto
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

	// Mettre à jour le StateManager SANS déclencher de notification
	// Cela évite la boucle : sync Voltalis -> StateManager -> API Voltalis
	mqttClient.StateManager.UpdateStateWithoutNotification(states)

	// Publier sur les topics d'ÉTAT (/get) pour afficher dans Home Assistant
	// NE PAS publier sur les topics de COMMANDE (/set) car cela déclencherait les listeners
	controllerStates := mqttClient.BuildControllerStateTopic()

	mqttClient.PublishState(controllerStates.Duration, states.ControllerState.Duration)
	mqttClient.PublishState(controllerStates.Mode, string(states.ControllerState.Mode))
	mqttClient.PublishState(controllerStates.Program, states.ControllerState.Program)
	for id, heaterState := range states.HeaterState {
		heaterStates := mqttClient.BuildHeaterStateTopic(id)
		mqttClient.PublishState(heaterStates.SingleDuration, heaterState.Duration)
		// Toujours publier le mode
		if heaterState.Mode != "" {
			mqttClient.PublishState(heaterStates.Mode, string(heaterState.Mode))
		}
		// Toujours publier le preset (même si mode est heat, pour garder l'état à jour)
		if heaterState.PresetMode != "" {
			mqttClient.PublishState(heaterStates.PresetMode, string(heaterState.PresetMode))
		}
		if heaterState.Temperature != 0 {
			mqttClient.PublishState(heaterStates.Temperature, heaterState.Temperature)
		}
		// Publier l'action en fonction du preset (pour l'indicateur visuel)
		action := presetToAction(heaterState.PresetMode, heaterState.Mode)
		mqttClient.PublishState(heaterStates.Action, string(action))
	}
	return nil
}

func mapEndDate(appliance api.Appliance, heaterState *state.HeaterState) {
	if appliance.Programming.EndDate != nil {
		heaterState.Duration = *appliance.Programming.EndDate
	} else {
		heaterState.Duration = "Jusqu'à ce que je change d'avis"
	}
}

func mapPreset(appliance api.Appliance, heaterState *state.HeaterState) {
	switch appliance.Programming.Mode {
	case "CONFORT":
		heaterState.PresetMode = state.HeaterPresetModeConfort
		heaterState.Mode = state.HeaterModeAuto
	case "ECO":
		heaterState.PresetMode = state.HeaterPresetModeEco
		heaterState.Mode = state.HeaterModeAuto
	case "HORS_GEL":
		heaterState.PresetMode = state.HeaterPresetModeHorsGel
		heaterState.Mode = state.HeaterModeAuto
	case "TEMPERATURE":
		// Mode température personnalisée = mode "heat" dans HA
		heaterState.Mode = state.HeaterModeHeat
		heaterState.Temperature = appliance.Programming.TemperatureTarget
	default:
		heaterState.PresetMode = state.HeaterPresetModeAucunMode
		heaterState.Mode = state.HeaterModeAuto
	}
}

// presetToAction convertit un preset en action pour l'indicateur visuel HA
func presetToAction(preset state.HeaterPresetMode, mode state.HeaterMode) mqtt.HeaterAction {
	if mode == state.HeaterModeOff {
		return mqtt.HeaterActionOff
	}
	if mode == state.HeaterModeHeat {
		return mqtt.HeaterActionHeating
	}
	switch preset {
	case state.HeaterPresetModeConfort:
		return mqtt.HeaterActionHeating
	case state.HeaterPresetModeEco:
		return mqtt.HeaterActionCooling
	case state.HeaterPresetModeHorsGel:
		return mqtt.HeaterActionIdle
	default:
		return mqtt.HeaterActionIdle
	}
}
