package main

import (
	"context"
	"fmt"
	"time"

	"github.com/francois76/voltalis-integration/voltalis/internal/logger"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
	"github.com/francois76/voltalis-integration/voltalis/internal/scheduler"
	"github.com/francois76/voltalis-integration/voltalis/internal/transform"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger.InitLogs()
	client, err := mqtt.InitClient("tcp://localhost:1883", "voltalis-addon")
	if err != nil {
		panic(err)
	}

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		return scheduler.Run(15*time.Second, func() error {
			return transform.SyncVoltalisHeatersToHA(client)
		})
	})

	g.Go(func() error {
		return transform.Start(ctx, client)
	})

	// Attendre que toutes les goroutines se terminent
	// Si une goroutine retourne une erreur, les autres seront annulées via le contexte
	if err := g.Wait(); err != nil {
		panic(fmt.Sprintf("Une goroutine a échoué: %v", err))
	}

}
