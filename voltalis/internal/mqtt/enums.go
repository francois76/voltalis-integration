package mqtt

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
)

const TEMPERATURE_NONE = "None"
