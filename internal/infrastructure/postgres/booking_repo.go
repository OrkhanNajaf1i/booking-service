// File: internal/infrastructure/postgres/booking_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type BookingRepository struct {
	db *sqlx.DB
}

func NewBookingRepository(db *sqlx.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create - Yeni booking əlavə edir
func (r *BookingRepository) Create(ctx context.Context, b *booking.Booking) error {
	query := `
		INSERT INTO bookings (
			id, business_id, customer_id, staff_id, service_id, slot_id,
			start_time, end_time, status, notes, created_at, updated_at
		) VALUES (
			:id, :business_id, :customer_id, :staff_id, :service_id, :slot_id,
			:start_time, :end_time, :status, :notes, :created_at, :updated_at
		)
	`
	_, err := r.db.NamedExecContext(ctx, query, b)
	if err != nil {
		return fmt.Errorf("failed to create booking: %w", err)
	}
	return nil
}

// GetByID - ID və BusinessID əsasında tapır
func (r *BookingRepository) GetByID(ctx context.Context, businessID, bookingID uuid.UUID) (*booking.Booking, error) {
	var b booking.Booking
	query := `
		SELECT * FROM bookings 
		WHERE id = $1 AND business_id = $2 AND is_deleted = false
	`
	// is_deleted sütunu varsa (soft delete üçün), yoxdursa silin.

	err := r.db.GetContext(ctx, &b, query, bookingID, businessID)
	if err == sql.ErrNoRows {
		return nil, nil // Not found handled as nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	return &b, nil
}

// GetByCustomer - Müştərinin rezervasiyaları
func (r *BookingRepository) GetByCustomer(ctx context.Context, businessID, customerID uuid.UUID) ([]*booking.Booking, error) {
	var bookings []*booking.Booking
	query := `
		SELECT * FROM bookings 
		WHERE business_id = $1 AND customer_id = $2 
		ORDER BY start_time DESC
	`
	err := r.db.SelectContext(ctx, &bookings, query, businessID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list customer bookings: %w", err)
	}
	return bookings, nil
}

// GetByStaff - İşçinin rezervasiyaları
func (r *BookingRepository) GetByStaff(ctx context.Context, businessID, staffID uuid.UUID) ([]*booking.Booking, error) {
	var bookings []*booking.Booking
	query := `
		SELECT * FROM bookings 
		WHERE business_id = $1 AND staff_id = $2 
		ORDER BY start_time ASC
	`
	err := r.db.SelectContext(ctx, &bookings, query, businessID, staffID)
	if err != nil {
		return nil, fmt.Errorf("failed to list staff bookings: %w", err)
	}
	return bookings, nil
}

// GetByBusiness - Bütün rezervasiyalar (Admin panel üçün)
func (r *BookingRepository) GetByBusiness(ctx context.Context, businessID uuid.UUID) ([]*booking.Booking, error) {
	var bookings []*booking.Booking
	query := `
		SELECT * FROM bookings 
		WHERE business_id = $1 
		ORDER BY created_at DESC
	`
	err := r.db.SelectContext(ctx, &bookings, query, businessID)
	if err != nil {
		return nil, fmt.Errorf("failed to list business bookings: %w", err)
	}
	return bookings, nil
}

// Update - Status və ya qeydləri yeniləyir
func (r *BookingRepository) Update(ctx context.Context, b *booking.Booking) error {
	query := `
		UPDATE bookings 
		SET 
			status = :status,
			notes = :notes,
			start_time = :start_time,
			end_time = :end_time,
			updated_at = :updated_at
		WHERE id = :id AND business_id = :business_id
	`
	result, err := r.db.NamedExecContext(ctx, query, b)
	if err != nil {
		return fmt.Errorf("failed to update booking: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("booking not found or no changes made")
	}

	return nil
}

// CountByStatus - Dashboard statistikası üçün
func (r *BookingRepository) CountByStatus(ctx context.Context, businessID uuid.UUID, status booking.BookingStatus) (int, error) {
	var count int
	query := `
		SELECT COUNT(*) FROM bookings 
		WHERE business_id = $1 AND status = $2
	`
	err := r.db.GetContext(ctx, &count, query, businessID, status)
	if err != nil {
		return 0, fmt.Errorf("failed to count bookings: %w", err)
	}
	return count, nil
}
