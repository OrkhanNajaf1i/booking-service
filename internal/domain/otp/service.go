// File: internal/domain/otp/service.go
package otp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo   Repository
	sender CodeSender
	policy Policy
	clock  Clock
}

func NewService(repo Repository, sender CodeSender, policy Policy) *Service {
	return &Service{
		repo:   repo,
		sender: sender,
		policy: policy,
		clock:  systemClock{},
	}
}

// WithClock – testde vaxti sabitlemek ucun.
func (s *Service) WithClock(clock Clock) *Service {
	s.clock = clock
	return s
}

// RequestResult – kod gonderildikden sonra tetbiqe qaytarilan melumat.
type RequestResult struct {
	// PhoneE164 – normallasdirilmis nomre; tetbiq tesdiq addiminda
	// eynisini geri gonderir.
	PhoneE164 string
	// MaskedPhone – ekranda gosterilir: "+994 50 ** ** 33".
	MaskedPhone string
	// ExpiresIn – kodun necə saniyə etibarli oldugu; geri sayim ucun.
	ExpiresIn int
	// ResendAfter – tekrar kod istemek ucun neçə saniyə gozlemeli.
	ResendAfter int
	// Channel – kodun hansi yolla getdiyi.
	Channel Channel
	// DebugCode – YALNIZ jurnal kanalinda dolur (inkisaf rejimi).
	// Real kanalda hemise bosdur.
	DebugCode string
}

// RequestCode – nomreye yeni kod gonderir.
func (s *Service) RequestCode(ctx context.Context, rawPhone string) (*RequestResult, error) {
	phone, err := NormalizePhone(rawPhone)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now()

	// Saatliq limit — nomrenin bombalanmasinin ve pul xerclenmesinin
	// qarsisini alir.
	sent, err := s.repo.CountSince(ctx, phone, now.Add(-time.Hour))
	if err != nil {
		return nil, fmt.Errorf("otp: failed to count recent codes: %w", err)
	}
	if sent >= s.policy.MaxPerHour {
		return nil, ErrTooManyRequests
	}

	// Ard-arda basmagin qarsisi: son kod cox tezedirse gozlemek lazimdir.
	previous, err := s.repo.LatestActive(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("otp: failed to read last code: %w", err)
	}
	if previous != nil && now.Sub(previous.CreatedAt) < s.policy.ResendAfter {
		return nil, ErrResendTooSoon
	}

	code, err := generateCode()
	if err != nil {
		return nil, fmt.Errorf("otp: failed to generate code: %w", err)
	}

	// Kod ƏVVƏLCƏ gonderilir, sonra yazilir: gonderme alinmasa
	// istifadeci bosuna limit itirmemelidir.
	if err := s.sender.Send(ctx, phone, code); err != nil {
		return nil, fmt.Errorf("otp: failed to send code: %w", err)
	}

	verification := &Verification{
		ID:        uuid.New(),
		PhoneE164: phone,
		CodeHash:  hashCode(code),
		Channel:   s.sender.Channel(),
		ExpiresAt: now.Add(s.policy.CodeTTL),
		CreatedAt: now,
	}
	if err := s.repo.Create(ctx, verification); err != nil {
		return nil, fmt.Errorf("otp: failed to save code: %w", err)
	}

	result := &RequestResult{
		PhoneE164:   phone,
		MaskedPhone: MaskPhone(phone),
		ExpiresIn:   int(s.policy.CodeTTL.Seconds()),
		ResendAfter: int(s.policy.ResendAfter.Seconds()),
		Channel:     verification.Channel,
	}

	// Jurnal kanalinda kod cavabda da qaytarilir ki, inkisaf zamani
	// jurnala baxmaq lazim gelmesin. Real kanalda bu bos qalir.
	if verification.Channel == ChannelLog {
		result.DebugCode = code
	}

	return result, nil
}

// VerifyCode – kodu yoxlayir. Ugurlu olsa normallasdirilmis nomre
// qaytarilir; cagiran teref hemin nomre uzre hesabi tapir/yaradir.
func (s *Service) VerifyCode(ctx context.Context, rawPhone, code string) (string, error) {
	phone, err := NormalizePhone(rawPhone)
	if err != nil {
		return "", err
	}

	verification, err := s.repo.LatestActive(ctx, phone)
	if err != nil {
		return "", fmt.Errorf("otp: failed to read code: %w", err)
	}
	if verification == nil {
		return "", ErrCodeNotFound
	}

	now := s.clock.Now()
	if now.After(verification.ExpiresAt) {
		return "", ErrCodeExpired
	}
	if verification.Attempts >= s.policy.MaxAttempts {
		return "", ErrTooManyAttempts
	}

	// Sabit vaxtli muqayise: cavab muddetinden kod haqqinda melumat
	// sizmamalidir.
	expected := hashCode(code)
	if subtle.ConstantTimeCompare([]byte(expected), []byte(verification.CodeHash)) != 1 {
		if err := s.repo.IncrementAttempts(ctx, verification.ID); err != nil {
			return "", fmt.Errorf("otp: failed to record attempt: %w", err)
		}
		return "", ErrCodeInvalid
	}

	if err := s.repo.MarkConsumed(ctx, verification.ID); err != nil {
		return "", fmt.Errorf("otp: failed to consume code: %w", err)
	}

	return phone, nil
}

// generateCode – 6 reqemli kod.
//
// `crypto/rand` isledilir: `math/rand` proqnozlasdirila bilir və kod
// tehlukesizlik elementidir.
func generateCode() (string, error) {
	limit := big.NewInt(1_000_000)

	value, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", err
	}

	// Basdaki sifirlar saxlanilir: "004521" de etibarli koddur.
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func hashCode(code string) string {
	sum := sha256.Sum256([]byte(code))
	return hex.EncodeToString(sum[:])
}
