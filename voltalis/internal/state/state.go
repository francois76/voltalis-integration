package state

// ResourceState représente l'état global de votre ressource
type ResourceState struct {
	ControllerState ControllerState
	HeaterState     map[int64]HeaterState
}

type ControllerState struct {
	Duration string
	Mode     HeaterPresetMode
	Program  string
}

type HeaterState struct {
	Duration    string
	PresetMode  HeaterPresetMode
	Mode        HeaterMode
	Temperature float64
}

// Modes pour les radiateurs Voltalis
type HeaterMode string

const (
	HeaterModeOff  HeaterMode = "off"
	HeaterModeAuto HeaterMode = "auto"
	HeaterModeHeat HeaterMode = "heat"
	HeaterModeNone HeaterMode = "none"
)

type HeaterPresetMode string

const (
	HeaterPresetModeConfort   HeaterPresetMode = "Confort"
	HeaterPresetModeEco       HeaterPresetMode = "Eco"
	HeaterPresetModeHorsGel   HeaterPresetMode = "Hors-Gel"
	HeaterPresetModeAucunMode HeaterPresetMode = "Aucun mode"
)

// Interface pour les types qui peuvent être comparés
type Comparable interface {
	Compare(other Comparable) map[string]interface{}
}

// Implémentation de Compare pour ControllerState
func (cs ControllerState) Compare(other Comparable) map[string]interface{} {
	otherCS := other.(ControllerState)
	changes := make(map[string]interface{})

	if cs.Duration != otherCS.Duration {
		changes["Duration"] = otherCS.Duration
	}
	if cs.Mode != otherCS.Mode {
		changes["Mode"] = otherCS.Mode
	}
	if cs.Program != otherCS.Program {
		changes["Program"] = otherCS.Program
	}

	return changes
}

// Implémentation de Compare pour HeaterState
func (hs HeaterState) Compare(other Comparable) map[string]interface{} {
	otherHS := other.(HeaterState)
	changes := make(map[string]interface{})
	if hs.Duration != otherHS.Duration {
		changes["Duration"] = otherHS.Duration
	}
	if hs.PresetMode != otherHS.PresetMode {
		changes["PresetMode"] = otherHS.PresetMode
	}
	if hs.Mode != otherHS.Mode {
		changes["Mode"] = otherHS.Mode
	}
	if hs.Temperature != otherHS.Temperature {
		changes["Temperature"] = otherHS.Temperature
	}

	return changes
}
