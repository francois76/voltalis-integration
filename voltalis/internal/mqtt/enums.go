package mqtt

import (
	"fmt"
	"time"
)

func init() {
	during := func(n int64, unit string) string {
		plural := "s"
		if n == 1 {
			plural = ""
		}
		return fmt.Sprintf("Pendant %d %s%s", n, unit, plural)
	}
	for _, i := range []int64{1, 2, 3, 4} {
		DURATION_NAMES_TO_VALUES[during(i, "heure")] = time.Duration(i) * time.Hour
	}
	DURATION_NAMES_TO_VALUES["Jusqu'à ce que je change d'avis"] = time.Duration(0)

}

// Modes pour les radiateurs Voltalis
type HeaterMode string

const (
	HeaterModeOff  HeaterMode = "off"
	HeaterModeAuto HeaterMode = "auto"
	HeaterModeHeat HeaterMode = "heat"
	HeaterModeNone HeaterMode = "none"
)

// PresetModes pour les radiateurs Voltalis
type HeaterPresetMode string

const (
	HeaterPresetModeConfort   HeaterPresetMode = "Confort"
	HeaterPresetModeEco       HeaterPresetMode = "Eco"
	HeaterPresetModeHorsGel   HeaterPresetMode = "Hors-Gel"
	HeaterPresetModeManuel    HeaterPresetMode = "Manuel"
	HeaterPresetModeAucunMode HeaterPresetMode = "Aucun mode"
	HeaterPresetModeNone      HeaterPresetMode = "none"
)

var PRESET_SELECT_ONE_HEATER []HeaterPresetMode = []HeaterPresetMode{HeaterPresetModeConfort, HeaterPresetModeEco, HeaterPresetModeHorsGel}
var PRESET_SELECT_CONTROLLER []HeaterPresetMode = []HeaterPresetMode{HeaterPresetModeConfort, HeaterPresetModeEco, HeaterPresetModeHorsGel, HeaterPresetModeAucunMode}

type HeaterAction string

const (
	HeaterActionOff        HeaterAction = "off"
	HeaterActionIdle       HeaterAction = "idle"
	HeaterActionHeating    HeaterAction = "heating"
	HeaterActionPreheating HeaterAction = "preheating"
	HeaterActionCooling    HeaterAction = "cooling"
)

// Modes pour les radiateurs Voltalis

type component string

const (
	ComponentClimate component = "climate"
	ComponentSelect  component = "select"
	ComponentSensor  component = "sensor"
	ComponentButton  component = "button"
)

var DURATION_NAMES_TO_VALUES = map[string]time.Duration{}

const TEMPERATURE_NONE = "None"
