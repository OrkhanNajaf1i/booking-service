package api

import (
	"fmt"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/user"
	httpapi "github.com/OrkhanNajaf1i/booking-service/internal/http"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

type App struct {
	cfg    *config.AppConfig
	logger logger.Logger
	server *http.Server
}

func New(cfg *config.AppConfig) (*App, error) {
	appLogger, err := logger.New(cfg)
	if err != nil {
		return nil, err
	}
	db, err := postgres.New(*cfg)
	if err != nil {
		return nil, err
	}
	businessRepository := postgres.NewBusinessRepository(db)
	userRepository := postgres.NewUserRepository(db)

	businessSvc := business.NewService(businessRepository)
	userSvc := user.NewService(userRepository)

	businessHandler := handlers.NewBusinessHandler(businessSvc)
	userHandler := handlers.NewUserHandler(userSvc)

	router := httpapi.NewRouter(httpapi.Handlers{
		Business: businessHandler,
		User:     userHandler,
	})

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	return &App{
		cfg:    cfg,
		logger: appLogger,
		server: server,
	}, nil
}
func (a *App) Run() error {
	a.logger.Info("API server starting", logger.Field{Key: "addr", Value: a.server.Addr})
	return a.server.ListenAndServe()
}
