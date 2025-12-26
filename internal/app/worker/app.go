package worker

import (
	"context"
	"database/sql"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

type App struct {
	config       *config.AppConfig
	logger       logger.Logger
	db           *sql.DB
	pollInterval time.Duration
}

func New(cfg *config.AppConfig) (*App, error) {
	logg, err := logger.New(cfg)
	if err != nil {
		return nil, err
	}
	db, err := postgres.New(*cfg)
	if err != nil {
		return nil, err
	}

	return &App{
		config:       cfg,
		logger:       logg,
		db:           db,
		pollInterval: time.Second * 10,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	a.logger.Info("Worker starting", logger.Field{Key: "pollInterval", Value: a.pollInterval.String()})

	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Worker stopping")
			return ctx.Err()
		case <-ticker.C:
			// Worker işi burada olacaq
			a.logger.Info("Worker polling")
		}
	}
}
