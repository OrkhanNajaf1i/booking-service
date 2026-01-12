// File: internal/infrastructure/postgres/business_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BusinessRepository struct {
	database *sqlx.DB
}

func NewBusinessRepository(database *sqlx.DB) *BusinessRepository {
	return &BusinessRepository{
		database: database,
	}
}

func (repository *BusinessRepository) Create(ctx context.Context, business *business.Business) error {
	query := `
		INSERT INTO businesses (
			id, name, owner_id, industry, service_category, 
			phone, business_type, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := repository.database.ExecContext(
		ctx, query,
		business.ID,
		business.Name,
		business.OwnerID,
		business.Industry,
		business.ServiceCategory,
		business.Phone,
		business.BusinessType,
		business.IsActive,
		business.CreatedAt,
		business.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres: failed to insert business: %w", err)
	}

	return nil
}

func (repository *BusinessRepository) GetByID(ctx context.Context, id uuid.UUID) (*business.Business, error) {
	query := `
		SELECT 
			id, name, owner_id, industry, service_category, 
			phone, business_type, is_active, created_at, updated_at
		FROM businesses
		WHERE id = $1
	`

	var businessEntity business.Business
	err := repository.database.GetContext(ctx, &businessEntity, query, id)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get business by ID: %w", err)
	}

	return &businessEntity, nil
}

func (repository *BusinessRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) (*business.Business, error) {
	query := `
		SELECT 
			id, name, owner_id, industry, service_category, 
			phone, business_type, is_active, created_at, updated_at
		FROM businesses
		WHERE owner_id = $1 AND is_active = true
		ORDER BY created_at DESC
		LIMIT 1
	`

	var businessEntity business.Business
	err := repository.database.GetContext(ctx, &businessEntity, query, ownerID)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get business by owner ID: %w", err)
	}

	return &businessEntity, nil
}

func (repository *BusinessRepository) Update(ctx context.Context, business *business.Business) error {
	query := `
		UPDATE businesses
		SET 
			name = $1,
			industry = $2,
			phone = $3,
			updated_at = $4
		WHERE id = $5
	`

	result, err := repository.database.ExecContext(
		ctx, query,
		business.Name,
		business.Industry,
		business.Phone,
		business.UpdatedAt,
		business.ID,
	)

	if err != nil {
		return fmt.Errorf("postgres: failed to update business: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("postgres: business not found for update")
	}

	return nil
}

func (repository *BusinessRepository) UpdateOwner(ctx context.Context, businessID, ownerID uuid.UUID) error {
	query := `
		UPDATE businesses
		SET 
			owner_id = $1,
			updated_at = NOW()
		WHERE id = $2
	`

	result, err := repository.database.ExecContext(ctx, query, ownerID, businessID)

	if err != nil {
		return fmt.Errorf("postgres: failed to update business owner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("postgres: business not found for owner update")
	}

	return nil
}
