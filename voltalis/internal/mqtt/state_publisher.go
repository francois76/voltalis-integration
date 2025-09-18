package mqtt

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"maps"
	"sync"
)

// ResourceState représente l'état global de votre ressource
type ResourceState struct {
	ControllerState ControllerState
	HeaterState     map[int64]HeaterState
}

type ControllerState struct {
	Duration string
	Mode     string
	Program  string
}

type HeaterState struct {
	Duration    string
	Mode        string
	Temperature float64
}

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

// Fonction générique pour comparer des pointeurs
func comparePointers[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

// Fonction générique pour comparer des maps
func compareMaps[K comparable, V Comparable](previous, current map[K]V) map[string]interface{} {
	changes := make(map[string]interface{})

	// Éléments supprimés
	var removed []K
	for key := range previous {
		if _, exists := current[key]; !exists {
			removed = append(removed, key)
		}
	}
	if len(removed) > 0 {
		changes["removed"] = removed
	}

	// Éléments ajoutés et modifiés
	added := make(map[K]V)
	modified := make(map[K]map[string]interface{})

	for key, currentValue := range current {
		if previousValue, exists := previous[key]; exists {
			// Élément existant - vérifier les changements
			if itemChanges := previousValue.Compare(currentValue); len(itemChanges) > 0 {
				modified[key] = itemChanges
			}
		} else {
			// Nouvel élément
			added[key] = currentValue
		}
	}

	if len(added) > 0 {
		changes["added"] = added
	}
	if len(modified) > 0 {
		changes["modified"] = modified
	}

	return changes
}

func compareResourceState(previous, current ResourceState) map[string]interface{} {
	changes := make(map[string]interface{})

	// Comparaison du ControllerState
	if controllerChanges := previous.ControllerState.Compare(current.ControllerState); len(controllerChanges) > 0 {
		changes["ControllerState"] = controllerChanges
	}

	// Comparaison des HeaterState avec la fonction générique
	if heaterChanges := compareMaps(previous.HeaterState, current.HeaterState); len(heaterChanges) > 0 {
		changes["HeaterState"] = heaterChanges
	}

	return changes
}

// StateChange représente un changement d'état avec les valeurs modifiées
type StateChange struct {
	CurrentState  ResourceState  `json:"current_state"`
	ChangedFields map[string]any `json:"changed_fields"`
	PreviousHash  string         `json:"previous_hash"`
	CurrentHash   string         `json:"current_hash"`
}

// StateManager gère les états et détecte les changements
type StateManager struct {
	mu           sync.RWMutex
	currentState *ResourceState
	previousHash string
	stateChannel chan StateChange
	subscribers  []chan StateChange
}

// NewStateManager crée une nouvelle instance du gestionnaire d'état
func NewStateManager() *StateManager {
	return &StateManager{
		stateChannel: make(chan StateChange, 100), // Buffer pour éviter les blocages
		subscribers:  make([]chan StateChange, 0),
	}
}

// Subscribe permet de s'abonner aux changements d'état
func (sm *StateManager) Subscribe() <-chan StateChange {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	subscriber := make(chan StateChange, 10)
	sm.subscribers = append(sm.subscribers, subscriber)
	return subscriber
}

// computeStateHash calcule un hash de l'état pour la déduplication
func (sm *StateManager) computeStateHash(state ResourceState) string {
	data, _ := json.Marshal(state)
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// UpdateState met à jour l'état et notifie les changements si nécessaire
func (sm *StateManager) UpdateState(newState ResourceState) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	currentHash := sm.computeStateHash(newState)

	// Vérification de déduplication
	if currentHash == sm.previousHash {
		log.Println("État identique détecté, pas de notification")
		return
	}

	var changedFields map[string]any

	// Si on a un état précédent, on calcule les différences
	if sm.currentState != nil {
		changedFields = compareResourceState(*sm.currentState, newState)
	} else {
		// Premier état : tous les champs sont "nouveaux"
		changedFields = map[string]any{
			"initial_state": true,
		}
	}

	// Création du StateChange
	stateChange := StateChange{
		CurrentState:  newState,
		ChangedFields: changedFields,
		PreviousHash:  sm.previousHash,
		CurrentHash:   currentHash,
	}

	// Mise à jour de l'état interne
	sm.currentState = &newState
	sm.previousHash = currentHash

	// Notification des abonnés
	sm.notifySubscribers(stateChange)
}

// notifySubscribers envoie le changement à tous les abonnés
func (sm *StateManager) notifySubscribers(change StateChange) {
	for _, subscriber := range sm.subscribers {
		select {
		case subscriber <- change:
			// Envoi réussi
		default:
			slog.Info("Abonné occupé, changement ignoré")
		}
	}
}

// GetCurrentState retourne l'état actuel (thread-safe)
func (sm *StateManager) GetCurrentState() ResourceState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.currentState == nil {
		return ResourceState{}
	}
	stateCopy := *sm.currentState
	stateCopy.HeaterState = maps.Clone(sm.currentState.HeaterState)
	return stateCopy
}
