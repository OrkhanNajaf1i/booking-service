package auth

import (
	"context"

	"github.com/google/uuid"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	SaveRefreshToken(ctx context.Context, token *RefreshToken) error
	GetRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error
	SavePasswordReset(ctx context.Context, reset *PasswordReset) error
	GetPasswordReset(ctx context.Context, token string) (*PasswordReset, error)
	UpdatePassword(ctx context.Context, userID string, hashedPassword string) error
	EmailExists(ctx context.Context, email string) (bool, error)
	UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error
}

type PasswordHasher interface {
	HashPassword(password string) (string, error)
	VerifyPassword(hash, password string) error
}
type EmailService interface {
	SendPasswordResetEmail(email string, resetURL string) error
}
type TokenManager interface {
	GenerateAccessToken(claims *JWTClaims) (string, error)
	GenerateRefreshToken() (string, error)
	ValidateAccessToken(token string) (*JWTClaims, error)
}
