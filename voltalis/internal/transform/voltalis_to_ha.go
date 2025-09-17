package transform

import (
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func SyncVoltalisHeatersToHA(client *mqtt.Client) error {
	// client.PublishState(heaterGetTopics[12345678901234].CurrentTemperature, "19.5")
	// client.PublishState(heaterGetTopics[12345678901234].Mode, "heat")
	// client.PublishState(heaterGetTopics[12345678901234].Temperature, "21")
	// client.PublishState(heaterGetTopics[23456789012345].CurrentTemperature, mqtt.TEMPERATURE_NONE)
	// client.PublishState(heaterGetTopics[23456789012345].Mode, "auto")
	// client.PublishState(heaterGetTopics[23456789012345].Temperature, mqtt.TEMPERATURE_NONE))
	return nil
}
