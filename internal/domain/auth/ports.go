package auth

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
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
	UpdateUserStatus(ctx context.Context, userID string, status string) error
	CreateStaffProfile(ctx context.Context, profile *StaffProfile) error
	GetStaffProfile(ctx context.Context, userID string) (*StaffProfile, error)
	UpdateStaffProfile(ctx context.Context, staffID string, profile *StaffProfile) error
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

type BusinessService interface {
	CreateSoloPractitionerBusiness(ctx context.Context, ownerID string, name string, industry string) (businessID string, err error)
	CreateMultiStaffBusiness(ctx context.Context, ownerID string, name string, industry string) (businessID string, err error)
	CreateDefaultLocation(ctx context.Context, businessID uuid.UUID) (locationID string, err error)
}

// type IDGenerator interface {
// 	Generate() uuid.UUID
// }

// type AuthService interface {
// 	Register(ctx context.Context, email, password, businessName, locationName string, flowType BusinessType) (*User, error)
// 	GetUserRole(ctx context.Context, UserID uuid.UUID) (string, error)
// }
// type BusinessRepository interface {
// 	CreateBusiness(ctx context.Context, business *Business) error
// 	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
// }

// type LocationRepository interface {
// 	CreateLocation(ctx context.Context, location *Location) error
// }

//	type StaffRepository interface {
//		CreateStaff(ctx context.Context, staff *Staff) error
//	}
