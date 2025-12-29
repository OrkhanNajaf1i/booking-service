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

func (s *Service) AssignOwner(ctx context.Context, businessID uuid.UUID, ownerID uuid.UUID) error {
	if businessID == uuid.Nil {
		return fmt.Errorf("business ID cannot be empty")
	}
	if ownerID == uuid.Nil {
		return fmt.Errorf("owner ID cannot be empty")
	}

	if err := s.repo.UpdateOwner(ctx, businessID, ownerID); err != nil {
		return fmt.Errorf("failed to assign owner: %w", err)
	}
	return nil
}

func (s *Service) createBusinessInternal(ctx context.Context, name string, bType BusinessType) (string, error) {
	if name == "" {
		return "", fmt.Errorf("business name cannot be empty")
	}
	business := &Business{
		ID:           uuid.New(),
		Name:         name,
		OwnerID:      uuid.Nil,
		BusinessType: bType,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateBusiness(ctx, business); err != nil {
		return "", fmt.Errorf("failed to create business: %w", err)
	}

	return business.ID.String(), nil
}

func (s *Service) CreateSoloPractitionerBusiness(
	ctx context.Context,
	ownerID string,
	name string,
	industry string,
) (string, error) {
	return s.createBusinessInternal(ctx, name, BusinessTypeSolo)
}

func (s *Service) CreateMultiStaffBusiness(
	ctx context.Context,
	ownerID string,
	name string,
	industry string,
) (string, error) {
	return s.createBusinessInternal(ctx, name, BusinessTypeMulti)
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

func (s *Service) CreateBusiness(ctx context.Context, name, phone string) (*Business, error) {
	if name == "" {
		return nil, fmt.Errorf("business name cannot be empty")
	}
	b := &Business{
		ID:        uuid.New(),
		Name:      name,
		Phone:     phone,
		OwnerID:   uuid.Nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
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
