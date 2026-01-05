// File: internal/domain/business/ports.go
package business

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID
	Email      string
	FullName   string
	Phone      string
	IsOwner    bool
	BusinessID uuid.UUID
}
type Repository interface {
	CreateBusiness(ctx context.Context, business *Business) error
	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
	UpdateOwner(ctx context.Context, businessID uuid.UUID, ownerID uuid.UUID) error
	GetBusinessByOwnerID(ctx context.Context, ownerID uuid.UUID) (*Business, error)
	CreateLocation(ctx context.Context, location *Location) error
	GetLocationsByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Location, error)
	GetLocationByID(ctx context.Context, id uuid.UUID, businessID uuid.UUID) (*Location, error)
	CreateStaffProfile(ctx context.Context, profile *StaffProfile) error
	GetStaffProfileByUserID(ctx context.Context, userID uuid.UUID, businessID uuid.UUID) (*StaffProfile, error)
	UpdateStaffProfile(ctx context.Context, profile *StaffProfile) error
	GetStaffByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*StaffProfile, error)
	CreateInvite(ctx context.Context, invite *BusinessInvite) error
	GetInviteByToken(ctx context.Context, token string, businessID uuid.UUID) (*BusinessInvite, error)
	UseInvite(ctx context.Context, tokenID uuid.UUID) error
	GetInvitesByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*BusinessInvite, error)
	UpdateUserBusinessID(ctx context.Context, userID uuid.UUID, businessID uuid.UUID) error
}

type UserService interface {
	GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error)
	UpdateUserBusinessID(ctx context.Context, userID uuid.UUID, businessID uuid.UUID, isOwner bool) error
}

type BusinessService interface {
	CreateSoloBusiness(ctx context.Context, ownerID uuid.UUID, req *CreateBusinessRequest) (uuid.UUID, error)
	CreateMultiBusiness(ctx context.Context, ownerID uuid.UUID, req *CreateBusinessRequest) (uuid.UUID, error)
	InviteStaff(ctx context.Context, businessID uuid.UUID, req *InviteStaffRequest) (string, error)
	JoinWithInvite(ctx context.Context, token string, password string) error
	CreateDefaultLocation(ctx context.Context, businessID uuid.UUID) (uuid.UUID, error)
}
