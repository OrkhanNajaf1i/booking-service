package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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
		log.Fatal("Failed to load config: %v", err)
	}
	if err := postgres.RunMigrations(*cfg); err != nil {
		log.Fatalf("migrations failed: %v", err)
	}
	appLogger, err := logger.New(cfg)
	if err != nil {
		log.Fatal("Failed to initialize logger: %v", err)
	}
	appLogger.Info("Worker starting")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			appLogger.Info("Processing reminders...")
		case <-stop:
			appLogger.Info("Worker shutting down gracefully")
		}
	}
}
