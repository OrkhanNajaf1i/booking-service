// File: internal/domain/slot/ports.go
package slot

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ============================================
// REPOSITORY INTERFACE
// ============================================

// Repository - Slot data layer (Adapter Port)
type Repository interface {
	// ============== SLOT OPERATIONS ==============

	// CreateSlot - Yeni slot yaratmaq
	CreateSlot(ctx context.Context, slot *Slot) error

	// GetSlotByID - ID-dən slot tap (multi-tenant: businessID check)
	GetSlotByID(ctx context.Context, businessID, slotID uuid.UUID) (*Slot, error)

	// ListSlots - Slot-ları list etmək (filtering ilə)
	ListSlots(
		ctx context.Context,
		businessID uuid.UUID,
		staffID *uuid.UUID,
		locationID *uuid.UUID,
		status *string,
		limit int,
		offset int,
	) ([]*Slot, error)

	// GetAvailableSlots - Müsait slot-ları tap (booking üçün)
	GetAvailableSlots(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		locationID uuid.UUID,
		limit int,
		offset int,
	) ([]*Slot, error)

	// UpdateSlot - Slot yeniləmə
	UpdateSlot(ctx context.Context, slot *Slot) error

	// BookSlot - Slot-u booking ilə bağla (ATOMIC)
	BookSlot(ctx context.Context, slotID uuid.UUID, bookingID uuid.UUID) error

	// UnbookSlot - Slot-u booking-dən ayır (ATOMIC)
	UnbookSlot(ctx context.Context, slotID uuid.UUID) error

	// DeleteSlot - Soft delete slot (status = unavailable)
	DeleteSlot(ctx context.Context, businessID, slotID uuid.UUID) error

	// CountAvailableSlots - Müsait slot sayı
	CountAvailableSlots(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		locationID uuid.UUID,
	) (int, error)

	// ============== WORKING HOURS OPERATIONS ==============

	// CreateWorkingHours - Staff üçün iş saatları təyin et
	CreateWorkingHours(ctx context.Context, wh *WorkingHours) error

	// GetWorkingHoursByStaff - Staff-ın iş saatlarını tap
	GetWorkingHoursByStaff(ctx context.Context, businessID, staffID uuid.UUID) ([]*WorkingHours, error)

	// GetWorkingHoursByDayOfWeek - Konkret gün üçün iş saatları
	GetWorkingHoursByDayOfWeek(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		dayOfWeek int,
	) (*WorkingHours, error)

	// UpdateWorkingHours - İş saatlarını yeniləmə
	UpdateWorkingHours(ctx context.Context, wh *WorkingHours) error

	// DeleteWorkingHours - Soft delete working hours
	DeleteWorkingHours(ctx context.Context, businessID, staffID uuid.UUID, dayOfWeek int) error

	// ============== BULK OPERATIONS ==============

	// GenerateSlots - Staff üçün slot-lar generate et (cronjob)
	GenerateSlots(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		slotDurationMins int,
		daysAhead int,
	) (int, error)

	// DeleteExpiredSlots - Keçmiş slot-ları sil
	DeleteExpiredSlots(ctx context.Context, businessID uuid.UUID) (int, error)
}

// ============================================
// SERVICE INTERFACE
// ============================================

// Service - Slot business logic (Domain Port)
type Service interface {
	// ============== CREATE ==============

	// CreateSlot - Yeni slot yarat
	CreateSlot(ctx context.Context, businessID uuid.UUID, req *CreateSlotRequest) (*Slot, error)

	// SetWorkingHours - İş saatları təyin et
	SetWorkingHours(
		ctx context.Context,
		businessID uuid.UUID,
		req *SetWorkingHoursRequest,
	) (*WorkingHours, error)

	// ============== READ ==============

	// GetSlot - Slot tap
	GetSlot(ctx context.Context, businessID, slotID uuid.UUID) (*Slot, error)

	// ListSlots - Slot-ları list et
	ListSlots(
		ctx context.Context,
		businessID uuid.UUID,
		query *ListSlotsQuery,
	) ([]*Slot, error)

	// GetAvailableSlots - Müsait slot-ları tap
	GetAvailableSlots(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		locationID uuid.UUID,
		page int,
		pageSize int,
	) ([]*Slot, error)

	// GetStaffWorkingHours - Staff iş saatlarını tap
	GetStaffWorkingHours(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
	) ([]*WorkingHours, error)

	// ============== UPDATE ==============

	// UpdateSlot - Slot yeniləmə
	UpdateSlot(ctx context.Context, businessID uuid.UUID, slot *Slot) error

	// UpdateWorkingHours - İş saatları yeniləmə
	UpdateWorkingHours(
		ctx context.Context,
		businessID uuid.UUID,
		req *SetWorkingHoursRequest,
	) (*WorkingHours, error)

	// ============== DELETE ==============

	// DeleteSlot - Slot silmə
	DeleteSlot(ctx context.Context, businessID, slotID uuid.UUID) error

	// DeleteWorkingHours - İş saatını silmə
	DeleteWorkingHours(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		dayOfWeek int,
	) error

	// ============== BOOKING INTEGRATION ==============

	// ReserveSlot - Booking üçün slot reserve et
	ReserveSlot(ctx context.Context, slotID uuid.UUID, bookingID uuid.UUID) error

	// ReleaseSlot - Booking iptal zamanı slot boş et
	ReleaseSlot(ctx context.Context, slotID uuid.UUID) error

	// ============== SCHEDULING ==============

	// GenerateSlots - Slot-lar generate et (cronjob)
	GenerateSlots(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		slotDurationMins int,
		daysAhead int,
	) (int, error)

	// CleanupExpiredSlots - Keçmiş slot-ları sil
	CleanupExpiredSlots(ctx context.Context, businessID uuid.UUID) (int, error)

	// ============== VALIDATION ==============

	// ValidateSlotCreation - Slot yaratmaq validation
	ValidateSlotCreation(ctx context.Context, businessID uuid.UUID, req *CreateSlotRequest) error

	// ValidateSlotAvailability - Slot müsait?
	ValidateSlotAvailability(ctx context.Context, slotID uuid.UUID) error

	// CheckConflicts - Conflict check
	CheckConflicts(
		ctx context.Context,
		businessID uuid.UUID,
		staffID uuid.UUID,
		locationID uuid.UUID,
		startTime time.Time,
		endTime time.Time,
	) (bool, error)
}
