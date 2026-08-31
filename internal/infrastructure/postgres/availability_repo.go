// File: internal/infrastructure/postgres/availability_repo.go
package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/availability"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// AvailabilityRepository – working_hours / schedule_settings / time_off
// cedvelleri ucun Postgres adapteri.
type AvailabilityRepository struct {
	database *sqlx.DB
}

func NewAvailabilityRepository(database *sqlx.DB) *AvailabilityRepository {
	return &AvailabilityRepository{database: database}
}

// ============================================================
// WORKING HOURS
// ============================================================

const workingHoursColumns = `
	id, business_id, staff_id, day_of_week,
	start_time, end_time,
	break_enabled, break_start, break_end,
	is_active, created_at, updated_at`

// UpsertWorkingHours – (business, staff, gun) uzre tekdir; varsa yenilenir.
func (r *AvailabilityRepository) UpsertWorkingHours(ctx context.Context, wh *availability.WorkingHours) error {
	query := `
		INSERT INTO working_hours (
			id, business_id, staff_id, day_of_week,
			start_time, end_time,
			break_enabled, break_start, break_end,
			is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (business_id, staff_id, day_of_week) DO UPDATE SET
			start_time    = EXCLUDED.start_time,
			end_time      = EXCLUDED.end_time,
			break_enabled = EXCLUDED.break_enabled,
			break_start   = EXCLUDED.break_start,
			break_end     = EXCLUDED.break_end,
			is_active     = EXCLUDED.is_active,
			updated_at    = EXCLUDED.updated_at
		RETURNING id`

	var id uuid.UUID
	err := r.database.QueryRowxContext(
		ctx, query,
		wh.ID, wh.BusinessID, wh.StaffID, wh.DayOfWeek,
		wh.StartTime, wh.EndTime,
		wh.BreakEnabled, wh.BreakStart, wh.BreakEnd,
		wh.IsActive, wh.CreatedAt, wh.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: working hours upsert failed: %w", err)
	}

	// Movcud setir yenilendise ID deyisir – cagirana dogru ID qayitsin.
	wh.ID = id
	return nil
}

func (r *AvailabilityRepository) ListWorkingHours(
	ctx context.Context,
	businessID, staffID uuid.UUID,
) ([]*availability.WorkingHours, error) {
	query := `
		SELECT ` + workingHoursColumns + `
		FROM working_hours
		WHERE business_id = $1 AND staff_id = $2
		ORDER BY day_of_week`

	rows := make([]*availability.WorkingHours, 0, 7)
	if err := r.database.SelectContext(ctx, &rows, query, businessID, staffID); err != nil {
		return nil, fmt.Errorf("postgres: list working hours failed: %w", err)
	}
	return rows, nil
}

func (r *AvailabilityRepository) GetWorkingHoursForDay(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	dayOfWeek int,
) (*availability.WorkingHours, error) {
	query := `
		SELECT ` + workingHoursColumns + `
		FROM working_hours
		WHERE business_id = $1 AND staff_id = $2 AND day_of_week = $3`

	var wh availability.WorkingHours
	err := r.database.GetContext(ctx, &wh, query, businessID, staffID, dayOfWeek)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get working hours failed: %w", err)
	}
	return &wh, nil
}

func (r *AvailabilityRepository) DeleteWorkingHours(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	dayOfWeek int,
) error {
	query := `DELETE FROM working_hours WHERE business_id = $1 AND staff_id = $2 AND day_of_week = $3`
	if _, err := r.database.ExecContext(ctx, query, businessID, staffID, dayOfWeek); err != nil {
		return fmt.Errorf("postgres: delete working hours failed: %w", err)
	}
	return nil
}

// ============================================================
// SCHEDULE SETTINGS
// ============================================================

const scheduleSettingsColumns = `
	id, business_id, staff_id, timezone,
	slot_step_mins, default_duration_mins,
	buffer_before_mins, buffer_after_mins,
	min_notice_mins, max_advance_days,
	auto_confirm, allow_reschedule_proposal,
	pending_expires_mins, cancellation_window_mins,
	allow_customer_reschedule, reschedule_window_mins,
	created_at, updated_at`

// GetSettings – staffID nil olanda biznesin default setiri qaytarilir.
// Setir yoxdursa (nil, nil) donur; qerar vermek domain qatinin isidir.
func (r *AvailabilityRepository) GetSettings(
	ctx context.Context,
	businessID uuid.UUID,
	staffID *uuid.UUID,
) (*availability.ScheduleSettings, error) {
	var (
		query    string
		args     []interface{}
		settings availability.ScheduleSettings
	)

	if staffID == nil || *staffID == uuid.Nil {
		query = `SELECT ` + scheduleSettingsColumns + `
			FROM schedule_settings
			WHERE business_id = $1 AND staff_id IS NULL`
		args = []interface{}{businessID}
	} else {
		query = `SELECT ` + scheduleSettingsColumns + `
			FROM schedule_settings
			WHERE business_id = $1 AND staff_id = $2`
		args = []interface{}{businessID, *staffID}
	}

	err := r.database.GetContext(ctx, &settings, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get schedule settings failed: %w", err)
	}
	return &settings, nil
}

// UpsertSettings – iki ayri partial unique index oldugu ucun
// ON CONFLICT hedefi staff_id-nin dolu olub-olmamasina gore secilir.
func (r *AvailabilityRepository) UpsertSettings(
	ctx context.Context,
	settings *availability.ScheduleSettings,
) error {
	const updateClause = `DO UPDATE SET
			timezone                  = EXCLUDED.timezone,
			slot_step_mins            = EXCLUDED.slot_step_mins,
			default_duration_mins     = EXCLUDED.default_duration_mins,
			buffer_before_mins        = EXCLUDED.buffer_before_mins,
			buffer_after_mins         = EXCLUDED.buffer_after_mins,
			min_notice_mins           = EXCLUDED.min_notice_mins,
			max_advance_days          = EXCLUDED.max_advance_days,
			auto_confirm              = EXCLUDED.auto_confirm,
			allow_reschedule_proposal = EXCLUDED.allow_reschedule_proposal,
			pending_expires_mins      = EXCLUDED.pending_expires_mins,
			cancellation_window_mins  = EXCLUDED.cancellation_window_mins,
			allow_customer_reschedule = EXCLUDED.allow_customer_reschedule,
			reschedule_window_mins    = EXCLUDED.reschedule_window_mins,
			updated_at                = EXCLUDED.updated_at
		RETURNING id`

	conflictTarget := `(business_id) WHERE staff_id IS NULL`
	if settings.StaffID != nil && *settings.StaffID != uuid.Nil {
		conflictTarget = `(business_id, staff_id) WHERE staff_id IS NOT NULL`
	}

	query := `
		INSERT INTO schedule_settings (
			id, business_id, staff_id, timezone,
			slot_step_mins, default_duration_mins,
			buffer_before_mins, buffer_after_mins,
			min_notice_mins, max_advance_days,
			auto_confirm, allow_reschedule_proposal,
			pending_expires_mins, cancellation_window_mins,
			allow_customer_reschedule, reschedule_window_mins,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
		ON CONFLICT ` + conflictTarget + ` ` + updateClause

	var id uuid.UUID
	err := r.database.QueryRowxContext(
		ctx, query,
		settings.ID, settings.BusinessID, settings.StaffID, settings.Timezone,
		settings.SlotStepMins, settings.DefaultDurationMins,
		settings.BufferBeforeMins, settings.BufferAfterMins,
		settings.MinNoticeMins, settings.MaxAdvanceDays,
		settings.AutoConfirm, settings.AllowRescheduleProposal,
		settings.PendingExpiresMins, settings.CancellationWindowMins,
		settings.AllowCustomerReschedule, settings.RescheduleWindowMins,
		settings.CreatedAt, settings.UpdatedAt,
	).Scan(&id)
	if err != nil {
		return fmt.Errorf("postgres: upsert schedule settings failed: %w", err)
	}

	settings.ID = id
	return nil
}

// ============================================================
// TIME OFF
// ============================================================

func (r *AvailabilityRepository) CreateTimeOff(ctx context.Context, timeOff *availability.TimeOff) error {
	query := `
		INSERT INTO time_off (id, business_id, staff_id, start_at, end_at, reason, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.database.ExecContext(
		ctx, query,
		timeOff.ID, timeOff.BusinessID, timeOff.StaffID,
		timeOff.StartAt, timeOff.EndAt, timeOff.Reason,
		timeOff.CreatedAt, timeOff.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres: create time off failed: %w", err)
	}
	return nil
}

// ListTimeOff – araliqla kesisen butun bloklari qaytarir
// (tam icinde olmasi teleb olunmur).
func (r *AvailabilityRepository) ListTimeOff(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	from, to time.Time,
) ([]*availability.TimeOff, error) {
	query := `
		SELECT id, business_id, staff_id, start_at, end_at, reason, created_at, updated_at
		FROM time_off
		WHERE business_id = $1
		  AND staff_id    = $2
		  AND start_at    < $4
		  AND end_at      > $3
		ORDER BY start_at`

	rows := make([]*availability.TimeOff, 0, 4)
	if err := r.database.SelectContext(ctx, &rows, query, businessID, staffID, from, to); err != nil {
		return nil, fmt.Errorf("postgres: list time off failed: %w", err)
	}
	return rows, nil
}

func (r *AvailabilityRepository) DeleteTimeOff(ctx context.Context, businessID, timeOffID uuid.UUID) error {
	query := `DELETE FROM time_off WHERE business_id = $1 AND id = $2`
	if _, err := r.database.ExecContext(ctx, query, businessID, timeOffID); err != nil {
		return fmt.Errorf("postgres: delete time off failed: %w", err)
	}
	return nil
}

// ============================================================
// BUSY INTERVALS
// ============================================================

// ListBookedIntervals – iscinin araliqla kesisen aktiv randevulari.
// Legv edilmis / tamamlanmis / gelmemis randevular vaxti tutmur.
func (r *AvailabilityRepository) ListBookedIntervals(
	ctx context.Context,
	businessID, staffID uuid.UUID,
	from, to time.Time,
) ([]availability.BusyInterval, error) {
	query := `
		SELECT start_time, end_time
		FROM bookings
		WHERE business_id = $1
		  AND staff_id    = $2
		  AND status IN ('pending', 'confirmed', 'reschedule_proposed')
		  AND start_time  < $4
		  AND end_time    > $3
		ORDER BY start_time`

	rows, err := r.database.QueryxContext(ctx, query, businessID, staffID, from, to)
	if err != nil {
		return nil, fmt.Errorf("postgres: list booked intervals failed: %w", err)
	}
	defer rows.Close()

	intervals := make([]availability.BusyInterval, 0, 8)
	for rows.Next() {
		var start, end time.Time
		if err := rows.Scan(&start, &end); err != nil {
			return nil, fmt.Errorf("postgres: scan booked interval failed: %w", err)
		}
		intervals = append(intervals, availability.BusyInterval{
			Start:  start,
			End:    end,
			Reason: "booking",
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: iterate booked intervals failed: %w", err)
	}

	return intervals, nil
}

// EnsureDefaultSchedule – yeni isciye baslangic is qrafiki qurur.
//
// business.ScheduleProvisioner realizasiyasidir. Idempotentdir: qrafik
// artiq varsa toxunulmur, cunki onboarding tekrarlana biler ve sahibin
// oz qurdugu saatlar ustunden yazilmamalidir.
//
// Default: Bazar ertesi–Cume 09:00–18:00, nahar 13:00–14:00. Hefte sonu
// setri de yaranir, amma sondurulmus — sahib qrafik ekraninda yeddi gunu
// de gorup istediyini acmalidir.
func (r *AvailabilityRepository) EnsureDefaultSchedule(
	ctx context.Context,
	businessID, staffID uuid.UUID,
) error {
	const hoursQuery = `
		INSERT INTO working_hours (
			business_id, staff_id, day_of_week,
			start_time, end_time,
			break_enabled, break_start, break_end,
			is_active
		)
		SELECT $1, $2, d.day_of_week,
		       '09:00', '18:00',
		       TRUE, '13:00', '14:00',
		       d.day_of_week BETWEEN 1 AND 5
		FROM generate_series(0, 6) AS d(day_of_week)
		WHERE NOT EXISTS (
			SELECT 1 FROM working_hours w WHERE w.staff_id = $2
		)`

	if _, err := r.database.ExecContext(ctx, hoursQuery, businessID, staffID); err != nil {
		return fmt.Errorf("postgres: create default working hours failed: %w", err)
	}

	// Ayarlar cedvelinin butun sutunlari default dasiyir.
	const settingsQuery = `
		INSERT INTO schedule_settings (business_id, staff_id)
		SELECT $1, $2
		WHERE NOT EXISTS (
			SELECT 1 FROM schedule_settings s WHERE s.staff_id = $2
		)`

	if _, err := r.database.ExecContext(ctx, settingsQuery, businessID, staffID); err != nil {
		return fmt.Errorf("postgres: create default schedule settings failed: %w", err)
	}

	return nil
}
