package main

import (
	"context"
	"fmt"
	"os"

	"github.com/container-registry/harbor-satellite/internal/config"
	"github.com/container-registry/harbor-satellite/internal/logger"
	"github.com/container-registry/harbor-satellite/internal/satellite"
	"github.com/container-registry/harbor-satellite/internal/state"
	"github.com/container-registry/harbor-satellite/internal/utils"
	"github.com/container-registry/harbor-satellite/registry"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := utils.SetupContext(context.Background())
	defer cancel()

	ctx, wg, scheduler, err := utils.Init(ctx)
	if err != nil {
		return err
	}
	log := logger.FromContext(ctx)

	go scheduler.ListenForProcessEvent()

	// Handle registry setup
	if err := handleRegistrySetup(wg, log, cancel); err != nil {
		log.Error().Err(err).Msg("Error setting up local registry")
		return err
	}

	err = scheduler.Start()
	if err != nil {
		log.Error().Err(err).Msg("Error starting scheduler")
		return err
	}
	defer scheduler.Stop()

	localRegistryConfig := state.NewRegistryConfig(config.GetRemoteRegistryURL(), config.GetRemoteRegistryUsername(), config.GetRemoteRegistryPassword())
	sourceRegistryConfig := state.NewRegistryConfig(config.GetSourceRegistryURL(), config.GetSourceRegistryUsername(), config.GetSourceRegistryPassword())
	satelliteService := satellite.NewSatellite(ctx, scheduler.GetSchedulerKey(), localRegistryConfig, sourceRegistryConfig, config.UseUnsecure(), config.GetState())

	wg.Go(func() error {
		return satelliteService.Run(ctx)
	})

	return wg.Wait()
}

func handleRegistrySetup(g *errgroup.Group, log *zerolog.Logger, cancel context.CancelFunc) error {
	log.Debug().Msg("Setting up local registry")
	if config.GetOwnRegistry() {
		log.Info().Msg("Configuring own registry")
		if err := utils.HandleOwnRegistry(); err != nil {
			log.Error().Err(err).Msg("Error handling own registry")
			return err
		}
	} else {
		var defaultZotConfig registry.DefaultZotConfig
		err := registry.ReadConfig(config.GetZotConfigPath(), &defaultZotConfig)
		if err != nil {
			return fmt.Errorf("error reading config: %w", err)
		}
		defaultZotURL := defaultZotConfig.GetLocalRegistryURL()
		if err := config.SetRemoteRegistryURL(defaultZotURL); err != nil {
			log.Error().Err(err).Msg("Error writing the remote registry URL")
			cancel()
			return err
		}
		g.Go(func() error {
			log.Info().Msg("Launching default registry")
			if err := utils.LaunchDefaultZotRegistry(); err != nil {
				log.Error().Err(err).Msg("Error launching default registry")
				cancel()
				return err
			}
			cancel()
			return nil
		})
	}
	return nil
}
