package business

import (
	"context"

	"github.com/google/uuid"
)

type Repository interface {
	CreateBusiness(ctx context.Context, business *Business) error
	GetBusinessByID(ctx context.Context, id uuid.UUID) (*Business, error)
}
