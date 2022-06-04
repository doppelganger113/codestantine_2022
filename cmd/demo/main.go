package main

import (
	"api"
	"api/logger"
)

func main() {
	log := logger.NewLogger(logger.WithPretty())
	app, err := api.InitializeApp(log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed initializing app")
	}
	api.Bootstrap(app, log)
}
