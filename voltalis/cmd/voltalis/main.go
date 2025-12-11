package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/francois76/voltalis-integration/voltalis/internal/api"
	"github.com/francois76/voltalis-integration/voltalis/internal/config"
	"github.com/francois76/voltalis-integration/voltalis/internal/logger"
	"github.com/francois76/voltalis-integration/voltalis/internal/mqtt"
	"github.com/francois76/voltalis-integration/voltalis/internal/scheduler"
	"github.com/francois76/voltalis-integration/voltalis/internal/transform"
	"golang.org/x/sync/errgroup"
)

func main() {
	logger.InitLogs()
	opts, err := config.LoadOptions()
	if err != nil {
		panic(err)
	}
	slog.With("options", opts).Info("loading options")
	mqttClient, err := mqtt.InitClient("tcp://"+opts.MqttURL, "voltalis-addon", opts.MqttPassword)
	if err != nil {
		panic(err)
	}

	apiClient, err := api.NewClient("https://api.myvoltalis.com", opts.VoltalisLogin, opts.VoltalisPassword)
	if err != nil {
		panic(err)
	}

	g, ctx := errgroup.WithContext(context.Background())

	s := scheduler.New(15*time.Second, func() error {
		return transform.SyncVoltalisHeatersToHA(mqttClient, apiClient)
	})
	s.Trigger()
	g.Go(func() error {

		s.Start()
		return <-s.Err()
	})

	g.Go(func() error {
		return transform.Start(ctx, mqttClient, apiClient, s)
	})

	// Attendre que toutes les goroutines se terminent
	// Si une goroutine retourne une erreur, les autres seront annulées via le contexte
	if err := g.Wait(); err != nil {
		panic(fmt.Sprintf("Une goroutine a échoué: %v", err))
	}

}
