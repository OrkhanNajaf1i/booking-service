// File: internal/domain/location/ports.go
package location

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, location *Location) error
	GetByID(ctx context.Context, id, businessID uuid.UUID) (*Location, error)
	ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]*Location, error)
	Update(ctx context.Context, location *Location) error
	Deactivate(ctx context.Context, id, businessID uuid.UUID) error
}

type Service interface {
	CreateLocation(ctx context.Context, businessID uuid.UUID, req *CreateLocationRequest) (*Location, error)
	CreateDefaultLocation(ctx context.Context, businessID uuid.UUID) (*Location, error)
	GetLocation(ctx context.Context, id, businessID uuid.UUID) (*Location, error)
	ListLocations(ctx context.Context, businessID uuid.UUID) ([]*Location, error)
	UpdateLocation(ctx context.Context, id, businessID uuid.UUID, req *UpdateLocationRequest) error
	DeactivateLocation(ctx context.Context, id, businessID uuid.UUID) error
}
