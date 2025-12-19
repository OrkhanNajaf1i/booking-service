package business

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

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
	b, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get business: %w", err)
	}
	if b == nil {
		return nil, fmt.Errorf("business not found")
	}
	return b, nil
}
