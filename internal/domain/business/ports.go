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

	// ListBookable – yalniz musterinin bron ede bileceyi bizneslər:
	// aktiv olmaqla yanasi en azi bir aktiv iscisi olmalidir.
	// Isci olmayan biznes musteriye gosterilse, o, secim edib bos
	// ekranla qarsilasir.
	ListBookable(ctx context.Context) ([]*BookableBusiness, error)
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
	// Yaradilan (ve ya movcud) profilin ID-sini qaytarir: qrafik
	// hemin isciye baglanir.
	EnsureOwnerProfile(ctx context.Context, businessID, userID uuid.UUID, title string) (uuid.UUID, error)
}

// ScheduleProvisioner – yeni isciye baslangic is qrafiki qurur.
//
// Bos vaxtlar sorgu aninda qrafikden hesablanir. Qrafik olmasa musteri
// HER gun ucun "is gunu deyil" gorur, sahib ise neyin catmadigini
// bilmir — yeni biznes sinmis kimi gorunur.
type ScheduleProvisioner interface {
	EnsureDefaultSchedule(ctx context.Context, businessID, staffID uuid.UUID) error
}

type Service interface {
	CreateBusiness(ctx context.Context, ownerID uuid.UUID, request *CreateBusinessRequest) (*Business, error)
	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
	GetBusinessByOwner(ctx context.Context, ownerID uuid.UUID) (*Business, error)
	ListBusinesses(ctx context.Context) ([]*Business, error)
	ListBookableBusinesses(ctx context.Context) ([]*BookableBusiness, error)
	UpdateBusiness(ctx context.Context, businessID uuid.UUID, ownerID uuid.UUID, request *UpdateBusinessRequest) error
}
