// File: internal/infrastructure/postgres/slot_repo.go
package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/slot"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ============================================
// SLOT REPOSITORY IMPLEMENTATION
// ============================================

// SlotRepository - PostgreSQL implementation of slot.Repository
type SlotRepository struct {
	database *sqlx.DB
}

// NewSlotRepository - Repository instance yaratır
func NewSlotRepository(database *sqlx.DB) *SlotRepository {
	return &SlotRepository{
		database: database,
	}
}

// ============================================
// SLOT OPERATIONS
// ============================================

// CreateSlot - Yeni slot yaratmaq
func (repository *SlotRepository) CreateSlot(ctx context.Context, s *slot.Slot) error {
	query := `
		INSERT INTO slots (
			id, business_id, staff_id, location_id,
			start_time, end_time, duration_mins,
			status, booking_id, notes,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := repository.database.ExecContext(
		ctx,
		query,
		s.ID,
		s.BusinessID,
		s.StaffID,
		s.LocationID,
		s.StartTime,
		s.EndTime,
		s.DurationMins,
		string(s.Status),
		s.BookingID,
		s.Notes,
		s.CreatedAt,
		s.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres: failed to create slot: %w", err)
	}

	return nil
}

// GetSlotByID - ID-dən slot tap (multi-tenant: businessID check)
func (repository *SlotRepository) GetSlotByID(ctx context.Context, businessID, slotID uuid.UUID) (*slot.Slot, error) {
	query := `
		SELECT
			id, business_id, staff_id, location_id,
			start_time, end_time, duration_mins,
			status, booking_id, notes,
			created_at, updated_at
		FROM slots
		WHERE id = $1 AND business_id = $2 AND deleted_at IS NULL
	`

	var s slot.Slot
	err := repository.database.GetContext(ctx, &s, query, slotID, businessID)

	if err == sql.ErrNoRows {
		return nil, slot.ErrSlotNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get slot by ID: %w", err)
	}

	return &s, nil
}

// ListSlots - Slot-ları list etmək (filtering ilə)
func (repository *SlotRepository) ListSlots(
	ctx context.Context,
	businessID uuid.UUID,
	staffID *uuid.UUID,
	locationID *uuid.UUID,
	status *string,
	limit int,
	offset int,
) ([]*slot.Slot, error) {
	query := `
		SELECT
			id, business_id, staff_id, location_id,
			start_time, end_time, duration_mins,
			status, booking_id, notes,
			created_at, updated_at
		FROM slots
		WHERE business_id = $1 AND deleted_at IS NULL
	`
	args := []interface{}{businessID}
	argIndex := 2

	// Filter by staff_id
	if staffID != nil {
		query += fmt.Sprintf(" AND staff_id = $%d", argIndex)
		args = append(args, staffID)
		argIndex++
	}

	// Filter by location_id
	if locationID != nil {
		query += fmt.Sprintf(" AND location_id = $%d", argIndex)
		args = append(args, locationID)
		argIndex++
	}

	// Filter by status
	if status != nil && *status != "" {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	query += " ORDER BY start_time ASC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	var slots []*slot.Slot
	err := repository.database.SelectContext(ctx, &slots, query, args...)

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to list slots: %w", err)
	}

	return slots, nil
}

// GetAvailableSlots - Müsait slot-ları tap (booking üçün)
func (repository *SlotRepository) GetAvailableSlots(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	locationID uuid.UUID,
	limit int,
	offset int,
) ([]*slot.Slot, error) {
	query := `
		SELECT
			id, business_id, staff_id, location_id,
			start_time, end_time, duration_mins,
			status, booking_id, notes,
			created_at, updated_at
		FROM slots
		WHERE
			business_id = $1
			AND staff_id = $2
			AND location_id = $3
			AND status = 'available'
			AND booking_id IS NULL
			AND start_time > NOW()
			AND deleted_at IS NULL
		ORDER BY start_time ASC
		LIMIT $4 OFFSET $5
	`

	var slots []*slot.Slot
	err := repository.database.SelectContext(
		ctx,
		&slots,
		query,
		businessID,
		staffID,
		locationID,
		limit,
		offset,
	)

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get available slots: %w", err)
	}

	return slots, nil
}

// UpdateSlot - Slot yeniləmə
func (repository *SlotRepository) UpdateSlot(ctx context.Context, s *slot.Slot) error {
	query := `
		UPDATE slots
		SET
			status = $1,
			booking_id = $2,
			notes = $3,
			updated_at = $4
		WHERE id = $5 AND business_id = $6 AND deleted_at IS NULL
	`

	result, err := repository.database.ExecContext(
		ctx,
		query,
		string(s.Status),
		s.BookingID,
		s.Notes,
		time.Now(),
		s.ID,
		s.BusinessID,
	)

	if err != nil {
		return fmt.Errorf("postgres: failed to update slot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return slot.ErrSlotNotFound
	}

	return nil
}

// BookSlot - Slot-u booking ilə bağla (ATOMIC)
func (repository *SlotRepository) BookSlot(ctx context.Context, slotID uuid.UUID, bookingID uuid.UUID) error {
	query := `
		UPDATE slots
		SET
			status = 'booked',
			booking_id = $1,
			updated_at = NOW()
		WHERE
			id = $2
			AND status = 'available'
			AND booking_id IS NULL
			AND deleted_at IS NULL
	`

	result, err := repository.database.ExecContext(
		ctx,
		query,
		bookingID,
		slotID,
	)

	if err != nil {
		return fmt.Errorf("postgres: failed to book slot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return slot.ErrSlotUnavailable
	}

	return nil
}

// UnbookSlot - Slot-u booking-dən ayır (ATOMIC)
func (repository *SlotRepository) UnbookSlot(ctx context.Context, slotID uuid.UUID) error {
	query := `
		UPDATE slots
		SET
			status = 'available',
			booking_id = NULL,
			updated_at = NOW()
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := repository.database.ExecContext(ctx, query, slotID)

	if err != nil {
		return fmt.Errorf("postgres: failed to unbook slot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return slot.ErrSlotNotFound
	}

	return nil
}

// DeleteSlot - Soft delete slot (status = unavailable)
func (repository *SlotRepository) DeleteSlot(ctx context.Context, businessID, slotID uuid.UUID) error {
	query := `
		UPDATE slots
		SET
			status = 'unavailable',
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE id = $1 AND business_id = $2 AND deleted_at IS NULL
	`

	result, err := repository.database.ExecContext(ctx, query, slotID, businessID)

	if err != nil {
		return fmt.Errorf("postgres: failed to delete slot: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return slot.ErrSlotNotFound
	}

	return nil
}

// CountAvailableSlots - Müsait slot sayı
func (repository *SlotRepository) CountAvailableSlots(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	locationID uuid.UUID,
) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM slots
		WHERE
			business_id = $1
			AND staff_id = $2
			AND location_id = $3
			AND status = 'available'
			AND booking_id IS NULL
			AND start_time > NOW()
			AND deleted_at IS NULL
	`

	var count int
	err := repository.database.GetContext(ctx, &count, query, businessID, staffID, locationID)

	if err != nil {
		return 0, fmt.Errorf("postgres: failed to count available slots: %w", err)
	}

	return count, nil
}

// ============================================
// WORKING HOURS OPERATIONS
// ============================================

// CreateWorkingHours - Staff üçün iş saatları təyin et
func (repository *SlotRepository) CreateWorkingHours(ctx context.Context, wh *slot.WorkingHours) error {
	query := `
		INSERT INTO working_hours (
			id, business_id, staff_id, day_of_week,
			start_time, end_time, is_active,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := repository.database.ExecContext(
		ctx,
		query,
		wh.ID,
		wh.BusinessID,
		wh.StaffID,
		wh.DayOfWeek,
		wh.StartTime,
		wh.EndTime,
		wh.IsActive,
		wh.CreatedAt,
		wh.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("postgres: failed to create working hours: %w", err)
	}

	return nil
}

// GetWorkingHoursByStaff - Staff-ın iş saatlarını tap
func (repository *SlotRepository) GetWorkingHoursByStaff(ctx context.Context, businessID, staffID uuid.UUID) ([]*slot.WorkingHours, error) {
	query := `
		SELECT
			id, business_id, staff_id, day_of_week,
			start_time, end_time, is_active,
			created_at, updated_at
		FROM working_hours
		WHERE business_id = $1 AND staff_id = $2 AND deleted_at IS NULL
		ORDER BY day_of_week ASC
	`

	var whs []*slot.WorkingHours
	err := repository.database.SelectContext(ctx, &whs, query, businessID, staffID)

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get working hours by staff: %w", err)
	}

	return whs, nil
}

// GetWorkingHoursByDayOfWeek - Konkret gün üçün iş saatları
func (repository *SlotRepository) GetWorkingHoursByDayOfWeek(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	dayOfWeek int,
) (*slot.WorkingHours, error) {
	query := `
		SELECT
			id, business_id, staff_id, day_of_week,
			start_time, end_time, is_active,
			created_at, updated_at
		FROM working_hours
		WHERE
			business_id = $1
			AND staff_id = $2
			AND day_of_week = $3
			AND deleted_at IS NULL
	`

	var wh slot.WorkingHours
	err := repository.database.GetContext(ctx, &wh, query, businessID, staffID, dayOfWeek)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	if err != nil {
		return nil, fmt.Errorf("postgres: failed to get working hours by day of week: %w", err)
	}

	return &wh, nil
}

// UpdateWorkingHours - İş saatlarını yeniləmə
func (repository *SlotRepository) UpdateWorkingHours(ctx context.Context, wh *slot.WorkingHours) error {
	query := `
		UPDATE working_hours
		SET
			start_time = $1,
			end_time = $2,
			is_active = $3,
			updated_at = NOW()
		WHERE
			id = $4
			AND business_id = $5
			AND deleted_at IS NULL
	`

	result, err := repository.database.ExecContext(
		ctx,
		query,
		wh.StartTime,
		wh.EndTime,
		wh.IsActive,
		wh.ID,
		wh.BusinessID,
	)

	if err != nil {
		return fmt.Errorf("postgres: failed to update working hours: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return slot.NewSlotError("NOT_FOUND", "working hours not found")
	}

	return nil
}

// DeleteWorkingHours - Soft delete working hours
func (repository *SlotRepository) DeleteWorkingHours(ctx context.Context, businessID, staffID uuid.UUID, dayOfWeek int) error {
	query := `
		UPDATE working_hours
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			business_id = $1
			AND staff_id = $2
			AND day_of_week = $3
			AND deleted_at IS NULL
	`

	result, err := repository.database.ExecContext(ctx, query, businessID, staffID, dayOfWeek)

	if err != nil {
		return fmt.Errorf("postgres: failed to delete working hours: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return slot.NewSlotError("NOT_FOUND", "working hours not found")
	}

	return nil
}

// ============================================
// BULK OPERATIONS
// ============================================

// GenerateSlots - Staff üçün slot-lar generate et (cronjob)
func (repository *SlotRepository) GenerateSlots(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	slotDurationMins int,
	daysAhead int,
) (int, error) {
	// Get working hours for this staff
	whs, err := repository.GetWorkingHoursByStaff(ctx, businessID, staffID)
	if err != nil {
		return 0, err
	}

	if len(whs) == 0 {
		return 0, slot.NewSlotError("NO_WORKING_HOURS", "staff has no working hours configured")
	}

	var totalGenerated int

	// For each day ahead
	for day := 1; day <= daysAhead; day++ {
		targetDate := time.Now().AddDate(0, 0, day)
		dayOfWeek := int(targetDate.Weekday())

		// Find working hours for this day
		var wh *slot.WorkingHours
		for _, w := range whs {
			if w.DayOfWeek == dayOfWeek {
				wh = w
				break
			}
		}

		if wh == nil || !wh.IsActive {
			continue // No working hours for this day
		}

		// Generate slots from start_time to end_time
		startTime, _ := time.Parse("15:04", wh.StartTime)
		endTime, _ := time.Parse("15:04", wh.EndTime)

		// Set dates
		slotStart := time.Date(
			targetDate.Year(),
			targetDate.Month(),
			targetDate.Day(),
			startTime.Hour(),
			startTime.Minute(),
			0, 0,
			time.Local,
		)
		slotEnd := time.Date(
			targetDate.Year(),
			targetDate.Month(),
			targetDate.Day(),
			endTime.Hour(),
			endTime.Minute(),
			0, 0,
			time.Local,
		)

		// Create slots
		for current := slotStart; current.Add(time.Duration(slotDurationMins)*time.Minute).Before(slotEnd) || current.Add(time.Duration(slotDurationMins)*time.Minute).Equal(slotEnd); current = current.Add(time.Duration(slotDurationMins) * time.Minute) {
			newSlot := &slot.Slot{
				ID:           uuid.New(),
				BusinessID:   businessID,
				StaffID:      staffID,
				LocationID:   uuid.Nil, // Should be set separately
				StartTime:    current,
				EndTime:      current.Add(time.Duration(slotDurationMins) * time.Minute),
				DurationMins: slotDurationMins,
				Status:       slot.SlotStatusAvailable,
				BookingID:    nil,
				Notes:        nil,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			if err := repository.CreateSlot(ctx, newSlot); err != nil {
				continue
			}

			totalGenerated++
		}
	}

	return totalGenerated, nil
}

// DeleteExpiredSlots - Keçmiş slot-ları sil
func (repository *SlotRepository) DeleteExpiredSlots(ctx context.Context, businessID uuid.UUID) (int, error) {
	query := `
		UPDATE slots
		SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE
			business_id = $1
			AND start_time < NOW()
			AND status != 'booked'
			AND deleted_at IS NULL
	`

	result, err := repository.database.ExecContext(ctx, query, businessID)

	if err != nil {
		return 0, fmt.Errorf("postgres: failed to delete expired slots: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("postgres: failed to get rows affected: %w", err)
	}

	return int(rowsAffected), nil
}
