package main

import (
	"time"

	"github.com/francois76/voltalis-integration/voltalis/internal/logger"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
	"github.com/francois76/voltalis-integration/voltalis/internal/scheduler"
	"github.com/francois76/voltalis-integration/voltalis/internal/transform"
)

func main() {
	logger.InitLogs()
	client, err := mqtt.InitClient("tcp://localhost:1883", "voltalis-addon")
	if err != nil {
		panic(err)
	}
	heaterTopics := transform.RegisterDevices(client)
	scheduler.Run(15*time.Second, func() {
		transform.SyncVoltalisHeaterToHA(client, heaterTopics)
	})

}
