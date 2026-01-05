// File: internal/api/app.go
package api

import (
	"fmt"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	httpapi "github.com/OrkhanNajaf1i/booking-service/internal/http"

	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	businessHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"

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
	if appLogger == nil {
		var err error
		appLogger, err = logger.New(cfg)
		if err != nil {
			return nil, err
		}
	}

	db, err := postgres.New(*cfg)
	if err != nil {
		return nil, err
	}

	businessRepository := postgres.NewBusinessRepository(db)
	authRepo := postgres.NewAuthRepository(db)

	passwordHasher := crypto.NewBcryptPasswordHasher()
	tokenManager := crypto.NewJWTSigner(cfg.JWTSecret)
	emailService := email.NewSMTPService(
		cfg.SMTPHost,
		cfg.SMTPPort,
		cfg.SMTPUser,
		cfg.SMTPPass,
		cfg.SMTPFrom,
	)

	userServiceAdapter := NewUserServiceAdapter(authRepo, businessRepository)

	businessSvc := business.NewService(businessRepository, userServiceAdapter)

	authSvc := auth.NewAuthService(
		authRepo,
		passwordHasher,
		emailService,
		tokenManager,
	)

	businessH := businessHandler.NewHandler(businessSvc)
	authH := authHandler.NewAuthHandler(authSvc, appLogger)

	router := httpapi.NewRouter(httpapi.Handlers{
		Business: businessH,
		Auth:     authH,
	}, tokenManager)

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
