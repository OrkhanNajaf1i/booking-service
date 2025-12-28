package api

import (
	"fmt"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	httpapi "github.com/OrkhanNajaf1i/booking-service/internal/http"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers"
	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/crypto"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/email"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

type App struct {
	cfg    *config.AppConfig
	logger logger.Logger
	server *http.Server
}

func New(cfg *config.AppConfig, appLogger logger.Logger) (*App, error) {
	appLogger, err := logger.New(cfg)
	if err != nil {
		return nil, err
	}
	db, err := postgres.New(*cfg)
	if err != nil {
		return nil, err
	}
	businessRepository := postgres.NewBusinessRepository(db)
	// userRepository := postgres.NewUserRepository(db)
	authRepo := postgres.NewAuthRepository(db)

	passwordHasher := crypto.NewBcryptPasswordHasher()
	tokenManager := crypto.NewJWTSigner(cfg.JWTSecret)
	emailService := email.NewDummyEmailService()

	businessSvc := business.NewService(businessRepository)

	authSvc := auth.NewAuthService(
		authRepo,
		passwordHasher,
		emailService,
		tokenManager,
		businessSvc,
	)

	businessHandler := handlers.NewBusinessHandler(businessSvc)
	// userHandler := handlers.NewUserHandler(userSvc)
	authHandlers := authHandler.NewAuthHandler(authSvc)
	router := httpapi.NewRouter(httpapi.Handlers{
		Business: businessHandler,
		// User:     userHandler,
		Auth: authHandlers,
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
