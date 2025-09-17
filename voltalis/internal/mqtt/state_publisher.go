package mqtt

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"sync"
)

// ResourceState représente l'état global de votre ressource
type ResourceState struct {
	ID int64 `json:"id"`
}

// StateChange représente un changement d'état avec les valeurs modifiées
type StateChange struct {
	CurrentState  ResourceState          `json:"current_state"`
	ChangedFields map[string]interface{} `json:"changed_fields"`
	PreviousHash  string                 `json:"previous_hash"`
	CurrentHash   string                 `json:"current_hash"`
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

// findChangedFields compare deux états et retourne les champs modifiés
func (sm *StateManager) findChangedFields(previous, current ResourceState) map[string]interface{} {
	changes := make(map[string]interface{})

	prevValue := reflect.ValueOf(previous)
	currValue := reflect.ValueOf(current)
	prevType := reflect.TypeOf(previous)

	for i := 0; i < prevValue.NumField(); i++ {
		field := prevType.Field(i)

		// Skip le timestamp qui change toujours
		if field.Name == "Timestamp" {
			continue
		}

		prevFieldValue := prevValue.Field(i)
		currFieldValue := currValue.Field(i)

		// Comparaison spéciale pour les maps
		if field.Name == "Metadata" {
			if !reflect.DeepEqual(prevFieldValue.Interface(), currFieldValue.Interface()) {
				changes[field.Name] = currFieldValue.Interface()
			}
		} else if !reflect.DeepEqual(prevFieldValue.Interface(), currFieldValue.Interface()) {
			changes[field.Name] = currFieldValue.Interface()
		}
	}

	return changes
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

	var changedFields map[string]interface{}

	// Si on a un état précédent, on calcule les différences
	if sm.currentState != nil {
		changedFields = sm.findChangedFields(*sm.currentState, newState)
	} else {
		// Premier état : tous les champs sont "nouveaux"
		changedFields = map[string]interface{}{
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
			fmt.Println(subscriber)
			// Envoi réussi
		default:
			log.Println("Abonné occupé, changement ignoré")
		}
	}
}

// GetCurrentState retourne l'état actuel (thread-safe)
func (sm *StateManager) GetCurrentState() *ResourceState {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if sm.currentState == nil {
		return nil
	}

	// Retour d'une copie pour éviter les modifications concurrentes
	state := *sm.currentState
	return &state
}
