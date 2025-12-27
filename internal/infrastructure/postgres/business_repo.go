package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
)

type BusinessRepository struct {
	db *sql.DB
}

func NewBusinessRepository(db *sql.DB) *BusinessRepository {
	return &BusinessRepository{db: db}
}
func (r *BusinessRepository) CreateBusiness(ctx context.Context, business *business.Business) error {
	query := `
        INSERT INTO businesses (
            id, name, owner_id, business_type, is_active, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7
        )
    `
	_, err := r.db.ExecContext(ctx, query, business.ID, business.Name, business.OwnerID, business.BusinessType, business.IsActive, business.CreatedAt, business.UpdatedAt)
	if err != nil {
		return fmt.Errorf("postgres: failed to create business: %w", err)
	}
	return nil
}

func (r *BusinessRepository) GetBusinessByID(ctx context.Context, id uuid.UUID) (*business.Business, error) {
	query := `
        SELECT id, name, owner_id, business_type, is_active, created_at, updated_at
        FROM businesses
        WHERE id = $1
        LIMIT 1
    `
	var business business.Business
	row := r.db.QueryRowContext(ctx, query, id)
	if err := row.Scan(&business.ID, &business.Name, &business.OwnerID, &business.BusinessType, &business.IsActive, &business.CreatedAt, &business.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("postgres: failed to get business: %w", err)
	}
	return &business, nil
}
