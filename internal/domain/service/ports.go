// File: internal/domain/service/ports.go
package service

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, s *Service) error
	GetByID(ctx context.Context, id, businessID uuid.UUID) (*Service, error)
	ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]*Service, error)
	Update(ctx context.Context, s *Service) error
	Deactivate(ctx context.Context, id, businessID uuid.UUID) error

	AssignServicesToStaff(ctx context.Context, businessID, staffID uuid.UUID, serviceIDs []uuid.UUID) error
	GetStaffServices(ctx context.Context, businessID, staffID uuid.UUID) ([]*Service, error)
	RemoveServiceFromStaff(ctx context.Context, businessID, staffID, serviceID uuid.UUID) error
}

type ServiceUseCase interface {
	CreateService(ctx context.Context, businessID uuid.UUID, req *CreateServiceRequest) (*Service, error)
	ListServices(ctx context.Context, businessID uuid.UUID) ([]*Service, error)
	GetService(ctx context.Context, id, businessID uuid.UUID) (*Service, error)
	UpdateService(ctx context.Context, id, businessID uuid.UUID, req *UpdateServiceRequest) error
	DeactivateService(ctx context.Context, id, businessID uuid.UUID) error

	AssignServicesToStaff(ctx context.Context, businessID, staffID uuid.UUID, serviceIDs []uuid.UUID) error
	GetStaffServices(ctx context.Context, businessID, staffID uuid.UUID) ([]*Service, error)
	RemoveServiceFromStaff(ctx context.Context, businessID, staffID, serviceID uuid.UUID) error
}
