package transform

import (
	"context"
	"log/slog"
	"time"

	"github.com/francois76/voltalis-integration/voltalis/internal/api"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
	"github.com/francois76/voltalis-integration/voltalis/internal/scheduler"
	"github.com/francois76/voltalis-integration/voltalis/internal/state"
)

// Mapping des modes HA vers les noms de quicksettings Voltalis
var modeToQuickSettingsName = map[state.HeaterPresetMode]string{
	state.HeaterPresetModeEco:     "quicksettings.shortleave",
	state.HeaterPresetModeConfort: "quicksettings.athome",
	state.HeaterPresetModeHorsGel: "quicksettings.longleave",
}

// Mapping inverse pour retrouver le mode depuis le nom
var quickSettingsNameToMode = map[string]state.HeaterPresetMode{
	"quicksettings.shortleave": state.HeaterPresetModeEco,
	"quicksettings.athome":     state.HeaterPresetModeConfort,
	"quicksettings.longleave":  state.HeaterPresetModeHorsGel,
}

// Mapping des modes HA vers les modes API Voltalis
var haPresetToVoltalisMode = map[state.HeaterPresetMode]string{
	state.HeaterPresetModeConfort: "CONFORT",
	state.HeaterPresetModeEco:     "ECO",
	state.HeaterPresetModeHorsGel: "HORS_GEL",
}

// Start est le point de démarrage de la fonction qui process les évenements MQTT de façon globalisée et appelle les APIs de voltalis pour répliquer les changements
func Start(ctx context.Context, mqttClient *mqtt.Client, apiClient *api.Client, schedule *scheduler.Scheduler) error {
	controller, err := mqttClient.RegisterController()
	if err != nil {
		return err
	}

	if err := syncPrograms(controller, apiClient); err != nil {
		return err
	}

	controller.ListenState(controller.SetTopics.Refresh, func(currentState *state.ResourceState, data string) {
		if err := syncPrograms(controller, apiClient); err != nil {
			slog.Error("failed to refresh programs: " + err.Error())
		}
		schedule.Trigger()
	})

	appliances, err := apiClient.GetAppliances()
	if err != nil {
		return err
	}
	for _, appliance := range appliances {
		if err := mqttClient.RegisterHeater(int64(appliance.ID), appliance.Name); err != nil {
			return err
		}
	}

	// Pré-charger les données nécessaires pour le traitement des changements
	programs, err := apiClient.GetPrograms()
	if err != nil {
		slog.Error("failed to load programs", "error", err)
	}
	quickSettings, err := apiClient.GetQuickSettings()
	if err != nil {
		slog.Error("failed to load quicksettings", "error", err)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stateChanges := mqttClient.StateManager.Subscribe()

	for {
		select {
		case change := <-stateChanges:
			// Ignorer l'état initial au démarrage - ce n'est pas un changement utilisateur
			if _, isInitial := change.ChangedFields["initial_state"]; isInitial {
				slog.Debug("État initial ignoré")
				continue
			}

			slog.With("change", change.ChangedFields).Debug("champs modifiés")

			// Détecter si on a des changements de radiateurs
			heaterChanges, hasHeaterChanges := change.ChangedFields["HeaterState"].(map[string]interface{})

			// Traitement des changements des radiateurs individuels EN PREMIER
			// (car ils peuvent causer des effets de bord sur le programme)
			var heaterApplied bool
			if hasHeaterChanges {
				var err error
				heaterApplied, err = handleHeaterChanges(apiClient, heaterChanges, change.CurrentState, quickSettings, appliances)
				if err != nil {
					slog.Error("failed to apply heater changes to Voltalis", "error", err)
				}
			}

			// Traitement des changements du contrôleur
			var controllerApplied bool
			if controllerChanges, ok := change.ChangedFields["ControllerState"].(map[string]interface{}); ok {
				var err error
				controllerApplied, err = handleControllerChanges(apiClient, controllerChanges, change.CurrentState, programs, quickSettings, appliances)
				if err != nil {
					slog.Error("failed to apply controller changes to Voltalis", "error", err)
				}
			}

			// Refresh après application des changements uniquement si des changements ont été faits
			if heaterApplied || controllerApplied {
				schedule.Trigger()
			}

		case <-ctx.Done():
			slog.Warn("context killed")
			return nil
		}
	}

}

// handleControllerChanges traite les changements du contrôleur global
// Retourne true si des changements ont été appliqués côté Voltalis
func handleControllerChanges(apiClient *api.Client, changes map[string]interface{}, currentState state.ResourceState, programs []api.Program, quickSettings []api.QuickSettings, appliances []api.Appliance) (bool, error) {
	applied := false

	// Changement de programme
	if newProgram, ok := changes["Program"].(string); ok {
		slog.Info("Programme changé", "nouveau", newProgram)
		if err := handleProgramChange(apiClient, newProgram, programs); err != nil {
			return false, err
		}
		applied = true
	}

	// Changement de mode global (quicksettings)
	if newMode, ok := changes["Mode"].(state.HeaterPresetMode); ok {
		slog.Info("Mode global changé", "nouveau", newMode)
		duration := currentState.ControllerState.Duration
		if err := handleModeChange(apiClient, newMode, duration, quickSettings, appliances); err != nil {
			return false, err
		}
		applied = true
	}

	// Changement de durée (sans changement de mode = mise à jour du quicksetting actuel)
	if _, hasDuration := changes["Duration"]; hasDuration {
		if _, hasMode := changes["Mode"]; !hasMode {
			// Durée changée sans changement de mode
			slog.Info("Durée changée", "nouvelle", currentState.ControllerState.Duration)
			if err := handleModeChange(apiClient, currentState.ControllerState.Mode, currentState.ControllerState.Duration, quickSettings, appliances); err != nil {
				return false, err
			}
			applied = true
		}
	}

	return applied, nil
}

// handleProgramChange gère le changement de programme
func handleProgramChange(apiClient *api.Client, programName string, programs []api.Program) error {
	// Désactiver tous les programmes d'abord
	for _, p := range programs {
		if p.Enabled {
			req := api.UpdateProgramRequest{
				ID:      p.ID,
				Name:    p.Name,
				Enabled: false,
			}
			if err := apiClient.UpdateProgram(p.ID, req); err != nil {
				slog.Error("failed to disable program", "program", p.Name, "error", err)
				return err
			}
			slog.Debug("Programme désactivé", "name", p.Name)
		}
	}

	// Si "Aucun programme" sélectionné, on s'arrête là
	if programName == "Aucun programme" {
		return nil
	}

	// Activer le nouveau programme
	for _, p := range programs {
		if p.Name == programName {
			req := api.UpdateProgramRequest{
				ID:      p.ID,
				Name:    p.Name,
				Enabled: true,
			}
			if err := apiClient.UpdateProgram(p.ID, req); err != nil {
				slog.Error("failed to enable program", "program", programName, "error", err)
				return err
			}
			slog.Info("Programme activé", "name", programName)
			return nil
		}
	}

	slog.Warn("Programme non trouvé", "name", programName)
	return nil
}

// handleModeChange gère le changement de mode global (quicksettings)
func handleModeChange(apiClient *api.Client, mode state.HeaterPresetMode, duration string, quickSettings []api.QuickSettings, appliances []api.Appliance) error {
	// Si mode "Aucun mode", désactiver tous les quicksettings
	if mode == state.HeaterPresetModeAucunMode {
		for _, qs := range quickSettings {
			if qs.Enabled {
				if err := apiClient.EnableQuickSetting(qs.ID, false); err != nil {
					slog.Error("failed to disable quicksetting", "name", qs.Name, "error", err)
					return err
				}
				slog.Debug("QuickSetting désactivé", "name", qs.Name)
			}
		}
		return nil
	}

	// Trouver le quicksetting correspondant au mode
	qsName, exists := modeToQuickSettingsName[mode]
	if !exists {
		slog.Warn("Mode non mappé vers un quicksetting", "mode", mode)
		return nil
	}

	// Trouver le quicksetting par son nom
	var targetQS *api.QuickSettings
	for _, qs := range quickSettings {
		if qs.Name == qsName {
			targetQS = &qs
			break
		}
	}

	if targetQS == nil {
		slog.Warn("QuickSetting non trouvé", "name", qsName)
		return nil
	}

	// Construire la liste des réglages d'appareils
	appSettings := make([]api.ApplianceSetting, 0, len(appliances))
	voltalisMode := haPresetToVoltalisMode[mode]

	for _, app := range appliances {
		appSettings = append(appSettings, api.ApplianceSetting{
			IDAppliance:       app.ID,
			ApplianceName:     app.Name,
			ApplianceType:     app.ApplianceType,
			Mode:              voltalisMode,
			TemperatureTarget: app.Programming.DefaultTemperature,
			IsOn:              true,
		})
	}

	// Calculer untilFurtherNotice et modeEndDate en fonction de la durée
	untilFurtherNotice := true
	var modeEndDate *string

	if duration != "" && duration != "Jusqu'à ce que je change d'avis" {
		untilFurtherNotice = false
		// Parser la durée depuis les constantes MQTT
		parsedDuration := parseDuration(duration)
		if parsedDuration > 0 {
			// Format sans timezone comme attendu par l'API Voltalis
			end := time.Now().Add(parsedDuration).Format("2006-01-02T15:04:05")
			modeEndDate = &end
		}
	}

	// Étape 1: Mettre à jour le quicksetting (sans enabled)
	updatedQS := api.QuickSettings{
		UntilFurtherNotice: untilFurtherNotice,
		AppliancesSettings: appSettings,
		ModeEndDate:        modeEndDate,
	}

	if err := apiClient.UpdateQuickSettings(targetQS.ID, updatedQS); err != nil {
		slog.Error("failed to update quicksetting", "name", qsName, "error", err)
		return err
	}
	slog.Debug("QuickSetting mis à jour", "name", qsName, "untilFurtherNotice", untilFurtherNotice)

	// Étape 2: Activer le quicksetting via l'endpoint /enable
	if err := apiClient.EnableQuickSetting(targetQS.ID, true); err != nil {
		slog.Error("failed to enable quicksetting", "name", qsName, "error", err)
		return err
	}

	slog.Info("QuickSetting activé", "name", qsName, "untilFurtherNotice", untilFurtherNotice)
	return nil
}

// handleHeaterChanges traite les changements individuels des radiateurs
// Retourne true si des changements ont été appliqués côté Voltalis
func handleHeaterChanges(apiClient *api.Client, changes map[string]interface{}, currentState state.ResourceState, quickSettings []api.QuickSettings, appliances []api.Appliance) (bool, error) {
	applied := false

	// Charger les manualSettings pour avoir les IDs
	manualSettings, err := apiClient.GetManualSettings()
	if err != nil {
		slog.Error("failed to load manual settings", "error", err)
		return false, err
	}

	// Traiter les modifications
	if modified, ok := changes["modified"].(map[int64]map[string]interface{}); ok {
		for heaterID, heaterChanges := range modified {
			slog.Info("Radiateur modifié", "id", heaterID, "changes", heaterChanges)

			heaterState := currentState.HeaterState[heaterID]

			// Pour les changements individuels de radiateurs, on utilise manualsetting
			wasApplied, err := handleSingleHeaterChange(apiClient, int(heaterID), heaterState, heaterChanges, manualSettings, appliances)
			if err != nil {
				slog.Error("failed to apply heater change", "heaterID", heaterID, "error", err)
			}
			if wasApplied {
				applied = true
			}
		}
	}

	return applied, nil
}

// handleSingleHeaterChange traite le changement d'un seul radiateur
// Retourne true si des changements ont été appliqués côté Voltalis
func handleSingleHeaterChange(apiClient *api.Client, heaterID int, heaterState state.HeaterState, changes map[string]interface{}, manualSettings []api.ManualSetting, appliances []api.Appliance) (bool, error) {
	slog.Debug("Changement individuel de radiateur détecté",
		"heaterID", heaterID,
		"presetMode", heaterState.PresetMode,
		"mode", heaterState.Mode,
		"temperature", heaterState.Temperature,
		"duration", heaterState.Duration,
	)

	// Trouver le manualSetting existant pour ce radiateur
	var existingMS *api.ManualSetting
	for _, ms := range manualSettings {
		if ms.IDAppliance == heaterID {
			existingMS = &ms
			break
		}
	}

	// Trouver l'appliance pour récupérer la température par défaut
	var appliance *api.Appliance
	for _, app := range appliances {
		if app.ID == heaterID {
			appliance = &app
			break
		}
	}

	if appliance == nil {
		slog.Warn("Appliance non trouvée", "heaterID", heaterID)
		return false, nil
	}

	// Vérifier si c'est un changement de preset (important: à vérifier AVANT le mode auto)
	_, hasPresetChange := changes["PresetMode"]
	_, hasModeChange := changes["Mode"]

	slog.Debug("Analyse des changements",
		"hasPresetChange", hasPresetChange,
		"hasModeChange", hasModeChange,
		"changes", changes,
	)

	// Cas spécial: mode "auto" SANS changement de preset = désactiver le manualSetting pour revenir à la programmation
	// Si on a un changement de preset, on veut appliquer ce preset même si le mode est "auto"
	if heaterState.Mode == state.HeaterModeAuto && !hasPresetChange {
		slog.Debug("Mode auto détecté sans changement de preset, tentative de désactivation du manualSetting", "heaterID", heaterID)
		if existingMS != nil {
			slog.Debug("ManualSetting trouvé", "id", existingMS.ID, "enabled", existingMS.Enabled)
			if existingMS.Enabled {
				slog.Info("Désactivation du manualSetting pour revenir à la programmation",
					"manualSettingID", existingMS.ID,
					"heaterID", heaterID,
				)
				request := api.UpdateManualSettingRequest{
					Enabled:            false,
					IDAppliance:        heaterID,
					UntilFurtherNotice: true,
					IsOn:               false,
					Mode:               existingMS.Mode,
					TemperatureTarget:  existingMS.TemperatureTarget,
				}
				if err := apiClient.UpdateManualSetting(existingMS.ID, request); err != nil {
					return false, err
				}
				slog.Info("ManualSetting désactivé, retour à la programmation", "heaterID", heaterID)
				return true, nil
			} else {
				slog.Debug("ManualSetting déjà désactivé", "heaterID", heaterID)
			}
		} else {
			slog.Debug("Pas de manualSetting à désactiver", "heaterID", heaterID)
		}
		return false, nil
	}

	// Cas spécial: mode "off" = éteindre le radiateur (IsOn: false)
	if heaterState.Mode == state.HeaterModeOff {
		slog.Debug("Mode off détecté, extinction du radiateur", "heaterID", heaterID)

		// Déterminer le mode à utiliser (garder l'ancien ou utiliser HORS_GEL par défaut)
		mode := "HORS_GEL"
		tempTarget := appliance.Programming.DefaultTemperature
		if existingMS != nil {
			mode = existingMS.Mode
			tempTarget = existingMS.TemperatureTarget
		}

		request := api.UpdateManualSettingRequest{
			Enabled:            true,
			IDAppliance:        heaterID,
			UntilFurtherNotice: true,
			IsOn:               false, // Radiateur éteint
			Mode:               mode,
			TemperatureTarget:  tempTarget,
		}

		if existingMS != nil {
			slog.Info("Extinction du radiateur via manualSetting existant",
				"manualSettingID", existingMS.ID,
				"heaterID", heaterID,
			)
			if err := apiClient.UpdateManualSetting(existingMS.ID, request); err != nil {
				return false, err
			}
		} else {
			slog.Info("Création d'un manualSetting pour éteindre le radiateur", "heaterID", heaterID)
			if _, err := apiClient.CreateManualSetting(request); err != nil {
				return false, err
			}
		}
		slog.Info("Radiateur éteint avec succès", "heaterID", heaterID)
		return true, nil
	}

	// Déterminer le mode Voltalis
	var voltalisMode string
	var modeExists bool

	// Récupérer la nouvelle valeur du preset si elle a changé
	presetMode := heaterState.PresetMode
	if hasPresetChange {
		// Si le preset a changé, utiliser la nouvelle valeur du changement
		if newPreset, ok := changes["PresetMode"].(state.HeaterPresetMode); ok {
			presetMode = newPreset
			slog.Debug("Utilisation du nouveau preset depuis le changement", "preset", presetMode)
		} else if newPresetStr, ok := changes["PresetMode"].(string); ok {
			presetMode = state.HeaterPresetMode(newPresetStr)
			slog.Debug("Utilisation du nouveau preset depuis le changement (string)", "preset", presetMode)
		}
	}

	// Cas 1: Mode "heat" SANS changement de preset = température manuelle
	if heaterState.Mode == state.HeaterModeHeat && !hasPresetChange {
		voltalisMode = "TEMPERATURE"
		modeExists = true
	} else if presetMode != "" && presetMode != state.HeaterPresetModeAucunMode {
		// Cas 2: Preset mode explicite (Confort, Eco, Hors-Gel)
		voltalisMode, modeExists = haPresetToVoltalisMode[presetMode]
	} else if heaterState.Mode == state.HeaterModeHeat {
		// Cas 3: Mode heat sans preset = température manuelle
		voltalisMode = "TEMPERATURE"
		modeExists = true
	}

	if !modeExists {
		// Ignorer les changements qui ne correspondent pas à un mode Voltalis valide
		// Par exemple: changement de durée sans changement de mode
		// On vérifie si c'est juste un changement de durée
		if _, hasDuration := changes["Duration"]; hasDuration && len(changes) == 1 {
			slog.Debug("Changement de durée uniquement, pas d'action nécessaire", "heaterID", heaterID)
			return false, nil
		}
		slog.Debug("Mode non reconnu, pas d'action", "presetMode", heaterState.PresetMode, "mode", heaterState.Mode)
		return false, nil
	}

	// Calculer untilFurtherNotice et endDate
	untilFurtherNotice := true
	var endDate *string

	if heaterState.Duration != "" && heaterState.Duration != "Jusqu'à ce que je change d'avis" {
		untilFurtherNotice = false
		parsedDuration := parseDuration(heaterState.Duration)
		if parsedDuration > 0 {
			end := time.Now().Add(parsedDuration).Format("2006-01-02T15:04:05")
			endDate = &end
		}
	}

	// Déterminer la température cible
	// Pour le mode TEMPERATURE, utiliser la température de l'état HA
	// Pour les autres modes (CONFORT, ECO, HORS_GEL), utiliser la température par défaut
	tempTarget := appliance.Programming.DefaultTemperature
	if voltalisMode == "TEMPERATURE" && heaterState.Temperature > 0 {
		tempTarget = heaterState.Temperature
	}

	// Construire la requête
	request := api.UpdateManualSettingRequest{
		Enabled:            true,
		IDAppliance:        heaterID,
		UntilFurtherNotice: untilFurtherNotice,
		IsOn:               true,
		Mode:               voltalisMode,
		EndDate:            endDate,
		TemperatureTarget:  tempTarget,
	}

	// Si un manualSetting existe, le mettre à jour, sinon en créer un nouveau
	if existingMS != nil {
		slog.Info("Mise à jour du manualSetting existant",
			"manualSettingID", existingMS.ID,
			"heaterID", heaterID,
			"mode", voltalisMode,
			"tempTarget", tempTarget,
		)
		if err := apiClient.UpdateManualSetting(existingMS.ID, request); err != nil {
			return false, err
		}
	} else {
		slog.Info("Création d'un nouveau manualSetting",
			"heaterID", heaterID,
			"mode", voltalisMode,
			"tempTarget", tempTarget,
		)
		if _, err := apiClient.CreateManualSetting(request); err != nil {
			return false, err
		}
	}

	slog.Info("Changement de radiateur appliqué avec succès",
		"heaterID", heaterID,
		"mode", voltalisMode,
		"tempTarget", tempTarget,
		"untilFurtherNotice", untilFurtherNotice,
	)

	return true, nil
}

// parseDuration convertit une chaîne de durée en time.Duration
func parseDuration(durationStr string) time.Duration {
	// Mapping basé sur les valeurs définies dans mqtt/enums.go
	durationMap := map[string]time.Duration{
		"Pendant 1 heure":  1 * time.Hour,
		"Pendant 2 heures": 2 * time.Hour,
		"Pendant 3 heures": 3 * time.Hour,
		"Pendant 4 heures": 4 * time.Hour,
	}

	if d, ok := durationMap[durationStr]; ok {
		return d
	}

	return 0
}
