// File: internal/http/handlers/auth/dto.go
package auth

import (
	"strings"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/google/uuid"
)

type RegisterHTTPRequest struct {
	Email    string `json:"email" example:"orkhan@example.com"`
	Password string `json:"password" example:"StrongPass123!"`
	FullName string `json:"full_name" example:"Orkhan Najafli"`
	Phone    string `json:"phone" example:"+994501234567"`
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
	Avatar        *string       `json:"avatar"`
	Role          auth.UserRole `json:"role"`
	BusinessID    *uuid.UUID    `json:"business_id"`
	IsActive      bool          `json:"is_active"`
	IsOwner       bool          `json:"is_owner"`
	EmailVerified bool          `json:"email_verified"`
	CreatedAt     time.Time     `json:"created_at"`
}

type AuthResponseDTO struct {
	AccessToken  string          `json:"access_token" example:"eyJhbGci..."`
	RefreshToken string          `json:"refresh_token" example:"def456..."`
	User         UserResponseDTO `json:"user"`
	ExpiresIn    int             `json:"expires_in" example:"900"`
	TokenType    string          `json:"token_type" example:"Bearer"`
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

func ToDomainRegister(httpReq *RegisterHTTPRequest) *auth.RegisterRequest {
	return &auth.RegisterRequest{
		Email:    strings.TrimSpace(httpReq.Email),
		Password: httpReq.Password,
		FullName: strings.TrimSpace(httpReq.FullName),
		Phone:    strings.TrimSpace(httpReq.Phone),
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

// Tam error kod xəritəsi (service + validation-la uyğun)
var ErrorMessages = map[string]string{
	// Register / login
	"EMAIL_EXISTS":         "Bu email artıq mövcuddur",
	"EMAIL_REQUIRED":       "Email tələb olunur",
	"EMAIL_TOO_LONG":       "Email çox uzundur",
	"INVALID_EMAIL_FORMAT": "Email formatı yanlışdır",

	"PASSWORD_REQUIRED":  "Parol tələb olunur",
	"PASSWORD_TOO_SHORT": "Parol minimum 8 simvol olmalıdır",
	"PASSWORD_TOO_LONG":  "Parol çox uzundur",
	"PASSWORD_WEAK":      "Parol kifayət qədər güclü deyil",

	"FULLNAME_REQUIRED":  "Tam ad tələb olunur",
	"FULLNAME_TOO_SHORT": "Tam ad çox qısadır",
	"FULLNAME_TOO_LONG":  "Tam ad çox uzundur",

	"PHONE_REQUIRED": "Telefon nömrəsi tələb olunur",

	"INVALID_CREDENTIALS": "Email və ya parol yanlışdır",
	"USER_INACTIVE":       "Akkaunt deaktivdir",

	"INVALID_REFRESH_TOKEN": "Refresh token yanlışdır",
	"REFRESH_TOKEN_EXPIRED": "Refresh token vaxtı çıxıb",
	"REFRESH_TOKEN_REVOKED": "Refresh token ləğv edilib",
	"USER_NOT_FOUND":        "İstifadəçi tapılmadı",

	"INVALID_TOKEN":           "Token yanlış və ya mövcud deyil",
	"TOKEN_EXPIRED":           "Token vaxtı çıxıb (24 saat)",
	"TOKEN_ALREADY_USED":      "Token artıq istifadə edilib",
	"RESET_TOKEN_SAVE_FAILED": "Reset token yadda saxlanmadı",
	"PASSWORD_HASH_FAILED":    "Parol işlənərkən xəta baş verdi",
	"PASSWORD_UPDATE_FAILED":  "Parolu yeniləmək alınmadı",

	"VALIDATION_ERROR": "Giriş məlumatları yanlışdır",
	"INTERNAL_ERROR":   "Daxili server xətası",
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
