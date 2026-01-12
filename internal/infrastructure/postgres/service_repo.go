// File: internal/infrastructure/postgres/service_repo.go
package postgres

import (
	"context"
	"fmt"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ServiceRepository struct {
	db *sqlx.DB
}

func NewServiceRepository(db *sqlx.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) Create(ctx context.Context, s *domain.Service) error {
	query := `
        INSERT INTO services (
            id, business_id, name, description, duration_minutes, price,
            is_active, created_at, updated_at
        ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `
	_, err := r.db.ExecContext(
		ctx, query,
		s.ID, s.BusinessID, s.Name, s.Description, s.DurationMinutes, s.Price,
		s.IsActive, s.CreatedAt, s.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert service: %w", err)
	}
	return nil
}

func (r *ServiceRepository) GetByID(ctx context.Context, id, businessID uuid.UUID) (*domain.Service, error) {
	query := `
        SELECT id, business_id, name, description, duration_minutes, price,
               is_active, created_at, updated_at
        FROM services
        WHERE id = $1 AND business_id = $2 AND is_active = true
    `
	var svc domain.Service
	err := r.db.GetContext(ctx, &svc, query, id, businessID)
	if err != nil {
		if isNoRowsError(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get service: %w", err)
	}
	return &svc, nil
}

func (r *ServiceRepository) ListByBusiness(ctx context.Context, businessID uuid.UUID) ([]*domain.Service, error) {
	query := `
        SELECT id, business_id, name, description, duration_minutes, price,
               is_active, created_at, updated_at
        FROM services
        WHERE business_id = $1 AND is_active = true
        ORDER BY created_at DESC
    `
	var list []*domain.Service
	if err := r.db.SelectContext(ctx, &list, query, businessID); err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	return list, nil
}

func (r *ServiceRepository) Update(ctx context.Context, s *domain.Service) error {
	query := `
        UPDATE services
        SET name = $1, description = $2, duration_minutes = $3, price = $4,
            updated_at = $5
        WHERE id = $6 AND business_id = $7
    `
	result, err := r.db.ExecContext(
		ctx, query,
		s.Name, s.Description, s.DurationMinutes, s.Price, s.UpdatedAt,
		s.ID, s.BusinessID,
	)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("service not found")
	}
	return nil
}

func (r *ServiceRepository) Deactivate(ctx context.Context, id, businessID uuid.UUID) error {
	query := `
        UPDATE services
        SET is_active = false, updated_at = NOW()
        WHERE id = $1 AND business_id = $2
    `
	result, err := r.db.ExecContext(ctx, query, id, businessID)
	if err != nil {
		return fmt.Errorf("failed to deactivate service: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("service not found")
	}
	return nil
}

func (r *ServiceRepository) AssignServicesToStaff(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	serviceIDs []uuid.UUID,
) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // nolint:errcheck

	query := `
        INSERT INTO staff_services (
            staff_id, business_id, service_id, created_at
        ) VALUES ($1,$2,$3,NOW())
        ON CONFLICT (staff_id, business_id, service_id) DO NOTHING
    `
	for _, sid := range serviceIDs {
		if _, err := tx.ExecContext(ctx, query, staffID, businessID, sid); err != nil {
			return fmt.Errorf("failed to assign service %s to staff %s: %w", sid, staffID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit assign services tx: %w", err)
	}
	return nil
}

func (r *ServiceRepository) GetStaffServices(
	ctx context.Context,
	businessID, staffID uuid.UUID,
) ([]*domain.Service, error) {
	query := `
        SELECT s.id, s.business_id, s.name, s.description, s.duration_minutes, s.price,
               s.is_active, s.created_at, s.updated_at
        FROM services s
        JOIN staff_services ss
          ON ss.service_id = s.id
        WHERE ss.staff_id = $1 AND ss.business_id = $2 AND s.is_active = true
        ORDER BY s.name ASC
    `
	var list []*domain.Service
	if err := r.db.SelectContext(ctx, &list, query, staffID, businessID); err != nil {
		return nil, fmt.Errorf("failed to get staff services: %w", err)
	}
	return list, nil
}

func (r *ServiceRepository) RemoveServiceFromStaff(
	ctx context.Context,
	businessID, staffID, serviceID uuid.UUID,
) error {
	query := `
        DELETE FROM staff_services
        WHERE staff_id = $1 AND business_id = $2 AND service_id = $3
    `
	result, err := r.db.ExecContext(ctx, query, staffID, businessID, serviceID)
	if err != nil {
		return fmt.Errorf("failed to remove service from staff: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("staff-service relation not found")
	}
	return nil
}

func isNoRowsError(err error) bool {
	type causer interface{ Error() string }
	if err == nil {
		return false
	}
	return err.Error() == "sql: no rows in result set"
}
