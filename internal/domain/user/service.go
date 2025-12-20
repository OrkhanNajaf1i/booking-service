package user

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

func (s *Service) CreateUser(ctx context.Context, businessID uuid.UUID, name, phone string) (*User, error) {
	if name == "" {
		return nil, fmt.Errorf("user name cannot be empty")
	}
	if phone == "" {
		return nil, fmt.Errorf("phone cannot be empty")
	}
	existing, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("user with this phone already exist")
	}

	u := New(businessID, name, phone)
	return u, nil
}

func (s *Service) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	if phone == "" {
		return nil, fmt.Errorf("phone cannot be empty")
	}

	u, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if u == nil {
		return nil, fmt.Errorf("user not found")
	}
	return u, nil
}
