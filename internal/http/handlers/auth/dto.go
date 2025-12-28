// package auth

// import "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"

// type UserDTO struct {
// 	ID           string            `json:"id"`
// 	Email        string            `json:"email"`
// 	FullName     string            `json:"full_name"`
// 	Phone        string            `json:"phone"`
// 	Avatar       string            `json:"avatar"`
// 	Role         auth.UserRole     `json:"user_role"`
// 	BusinessID   string            `json:"business_id"`
// 	BusinessType auth.BusinessType `json:"business_type"`
// 	Status       string            `json:"status"`
// 	IsOwner      bool              `json:"is_owner"`
// }

// type StaffProfileDTO struct {
// 	ID         string  `json:"id"`
// 	UserID     string  `json:"user_id"`
// 	FullName   string  `json:"full_name"`
// 	Title      string  `json:"title"`
// 	Department string  `json:"department"`
// 	Bio        string  `json:"bio"`
// 	Avatar     string  `json:"avatar"`
// 	HourlyRate float64 `json:"hourly_rate"`
// 	Status     string  `json:"status"`
// }
// type AuthResponseDTO struct {
// 	AccessToken  string  `json:"access_token"`
// 	RefreshToken string  `json:"refresh_token"`
// 	User         UserDTO `json:"user"`
// 	ExpiresIn    int     `json:"expires_in"`
// 	TokenType    string  `json:"token_type"`
// }

//	var ErrorMessages = map[string]string{
//		"EMAIL_EXISTS":          "Bu email artıq mövcuddur",
//		"INVALID_CREDENTIALS":   "Email və ya parol yanlışdır",
//		"USER_INACTIVE":         "Akkaunt deaktivdir, zəhmət olmasa admin-ə müraciət edin",
//		"INVALID_TOKEN":         "Token yanlış və ya mövcud deyil",
//		"TOKEN_EXPIRED":         "Token vaxtı çıxıb, zəhmət olmasa yeni reset tələb edin",
//		"TOKEN_ALREADY_USED":    "Bu token artıq istifadə edilib, zəhmət olmasa yeni reset tələb edin",
//		"REFRESH_TOKEN_EXPIRED": "Refresh token vaxtı çıxıb, zəhmət olmasa yenidən login edin",
//		"REFRESH_TOKEN_REVOKED": "Refresh token ləğv edilib (logout olunub), yenidən login edin",
//	}
//
// File: internal/http/handlers/auth/dto.go
package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/google/uuid"
)

type RegisterHTTPRequest struct {
	Email           string        `json:"email"`
	Password        string        `json:"password"`
	FullName        string        `json:"full_name"`
	Phone           string        `json:"phone"`
	Role            auth.UserRole `json:"role"`
	BusinessType    string        `json:"business_type,omitempty"`
	BusinessName    string        `json:"business_name,omitempty"`
	Industry        string        `json:"industry,omitempty"`
	ServiceCategory string        `json:"service_category,omitempty"`
}

type LoginHTTPRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenHTTPRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type ForgotPasswordHTTPRequest struct {
	Email string `json:"email"`
}

type ResetPasswordHTTPRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type UserResponseDTO struct {
	ID            uuid.UUID     `json:"id"`
	Email         string        `json:"email"`
	FullName      string        `json:"full_name"`
	Phone         string        `json:"phone"`
	Avatar        string        `json:"avatar"`
	Role          auth.UserRole `json:"role"`
	BusinessID    uuid.UUID     `json:"business_id"`
	IsActive      bool          `json:"is_active"`
	IsOwner       bool          `json:"is_owner"`
	EmailVerified bool          `json:"email_verified"`
	CreatedAt     time.Time     `json:"created_at"` // time.Time → ISO string
}

type AuthResponseDTO struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	User         UserResponseDTO `json:"user"`
	ExpiresIn    int             `json:"expires_in"`
	TokenType    string          `json:"token_type"`
}

type SuccessResponseDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponseDTO struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// dto.go - ToDomainRegister düzəlişi
func ToDomainRegister(httpReq *RegisterHTTPRequest) *auth.RegisterRequest {
	// Validation SİLİNDİ (service-də var)
	return &auth.RegisterRequest{
		Email:           strings.TrimSpace(httpReq.Email),
		Password:        httpReq.Password,
		FullName:        strings.TrimSpace(httpReq.FullName),
		Phone:           strings.TrimSpace(httpReq.Phone),
		Role:            httpReq.Role,
		BusinessType:    auth.BusinessType(httpReq.BusinessType),
		BusinessName:    strings.TrimSpace(httpReq.BusinessName),
		Industry:        strings.TrimSpace(httpReq.Industry),
		ServiceCategory: strings.TrimSpace(httpReq.ServiceCategory),
	}
}

func FromDomainUser(user *auth.User) *UserResponseDTO {
	return &UserResponseDTO{
		ID:            user.ID,
		Email:         user.Email,
		FullName:      user.FullName,
		Phone:         user.Phone,
		Avatar:        user.Avatar,
		Role:          user.Role,
		BusinessID:    user.BusinessID,
		IsActive:      user.IsActive,
		IsOwner:       user.IsOwner,
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt,
	}
}

func FromDomainAuthResponse(resp *auth.AuthResponse) *AuthResponseDTO {
	return &AuthResponseDTO{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		User:         *FromDomainUser(resp.User),
		ExpiresIn:    resp.ExpiresIn,
		TokenType:    resp.TokenType,
	}
}

var ErrorMessages = map[string]string{
	"EMAIL_EXISTS":          "Bu email artıq mövcuddur",
	"INVALID_CREDENTIALS":   "Email və ya parol yanlışdır",
	"USER_INACTIVE":         "Akkaunt deaktivdir",
	"INVALID_TOKEN":         "Token yanlış və ya mövcud deyil",
	"TOKEN_EXPIRED":         "Token vaxtı çıxıb (24 saat)",
	"TOKEN_ALREADY_USED":    "Token artıq istifadə edilib",
	"REFRESH_TOKEN_EXPIRED": "Refresh token vaxtı çıxıb",
	"REFRESH_TOKEN_REVOKED": "Refresh token ləğv edilib",
	"VALIDATION_ERROR":      "Giriş məlumatları yanlışdır",
}

func GetErrorResponse(code string) *ErrorResponseDTO {
	msg, ok := ErrorMessages[code]
	if !ok {
		msg = "Internal server error"
	}
	return &ErrorResponseDTO{
		Success: false,
		Code:    code,
		Message: msg,
	}
}
