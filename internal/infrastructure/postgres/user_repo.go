package postgres

import (
	// domainuser "booking-service/internal/domain/user"
	"errors"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/google/uuid"

	"context"
	"database/sql"
	"fmt"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}
func (r *UserRepository) CreateUser(ctx context.Context, user *auth.User) error {
	query := `
        INSERT INTO users (
            id, email, password_hash, role, business_id, is_active, created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8
        )
    `
	_, err := r.db.ExecContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.BusinessID,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create to user: %w", err)
	}
	return nil
}
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
        SELECT id, email, password_hash, role, business_id, is_active, created_at, updated_at
        FROM users
        WHERE email = $1
        LIMIT 1
    `
	var user auth.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.BusinessID,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user: %w", err)
	}
	return &user, nil
}
func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*auth.User, error) {
	query := `
        SELECT id, email, password_hash, role, business_id, is_active, created_at, updated_at
        FROM users
        WHERE id = $1
        LIMIT 1
    `
	var user auth.User
	row := r.db.QueryRowContext(ctx, query, userID)
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Role, &user.BusinessID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}
	return &user, nil
}
