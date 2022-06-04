package api

import (
	"api/core"
	"api/http_server"
	"context"
	"github.com/rs/zerolog"
	"time"
)

var (
	GitCommit string
	BuildDate string
	Version   string
)

func Bootstrap(app *core.App, logger *zerolog.Logger, configurations ...Configure) {
	boostrapConfig := newConfiguration(configurations...)

	logger.Info().Msgf("[Build]: Version %s BuildDate %s GitCommit %s", Version, BuildDate, GitCommit)

	appCtx := context.Background()
	defer appCtx.Done()

	errChannel := make(chan error)

	appInitTimeout, appInitCancel := context.WithTimeout(
		appCtx, boostrapConfig.initTimeout,
	)
	defer appInitCancel()

	logger.Info().Msg("Initializing app...")

	if err := app.Init(appInitTimeout, appCtx); err != nil {
		logger.Fatal().Msgf("failed initializing app: %s", err.Error())
		return
	}

	server, err := http_server.StartNewConfiguredAndListenChannel(logger,
		http_server.Handlers{
			ImagesHandler: app.ImagesService,
			Authenticator: app.Auth,
		}, errChannel)
	if err != nil {
		logger.Fatal().Msgf("failed starting the server: %s", err.Error())
		return
	}
	go boostrapConfig.interrupt(errChannel)

	fatalErr := <-errChannel
	logger.Info().Msgf("Closing server: %s", fatalErr.Error())

	shutdownGracefully(app, logger, server, boostrapConfig.shutdownTimeout)
}

func shutdownGracefully(app *core.App, logger *zerolog.Logger, server *http_server.Server, timeout time.Duration) {
	if app == nil && server == nil {
		return
	}
	logger.Info().Msg("Gracefully shutting down...")

	gracefullCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if server != nil {
		if err := server.Shutdown(gracefullCtx); err != nil {
			logger.Error().Msgf("Error shutting down the server: %s", err.Error())
		} else {
			logger.Info().Msg("HttpServer gracefully shut down")
		}
	}

	if app != nil {
		if err := app.Shutdown(gracefullCtx); err != nil {
			logger.Error().Msg(err.Error())
		} else {
			logger.Info().Msg("application shut down successfully")
		}
	}
}
