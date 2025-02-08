package main

import (
	"net/http"

	"receipt_processor/pkg/api"
	"receipt_processor/pkg/database"
	"receipt_processor/pkg/middleware"
	"receipt_processor/pkg/repository"
	"receipt_processor/pkg/service"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	// Initialize logger.
	initLogger()

	// Initialize configuration using Viper.
	initViper()

	// Set up the SQLite database.
	dbPath := viper.GetString("database.path")
	if dbPath == "" {
		log.Fatal().Msg("database.path is not set in configuration")
	}
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	// Initialize the repositories.
	receiptRepo := repository.NewReceiptRepository(db)

	// Initialize the service layer.
	receiptService := service.NewReceiptService(receiptRepo)

	// Prepare middleware.
	requestIDMiddleware := middleware.RequestIDMiddleware()

	// Set up the API router with handlers and middleware.
	router := api.NewRouter(receiptService, []middleware.Middleware{
		requestIDMiddleware,
	})

	// Determine the server port from configuration, or default to "8080".
	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}
	log.Info().Msgf("Server starting on port %s", port)

	// Start the HTTP server.
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal().Err(err).Msg("Server failed")
	}
}

func initLogger() {
	// Use Unix time for timestamps.
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// Set the default global logging level.
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// Include caller information in logs.
	log.Logger = log.With().Caller().Logger()
}

func initViper() {
	viper.SetConfigName("config") // Config file name (without extension)
	viper.AddConfigPath("config") // Look for the config file in the "config" directory
	viper.AutomaticEnv()          // Override config with environment variables if set
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal().Msg("No configuration file loaded; defaults will be used")
	}
}
