// File: internal/infrastructure/postgres/business_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

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
        INSERT INTO businesses (
            id, name, owner_id, industry, service_category, phone, 
            business_type, is_active, created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
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
		business.Industry,
		business.ServiceCategory,
		business.Phone,
		business.BusinessType,
		business.IsActive,
		business.CreatedAt,
		business.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create business: %w", err)
	}
	return nil
}

func (r *BusinessRepository) GetBusinessByID(
	ctx context.Context,
	id uuid.UUID,
) (*businessDomain.Business, error) {
	const query = `
        SELECT 
            id, name, owner_id, industry, service_category, phone, 
            business_type, is_active, created_at, updated_at
        FROM businesses
        WHERE id = $1
    `

	business := &businessDomain.Business{}
	var ownerID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&business.ID,
		&business.Name,
		&ownerID,
		&business.Industry,
		&business.ServiceCategory,
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
		return nil, fmt.Errorf("failed to get business by id: %w", err)
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

func (r *BusinessRepository) GetBusinessByOwnerID(
	ctx context.Context,
	ownerID uuid.UUID,
) (*businessDomain.Business, error) {
	const query = `
        SELECT 
            id, name, owner_id, industry, service_category, phone, 
            business_type, is_active, created_at, updated_at
        FROM businesses
        WHERE owner_id = $1
        LIMIT 1
    `

	business := &businessDomain.Business{}
	var ownerIDScanned sql.NullString

	err := r.db.QueryRowContext(ctx, query, ownerID).Scan(
		&business.ID,
		&business.Name,
		&ownerIDScanned,
		&business.Industry,
		&business.ServiceCategory,
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
		return nil, fmt.Errorf("failed to get business by owner_id: %w", err)
	}

	if ownerIDScanned.Valid {
		parsedUUID, _ := uuid.Parse(ownerIDScanned.String)
		business.OwnerID = parsedUUID
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
		return fmt.Errorf("failed to update owner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
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
        INSERT INTO locations (
            id, business_id, name, address, city, is_active, created_at, updated_at
        )
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
		return fmt.Errorf("failed to create location: %w", err)
	}
	return nil
}

func (r *BusinessRepository) GetLocationsByBusinessID(
	ctx context.Context,
	businessID uuid.UUID,
) ([]*businessDomain.Location, error) {
	const query = `
        SELECT 
            id, business_id, name, address, city, is_active, created_at, updated_at
        FROM locations
        WHERE business_id = $1
        ORDER BY created_at
    `

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to query locations: %w", err)
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
			return nil, fmt.Errorf("failed to scan location: %w", err)
		}
		locations = append(locations, loc)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return locations, nil
}

func (r *BusinessRepository) GetLocationByID(
	ctx context.Context,
	id uuid.UUID,
	businessID uuid.UUID,
) (*businessDomain.Location, error) {
	const query = `
        SELECT 
            id, business_id, name, address, city, is_active, created_at, updated_at
        FROM locations
        WHERE id = $1 AND business_id = $2
    `

	loc := &businessDomain.Location{}
	err := r.db.QueryRowContext(ctx, query, id, businessID).Scan(
		&loc.ID,
		&loc.BusinessID,
		&loc.Name,
		&loc.Address,
		&loc.City,
		&loc.IsActive,
		&loc.CreatedAt,
		&loc.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get location by id: %w", err)
	}

	return loc, nil
}

func (r *BusinessRepository) CreateStaffProfile(
	ctx context.Context,
	profile *businessDomain.StaffProfile,
) error {
	const query = `
        INSERT INTO staff_profiles (
            id, user_id, business_id, location_id, role, title, 
            status, joined_at, created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    `

	var locationID interface{}
	if profile.LocationID == nil || *profile.LocationID == uuid.Nil {
		locationID = nil
	} else {
		locationID = *profile.LocationID
	}

	_, err := r.db.ExecContext(
		ctx, query,
		profile.ID,
		profile.UserID,
		profile.BusinessID,
		locationID,
		profile.Role,
		profile.Title,
		profile.Status,
		profile.JoinedAt,
		profile.CreatedAt,
		profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create staff profile: %w", err)
	}
	return nil
}

func (r *BusinessRepository) GetStaffProfileByUserID(
	ctx context.Context,
	userID uuid.UUID,
	businessID uuid.UUID,
) (*businessDomain.StaffProfile, error) {
	const query = `
        SELECT 
            id, user_id, business_id, location_id, role, title, 
            status, joined_at, created_at, updated_at
        FROM staff_profiles
        WHERE user_id = $1 AND business_id = $2
        LIMIT 1
    `

	profile := &businessDomain.StaffProfile{}
	var locationID sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID, businessID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.BusinessID,
		&locationID,
		&profile.Role,
		&profile.Title,
		&profile.Status,
		&profile.JoinedAt,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get staff profile: %w", err)
	}

	if locationID.Valid {
		parsedUUID, err := uuid.Parse(locationID.String)
		if err == nil {
			profile.LocationID = &parsedUUID
		}
	}

	return profile, nil
}

func (r *BusinessRepository) UpdateStaffProfile(
	ctx context.Context,
	profile *businessDomain.StaffProfile,
) error {
	const query = `
        UPDATE staff_profiles
        SET 
            location_id = $1,
            role = $2,
            title = $3,
            status = $4,
            updated_at = $5
        WHERE id = $6 AND business_id = $7
    `

	var locationID interface{}
	if profile.LocationID == nil || *profile.LocationID == uuid.Nil {
		locationID = nil
	} else {
		locationID = *profile.LocationID
	}

	result, err := r.db.ExecContext(
		ctx, query,
		locationID,
		profile.Role,
		profile.Title,
		profile.Status,
		time.Now(),
		profile.ID,
		profile.BusinessID,
	)
	if err != nil {
		return fmt.Errorf("failed to update staff profile: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("staff profile not found or no change")
	}

	return nil
}

func (r *BusinessRepository) GetStaffByBusinessID(
	ctx context.Context,
	businessID uuid.UUID,
) ([]*businessDomain.StaffProfile, error) {
	const query = `
        SELECT 
            id, user_id, business_id, location_id, role, title, 
            status, joined_at, created_at, updated_at
        FROM staff_profiles
        WHERE business_id = $1
        ORDER BY joined_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to query staff profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*businessDomain.StaffProfile
	for rows.Next() {
		profile := &businessDomain.StaffProfile{}
		var locationID sql.NullString

		if err := rows.Scan(
			&profile.ID,
			&profile.UserID,
			&profile.BusinessID,
			&locationID,
			&profile.Role,
			&profile.Title,
			&profile.Status,
			&profile.JoinedAt,
			&profile.CreatedAt,
			&profile.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan staff profile: %w", err)
		}

		if locationID.Valid {
			parsedUUID, _ := uuid.Parse(locationID.String)
			profile.LocationID = &parsedUUID
		}

		profiles = append(profiles, profile)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return profiles, nil
}

func (r *BusinessRepository) CreateInvite(
	ctx context.Context,
	invite *businessDomain.BusinessInvite,
) error {
	const query = `
        INSERT INTO business_invites (
            id, business_id, invited_email, invited_phone, role, 
            location_id, token, expires_at, used, created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
    `

	var locationID interface{}
	if invite.LocationID == nil || *invite.LocationID == uuid.Nil {
		locationID = nil
	} else {
		locationID = *invite.LocationID
	}

	_, err := r.db.ExecContext(
		ctx, query,
		invite.ID,
		invite.BusinessID,
		invite.InvitedEmail,
		invite.InvitedPhone,
		invite.Role,
		locationID,
		invite.Token,
		invite.ExpiresAt,
		invite.Used,
		invite.CreatedAt,
		invite.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create invite: %w", err)
	}
	return nil
}

func (r *BusinessRepository) GetInviteByToken(
	ctx context.Context,
	token string,
	businessID uuid.UUID,
) (*businessDomain.BusinessInvite, error) {
	var query string
	var args []interface{}

	if businessID == uuid.Nil {
		query = `
            SELECT 
                id, business_id, invited_email, invited_phone, role, 
                location_id, token, expires_at, used, created_at, updated_at
            FROM business_invites
            WHERE token = $1
            LIMIT 1
        `
		args = []interface{}{token}
	} else {
		query = `
            SELECT 
                id, business_id, invited_email, invited_phone, role, 
                location_id, token, expires_at, used, created_at, updated_at
            FROM business_invites
            WHERE token = $1 AND business_id = $2
            LIMIT 1
        `
		args = []interface{}{token, businessID}
	}

	invite := &businessDomain.BusinessInvite{}
	var locationID sql.NullString

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&invite.ID,
		&invite.BusinessID,
		&invite.InvitedEmail,
		&invite.InvitedPhone,
		&invite.Role,
		&locationID,
		&invite.Token,
		&invite.ExpiresAt,
		&invite.Used,
		&invite.CreatedAt,
		&invite.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invite by token: %w", err)
	}

	if locationID.Valid {
		parsedUUID, _ := uuid.Parse(locationID.String)
		invite.LocationID = &parsedUUID
	}

	return invite, nil
}

func (r *BusinessRepository) UseInvite(
	ctx context.Context,
	inviteID uuid.UUID,
) error {
	const query = `
        UPDATE business_invites
        SET used = true, updated_at = NOW()
        WHERE id = $1
    `

	result, err := r.db.ExecContext(ctx, query, inviteID)
	if err != nil {
		return fmt.Errorf("failed to mark invite as used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invite not found")
	}

	return nil
}

func (r *BusinessRepository) GetInvitesByBusinessID(
	ctx context.Context,
	businessID uuid.UUID,
) ([]*businessDomain.BusinessInvite, error) {
	const query = `
        SELECT 
            id, business_id, invited_email, invited_phone, role, 
            location_id, token, expires_at, used, created_at, updated_at
        FROM business_invites
        WHERE business_id = $1
        ORDER BY created_at DESC
    `

	rows, err := r.db.QueryContext(ctx, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to query invites: %w", err)
	}
	defer rows.Close()

	var invites []*businessDomain.BusinessInvite
	for rows.Next() {
		invite := &businessDomain.BusinessInvite{}
		var locationID sql.NullString

		if err := rows.Scan(
			&invite.ID,
			&invite.BusinessID,
			&invite.InvitedEmail,
			&invite.InvitedPhone,
			&invite.Role,
			&locationID,
			&invite.Token,
			&invite.ExpiresAt,
			&invite.Used,
			&invite.CreatedAt,
			&invite.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan invite: %w", err)
		}

		if locationID.Valid {
			parsedUUID, _ := uuid.Parse(locationID.String)
			invite.LocationID = &parsedUUID
		}

		invites = append(invites, invite)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return invites, nil
}

func (r *BusinessRepository) UpdateUserBusinessID(
	ctx context.Context,
	userID uuid.UUID,
	businessID uuid.UUID,
) error {
	const query = `
        UPDATE users
        SET business_id = $1, updated_at = NOW()
        WHERE id = $2
    `

	result, err := r.db.ExecContext(ctx, query, businessID, userID)
	if err != nil {
		return fmt.Errorf("failed to update user business_id: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
