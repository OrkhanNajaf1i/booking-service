// File: internal/domain/staff/ports.go
package staff

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateStaffProfile(ctx context.Context, profile *StaffProfile) error
	GetStaffByID(ctx context.Context, id, businessID uuid.UUID) (*StaffProfile, error)
	GetStaffByUserID(ctx context.Context, userID, businessID uuid.UUID) (*StaffProfile, error)
	ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]*StaffWithUser, error)
	UpdateStaffProfile(ctx context.Context, profile *StaffProfile) error
	DeactivateStaff(ctx context.Context, id, businessID uuid.UUID) error
	CreateInvite(ctx context.Context, invite *BusinessInvite) error
	GetInviteByToken(ctx context.Context, token string) (*BusinessInvite, error)
	MarkInviteAsUsed(ctx context.Context, inviteID uuid.UUID) error
	ListInvitesByBusiness(ctx context.Context, businessID uuid.UUID) ([]*BusinessInvite, error)
}

type UserService interface {
	UpdateUserBusinessID(ctx context.Context, userID, businessID uuid.UUID, isOwner bool) error
}

// BusinessOwnerLookup – biznesin sahibini tapir.
//
// Sahib isci siyahisindan silinmemelidir; bunu yoxlamaq ucun staff
// domeninin biznesin sahibini bilmesi lazimdir. Butun business
// domenini asili etmemek ucun yalniz bu kicik port goturulur.
type BusinessOwnerLookup interface {
	OwnerUserID(ctx context.Context, businessID uuid.UUID) (uuid.UUID, error)
}

type Service interface {
	CreateStaffProfile(ctx context.Context, businessID uuid.UUID, req *CreateStaffRequest) (*StaffProfile, error)
	// IsOwner – hemin isci biznesin sahibidirmi.
	IsOwner(ctx context.Context, staffID, businessID uuid.UUID) (bool, error)
	GetStaff(ctx context.Context, staffID, businessID uuid.UUID) (*StaffProfile, error)
	ListStaff(ctx context.Context, businessID uuid.UUID) ([]*StaffWithUser, error)
	UpdateStaff(ctx context.Context, staffID, businessID uuid.UUID, req *UpdateStaffRequest) error
	DeactivateStaff(ctx context.Context, staffID, businessID uuid.UUID) error
	InviteStaff(ctx context.Context, businessID uuid.UUID, req *InviteStaffRequest) (string, error)
	ValidateInviteToken(ctx context.Context, token string) (*BusinessInvite, error)
	AcceptInvite(ctx context.Context, userID uuid.UUID, token, password string) error
}
