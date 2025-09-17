package transform

import (
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func RegisterDevices(client *mqtt.Client) error {
	if err := client.RegisterController(); err != nil {
		return err
	}
	if err := client.RegisterHeater(12345678901234, "Salon"); err != nil {
		return err
	}
	if err := client.RegisterHeater(23456789012345, "Chambre"); err != nil {
		return err
	}
	return nil
}
