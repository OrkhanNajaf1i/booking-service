package auth

import "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"

type UserDTO struct {
	ID           string            `json:"id"`
	Email        string            `json:"email"`
	FullName     string            `json:"full_name"`
	Phone        string            `json:"phone"`
	Avatar       string            `json:"avatar"`
	Role         auth.UserRole     `json:"user_role"`
	BusinessID   string            `json:"business_id"`
	BusinessType auth.BusinessType `json:"business_type"`
	Status       string            `json:"status"`
	IsOwner      bool              `json:"is_owner"`
}

type StaffProfileDTO struct {
	ID         string  `json:"id"`
	UserID     string  `json:"user_id"`
	FullName   string  `json:"full_name"`
	Title      string  `json:"title"`
	Department string  `json:"department"`
	Bio        string  `json:"bio"`
	Avatar     string  `json:"avatar"`
	HourlyRate float64 `json:"hourly_rate"`
	Status     string  `json:"status"`
}
type AuthResponseDTO struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	User         UserDTO `json:"user"`
	ExpiresIn    int     `json:"expires_in"`
	TokenType    string  `json:"token_type"`
}

var ErrorMessages = map[string]string{
	"EMAIL_EXISTS":          "Bu email artıq mövcuddur",
	"INVALID_CREDENTIALS":   "Email və ya parol yanlışdır",
	"USER_INACTIVE":         "Akkaunt deaktivdir, zəhmət olmasa admin-ə müraciət edin",
	"INVALID_TOKEN":         "Token yanlış və ya mövcud deyil",
	"TOKEN_EXPIRED":         "Token vaxtı çıxıb, zəhmət olmasa yeni reset tələb edin",
	"TOKEN_ALREADY_USED":    "Bu token artıq istifadə edilib, zəhmət olmasa yeni reset tələb edin",
	"REFRESH_TOKEN_EXPIRED": "Refresh token vaxtı çıxıb, zəhmət olmasa yenidən login edin",
	"REFRESH_TOKEN_REVOKED": "Refresh token ləğv edilib (logout olunub), yenidən login edin",
}
