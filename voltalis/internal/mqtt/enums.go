package mqtt

// Modes pour les radiateurs Voltalis
type HeaterMode string

const (
	HeaterModeOff  HeaterMode = "off"
	HeaterModeAuto HeaterMode = "auto"
	HeaterModeHeat HeaterMode = "heat"
)

// PresetModes pour les radiateurs Voltalis
type HeaterPresetMode string

const (
	HeaterPresetEco  HeaterPresetMode = "eco"
	HeaterPresetAway HeaterPresetMode = "away"
	HeaterPresetHome HeaterPresetMode = "home"
)

// Modes pour les radiateurs Voltalis

type component string

const (
	ComponentClimate component = "climate"
	ComponentSelect  component = "select"
)
