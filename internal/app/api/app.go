// File: internal/app/api/app.go
package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/availability"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/customer"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/location"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	serviceDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/service"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/staff"

	httpapi "github.com/OrkhanNajaf1i/booking-service/internal/http"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"

	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	availabilityHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/availability"
	bookingHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/booking"
	businessHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
	customerHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/customer"
	locationHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/location"
	notificationHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/notification"
	publicHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/public"
	serviceHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/service"
	staffHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/staff"

	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/adapters"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/crypto"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/email"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/notify"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/push"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/realtime"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

type App struct {
	cfg    *config.AppConfig
	logger logger.Logger
	server *http.Server
	hub    *realtime.Hub

	// hubCancel – Run() bitende hub goroutine-ini dayandirir.
	hubCancel context.CancelFunc
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

	// ---------- REPOSITORY-LER ----------
	businessRepo := postgres.NewBusinessRepository(db)
	authRepo := postgres.NewAuthRepository(db)
	locationRepo := postgres.NewLocationRepository(db)
	staffRepo := postgres.NewStaffRepository(db)
	serviceRepo := postgres.NewServiceRepository(db)
	customerRepo := postgres.NewCustomerRepository(db, appLogger)
	availabilityRepo := postgres.NewAvailabilityRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)
	notificationRepo := postgres.NewNotificationRepository(db)
	participantsRepo := postgres.NewParticipantsRepository(db)

	// ---------- KRIPTO / EMAIL ----------
	passwordHasher := crypto.NewBcryptPasswordHasher()
	tokenManager := crypto.NewJWTSigner(cfg.JWTSecret)

	var emailService auth.EmailService
	switch cfg.EmailProvider {
	case "brevo":
		emailService = email.NewBrevoAdapter(cfg.BrevoAPIKey, cfg.BrevoSenderEmail)
	default:
		emailService = email.NewSMTPService(
			cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom,
		)
	}

	// ---------- REALTIME + PUSH ----------
	hub := realtime.NewHub(appLogger)
	pushSender := push.NewFCMAdapter(cfg.FCMCredentialsFile, cfg.FCMCredentialsJSON, appLogger)

	notificationSvc := notification.NewService(notificationRepo, hub, pushSender, appLogger)
	bookingPublisher := notify.NewBookingPublisher(notificationSvc, appLogger, cfg.DefaultTimezone)

	// ---------- DOMAIN SERVIS-LERI ----------
	businessSvc := business.NewService(businessRepo)
	authSvc := auth.NewAuthService(authRepo, passwordHasher, emailService, tokenManager)
	locationSvc := location.NewService(locationRepo)
	staffSvc := staff.NewService(staffRepo, authRepo)
	serviceSvc := serviceDomain.NewServiceUseCase(serviceRepo)
	customerSvc := customer.NewService(customerRepo, authRepo)

	// Availability xidmet mueddetini kataloqdan oxuyur.
	availabilitySvc := availability.NewService(
		availabilityRepo,
		adapters.NewServiceDuration(serviceRepo),
		appLogger,
	)

	bookingSvc := booking.NewService(
		bookingRepo,
		availabilitySvc,
		participantsRepo,
		bookingPublisher,
		customerRepo,
		appLogger,
	)

	// ---------- HANDLER-LER ----------
	handlers := httpapi.Handlers{
		Business:     businessHandler.NewBusinessHandler(businessSvc, appLogger),
		Auth:         authHandler.NewAuthHandler(authSvc, appLogger),
		Location:     locationHandler.NewHandler(locationSvc),
		Staff:        staffHandler.NewHandler(staffSvc),
		Service:      serviceHandler.NewHandler(serviceSvc),
		Customer:     *customerHandler.NewHandler(customerSvc, appLogger),
		Availability: availabilityHandler.NewHandler(availabilitySvc, appLogger),
		Booking:      bookingHandler.NewHandler(bookingSvc, staffRepo, appLogger),
		Notification: notificationHandler.NewHandler(notificationSvc, appLogger),
		Public: publicHandler.NewHandler(
			businessSvc, staffSvc, serviceSvc, availabilitySvc, appLogger,
		),
		Realtime: realtime.NewHandler(hub, tokenManager, appLogger, splitOrigins(cfg.WSAllowedOrigins)),
	}

	router := httpapi.NewRouter(handlers, tokenManager)
	handlerWithCORS := middleware.CORSMiddleware(router)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: handlerWithCORS,
		// WebSocket uzun omurlu oldugu ucun write timeout qoyulmur.
		ReadHeaderTimeout: 15 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return &App{
		cfg:    cfg,
		logger: appLogger,
		server: server,
		hub:    hub,
	}, nil
}

func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	a.hubCancel = cancel
	go a.hub.Run(ctx)

	a.logger.Info("API server starting", logger.Field{Key: "addr", Value: a.server.Addr})
	return a.server.ListenAndServe()
}

// Shutdown – graceful dayandirma.
func (a *App) Shutdown(ctx context.Context) error {
	if a.hubCancel != nil {
		a.hubCancel()
	}
	return a.server.Shutdown(ctx)
}

// splitOrigins – "https://a.com,https://b.com" -> []string.
// Bos olarsa nil qaytarir (butun origin-lere icaze).
func splitOrigins(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	parts := strings.Split(trimmed, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if cleaned := strings.TrimSpace(part); cleaned != "" {
			origins = append(origins, cleaned)
		}
	}
	return origins
}
