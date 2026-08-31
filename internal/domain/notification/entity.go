// File: internal/domain/notification/entity.go
//
// Bildiris domeni uc kanali bir yerde idare edir:
//
//	in-app  – notifications cedveli (zeng ikonu, bildiris merkezi)
//	ws      – acıq WebSocket sessiyalarina aninda gonderis
//	push    – telefon bagli olanda FCM
//
// Butun kanallar eyni Envelope-dan qidalanir.
package notification

import (
	"time"

	"github.com/google/uuid"
)

// Type – bildirisin novu. Client bu deyere gore ekran/davranis secir.
type Type string

const (
	TypeBookingCreated            Type = "booking.created"             // musteri bron etdi -> provider-e
	TypeBookingConfirmed          Type = "booking.confirmed"           // provider tesdiqledi -> musteriye
	TypeBookingRescheduleProposed Type = "booking.reschedule_proposed" // provider basqa vaxt teklif etdi -> musteriye
	TypeBookingRescheduleAccepted Type = "booking.reschedule_accepted" // musteri teklifi qebul etdi -> provider-e
	TypeBookingRescheduleDeclined Type = "booking.reschedule_declined" // musteri teklifi redd etdi -> provider-e
	TypeBookingCancelled          Type = "booking.cancelled"           // her iki terefden ola biler
	TypeBookingCompleted          Type = "booking.completed"           // xidmet bitdi
	TypeBookingNoShow             Type = "booking.no_show"             // musteri gelmedi
	TypeBookingReminder           Type = "booking.reminder"            // worker-in xatirlatmasi
)

// Channel – gonderis kanali.
type Channel string

const (
	ChannelWS    Channel = "ws"
	ChannelPush  Channel = "push"
	ChannelEmail Channel = "email"
)

// Platform – device token-in aid oldugu platforma.
type Platform string

const (
	PlatformIOS     Platform = "ios"
	PlatformAndroid Platform = "android"
	PlatformWeb     Platform = "web"
)

func (p Platform) IsValid() bool {
	return p == PlatformIOS || p == PlatformAndroid || p == PlatformWeb
}

// ============================================================
// NOTIFICATION (in-app)
// ============================================================

type Notification struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	BusinessID *uuid.UUID `db:"business_id" json:"business_id,omitempty"`
	UserID     uuid.UUID  `db:"user_id"     json:"user_id"`
	BookingID  *uuid.UUID `db:"booking_id"  json:"booking_id,omitempty"`
	Type       Type       `db:"type"        json:"type"`
	Title      string     `db:"title"       json:"title"`
	Body       string     `db:"body"        json:"body"`
	Payload    JSONMap    `db:"payload"     json:"payload"`
	IsRead     bool       `db:"is_read"     json:"is_read"`
	ReadAt     *time.Time `db:"read_at"     json:"read_at,omitempty"`
	CreatedAt  time.Time  `db:"created_at"  json:"created_at"`
}

// ============================================================
// ENVELOPE – butun kanallarin ortaq giris formati
// ============================================================

// Envelope – bir hadisenin bir alici ucun tam tesviri.
type Envelope struct {
	Type       Type       `json:"type"`
	UserID     uuid.UUID  `json:"user_id"`
	BusinessID *uuid.UUID `json:"business_id,omitempty"`
	BookingID  *uuid.UUID `json:"booking_id,omitempty"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	Payload    JSONMap    `json:"payload,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`

	// SkipPush – yalniz in-app + WS lazim olanda.
	SkipPush bool `json:"-"`
}

// ============================================================
// DEVICE TOKEN (FCM)
// ============================================================

type DeviceToken struct {
	ID        uuid.UUID `db:"id"         json:"id"`
	UserID    uuid.UUID `db:"user_id"    json:"user_id"`
	Token     string    `db:"token"      json:"token"`
	Platform  Platform  `db:"platform"   json:"platform"`
	IsActive  bool      `db:"is_active"  json:"is_active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// ============================================================
// OUTBOX – worker-in emal etdiyi novbe
// ============================================================

type OutboxStatus string

const (
	OutboxPending OutboxStatus = "pending"
	OutboxSent    OutboxStatus = "sent"
	OutboxFailed  OutboxStatus = "failed"
)

// OutboxItem – push/email gonderisi API sorgusunu bloklamasin deye
// once baza yazilir, sonra worker gonderir.
type OutboxItem struct {
	ID          uuid.UUID    `db:"id"           json:"id"`
	UserID      uuid.UUID    `db:"user_id"      json:"user_id"`
	BookingID   *uuid.UUID   `db:"booking_id"   json:"booking_id,omitempty"`
	Channel     Channel      `db:"channel"      json:"channel"`
	Type        Type         `db:"type"         json:"type"`
	Payload     JSONMap      `db:"payload"      json:"payload"`
	Status      OutboxStatus `db:"status"       json:"status"`
	Attempts    int          `db:"attempts"     json:"attempts"`
	LastError   *string      `db:"last_error"   json:"last_error,omitempty"`
	ScheduledAt time.Time    `db:"scheduled_at" json:"scheduled_at"`
	SentAt      *time.Time   `db:"sent_at"      json:"sent_at,omitempty"`
	CreatedAt   time.Time    `db:"created_at"   json:"created_at"`
}

// ============================================================
// REQUEST DTOS
// ============================================================

type RegisterDeviceRequest struct {
	Token    string   `json:"token"`
	Platform Platform `json:"platform"`
}

// ============================================================
// ERRORS
// ============================================================

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

func NewError(code, message string) *Error { return &Error{Code: code, Message: message} }

var (
	ErrInvalidToken    = NewError("INVALID_TOKEN", "device token bosdur")
	ErrInvalidPlatform = NewError("INVALID_PLATFORM", "platform ios/android/web olmalidir")
	ErrNotFound        = NewError("NOT_FOUND", "bildiris tapilmadi")
)
