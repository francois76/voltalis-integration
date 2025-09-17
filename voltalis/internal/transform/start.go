package transform

import (
	"context"
	"fmt"
	"time"

	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func Start(client *mqtt.Client) error {
	if err := client.RegisterController(); err != nil {
		return err
	}
	if err := client.RegisterHeater(12345678901234, "Salon"); err != nil {
		return err
	}
	if err := client.RegisterHeater(23456789012345, "Chambre"); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stateChanges := client.StateManager.Subscribe()

	// Goroutine pour traiter les changements d'état
	go func() {
		for {
			select {
			case change := <-stateChanges:
				fmt.Printf("🔄 Changement détecté!\n")
				fmt.Printf("   Hash précédent: %s\n", change.PreviousHash[:8]+"...")
				fmt.Printf("   Hash actuel: %s\n", change.CurrentHash[:8]+"...")
				fmt.Printf("   État actuel: %+v\n", change.CurrentState)
				fmt.Printf("   Champs modifiés: %+v\n", change.ChangedFields)

				// Ici vous pourriez traiter les changements
				// - Sauvegarder en base
				// - Envoyer des alertes
				// - Déclencher des actions

			case <-ctx.Done():
				return
			}
		}
	}()

	// Simulation des messages MQTT
	go simulateMQTTMessages(client.StateManager)

	// Laisser tourner pendant quelques secondes
	time.Sleep(10 * time.Second)

	fmt.Println("\n--- État final ---")
	finalState := client.StateManager.GetCurrentState()
	if finalState != nil {
		fmt.Printf("État final: %+v\n", *finalState)
	}
	return nil
}

// Simulation d'un listener MQTT
func simulateMQTTMessages(stateManager *mqtt.StateManager) {
	// Simulation de messages MQTT avec différents états
	states := []mqtt.ResourceState{
		{
			ID:          "sensor_001",
			Status:      "active",
			Temperature: 23.5,
			Humidity:    60.0,
			Metadata:    map[string]interface{}{"location": "living_room"},
		},
		{
			ID:          "sensor_001",
			Status:      "active",
			Temperature: 24.1, // Changement
			Humidity:    60.0,
			Metadata:    map[string]interface{}{"location": "living_room"},
		},
		{
			ID:          "sensor_001",
			Status:      "active",
			Temperature: 24.1, // Pas de changement
			Humidity:    60.0,
			Metadata:    map[string]interface{}{"location": "living_room"},
		},
		{
			ID:          "sensor_001",
			Status:      "warning", // Changement
			Temperature: 24.1,
			Humidity:    65.5, // Changement
			Metadata:    map[string]interface{}{"location": "living_room", "alert": "high_humidity"},
		},
	}

	for i, state := range states {
		fmt.Printf("\n--- Message MQTT %d ---\n", i+1)
		stateManager.UpdateState(state)
		time.Sleep(2 * time.Second)
	}
}
