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
			id, name, owner_id, industry, service_category, category_slug,
			phone, business_type, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := repository.database.ExecContext(
		ctx, query,
		business.ID,
		business.Name,
		business.OwnerID,
		business.Industry,
		business.ServiceCategory,
		business.CategorySlug,
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
			COALESCE(category_slug, '') AS category_slug,
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
			COALESCE(category_slug, '') AS category_slug,
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
func (repository *BusinessRepository) ListBusinesses(ctx context.Context) ([]*business.Business, error) {
	query := `
		SELECT
			id, name, owner_id, industry, service_category,
			COALESCE(category_slug, '') AS category_slug,
			phone, business_type, is_active, created_at, updated_at
		FROM businesses
		ORDER BY created_at DESC
	`
	var businesses []*business.Business
	err := repository.database.SelectContext(ctx, &businesses, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get business by owner ID: %w", err)
	}
	if businesses == nil {
		return []*business.Business{}, nil
	}
	return businesses, nil
}
func (repository *BusinessRepository) Update(ctx context.Context, business *business.Business) error {
	query := `
		UPDATE businesses
		SET
			name = $1,
			industry = $2,
			phone = $3,
			service_category = $4,
			category_slug = NULLIF($5, ''),
			updated_at = $6
		WHERE id = $7
	`

	result, err := repository.database.ExecContext(
		ctx, query,
		business.Name,
		business.Industry,
		business.Phone,
		business.ServiceCategory,
		business.CategorySlug,
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

// ListBookable – musterinin real olaraq bron ede bileceyi bizneslər.
//
// Sert: biznes aktivdir VE en azi bir aktiv iscisi var. Isci olmadan
// randevu yaradila bilmir (booking staff_id-ye baglanir), ona gore bele
// biznesi musteriye gostermek onu bos ekrana aparir.
//
// Filiallar da eyni sorgu ile gelir: "yaxinliqdakilar" filtri mesafeni
// biznesin en yaxin filialina gore olcur. LEFT JOIN-dir — filiali
// olmayan biznes siyahidan dusmemelidir, sadece mesafesiz gorunur.
func (repository *BusinessRepository) ListBookable(ctx context.Context) ([]*business.BookableBusiness, error) {
	query := `
		SELECT
			b.id, b.name, b.owner_id, b.industry, b.service_category,
			COALESCE(b.category_slug, '') AS category_slug,
			b.phone, b.business_type, b.is_active, b.created_at, b.updated_at,
			l.name      AS location_name,
			l.address   AS location_address,
			l.city      AS location_city,
			l.latitude  AS location_latitude,
			l.longitude AS location_longitude
		FROM businesses b
		LEFT JOIN locations l
		       ON l.business_id = b.id
		      AND l.is_active = TRUE
		WHERE b.is_active = TRUE
		  AND EXISTS (
			  SELECT 1 FROM staff_profiles sp
			  WHERE sp.business_id = b.id AND sp.status = 'active'
		  )
		ORDER BY b.created_at DESC, l.created_at ASC
	`

	// JOIN sebebi ile biznes her filiali ucun bir setir qaytarir;
	// asagida yenidan yigilir.
	type row struct {
		business.Business
		LocationName      *string  `db:"location_name"`
		LocationAddress   *string  `db:"location_address"`
		LocationCity      *string  `db:"location_city"`
		LocationLatitude  *float64 `db:"location_latitude"`
		LocationLongitude *float64 `db:"location_longitude"`
	}

	var rows []row
	if err := repository.database.SelectContext(ctx, &rows, query); err != nil {
		if err == sql.ErrNoRows {
			return []*business.BookableBusiness{}, nil
		}
		return nil, fmt.Errorf("postgres: failed to list bookable businesses: %w", err)
	}

	result := make([]*business.BookableBusiness, 0, len(rows))
	index := make(map[uuid.UUID]*business.BookableBusiness, len(rows))

	for _, item := range rows {
		entry, seen := index[item.Business.ID]
		if !seen {
			entry = &business.BookableBusiness{Business: item.Business}
			index[item.Business.ID] = entry
			result = append(result, entry)
		}

		if item.LocationName == nil {
			continue
		}
		entry.Locations = append(entry.Locations, business.LocationSummary{
			Name:      *item.LocationName,
			Address:   derefString(item.LocationAddress),
			City:      derefString(item.LocationCity),
			Latitude:  item.LocationLatitude,
			Longitude: item.LocationLongitude,
		})
	}

	return result, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

// OwnerUserID – biznesin sahibinin user_id-si.
//
// staff domeni bunu sahibin oz profilini silmesinin qarsisini almaq
// ucun isledir (staff.BusinessOwnerLookup).
func (repository *BusinessRepository) OwnerUserID(
	ctx context.Context,
	businessID uuid.UUID,
) (uuid.UUID, error) {
	var ownerID uuid.UUID

	err := repository.database.GetContext(
		ctx, &ownerID,
		`SELECT owner_id FROM businesses WHERE id = $1`,
		businessID,
	)
	if err == sql.ErrNoRows {
		return uuid.Nil, nil
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("postgres: failed to get business owner: %w", err)
	}

	return ownerID, nil
}
