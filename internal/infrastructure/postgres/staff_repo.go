// File: internal/infrastructure/postgres/staff_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/staff"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type StaffRepository struct {
	db *sqlx.DB
}

func NewStaffRepository(db *sqlx.DB) *StaffRepository {
	return &StaffRepository{db: db}
}

func (r *StaffRepository) CreateStaffProfile(ctx context.Context, profile *staff.StaffProfile) error {
	query := `
		INSERT INTO staff_profiles (
			id, user_id, business_id, location_id, role, title, 
			department, bio, hourly_rate, status, joined_at, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		profile.ID, profile.UserID, profile.BusinessID, profile.LocationID,
		profile.Role, profile.Title, profile.Department, profile.Bio,
		profile.HourlyRate, profile.Status, profile.JoinedAt,
		profile.CreatedAt, profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert staff profile: %w", err)
	}

	return nil
}

func (r *StaffRepository) GetStaffByID(ctx context.Context, id, businessID uuid.UUID) (*staff.StaffProfile, error) {
	query := `
		SELECT id, user_id, business_id, location_id, role, title, 
			   department, bio, hourly_rate, status, joined_at, 
			   created_at, updated_at
		FROM staff_profiles
		WHERE id = $1 AND business_id = $2 AND status != 'inactive'
	`

	var profile staff.StaffProfile
	err := r.db.GetContext(ctx, &profile, query, id, businessID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}

	return &profile, nil
}

func (r *StaffRepository) GetStaffByUserID(ctx context.Context, userID, businessID uuid.UUID) (*staff.StaffProfile, error) {
	query := `
		SELECT id, user_id, business_id, location_id, role, title, 
			   department, bio, hourly_rate, status, joined_at, 
			   created_at, updated_at
		FROM staff_profiles
		WHERE user_id = $1 AND business_id = $2 AND status != 'inactive'
	`

	var profile staff.StaffProfile
	err := r.db.GetContext(ctx, &profile, query, userID, businessID)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get staff by user: %w", err)
	}

	return &profile, nil
}

func (r *StaffRepository) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]*staff.StaffWithUser, error) {
	query := `
		SELECT sp.id, sp.user_id, sp.role, sp.title, sp.department, 
			sp.location_id, sp.status, sp.joined_at,
			u.full_name, u.email, u.phone, u.avatar
		FROM staff_profiles sp
		JOIN users u ON sp.user_id = u.id
		WHERE sp.business_id = $1 AND sp.status != 'inactive'
		ORDER BY sp.joined_at DESC
	`

	var staffList []*staff.StaffWithUser
	err := r.db.SelectContext(ctx, &staffList, query, businessID)

	if err != nil {
		return nil, fmt.Errorf("failed to list staff: %w", err)
	}

	return staffList, nil
}

func (r *StaffRepository) UpdateStaffProfile(ctx context.Context, profile *staff.StaffProfile) error {
	query := `
		UPDATE staff_profiles
		SET role = $1, title = $2, department = $3, bio = $4, 
			hourly_rate = $5, location_id = $6, updated_at = $7
		WHERE id = $8 AND business_id = $9
	`

	result, err := r.db.ExecContext(
		ctx, query,
		profile.Role, profile.Title, profile.Department, profile.Bio,
		profile.HourlyRate, profile.LocationID, profile.UpdatedAt,
		profile.ID, profile.BusinessID,
	)

	if err != nil {
		return fmt.Errorf("failed to update staff: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("staff not found")
	}

	return nil
}

func (r *StaffRepository) DeactivateStaff(ctx context.Context, id, businessID uuid.UUID) error {
	query := `
		UPDATE staff_profiles
		SET status = 'inactive', updated_at = NOW()
		WHERE id = $1 AND business_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, businessID)

	if err != nil {
		return fmt.Errorf("failed to deactivate staff: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("staff not found")
	}

	return nil
}

func (r *StaffRepository) CreateInvite(ctx context.Context, invite *staff.BusinessInvite) error {
	query := `
		INSERT INTO business_invites (
			id, business_id, invited_email, invited_phone, role, 
			location_id, token, expires_at, used, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := r.db.ExecContext(
		ctx, query,
		invite.ID, invite.BusinessID, invite.InvitedEmail, invite.InvitedPhone,
		invite.Role, invite.LocationID, invite.Token, invite.ExpiresAt,
		invite.Used, invite.CreatedAt, invite.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert invite: %w", err)
	}

	return nil
}

func (r *StaffRepository) GetInviteByToken(ctx context.Context, token string) (*staff.BusinessInvite, error) {
	query := `
		SELECT id, business_id, invited_email, invited_phone, role, 
			   location_id, token, expires_at, used, created_at, updated_at
		FROM business_invites
		WHERE token = $1
	`

	var invite staff.BusinessInvite
	err := r.db.GetContext(ctx, &invite, query, token)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invite: %w", err)
	}

	return &invite, nil
}

func (r *StaffRepository) MarkInviteAsUsed(ctx context.Context, inviteID uuid.UUID) error {
	query := `
		UPDATE business_invites
		SET used = true, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, inviteID)

	if err != nil {
		return fmt.Errorf("failed to mark invite as used: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("invite not found")
	}

	return nil
}

func (r *StaffRepository) ListInvitesByBusiness(ctx context.Context, businessID uuid.UUID) ([]*staff.BusinessInvite, error) {
	query := `
		SELECT id, business_id, invited_email, invited_phone, role, 
			   location_id, token, expires_at, used, created_at, updated_at
		FROM business_invites
		WHERE business_id = $1
		ORDER BY created_at DESC
	`

	var invites []*staff.BusinessInvite
	err := r.db.SelectContext(ctx, &invites, query, businessID)

	if err != nil {
		return nil, fmt.Errorf("failed to list invites: %w", err)
	}

	return invites, nil
}
