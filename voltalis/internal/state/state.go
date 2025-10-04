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
	Mode        HeaterMode
	Temperature float64
}

// HeaterMode correspond aux modes possibles fondamentalement pour un radiateur, il s'agit d'une simplification qui donne lieu à des traductions en terme de modélisation
// à la fois coté mqtt et coté voltalis
type HeaterMode string

const (
	HeaterModeOff     HeaterMode = "off"
	HeaterModeManual  HeaterMode = "heat" // correspond au mode manuel, mais nommé tel quel pour faciliter le cast côté voltalis
	HeaterModeConfort HeaterMode = "Confort"
	HeaterModeEco     HeaterMode = "Eco"
	HeaterModeHorsGel HeaterMode = "Hors-Gel"
)

type HeaterPresetMode string

const (
	HeaterPresetModeConfort   HeaterPresetMode = "Confort"
	HeaterPresetModeEco       HeaterPresetMode = "Eco"
	HeaterPresetModeHorsGel   HeaterPresetMode = "Hors-Gel"
	HeaterPresetModeAucunMode HeaterPresetMode = "none"
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
	if hs.Mode != otherHS.Mode {
		changes["Mode"] = otherHS.Mode
	}
	if hs.Temperature != otherHS.Temperature {
		changes["Temperature"] = otherHS.Temperature
	}

	return changes
}
