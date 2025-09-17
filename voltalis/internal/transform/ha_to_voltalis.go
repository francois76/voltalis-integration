package transform

import (
	"context"
	"fmt"

	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

// Start est le point de démarrage de la fonction qui process les évenements MQTT de façon globalisée et appelle les APIs de voltalis pour répliquer les changements
func Start(ctx context.Context, mqttClient *mqtt.Client) error {
	if err := mqttClient.RegisterController(); err != nil {
		return err
	}
	if err := mqttClient.RegisterHeater(12345678901234, "Salon"); err != nil {
		return err
	}
	if err := mqttClient.RegisterHeater(23456789012345, "Chambre"); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	stateChanges := mqttClient.StateManager.Subscribe()

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
			fmt.Println("context killed")
			return nil
		}
	}

}
