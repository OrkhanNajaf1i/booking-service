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
}

type Service interface {
	CreateBusiness(ctx context.Context, ownerID uuid.UUID, request *CreateBusinessRequest) (*Business, error)
	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
	GetBusinessByOwner(ctx context.Context, ownerID uuid.UUID) (*Business, error)
	UpdateBusiness(ctx context.Context, businessID uuid.UUID, request *UpdateBusinessRequest) error
}
