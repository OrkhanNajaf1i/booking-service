// File: internal/domain/location/service.go
package location

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LocationService struct {
	repo Repository
}

func NewService(repo Repository) *LocationService {
	return &LocationService{repo: repo}
}

func (s *LocationService) CreateLocation(
	ctx context.Context,
	businessID uuid.UUID,
	req *CreateLocationRequest,
) (*Location, error) {
	if businessID == uuid.Nil {
		return nil, &LocationError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if req == nil {
		return nil, &LocationError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	location := NewLocation(businessID, req.Name)
	location.Address = req.Address
	location.City = req.City
	location.Phone = req.Phone

	if err := s.validateLocation(location); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, location); err != nil {
		return nil, fmt.Errorf("failed to create location: %w", err)
	}

	return location, nil
}

func (s *LocationService) CreateDefaultLocation(
	ctx context.Context,
	businessID uuid.UUID,
) (*Location, error) {
	if businessID == uuid.Nil {
		return nil, &LocationError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}

	location := NewLocation(businessID, "Default Location")

	if err := s.repo.Create(ctx, location); err != nil {
		return nil, fmt.Errorf("failed to create default location: %w", err)
	}

	return location, nil
}

func (s *LocationService) GetLocation(
	ctx context.Context,
	id, businessID uuid.UUID,
) (*Location, error) {
	if id == uuid.Nil || businessID == uuid.Nil {
		return nil, &LocationError{Code: "INVALID_ID", Message: "Location ID and Business ID are required"}
	}

	location, err := s.repo.GetByID(ctx, id, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}
	if location == nil {
		return nil, &LocationError{Code: "NOT_FOUND", Message: "Location not found"}
	}

	return location, nil
}

func (s *LocationService) ListLocations(
	ctx context.Context,
	businessID uuid.UUID,
) ([]*Location, error) {
	if businessID == uuid.Nil {
		return nil, &LocationError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}

	locations, err := s.repo.ListByBusiness(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	return locations, nil
}

func (s *LocationService) UpdateLocation(
	ctx context.Context,
	id, businessID uuid.UUID,
	req *UpdateLocationRequest,
) error {
	if id == uuid.Nil || businessID == uuid.Nil {
		return &LocationError{Code: "INVALID_ID", Message: "Location ID and Business ID are required"}
	}

	location, err := s.repo.GetByID(ctx, id, businessID)
	if err != nil {
		return fmt.Errorf("failed to get location: %w", err)
	}
	if location == nil {
		return &LocationError{Code: "NOT_FOUND", Message: "Location not found"}
	}

	location.Name = req.Name
	location.Address = req.Address
	location.City = req.City
	location.Phone = req.Phone
	location.UpdatedAt = time.Now()

	if err := s.validateLocation(location); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, location); err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	return nil
}

func (s *LocationService) DeactivateLocation(
	ctx context.Context,
	id, businessID uuid.UUID,
) error {
	if id == uuid.Nil || businessID == uuid.Nil {
		return &LocationError{Code: "INVALID_ID", Message: "Location ID and Business ID are required"}
	}

	if err := s.repo.Deactivate(ctx, id, businessID); err != nil {
		return fmt.Errorf("failed to deactivate location: %w", err)
	}

	return nil
}
