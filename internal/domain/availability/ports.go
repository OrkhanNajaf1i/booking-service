// File: internal/domain/availability/ports.go
package availability

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository – availability qaydalarinin saxlanma qati.
type Repository interface {
	// ---------- WORKING HOURS ----------

	// UpsertWorkingHours – (business, staff, day_of_week) uzre yazir/yenileyir.
	UpsertWorkingHours(ctx context.Context, wh *WorkingHours) error
	ListWorkingHours(ctx context.Context, businessID, staffID uuid.UUID) ([]*WorkingHours, error)
	GetWorkingHoursForDay(ctx context.Context, businessID, staffID uuid.UUID, dayOfWeek int) (*WorkingHours, error)
	DeleteWorkingHours(ctx context.Context, businessID, staffID uuid.UUID, dayOfWeek int) error

	// ---------- SCHEDULE SETTINGS ----------

	// GetSettings – staffID nil olanda biznesin default setirini qaytarir.
	// Setir yoxdursa (nil, nil) qaytarir.
	GetSettings(ctx context.Context, businessID uuid.UUID, staffID *uuid.UUID) (*ScheduleSettings, error)
	UpsertSettings(ctx context.Context, settings *ScheduleSettings) error

	// ---------- TIME OFF ----------

	CreateTimeOff(ctx context.Context, timeOff *TimeOff) error
	ListTimeOff(ctx context.Context, businessID, staffID uuid.UUID, from, to time.Time) ([]*TimeOff, error)
	DeleteTimeOff(ctx context.Context, businessID, timeOffID uuid.UUID) error

	// ---------- BUSY ----------

	// ListBookedIntervals – verilmis araliqda iscinin aktiv booking-leri
	// (pending / confirmed / reschedule_proposed).
	ListBookedIntervals(ctx context.Context, businessID, staffID uuid.UUID, from, to time.Time) ([]BusyInterval, error)
}

// ServiceDurationProvider – xidmet kataloqundan mueddet oxumaq ucun.
// Availability paketinin service kataloquna birbasa asililigi olmasin deye
// bu kicik port ayrilib.
type ServiceDurationProvider interface {
	GetDurationMins(ctx context.Context, businessID, serviceID uuid.UUID) (int, error)
}

// Service – availability use-case-leri.
type Service interface {
	// GetAvailability – tarix araliginda hesablanmis bosluqlar.
	GetAvailability(ctx context.Context, businessID uuid.UUID, query *AvailabilityQuery) (*AvailabilityResult, error)

	// CheckSlot – konkret baslangic vaxti bron ucun etibarlidirmi?
	// Booking yaratmadan evvel cagirilir; ugurlu halda tam interval qaytarir.
	CheckSlot(
		ctx context.Context,
		businessID, staffID uuid.UUID,
		serviceID *uuid.UUID,
		start time.Time,
	) (*TimeSlot, error)

	// ResolveSettings – staff override + biznes default + sistem default zenciri.
	ResolveSettings(ctx context.Context, businessID uuid.UUID, staffID uuid.UUID) (*ScheduleSettings, error)

	// ---------- ADMIN ----------

	SetWorkingHours(ctx context.Context, businessID uuid.UUID, req *SetWorkingHoursRequest) (*WorkingHours, error)
	BulkSetWorkingHours(ctx context.Context, businessID uuid.UUID, req *BulkWorkingHoursRequest) ([]*WorkingHours, error)
	ListWorkingHours(ctx context.Context, businessID, staffID uuid.UUID) ([]*WorkingHours, error)
	DeleteWorkingHours(ctx context.Context, businessID, staffID uuid.UUID, dayOfWeek int) error

	GetSettings(ctx context.Context, businessID uuid.UUID, staffID *uuid.UUID) (*ScheduleSettings, error)
	UpdateSettings(ctx context.Context, businessID uuid.UUID, req *UpdateScheduleSettingsRequest) (*ScheduleSettings, error)

	CreateTimeOff(ctx context.Context, businessID uuid.UUID, req *CreateTimeOffRequest) (*TimeOff, error)
	ListTimeOff(ctx context.Context, businessID, staffID uuid.UUID, from, to time.Time) ([]*TimeOff, error)
	DeleteTimeOff(ctx context.Context, businessID, timeOffID uuid.UUID) error
}
