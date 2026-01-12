package auth

import (
	"time"

	"github.com/google/uuid"
)

type BusinessType string

const (
	BusinessTypeSolo  BusinessType = "solo_practitioner"
	BusinessTypeMulti BusinessType = "multi_staff_business"
)

type UserRole string

const (
	UserTypeCustomer         UserRole = "customer"
	UserTypeOwner            UserRole = "provider_owner"
	UserTypeStaff            UserRole = "staff"
	UserTypeSoloPractitioner UserRole = "solo_practitioner"
)

type StaffRole string

const (
	StaffRoleManager       StaffRole = "manager"
	StaffRoleStaff         StaffRole = "staff"
	StaffRoleAdministrator StaffRole = "admin"
)

type User struct {
	ID            uuid.UUID  `db:"id" json:"id"`
	Email         string     `db:"email" json:"email"`
	FullName      string     `db:"full_name" json:"full_name"`
	Phone         string     `db:"phone" json:"phone"`
	PasswordHash  string     `db:"password_hash" json:"-"`
	Role          UserRole   `db:"role" json:"role"`
	BusinessID    *uuid.UUID `db:"business_id" json:"business_id"`
	Avatar        *string    `db:"avatar" json:"avatar"`
	IsActive      bool       `db:"is_active" json:"is_active"`
	IsOwner       bool       `db:"is_owner" json:"is_owner"`
	EmailVerified bool       `db:"email_verified" json:"email_verified"`
	CreatedAt     time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at" json:"updated_at"`
}
type RefreshToken struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Token     string    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	Revoked   bool      `db:"revoked" json:"revoked"`
}
type PasswordReset struct {
	ID        uuid.UUID `db:"id" json:"id"`
	Email     string    `db:"email" json:"email"`
	Token     string    `db:"token" json:"token"`
	ExpiresAt time.Time `db:"expires_at" json:"expires_at"`
	Used      bool      `db:"used" json:"used"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}
type StaffProfile struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	BusinessID *uuid.UUID `db:"business_id"`
	LocationID *uuid.UUID `db:"location_id"`
	Role       StaffRole  `db:"role"`
	Title      string     `db:"title"`
	Department string     `db:"department"`
	Bio        string     `db:"bio"`
	HourlyRate float64    `db:"hourly_rate"`
	Status     string     `db:"status"`
	JoinedAt   time.Time  `db:"joined_at"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}
type JWTClaims struct {
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	Email      string     `db:"email" json:"email"`
	Role       UserRole   `db:"role" json:"role"`
	BusinessID *uuid.UUID `db:"business_id" json:"business_id"`
	IsOwner    bool       `db:"is_owner" json:"is_owner"`
	ExpiresAt  int64      `db:"expires_at" json:"expires_at"`
}
type RegisterRequest struct {
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
	FullName string `db:"full_name" json:"full_name"`
	Phone    string `db:"phone" json:"phone"`
}

type LoginRequest struct {
	Email    string `db:"email" json:"email"`
	Password string `db:"password" json:"password"`
}
type RefreshTokenRequest struct {
	RefreshToken string `db:"refresh_token" json:"refresh_token"`
}
type ForgotPasswordRequest struct {
	Email string `db:"email" json:"email"`
}
type ResetPasswordRequest struct {
	Token    string `db:"token" json:"token"`
	Password string `db:"password" json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         *User  `json:"user"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}
type RegistrationError struct {
	Code    string `db:"code" json:"code"`
	Message string `db:"message" json:"message"`
}

func (e *RegistrationError) Error() string {
	return e.Message
}
