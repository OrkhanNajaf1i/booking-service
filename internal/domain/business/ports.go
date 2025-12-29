// File: internal/domain/business/ports.go
package business

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateBusiness(ctx context.Context, business *Business) error
	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
	UpdateOwner(ctx context.Context, businessID uuid.UUID, ownerID uuid.UUID) error
	CreateLocation(ctx context.Context, location *Location) error
	GetLocationsByBusinessID(ctx context.Context, businessID uuid.UUID) ([]*Location, error)
}
