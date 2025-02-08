package main

import (
	"net/http"

	"receipt_processor/pkg/api"
	"receipt_processor/pkg/database"
	"receipt_processor/pkg/middleware"
	"receipt_processor/pkg/redis"
	"receipt_processor/pkg/repository"
	"receipt_processor/pkg/service"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func main() {
	// Initialize the logger.
	initLogger()

	// Initialize configuration using Viper.
	initViper()

	// Initialize Redis client.
	redisClient, err := redis.New()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize Redis")
	}

	// Set up the SQLite database.
	dbPath := viper.GetString("database.path")
	if dbPath == "" {
		log.Fatal().Msg("database.path is not set in configuration")
	}
	db, err := database.New(dbPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to the database")
	}

	// Initialize the receipt repository and service.
	receiptRepo := repository.NewReceiptRepository(db)
	receiptService := service.NewReceiptService(receiptRepo)

	// Initialize the rate limiter repository and middleware.
	rateLimiterRepo := repository.NewRateLimiterRepository(redisClient.Rdb)
	rateLimiterMiddleware := middleware.RateLimitMiddleware(rateLimiterRepo)

	// Combine middleware: e.g., request ID and rate limiter.
	middlewares := []middleware.Middleware{
		middleware.RequestIDMiddleware(),
		rateLimiterMiddleware,
	}

	// Set up the API router with handlers and the middleware chain.
	router := api.NewRouter(receiptService, middlewares)

	// Determine the server port.
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
	// Set the global logging level.
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	// Include caller information in logs.
	log.Logger = log.With().Caller().Logger()
}

func initViper() {
	// Set the config file name and path.
	viper.SetConfigName("config")
	viper.AddConfigPath("config")
	viper.AutomaticEnv() // Override config with environment variables if set.
	if err := viper.ReadInConfig(); err != nil {
		log.Warn().Msg("No configuration file loaded; defaults will be used")
	}
}
