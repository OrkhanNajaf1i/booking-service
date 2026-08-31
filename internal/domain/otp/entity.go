// File: internal/domain/otp/entity.go
//
// Telefon nomresinin tesdiqi.
//
// Musteri nomresini yazir, 6 reqemli kod gelir, kodu yazib daxil olur.
// Sifre yoxdur — sifre unudulur, telefon isə əldədir.
//
// Kanal (SMS / WhatsApp) burada secilmir: domen yalniz "kod gonderilsin"
// deyir, hansi yolla getdiyi `CodeSender` tetbiqinin isidir.
package otp

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Channel – kodun hansi yolla getdiyi.
type Channel string

const (
	// ChannelLog – kod yalniz jurnala yazilir. Inkisaf ve test ucun;
	// pulsuzdur, xarici xidmet teleb etmir.
	ChannelLog Channel = "log"
	// ChannelSMS – operator vasitesile SMS.
	ChannelSMS Channel = "sms"
	// ChannelWhatsApp – WhatsApp Cloud API.
	ChannelWhatsApp Channel = "whatsapp"
)

func (c Channel) IsValid() bool {
	return c == ChannelLog || c == ChannelSMS || c == ChannelWhatsApp
}

// Verification – gonderilmis bir kod.
type Verification struct {
	ID         uuid.UUID  `db:"id"`
	PhoneE164  string     `db:"phone_e164"`
	CodeHash   string     `db:"code_hash"`
	Channel    Channel    `db:"channel"`
	Attempts   int        `db:"attempts"`
	ExpiresAt  time.Time  `db:"expires_at"`
	ConsumedAt *time.Time `db:"consumed_at"`
	CreatedAt  time.Time  `db:"created_at"`
}

// Policy – kodun omru ve sui-istifadeye qarsi hedler.
type Policy struct {
	// CodeTTL – kodun neqeder muddet etibarli oldugu.
	CodeTTL time.Duration
	// MaxAttempts – bir kod uzre nece sehv cehde icaze verilir.
	MaxAttempts int
	// MaxPerHour – bir nomreye saatda nece kod gonderile biler.
	MaxPerHour int
	// ResendAfter – tekrar kod istemek ucun gozleme muddeti.
	ResendAfter time.Duration
}

// DefaultPolicy – gundelik istifade ucun oturusmus deyerler.
//
// 5 deqiqe: SMS gecikə bilir, amma kod uzun yasamamalidir.
// 5 cehd: adam sehv yaza bilir, brute-force ise 10^6 fezada bu limitlə
// praktiki olaraq mumkunsuzdur.
// Saatda 5 kod: pul xerclenmesinin ve nomrenin bombalanmasinin qarsisi.
func DefaultPolicy() Policy {
	return Policy{
		CodeTTL:     5 * time.Minute,
		MaxAttempts: 5,
		MaxPerHour:  5,
		ResendAfter: 60 * time.Second,
	}
}

// ── Xetalar ──────────────────────────────────────────────────

var (
	ErrInvalidPhone = &Error{
		Code:    "PHONE_INVALID",
		Message: "Nömrə düzgün deyil. Nümunə: 050 111 22 33",
	}
	ErrTooManyRequests = &Error{
		Code:    "TOO_MANY_REQUESTS",
		Message: "Çox sayda kod istənilib. Bir qədər sonra yenidən cəhd edin.",
	}
	ErrResendTooSoon = &Error{
		Code:    "RESEND_TOO_SOON",
		Message: "Yeni kod üçün bir az gözləyin.",
	}
	ErrCodeNotFound = &Error{
		Code:    "CODE_NOT_FOUND",
		Message: "Bu nömrə üçün aktiv kod yoxdur. Yenidən kod istəyin.",
	}
	ErrCodeExpired = &Error{
		Code:    "CODE_EXPIRED",
		Message: "Kodun vaxtı bitib. Yenidən kod istəyin.",
	}
	ErrCodeInvalid = &Error{
		Code:    "CODE_INVALID",
		Message: "Kod yanlışdır.",
	}
	ErrTooManyAttempts = &Error{
		Code:    "TOO_MANY_ATTEMPTS",
		Message: "Çox sayda yanlış cəhd. Yenidən kod istəyin.",
	}
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Code + ": " + e.Message }

// AsError – zencirdeki otp xetasini cixarir.
func AsError(err error) (*Error, bool) {
	var target *Error
	if errors.As(err, &target) {
		return target, true
	}
	return nil, false
}
