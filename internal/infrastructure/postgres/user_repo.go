package postgres

import (
	// domainuser "booking-service/internal/domain/user"
	domainuser "booking-service/internal/domain/user"

	"context"
	"database/sql"
	"fmt"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}
func (r *UserRepo) GetUserByPhone(ctx context.Context, phone string) (*domainuser.User, error) {
	query := `
		SELECT id, business_id, name, phone, created_at
		FROM users
		WHERE phone = $1
	`
	row := r.db.QueryRowContext(ctx, query, phone)
	var u domainuser.User
	if err := row.Scan(&u.ID, &u.BusinessID, &u.Name, &u.Phone, &u.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("postgress: failed to get user by phone: %w", err)
	}
	return &u, nil
}
