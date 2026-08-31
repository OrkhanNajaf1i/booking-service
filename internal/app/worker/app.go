// File: internal/app/worker/app.go
//
// Worker iki isi gorur:
//
//  1. notification_outbox – novbedeki push bildirislerini gonderir
//  2. reminder          – yaxinlasan randevular ucun xatirlatma yaradir
//
// Her ikisi de eyni poll dovresinde islenir. Worker olmasa da sistem
// isleyir: in-app bildiris ve WebSocket API prosesinde gonderilir,
// worker yalniz push ve xatirlatmani elave edir.
package worker

import (
	"context"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/config"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/push"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/jmoiron/sqlx"
)

type App struct {
	config        *config.AppConfig
	logger        logger.Logger
	db            *sqlx.DB
	notifications notification.Service
	reminders     *ReminderJob
	expiry        *PendingExpiryJob
	pollInterval  time.Duration
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

	notificationRepo := postgres.NewNotificationRepository(db)
	bookingRepo := postgres.NewBookingRepository(db)
	participantsRepo := postgres.NewParticipantsRepository(db)
	pushSender := push.NewFCMAdapter(cfg.FCMCredentialsFile, cfg.FCMCredentialsJSON, appLogger)

	// Worker-de realtime publisher yoxdur: WebSocket sessiyalari API
	// prosesinde yasayir. Push kanali burada islenir.
	notificationSvc := notification.NewService(notificationRepo, nil, pushSender, appLogger)

	pollInterval := cfg.WorkerPollInterval
	if pollInterval <= 0 {
		pollInterval = 15 * time.Second
	}

	return &App{
		config:        cfg,
		logger:        appLogger,
		db:            db,
		notifications: notificationSvc,
		reminders:     NewReminderJob(db, notificationSvc, appLogger, cfg.DefaultTimezone),
		expiry: NewPendingExpiryJob(
			bookingRepo, participantsRepo, notificationSvc, appLogger, cfg.DefaultTimezone,
		),
		pollInterval: pollInterval,
	}, nil
}

// Run – ctx legv edilene qeder dovre vurur.
func (a *App) Run(ctx context.Context) error {
	a.logger.Info("Worker basladi",
		logger.Field{Key: "poll_interval", Value: a.pollInterval.String()},
	)

	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

	// Ilk dovreni gozlemeden isledirik.
	a.tick(ctx)

	for {
		select {
		case <-ctx.Done():
			a.logger.Info("Worker dayanir")
			return ctx.Err()
		case <-ticker.C:
			a.tick(ctx)
		}
	}
}

// tick – bir dovre. Xetalar loglanır, dovre dayanmir.
func (a *App) tick(ctx context.Context) {
	sent, err := a.notifications.ProcessOutbox(ctx, 50)
	if err != nil {
		a.logger.Error("Outbox emali ugursuz",
			logger.Field{Key: "error", Value: err.Error()},
		)
	} else if sent > 0 {
		a.logger.Info("Push gonderildi", logger.Field{Key: "count", Value: sent})
	}

	created, err := a.reminders.Run(ctx)
	if err != nil {
		a.logger.Error("Xatirlatma emali ugursuz",
			logger.Field{Key: "error", Value: err.Error()},
		)
	} else if created > 0 {
		a.logger.Info("Xatirlatma yaradildi", logger.Field{Key: "count", Value: created})
	}

	// Cavabsiz qalmis bronlar slotu bloklamamalidir.
	expired, err := a.expiry.Run(ctx)
	if err != nil {
		a.logger.Error("Cavabsiz bron emali ugursuz",
			logger.Field{Key: "error", Value: err.Error()},
		)
	} else if expired > 0 {
		a.logger.Info("Cavabsiz bron legv edildi", logger.Field{Key: "count", Value: expired})
	}
}

// Close – baglantilari baglayir.
func (a *App) Close() error {
	if a.db != nil {
		return a.db.Close()
	}
	return nil
}
