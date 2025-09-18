package transform

import (
	"context"
	"log/slog"

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
			slog.With("change", change.ChangedFields).Debug("champs modifiés")

		case <-ctx.Done():
			slog.Warn("context killed")
			return nil
		}
	}

}
