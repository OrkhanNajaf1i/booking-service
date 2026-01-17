// File: internal/infrastructure/postgres/customer_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/customer"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

// ============================================
// HELPER FUNCTIONS - TYPE CONVERSIONS
// ============================================

// nullStringToString - NULL pointer-u empty string-ə çevir
func nullStringToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// stringToNullString - String-i *string pointer-ə çevir (NULL ola bilərsə)
func stringToNullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// stringToCustomerStatus - String-i CustomerStatus-a çevir (type-safe)
func stringToCustomerStatus(s string) customer.CustomerStatus {
	switch s {
	case "active":
		return customer.StatusActive
	case "inactive":
		return customer.StatusInactive
	case "blocked":
		return customer.StatusBlocked
	default:
		return customer.StatusActive // default fallback
	}
}

// customerStatusToString - CustomerStatus-u string-ə çevir
func customerStatusToString(s customer.CustomerStatus) string {
	return string(s)
}

// ============================================
// CUSTOMERROW STRUCT - DATABASE ROW MAPPING
// ============================================

// CustomerRow - Database row mappi (internal use only)
// Used by sqlx for automatic struct scanning with db tags
type CustomerRow struct {
	ID            uuid.UUID  `db:"id"`
	BusinessID    uuid.UUID  `db:"business_id"`
	UserID        *uuid.UUID `db:"user_id"` // nullable
	FullName      string     `db:"full_name"`
	Email         string     `db:"email"`
	Phone         string     `db:"phone"`
	Notes         *string    `db:"notes"`  // nullable - convert to string
	Status        string     `db:"status"` // raw string - convert to CustomerStatus
	TotalBookings int        `db:"total_bookings"`
	LastBookingAt *time.Time `db:"last_booking_at"` // nullable
	CreatedAt     time.Time  `db:"created_at"`
	UpdatedAt     time.Time  `db:"updated_at"`
}

// toDomain - Database row-u Domain entityyə çevir (TYPE CONVERSIONS)
func (cr *CustomerRow) toDomain() *customer.Customer {
	return &customer.Customer{
		ID:            cr.ID,
		BusinessID:    cr.BusinessID,
		UserID:        cr.UserID,
		FullName:      cr.FullName,
		Email:         cr.Email,
		Phone:         cr.Phone,
		Notes:         nullStringToString(cr.Notes),      // ✅ *string → string
		Status:        stringToCustomerStatus(cr.Status), // ✅ string → CustomerStatus
		TotalBookings: cr.TotalBookings,
		LastBookingAt: cr.LastBookingAt,
		CreatedAt:     cr.CreatedAt,
		UpdatedAt:     cr.UpdatedAt,
	}
}

// ============================================
// CUSTOMERREPOSITORY - MAIN STRUCT
// ============================================

// CustomerRepository - PostgreSQL adapter (sqlx version)
type CustomerRepository struct {
	db     *sqlx.DB
	logger logger.Logger
}

// NewCustomerRepository - Repository instance yaratır
func NewCustomerRepository(db *sqlx.DB, logger logger.Logger) customer.Repository {
	return &CustomerRepository{
		db:     db,
		logger: logger,
	}
}

// ============================================
// CREATE
// ============================================

// Create - Yeni müştəri əlavə etmə
func (r *CustomerRepository) Create(ctx context.Context, cust *customer.Customer) error {
	query := `
		INSERT INTO customers (
			id, business_id, user_id, full_name, email, phone, 
			notes, status, total_bookings, last_booking_at, created_at, updated_at
		) VALUES (
			:id, :business_id, :user_id, :full_name, :email, :phone, 
			:notes, :status, :total_bookings, :last_booking_at, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":              cust.ID,
		"business_id":     cust.BusinessID,
		"user_id":         cust.UserID,
		"full_name":       cust.FullName,
		"email":           cust.Email,
		"phone":           cust.Phone,
		"notes":           stringToNullString(cust.Notes),
		"status":          customerStatusToString(cust.Status),
		"total_bookings":  cust.TotalBookings,
		"last_booking_at": cust.LastBookingAt,
		"created_at":      cust.CreatedAt,
		"updated_at":      cust.UpdatedAt,
	})

	if err != nil {
		// Email uniqueness violation (PostgreSQL constraint)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			r.logger.Warn("Create: Email already exists",
				logger.Field{Key: "business_id", Value: cust.BusinessID.String()},
				logger.Field{Key: "email", Value: cust.Email},
			)
			return customer.ErrEmailAlreadyExists
		}

		r.logger.Error("Create: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "business_id", Value: cust.BusinessID.String()},
			logger.Field{Key: "email", Value: cust.Email},
		)
		return fmt.Errorf("failed to create customer: %w", err)
	}

	r.logger.Info("Create: Customer created successfully",
		logger.Field{Key: "customer_id", Value: cust.ID.String()},
		logger.Field{Key: "email", Value: cust.Email},
	)

	return nil
}

func (r *CustomerRepository) GetByID(ctx context.Context, businessID, id uuid.UUID) (*customer.Customer, error) {
	query := `
		SELECT 
			id, business_id, user_id, full_name, email, phone, 
			notes, status, total_bookings, last_booking_at, created_at, updated_at
		FROM customers
		WHERE id = $1 AND business_id = $2 AND status != 'inactive'
	`

	var row CustomerRow
	err := r.db.GetContext(ctx, &row, query, id, businessID)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("GetByID: Customer not found",
				logger.Field{Key: "customer_id", Value: id.String()},
				logger.Field{Key: "business_id", Value: businessID.String()},
			)
			return nil, customer.ErrCustomerNotFound
		}

		r.logger.Error("GetByID: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: id.String()},
		)
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	r.logger.Info("GetByID: Customer retrieved successfully",
		logger.Field{Key: "customer_id", Value: id.String()},
		logger.Field{Key: "email", Value: row.Email},
	)

	return row.toDomain(), nil
}

func (r *CustomerRepository) GetByEmail(ctx context.Context, businessID uuid.UUID, email string) (*customer.Customer, error) {
	query := `
		SELECT 
			id, business_id, user_id, full_name, email, phone, 
			notes, status, total_bookings, last_booking_at, created_at, updated_at
		FROM customers
		WHERE email = $1 AND business_id = $2
	`

	var row CustomerRow
	err := r.db.GetContext(ctx, &row, query, email, businessID)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("GetByEmail: Customer not found",
				logger.Field{Key: "email", Value: email},
				logger.Field{Key: "business_id", Value: businessID.String()},
			)
			return nil, nil
		}

		r.logger.Error("GetByEmail: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "email", Value: email},
		)
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return row.toDomain(), nil
}

func (r *CustomerRepository) GetByUserID(ctx context.Context, businessID, userID uuid.UUID) (*customer.Customer, error) {
	query := `
		SELECT 
			id, business_id, user_id, full_name, email, phone, 
			notes, status, total_bookings, last_booking_at, created_at, updated_at
		FROM customers
		WHERE user_id = $1 AND business_id = $2 AND status != 'inactive'
	`

	var row CustomerRow
	err := r.db.GetContext(ctx, &row, query, userID, businessID)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Warn("GetByUserID: Customer not found",
				logger.Field{Key: "user_id", Value: userID.String()},
				logger.Field{Key: "business_id", Value: businessID.String()},
			)
			return nil, customer.ErrCustomerNotFound
		}

		r.logger.Error("GetByUserID: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "user_id", Value: userID.String()},
		)
		return nil, fmt.Errorf("failed to get customer by user_id: %w", err)
	}

	return row.toDomain(), nil
}

func (r *CustomerRepository) List(ctx context.Context, businessID uuid.UUID, limit, offset int) ([]*customer.Customer, error) {
	query := `
		SELECT 
			id, business_id, user_id, full_name, email, phone, 
			notes, status, total_bookings, last_booking_at, created_at, updated_at
		FROM customers
		WHERE business_id = $1 AND status != 'inactive'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	var rows []CustomerRow
	err := r.db.SelectContext(ctx, &rows, query, businessID, limit, offset)

	if err != nil {
		r.logger.Error("List: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	customers := make([]*customer.Customer, 0, len(rows))
	for i := range rows {
		customers = append(customers, rows[i].toDomain())
	}

	r.logger.Info("List: Retrieved successfully",
		logger.Field{Key: "business_id", Value: businessID.String()},
		logger.Field{Key: "count", Value: len(customers)},
	)

	return customers, nil
}

func (r *CustomerRepository) Count(ctx context.Context, businessID uuid.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM customers
		WHERE business_id = $1 AND status != 'inactive'
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, businessID)

	if err != nil {
		r.logger.Error("Count: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)
		return 0, fmt.Errorf("failed to count customers: %w", err)
	}

	return count, nil
}

func (r *CustomerRepository) Update(ctx context.Context, cust *customer.Customer) error {
	query := `
		UPDATE customers
		SET 
			full_name = :full_name,
			email = :email,
			phone = :phone,
			notes = :notes,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id AND business_id = :business_id
	`

	result, err := r.db.NamedExecContext(ctx, query, map[string]interface{}{
		"id":          cust.ID,
		"business_id": cust.BusinessID,
		"full_name":   cust.FullName,
		"email":       cust.Email,
		"phone":       cust.Phone,
		"notes":       stringToNullString(cust.Notes),
		"status":      customerStatusToString(cust.Status),
		"updated_at":  cust.UpdatedAt,
	})

	if err != nil {
		// Email uniqueness violation
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			r.logger.Warn("Update: Email already exists",
				logger.Field{Key: "business_id", Value: cust.BusinessID.String()},
				logger.Field{Key: "email", Value: cust.Email},
			)
			return customer.ErrEmailAlreadyExists
		}

		r.logger.Error("Update: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: cust.ID.String()},
		)
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Update: RowsAffected error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Update: Customer not found",
			logger.Field{Key: "customer_id", Value: cust.ID.String()},
			logger.Field{Key: "business_id", Value: cust.BusinessID.String()},
		)
		return customer.ErrCustomerNotFound
	}

	r.logger.Info("Update: Customer updated successfully",
		logger.Field{Key: "customer_id", Value: cust.ID.String()},
		logger.Field{Key: "email", Value: cust.Email},
	)

	return nil
}

func (r *CustomerRepository) Delete(ctx context.Context, businessID, id uuid.UUID) error {
	query := `
		UPDATE customers
		SET status = $1, updated_at = $2
		WHERE id = $3 AND business_id = $4
	`

	result, err := r.db.ExecContext(ctx, query,
		customerStatusToString(customer.StatusInactive),
		time.Now(),
		id,
		businessID,
	)

	if err != nil {
		r.logger.Error("Delete: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: id.String()},
		)
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Delete: RowsAffected error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("Delete: Customer not found",
			logger.Field{Key: "customer_id", Value: id.String()},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)
		return customer.ErrCustomerNotFound
	}

	r.logger.Info("Delete: Customer soft-deleted successfully",
		logger.Field{Key: "customer_id", Value: id.String()},
	)

	return nil
}

func (r *CustomerRepository) IncrementBookingCount(ctx context.Context, customerID uuid.UUID) error {
	query := `
		UPDATE customers
		SET total_bookings = total_bookings + 1
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, customerID)
	if err != nil {
		r.logger.Error("IncrementBookingCount: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: customerID.String()},
		)
		return fmt.Errorf("failed to increment booking count: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("IncrementBookingCount: Customer not found",
			logger.Field{Key: "customer_id", Value: customerID.String()},
		)
		return customer.ErrCustomerNotFound
	}

	return nil
}

func (r *CustomerRepository) UpdateLastBookingTime(ctx context.Context, customerID uuid.UUID, timestamp time.Time) error {
	query := `
		UPDATE customers
		SET last_booking_at = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.ExecContext(ctx, query, timestamp, time.Now(), customerID)
	if err != nil {
		r.logger.Error("UpdateLastBookingTime: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: customerID.String()},
		)
		return fmt.Errorf("failed to update last booking time: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		r.logger.Warn("UpdateLastBookingTime: Customer not found",
			logger.Field{Key: "customer_id", Value: customerID.String()},
		)
		return customer.ErrCustomerNotFound
	}

	return nil
}
