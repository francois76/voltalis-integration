package transform

import (
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func SyncVoltalisHeatersToHA(client *mqtt.Client, heaterWriteTopics map[int64]mqtt.HeaterWriteTopics) {
	// client.PublishState(heaterWriteTopics[12345678901234].CurrentTemperature, "19.5")
	// client.PublishState(heaterWriteTopics[12345678901234].Mode, "heat")
	// client.PublishState(heaterWriteTopics[12345678901234].Temperature, "21")
	// client.PublishState(heaterWriteTopics[23456789012345].CurrentTemperature, mqtt.TEMPERATURE_NONE)
	// client.PublishState(heaterWriteTopics[23456789012345].Mode, "auto")
	// client.PublishState(heaterWriteTopics[23456789012345].Temperature, mqtt.TEMPERATURE_NONE))
}
