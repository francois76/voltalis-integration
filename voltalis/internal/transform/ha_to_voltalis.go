package transform

import (
	"context"
	"log/slog"

	"github.com/francois76/voltalis-integration/voltalis/internal/api"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

// Start est le point de démarrage de la fonction qui process les évenements MQTT de façon globalisée et appelle les APIs de voltalis pour répliquer les changements
func Start(ctx context.Context, mqttClient *mqtt.Client, apiClient *api.Client) error {
	if err := mqttClient.RegisterController(); err != nil {
		return err
	}
	appliances, err := apiClient.GetAppliances()
	if err != nil {
		return err
	}
	for _, appliance := range appliances {
		if err := mqttClient.RegisterHeater(int64(appliance.ID), appliance.Name); err != nil {
			return err
		}
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
