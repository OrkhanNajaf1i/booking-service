// File: internal/domain/auth/phone_login.go
//
// Telefon nomresi ile giris.
//
// Kod tesdiqlendikden SONRA cagirilir — kodun ozunu `domain/otp`
// yoxlayir. Burada yalniz "bu nomrenin sahibi hansi hesabdir"
// suali hell olunur.
//
// Bu axin YALNIZ musteriler ucundur. Xidmet gostərənlər admin panele
// e-poct/sifre ile girir (bax: CLAUDE.md, rol ayrimi) — telefonla
// giris onlarin hesabina yol acmamalidir.
package auth

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// PhoneAccountFinder – nomre uzre hesab tapmaq/yaratmaq ucun.
type PhoneAccountFinder interface {
	// GetUserByPhoneE164 – tesdiqlenmis nomre uzre hesab.
	GetUserByPhoneE164(ctx context.Context, phoneE164 string) (*User, error)
	// MarkPhoneVerified – movcud hesabin nomresini tesdiqlenmis edir.
	MarkPhoneVerified(ctx context.Context, userID uuid.UUID, phoneE164 string) error
}

// LoginWithPhone – tesdiqlenmis nomre uzre hesabi tapir, yoxdursa
// yaradir, sonra adi giris cavabini qaytarir.
//
// `fullName` yalniz YENI hesab ucun islenir; movcud hesabin adini
// dəyişmir — adam adini profilden idaresi altinda saxlamalidir.
func (s *Service) LoginWithPhone(
	ctx context.Context,
	phoneE164 string,
	fullName string,
) (*AuthResponse, error) {
	if s.phoneAccounts == nil {
		return nil, &RegistrationError{
			Code:    "PHONE_LOGIN_UNAVAILABLE",
			Message: "Telefonla giriş konfiqurasiya olunmayıb",
		}
	}

	existing, err := s.phoneAccounts.GetUserByPhoneE164(ctx, phoneE164)
	if err != nil {
		return nil, fmt.Errorf("failed to look up phone account: %w", err)
	}

	if existing != nil {
		if !existing.IsActive {
			return nil, &RegistrationError{
				Code:    "USER_INACTIVE",
				Message: "Hesab deaktivdir",
			}
		}

		// Xidmet gosterenin hesabi bu yolla acilmamalidir: admin
		// panelin acari e-poct/sifredir.
		if existing.Role != UserTypeCustomer {
			return nil, &RegistrationError{
				Code: "PHONE_LOGIN_NOT_ALLOWED",
				Message: "Bu nömrə xidmət göstərən hesabına aiddir. " +
					"E-poçt və şifrə ilə daxil olun.",
			}
		}

		// Nomre artiq tesdiqlenib, amma qeyd yenilenir: kohne hesab
		// ilk defe telefonla girende `phone_verified` false ola bilir.
		if err := s.phoneAccounts.MarkPhoneVerified(ctx, existing.ID, phoneE164); err != nil {
			return nil, fmt.Errorf("failed to mark phone verified: %w", err)
		}

		return s.generateAuthResponse(ctx, existing)
	}

	// Yeni musteri — sifre yoxdur.
	//
	// `password_hash` bos qalir: bu hesaba yalniz telefonla girilir.
	// Sifre ile giris `VerifyPassword`-da onsuz da ugursuz olur, ona
	// gore bos hash acilan qapi deyil.
	name := strings.TrimSpace(fullName)
	if name == "" {
		// Ad sonra profilden duzeldile bilir; bron ekraninda bos ad
		// qorxulu gorunmesin deye nomrenin son reqemleri qoyulur.
		name = "Müştəri " + lastDigits(phoneE164, 4)
	}

	now := time.Now()
	user := &User{
		ID: uuid.New(),
		// `users.email` NOT NULL ve UNIQUE-dir, telefon hesabinin ise
		// e-poctu yoxdur. Nomreden cixarilan sintetik unvan qoyulur:
		// `.invalid` RFC 2606-ya gore mehz bu mextsedle ayrilib ve heç
		// vaxt marsrutlasdirilmir. Adam sonra real e-poct elave ede biler.
		Email:    syntheticEmail(phoneE164),
		FullName: name,
		Phone:    phoneE164,
		// Sifresiz hesab.
		PasswordHash:  "",
		Role:          UserTypeCustomer,
		BusinessID:    nil,
		IsActive:      true,
		IsOwner:       false,
		EmailVerified: false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create phone user: %w", err)
	}
	if err := s.phoneAccounts.MarkPhoneVerified(ctx, user.ID, phoneE164); err != nil {
		return nil, fmt.Errorf("failed to mark phone verified: %w", err)
	}

	return s.generateAuthResponse(ctx, user)
}

// syntheticEmail – nomreden cixarilan, gonderilmeyen unvan.
func syntheticEmail(phoneE164 string) string {
	return strings.TrimPrefix(phoneE164, "+") + "@phone.invalid"
}

func lastDigits(text string, count int) string {
	if len(text) <= count {
		return text
	}
	return text[len(text)-count:]
}
