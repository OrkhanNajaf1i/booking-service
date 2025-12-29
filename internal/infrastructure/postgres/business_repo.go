package postgres

import (
	"context"
	"database/sql"
	"fmt"

	businessDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
)

type BusinessRepository struct {
	db *sql.DB
}

func NewBusinessRepository(db *sql.DB) *BusinessRepository {
	return &BusinessRepository{db: db}
}

func (r *BusinessRepository) CreateBusiness(
	ctx context.Context,
	business *businessDomain.Business,
) error {
	const query = `
		INSERT INTO businesses (id, name, owner_id, phone, business_type, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	var ownerID interface{}
	if business.OwnerID == uuid.Nil {
		ownerID = nil
	} else {
		ownerID = business.OwnerID
	}

	_, err := r.db.ExecContext(
		ctx, query,
		business.ID,
		business.Name,
		ownerID,
		business.Phone,
		business.BusinessType,
		business.IsActive,
		business.CreatedAt,
		business.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create business failed: %w", err)
	}
	return nil
}

func (r *BusinessRepository) GetBusinessByID(
	ctx context.Context,
	id uuid.UUID,
) (*businessDomain.Business, error) {
	const query = `
		SELECT id, name, owner_id, phone, business_type, is_active, created_at, updated_at
		FROM businesses
		WHERE id = $1
	`
	business := &businessDomain.Business{}
	var ownerID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&business.ID,
		&business.Name,
		&ownerID,
		&business.Phone,
		&business.BusinessType,
		&business.IsActive,
		&business.CreatedAt,
		&business.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("query business failed: %w", err)
	}

	if ownerID.Valid {
		parsedUUID, err := uuid.Parse(ownerID.String)
		if err == nil {
			business.OwnerID = parsedUUID
		}
	} else {
		business.OwnerID = uuid.Nil
	}

	return business, nil
}

func (r *BusinessRepository) UpdateOwner(
	ctx context.Context,
	businessID uuid.UUID,
	ownerID uuid.UUID,
) error {
	const query = `
		UPDATE businesses 
		SET owner_id = $1, updated_at = NOW() 
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, ownerID, businessID)
	if err != nil {
		return fmt.Errorf("update owner failed: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected failed: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("business not found or no change")
	}

	return nil
}

func (r *BusinessRepository) CreateLocation(
	ctx context.Context,
	location *businessDomain.Location,
) error {
	const query = `
		INSERT INTO locations (id, business_id, name, address, city, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		location.ID,
		location.BusinessID,
		location.Name,
		location.Address,
		location.City,
		location.IsActive,
		location.CreatedAt,
		location.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create location failed: %w", err)
	}
	return nil
}

func (r *BusinessRepository) GetLocationsByBusinessID(
	ctx context.Context,
	businessID uuid.UUID,
) ([]*businessDomain.Location, error) {
	const query = `
		SELECT id, business_id, name, address, city, is_active, created_at, updated_at
		FROM locations
		WHERE business_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("query locations failed: %w", err)
	}
	defer rows.Close()

	var locations []*businessDomain.Location
	for rows.Next() {
		loc := &businessDomain.Location{}
		if err := rows.Scan(
			&loc.ID,
			&loc.BusinessID,
			&loc.Name,
			&loc.Address,
			&loc.City,
			&loc.IsActive,
			&loc.CreatedAt,
			&loc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan location failed: %w", err)
		}
		locations = append(locations, loc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return locations, nil
}
