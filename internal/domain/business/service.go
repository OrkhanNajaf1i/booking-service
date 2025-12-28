// File: internal/domain/business/service.go
package business

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Mövcud metodlar
func (s *Service) CreateBusiness(ctx context.Context, name, phone string) (*Business, error) {
	if name == "" {
		return nil, fmt.Errorf("business name cannot be empty")
	}
	if phone == "" {
		return nil, fmt.Errorf("phone cannot be empty")
	}
	b := New(name, phone)
	if err := s.repo.CreateBusiness(ctx, b); err != nil {
		return nil, fmt.Errorf("failed to create business: %w", err)
	}
	return b, nil
}

func (s *Service) GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error) {
	b, err := s.repo.GetBusinessByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get business: %w", err)
	}
	if b == nil {
		return nil, fmt.Errorf("business not found")
	}
	return b, nil
}

// ✅ auth.BusinessService interfeysini implement edirik
// Bu metodlar Multi-tenant izolyasiyanı, business logic-i və location yaratmasını həyata keçirir

// CreateSoloPractitionerBusiness - Solo iş üçün business + default location yaradır
func (s *Service) CreateSoloPractitionerBusiness(
	ctx context.Context,
	ownerID string, // UUID string olaraq gəlir (auth qatından)
	name string, // Business adı
	industry string, // Sənayelə (məsələn: "Barber", "Consultant")
) (string, error) {
	// Input validation
	if ownerID == "" {
		return "", fmt.Errorf("owner ID cannot be empty")
	}
	if name == "" {
		return "", fmt.Errorf("business name cannot be empty")
	}

	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return "", fmt.Errorf("invalid owner ID: %w", err)
	}

	business := &Business{
		ID:           uuid.New(),
		Name:         name,
		OwnerID:      ownerUUID,
		BusinessType: BusinessTypeSolo,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateBusiness(ctx, business); err != nil {
		return "", fmt.Errorf("failed to create solo business: %w", err)
	}

	return business.ID.String(), nil
}

func (s *Service) CreateMultiStaffBusiness(
	ctx context.Context,
	ownerID string,
	name string,
	industry string,
) (string, error) {
	if ownerID == "" {
		return "", fmt.Errorf("owner ID cannot be empty")
	}
	if name == "" {
		return "", fmt.Errorf("business name cannot be empty")
	}

	ownerUUID, err := uuid.Parse(ownerID)
	if err != nil {
		return "", fmt.Errorf("invalid owner ID: %w", err)
	}

	business := &Business{
		ID:           uuid.New(),
		Name:         name,
		OwnerID:      ownerUUID,
		BusinessType: BusinessTypeMulti,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateBusiness(ctx, business); err != nil {
		return "", fmt.Errorf("failed to create multi-staff business: %w", err)
	}

	return business.ID.String(), nil
}

func (s *Service) CreateDefaultLocation(
	ctx context.Context,
	businessID uuid.UUID,
) (string, error) {
	if businessID == uuid.Nil {
		return "", fmt.Errorf("business ID cannot be nil")
	}

	business, err := s.repo.GetBusinessByID(ctx, businessID)
	if err != nil {
		return "", fmt.Errorf("failed to verify business: %w", err)
	}
	if business == nil {
		return "", fmt.Errorf("business not found")
	}

	location := &Location{
		ID:         uuid.New(),
		BusinessID: businessID,
		Name:       "Default Location",
		IsActive:   true,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := s.repo.CreateLocation(ctx, location); err != nil {
		return "", fmt.Errorf("failed to create default location: %w", err)
	}

	return location.ID.String(), nil
}
