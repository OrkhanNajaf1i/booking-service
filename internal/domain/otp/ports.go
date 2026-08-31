// File: internal/domain/otp/ports.go
package otp

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, verification *Verification) error
	// LatestActive – hemin nomre ucun hele istifade edilmemis son kod.
	LatestActive(ctx context.Context, phoneE164 string) (*Verification, error)
	// CountSince – saatliq limiti saymaq ucun.
	CountSince(ctx context.Context, phoneE164 string, since time.Time) (int, error)
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
	MarkConsumed(ctx context.Context, id uuid.UUID) error
}

// CodeSender – kodu istifadeciye catdiran kanal.
//
// Domen hansi xidmetin islendiyini bilmir: SMS operatoru, WhatsApp ve
// ya sadece jurnal. Provayder deyisende yalniz bu interfeysin tetbiqi
// evez olunur.
type CodeSender interface {
	// Send – kodu gonderir. Xeta qaytarsa kod yaradilmis sayilmir.
	Send(ctx context.Context, phoneE164, code string) error
	// Channel – hansi kanal oldugu; qeyd ucun saxlanilir.
	Channel() Channel
}

// Clock – testde vaxti idare etmek ucun.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }
