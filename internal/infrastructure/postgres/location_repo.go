// File: internal/infrastructure/postgres/location_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/location"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type LocationRepository struct {
	db *sqlx.DB
}

func NewLocationRepository(db *sqlx.DB) *LocationRepository {
	return &LocationRepository{db: db}
}

func (r *LocationRepository) Create(ctx context.Context, loc *location.Location) error {
	query := `
		INSERT INTO locations (
			id, business_id, name, address, city, phone, 
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		loc.ID, loc.BusinessID, loc.Name, loc.Address, loc.City, loc.Phone,
		loc.IsActive, loc.CreatedAt, loc.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert location: %w", err)
	}

	return nil
}

func (r *LocationRepository) GetByID(ctx context.Context, id, businessID uuid.UUID) (*location.Location, error) {
	query := `
		SELECT id, business_id, name, address, city, phone, 
			   is_active, created_at, updated_at
		FROM locations
		WHERE id = $1 AND business_id = $2 AND is_active = true
	`

	var loc location.Location
	err := r.db.GetContext(ctx, &loc, query, id, businessID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get location: %w", err)
	}

	return &loc, nil
}

func (r *LocationRepository) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]*location.Location, error) {
	query := `
		SELECT id, business_id, name, address, city, phone, 
			   is_active, created_at, updated_at
		FROM locations
		WHERE business_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`

	var locations []*location.Location
	err := r.db.SelectContext(ctx, &locations, query, businessID)

	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}

	return locations, nil
}

func (r *LocationRepository) Update(ctx context.Context, loc *location.Location) error {
	query := `
		UPDATE locations
		SET name = $1, address = $2, city = $3, phone = $4, updated_at = $5
		WHERE id = $6 AND business_id = $7
	`

	result, err := r.db.ExecContext(
		ctx, query,
		loc.Name, loc.Address, loc.City, loc.Phone, loc.UpdatedAt,
		loc.ID, loc.BusinessID,
	)

	if err != nil {
		return fmt.Errorf("failed to update location: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("location not found")
	}

	return nil
}

func (r *LocationRepository) Deactivate(ctx context.Context, id, businessID uuid.UUID) error {
	query := `
		UPDATE locations
		SET is_active = false, updated_at = NOW()
		WHERE id = $1 AND business_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, businessID)

	if err != nil {
		return fmt.Errorf("failed to deactivate location: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("location not found")
	}

	return nil
}
