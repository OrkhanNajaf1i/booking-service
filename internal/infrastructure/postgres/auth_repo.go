package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/google/uuid"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{
		db: db,
	}
}

func (r *AuthRepository) CreateUser(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (id, email, full_name, phone, password_hash, role, business_id, avatar, is_active, is_owner, email_verified, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`
	_, err := r.db.ExecContext(ctx, query, user.ID, user.Email, user.FullName, user.Phone, user.PasswordHash, user.Role, user.BusinessID, user.Avatar, user.IsActive, user.IsOwner, user.EmailVerified, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `SELECT id, email, full_name, phone, password_hash, role, business_id, avatar, is_active, is_owner, email_verified, created_at, updated_at 
	          FROM users WHERE email = $1 LIMIT 1`
	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Phone, &user.PasswordHash,
		&user.Role, &user.BusinessID, &user.Avatar, &user.IsActive, &user.IsOwner,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}
	return user, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	query := `SELECT id, email, full_name, phone, password_hash, role, business_id, avatar, is_active, is_owner, email_verified, created_at, updated_at 
	          FROM users WHERE id = $1`
	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Phone, &user.PasswordHash,
		&user.Role, &user.BusinessID, &user.Avatar, &user.IsActive, &user.IsOwner,
		&user.EmailVerified, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id %s: %w", id, err)
	}
	return user, nil
}

func (r *AuthRepository) SaveRefreshToken(ctx context.Context, token *auth.RefreshToken) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, revoked) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, token.ID, token.UserID, token.Token, token.ExpiresAt, token.CreatedAt, token.Revoked)
	if err != nil {
		return fmt.Errorf("failed to save refresh token for user %s: %w", token.UserID, err)
	}
	return nil
}

func (r *AuthRepository) GetRefreshToken(ctx context.Context, token string) (*auth.RefreshToken, error) {
	query := `SELECT id, user_id, token, expires_at, created_at, revoked FROM refresh_tokens WHERE token = $1`
	rt := &auth.RefreshToken{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(&rt.ID, &rt.UserID, &rt.Token, &rt.ExpiresAt, &rt.CreatedAt, &rt.Revoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	return rt, nil
}

func (r *AuthRepository) RevokeRefreshToken(ctx context.Context, tokenID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, tokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token %s: %w", tokenID, err)
	}
	return nil
}
func (r *AuthRepository) SavePasswordReset(ctx context.Context, reset *auth.PasswordReset) error {
	query := `INSERT INTO password_resets (id, email, token, expires_at, used, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.ExecContext(ctx, query, reset.ID, reset.Email, reset.Token, reset.ExpiresAt, reset.Used, reset.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to save password reset for %s: %w", reset.Email, err)
	}
	return nil
}

func (r *AuthRepository) GetPasswordReset(ctx context.Context, token string) (*auth.PasswordReset, error) {
	query := `SELECT id, email, token, expires_at, used, created_at FROM password_resets WHERE token = $1`
	pr := &auth.PasswordReset{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(&pr.ID, &pr.Email, &pr.Token, &pr.ExpiresAt, &pr.Used, &pr.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get password reset: %w", err)
	}
	return pr, nil
}

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = $2 WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password for user %s: %w", userID, err)
	}
	return nil
}

func (r *AuthRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence for %s: %w", email, err)
	}
	return exists, nil
}
func (r *AuthRepository) UpdateUserStatus(ctx context.Context, userID uuid.UUID, status string) error {
	// active := (status == "active")
	query := `UPDATE users SET is_active = $1, updated_at = $2, WHERE id = $3`
	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update status for user %s: %w", userID, err)
	}
	return nil
}

func (r *AuthRepository) CreateStaffProfile(ctx context.Context, profile *auth.StaffProfile) error {
	query := `INSERT INTO staff_profiles (id, user_id, business_id, location_id, role, title, department, bio, hourly_rate, status, joined_at, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`
	_, err := r.db.ExecContext(ctx, query,
		profile.ID, profile.UserID, profile.BusinessID, profile.LocationID,
		profile.Role, profile.Title, profile.Department, profile.Bio,
		profile.HourlyRate, profile.Status, profile.JoinedAt, profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create staff profile for user %s: %w", profile.UserID, err)
	}
	return nil
}

func (r *AuthRepository) GetStaffProfile(ctx context.Context, userID string) (*auth.StaffProfile, error) {
	query := `SELECT id, user_id, business_id, location_id, role, title, department, bio, hourly_rate, status, joined_at, created_at, updated_at 
	          FROM staff_profiles WHERE user_id = $1`
	sp := &auth.StaffProfile{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&sp.ID, &sp.UserID, &sp.BusinessID, &sp.LocationID, &sp.Role, &sp.Title,
		&sp.Department, &sp.Bio, &sp.HourlyRate, &sp.Status, &sp.JoinedAt, &sp.CreatedAt, &sp.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get staff profile for user %s: %w", userID, err)
	}
	return sp, nil
}

func (r *AuthRepository) UpdateStaffProfile(ctx context.Context, staffID string, profile *auth.StaffProfile) error {
	query := `UPDATE staff_profiles SET title = $1, department = $2, bio = $3, hourly_rate = $4, status = $5, updated_at = $6 WHERE id = $7`
	_, err := r.db.ExecContext(ctx, query, profile.Title, profile.Department, profile.Bio, profile.HourlyRate, profile.Status, time.Now(), staffID)
	if err != nil {
		return fmt.Errorf("failed to update staff profile %s: %w", staffID, err)
	}
	return nil
}
