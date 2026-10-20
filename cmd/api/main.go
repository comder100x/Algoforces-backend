package main

import (
	"algoforces/internal/app"
	"algoforces/internal/server"

	"github.com/rs/zerolog/log"
)

//	@title			Algoforces API
//	@version		1.0
//	@description	API for Algoforces application

//	@host		localhost:8080
//	@BasePath	/

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Enter your token only (without Bearer prefix)

func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize application")
	}
	defer application.Close()

	log.Info().Msg("Starting Algoforces API on :8080")
	if err := server.Run(application.Router, ":8080"); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
