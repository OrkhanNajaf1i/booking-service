// File: internal/api/app.go
package api

import (
	"fmt"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	httpapi "github.com/OrkhanNajaf1i/booking-service/internal/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"

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
		return nil, fmt.Errorf("postgres init failed: %w", err)
	}

	businessRepo := postgres.NewBusinessRepository(db)
	businessSvc := business.NewService(businessRepo)

	authRepo := postgres.NewAuthRepository(db)
	passwordHasher := crypto.NewBcryptPasswordHasher()
	tokenManager := crypto.NewJWTSigner(cfg.JWTSecret)

	var emailService auth.EmailService
	switch cfg.EmailProvider {
	case "brevo":
		emailService = email.NewBrevoAdapter(
			cfg.BrevoAPIKey,
			cfg.BrevoSenderEmail,
		)
	default:
		emailService = email.NewSMTPService(
			cfg.SMTPHost,
			cfg.SMTPPort,
			cfg.SMTPUser,
			cfg.SMTPPass,
			cfg.SMTPFrom,
		)
	}

	authSvc := auth.NewAuthService(
		authRepo,
		passwordHasher,
		emailService,
		tokenManager,
	)

	businessH := businessHandler.NewBusinessHandler(businessSvc)
	authH := authHandler.NewAuthHandler(authSvc, appLogger)

	router := httpapi.NewRouter(httpapi.Handlers{
		Business: businessH,
		Auth:     authH,
	}, tokenManager)

	handlerWithCORS := middleware.CORSMiddleware(router)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: handlerWithCORS,
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
