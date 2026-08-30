package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/OrkhanNajaf1i/booking-service/internal/app/worker"
	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("WORKER: .env file not found, using system envs")
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	appLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	if err := postgres.RunMigrations(*cfg, appLogger); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}

	app, err := worker.New(cfg, appLogger)
	if err != nil {
		log.Fatalf("failed to init worker: %v", err)
	}
	defer app.Close()

	// SIGINT/SIGTERM gelende ctx legv olunur ve Run() temiz cixir.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("worker error: %v", err)
	}

	appLogger.Info("Worker dayandi")
}
