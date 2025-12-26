package auth

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
}

type BusinessRepository interface {
	CreateBusiness(ctx context.Context, business *Business) error
	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
}

type LocationRepository interface {
	CreateLocation(ctx context.Context, location *Location) error
}

type StaffRepository interface {
	CreateStaff(ctx context.Context, staff *Staff) error
}
type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hash, password string) bool
}

type IDGenerator interface {
	Generate() uuid.UUID
}

type AuthService interface {
	Register(ctx context.Context, email, password, businessName, locationName string, flowType BusinessType) (*User, error)
	GetUserRole(ctx context.Context, UserID uuid.UUID) (string, error)
}
