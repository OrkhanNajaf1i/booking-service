package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/gofrs/uuid"
)

type BusinessRepo struct {
	db *sql.DB
}

func NewBusinessRepo(db *sql.DB) *BusinessRepo {
	return &BusinessRepo{db: db}
}
func (r *BusinessRepo) CreateBusiness(ctx context.Context, b *business.Business) error {
	query := `
		INSERT INTO businesses (id, name, phone, created_at)
		VALUES ($1, $2, $3, $4)
	`
	_, err := r.db.ExecContext(ctx, query, b.ID, b.Name, b.Phone, b.CreatedAt)
	if err != nil {
		fmt.Errorf("postgress: failed to create business: %w", err)
	}
	return nil
}

func (r *BusinessRepo) GetBusinessByID(ctx context.Context, id uuid.UUID) (*business.Business, error) {
	query := `SELECT id, name, phone, created_at FROM businesses WHERE id = $1`
	row := r.db.QueryRowContext(ctx, query, id)
	var b business.Business
	if err := row.Scan(&b.ID, &b.Name, &b.Phone, &b.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("postgres: failed to get business: %w", err)
	}
	return &b, nil
}
