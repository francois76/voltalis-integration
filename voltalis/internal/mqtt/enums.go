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
	HeaterPresetModeConfort HeaterPresetMode = "Confort"
	HeaterPresetModeEco     HeaterPresetMode = "Eco"
	HeaterPresetModeHorsGel HeaterPresetMode = "Hors-Gel"
	HeaterPresetModeManuel  HeaterPresetMode = "Manuel"
	HeaterPresetModeArret   HeaterPresetMode = "Arret"
	HeaterPresetModeNone    HeaterPresetMode = "none"
)

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
