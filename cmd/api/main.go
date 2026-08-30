package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/OrkhanNajaf1i/booking-service/docs"
	"github.com/OrkhanNajaf1i/booking-service/internal/app/api"
	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/joho/godotenv"
)

// @title           Booking Service API
// @version         2.0
// @description     Booking Platformasi ucun Backend API. Randevu vaxtlari is qrafiki, nahar fasilesi ve slot addimi esasinda runtime-da hesablanir; bildirisler WebSocket ve FCM ile realtime catdirilir.
// @host            booking-service-sld9.onrender.com
// @schemes         https
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("API: .env file not found, using system envs")
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

	app, err := api.New(cfg, appLogger)
	if err != nil {
		log.Fatalf("failed to init api app: %v", err)
	}

	// Server ayri goroutine-de qalxir ki, esas axin siqnali gozleye bilsin.
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- app.Run()
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErrors:
		if err != nil {
			log.Fatalf("API server error: %v", err)
		}
	case <-ctx.Done():
		appLogger.Info("Shutdown siqnali alindi")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := app.Shutdown(shutdownCtx); err != nil {
			appLogger.Error("Graceful shutdown ugursuz",
				logger.Field{Key: "error", Value: err.Error()},
			)
		}
	}

	appLogger.Info("API server stopped")
}
