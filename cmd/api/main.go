package main

import (
	"log"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/app/api"
	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("API: .env file not found, using system envs")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed no load config: %v", err)
	}

	appLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("Failled no initialize logger: %v", err)
	}

	if err := postgres.RunMigrations(*cfg, appLogger); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	app, err := api.New(cfg, appLogger)
	if err != nil {
		log.Fatalf("failed to init api app: %v", err)
	}
	if err := app.Run(); err != nil {
		log.Fatalf("API server error: %v", err)
	}

	appLogger.Info("API server starting", logger.Field{Key: "port", Value: cfg.Port})
	time.Sleep(time.Second * 1)
	appLogger.Info("API server stopped")
}
