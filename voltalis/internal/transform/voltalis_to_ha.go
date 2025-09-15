package transform

import (
	"fmt"

	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
)

func SyncVoltalisToHA(client *mqtt.Client, heaterWriteTopics mqtt.HeaterWriteTopics) {
	client.PublishState(heaterWriteTopics.CurrentTemperature, fmt.Sprintf("%.1f", 19.5))
	client.PublishState(heaterWriteTopics.Mode, "heat")
	client.PublishState(heaterWriteTopics.Temperature, "21")
}
