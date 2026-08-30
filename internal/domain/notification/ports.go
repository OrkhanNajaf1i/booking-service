// File: internal/domain/notification/ports.go
package notification

import (
	"context"

	"github.com/google/uuid"
)

// Repository – bildiris, device token ve outbox saxlama qati.
type Repository interface {
	// ---------- IN-APP ----------
	Create(ctx context.Context, item *Notification) error
	ListByUser(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit, offset int) ([]*Notification, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error

	// ---------- DEVICE TOKENS ----------
	SaveDeviceToken(ctx context.Context, token *DeviceToken) error
	DeactivateDeviceToken(ctx context.Context, token string) error
	ListActiveDeviceTokens(ctx context.Context, userID uuid.UUID) ([]*DeviceToken, error)

	// ---------- OUTBOX ----------
	EnqueueOutbox(ctx context.Context, item *OutboxItem) error
	// ClaimOutbox – novbeti pending setirleri atomik sekilde goturur
	// (SELECT ... FOR UPDATE SKIP LOCKED), bir nece worker ucun tehlukesizdir.
	ClaimOutbox(ctx context.Context, limit int) ([]*OutboxItem, error)
	MarkOutboxSent(ctx context.Context, id uuid.UUID) error
	MarkOutboxFailed(ctx context.Context, id uuid.UUID, reason string) error
}

// RealtimePublisher – acıq WebSocket sessiyalarina yayimlayir.
// Adapteri hub-dır; bagli sessiya yoxdursa sessizce kecir.
type RealtimePublisher interface {
	PublishToUser(ctx context.Context, userID uuid.UUID, envelope *Envelope) error
}

// PushSender – FCM (ve ya basqa push provayderi) adapteri.
type PushSender interface {
	Send(ctx context.Context, tokens []string, envelope *Envelope) error
	// Enabled – konfiqurasiya yoxdursa false; servis push-u atlayir.
	Enabled() bool
}

// Service – bildiris use-case-leri.
type Service interface {
	// Dispatch – bir hadiseni saxlayir, WS-e yayimlayir, push-u novbeye salir.
	// Kanal xetasi hadiseni itirmemelidir deye xetalar loglanır, geri qaytarilmir;
	// yalniz baza yazilisi ugursuz olsa error donur.
	Dispatch(ctx context.Context, envelope *Envelope) error
	DispatchMany(ctx context.Context, envelopes []*Envelope) error

	List(ctx context.Context, userID uuid.UUID, unreadOnly bool, limit, offset int) ([]*Notification, error)
	CountUnread(ctx context.Context, userID uuid.UUID) (int, error)
	MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error
	MarkAllRead(ctx context.Context, userID uuid.UUID) error

	RegisterDevice(ctx context.Context, userID uuid.UUID, req *RegisterDeviceRequest) (*DeviceToken, error)
	UnregisterDevice(ctx context.Context, token string) error

	// ProcessOutbox – worker terefinden cagirilir; gonderilen sayini qaytarir.
	ProcessOutbox(ctx context.Context, batchSize int) (int, error)
}
