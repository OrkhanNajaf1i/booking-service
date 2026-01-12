// File: internal/domain/service/service.go
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ServiceService struct {
	repo Repository
}

func NewServiceUseCase(repo Repository) *ServiceService {
	return &ServiceService{repo: repo}
}

// CreateService - Yeni xidmət yaratmaq
func (s *ServiceService) CreateService(
	ctx context.Context,
	businessID uuid.UUID,
	req *CreateServiceRequest,
) (*Service, error) {
	if businessID == uuid.Nil {
		return nil, &ServiceError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if req == nil {
		return nil, &ServiceError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	svc := NewService(businessID, req.Name, req.Description, req.DurationMinutes, req.Price)

	if err := s.validateService(svc); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, svc); err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	return svc, nil
}

func (s *ServiceService) ListServices(
	ctx context.Context,
	businessID uuid.UUID,
) ([]*Service, error) {
	if businessID == uuid.Nil {
		return nil, &ServiceError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}

	services, err := s.repo.ListByBusiness(ctx, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}

	return services, nil
}

func (s *ServiceService) GetService(
	ctx context.Context,
	id, businessID uuid.UUID,
) (*Service, error) {
	if id == uuid.Nil || businessID == uuid.Nil {
		return nil, &ServiceError{Code: "INVALID_ID", Message: "Service ID and Business ID are required"}
	}

	svc, err := s.repo.GetByID(ctx, id, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	if svc == nil {
		return nil, &ServiceError{Code: "NOT_FOUND", Message: "Service not found"}
	}

	return svc, nil
}

func (s *ServiceService) UpdateService(
	ctx context.Context,
	id, businessID uuid.UUID,
	req *UpdateServiceRequest,
) error {
	if id == uuid.Nil || businessID == uuid.Nil {
		return &ServiceError{Code: "INVALID_ID", Message: "Service ID and Business ID are required"}
	}
	if req == nil {
		return &ServiceError{Code: "INVALID_REQUEST", Message: "Request cannot be nil"}
	}

	svc, err := s.repo.GetByID(ctx, id, businessID)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}
	if svc == nil {
		return &ServiceError{Code: "NOT_FOUND", Message: "Service not found"}
	}

	svc.Name = req.Name
	svc.Description = req.Description
	svc.DurationMinutes = req.DurationMinutes
	svc.Price = req.Price
	svc.UpdatedAt = time.Now()

	if err := s.validateService(svc); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, svc); err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	return nil
}

func (s *ServiceService) DeactivateService(
	ctx context.Context,
	id, businessID uuid.UUID,
) error {
	if id == uuid.Nil || businessID == uuid.Nil {
		return &ServiceError{Code: "INVALID_ID", Message: "Service ID and Business ID are required"}
	}

	if err := s.repo.Deactivate(ctx, id, businessID); err != nil {
		return fmt.Errorf("failed to deactivate service: %w", err)
	}

	return nil
}

func (s *ServiceService) AssignServicesToStaff(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	serviceIDs []uuid.UUID,
) error {
	if businessID == uuid.Nil {
		return &ServiceError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if staffID == uuid.Nil {
		return &ServiceError{Code: "INVALID_STAFF", Message: "Staff ID cannot be empty"}
	}
	if err := s.validateAssignServicesRequest(serviceIDs); err != nil {
		return err
	}

	if err := s.repo.AssignServicesToStaff(ctx, businessID, staffID, serviceIDs); err != nil {
		return fmt.Errorf("failed to assign services to staff: %w", err)
	}

	return nil
}

func (s *ServiceService) GetStaffServices(
	ctx context.Context,
	businessID, staffID uuid.UUID,
) ([]*Service, error) {
	if businessID == uuid.Nil {
		return nil, &ServiceError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if staffID == uuid.Nil {
		return nil, &ServiceError{Code: "INVALID_STAFF", Message: "Staff ID cannot be empty"}
	}

	services, err := s.repo.GetStaffServices(ctx, businessID, staffID)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff services: %w", err)
	}

	return services, nil
}

func (s *ServiceService) RemoveServiceFromStaff(
	ctx context.Context,
	businessID, staffID, serviceID uuid.UUID,
) error {
	if businessID == uuid.Nil {
		return &ServiceError{Code: "INVALID_BUSINESS", Message: "Business ID cannot be empty"}
	}
	if staffID == uuid.Nil {
		return &ServiceError{Code: "INVALID_STAFF", Message: "Staff ID cannot be empty"}
	}
	if serviceID == uuid.Nil {
		return &ServiceError{Code: "INVALID_SERVICE", Message: "Service ID cannot be empty"}
	}

	if err := s.repo.RemoveServiceFromStaff(ctx, businessID, staffID, serviceID); err != nil {
		return fmt.Errorf("failed to remove service from staff: %w", err)
	}

	return nil
}
