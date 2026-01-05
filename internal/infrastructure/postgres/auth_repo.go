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
        INSERT INTO users (
            id, email, full_name, phone, password_hash, 
            role, business_id, avatar, is_active, is_owner, 
            email_verified, created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
    `
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.FullName,
		user.Phone,
		user.PasswordHash,
		user.Role,
		user.BusinessID,
		user.Avatar,
		user.IsActive,
		user.IsOwner,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
        SELECT id, email, full_name, phone, password_hash, role, 
               business_id, avatar, is_active, is_owner, email_verified, 
               created_at, updated_at 
        FROM users 
        WHERE email = $1 
        LIMIT 1
    `

	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.BusinessID,
		&user.Avatar,
		&user.IsActive,
		&user.IsOwner,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email %s: %w", email, err)
	}
	return user, nil
}

func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	query := `
        SELECT id, email, full_name, phone, password_hash, role, 
               business_id, avatar, is_active, is_owner, email_verified, 
               created_at, updated_at 
        FROM users 
        WHERE id = $1
    `

	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.FullName,
		&user.Phone,
		&user.PasswordHash,
		&user.Role,
		&user.BusinessID,
		&user.Avatar,
		&user.IsActive,
		&user.IsOwner,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
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
	query := `
        INSERT INTO password_resets (
            id, email, token, expires_at, used, created_at, updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT (id) DO UPDATE
        SET 
            used = EXCLUDED.used,
            updated_at = EXCLUDED.updated_at
    `

	updatedAt := reset.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = reset.CreatedAt
	}

	_, err := r.db.ExecContext(ctx, query,
		reset.ID,
		reset.Email,
		reset.Token,
		reset.ExpiresAt,
		reset.Used,
		reset.CreatedAt,
		updatedAt,
	)
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

func (r *AuthRepository) UpdatePassword(ctx context.Context, userID string, hashedPassword string) error {
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
	isActive := (status == "active")

	query := `
        UPDATE users 
        SET is_active = $1, updated_at = $2 
        WHERE id = $3
    `
	_, err := r.db.ExecContext(ctx, query,
		isActive,
		time.Now(),
		userID,
	)

	if err != nil {
		return fmt.Errorf("failed to update status for user %s: %w", userID, err)
	}
	return nil
}
