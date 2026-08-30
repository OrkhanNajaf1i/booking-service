// File: internal/domain/business/ports.go
package business

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, business *Business) error
	GetByID(ctx context.Context, id uuid.UUID) (*Business, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*Business, error)
	Update(ctx context.Context, business *Business) error
	UpdateOwner(ctx context.Context, businessID, ownerID uuid.UUID) error
	ListBusinesses(ctx context.Context) ([]*Business, error)
}

// OwnerLinker – istifadecini yaradilan biznese baglayir.
//
// Bu addim olmadan JWT-ye business_id dusmur ve butun biznes-kontekstli
// endpoint-ler ise dusmur: /bookings, /staff, /availability ve s.
type OwnerLinker interface {
	UpdateUserBusinessID(ctx context.Context, userID, businessID uuid.UUID, isOwner bool) error
}

// StaffProvisioner – biznes sahibi ucun staff profili yaradir.
//
// Solo hekim/berber ozu de xidmet gosteren sexsdir: randevu staff_id-ye
// baglanir, ona gore sahibin profili olmasa ne qrafik teyin edile bilir,
// ne de bron qebul olunur.
type StaffProvisioner interface {
	EnsureOwnerProfile(ctx context.Context, businessID, userID uuid.UUID, title string) error
}

type Service interface {
	CreateBusiness(ctx context.Context, ownerID uuid.UUID, request *CreateBusinessRequest) (*Business, error)
	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
	GetBusinessByOwner(ctx context.Context, ownerID uuid.UUID) (*Business, error)
	ListBusinesses(ctx context.Context) ([]*Business, error)
	UpdateBusiness(ctx context.Context, businessID uuid.UUID, ownerID uuid.UUID, request *UpdateBusinessRequest) error
}
