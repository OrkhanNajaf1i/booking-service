// File: internal/domain/slot/service.go
package slot

import (
	"context"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// ============================================
// SERVICE IMPLEMENTATION
// ============================================

// SlotService - Slot business logic
type SlotService struct {
	repo   Repository
	logger logger.Logger
}

// NewSlotService - Service instance yaratır
func NewSlotService(repo Repository, logger logger.Logger) Service {
	return &SlotService{
		repo:   repo,
		logger: logger,
	}
}

// ============================================
// CREATE OPERATIONS
// ============================================

// CreateSlot - Yeni slot yarat (validation ilə)
func (s *SlotService) CreateSlot(ctx context.Context, businessID uuid.UUID, req *CreateSlotRequest) (*Slot, error) {
	// Validate request
	if err := ValidateCreateSlotRequest(req); err != nil {
		s.logger.Warn("CreateSlot: Invalid request",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)
		return nil, err
	}

	// Create slot entity
	slot := NewSlot(
		businessID,
		req.StaffID,
		req.LocationID,
		req.StartTime,
		req.EndTime,
		req.DurationMins,
	)
	slot.Notes = req.Notes

	// Save to repository
	if err := s.repo.CreateSlot(ctx, slot); err != nil {
		s.logger.Error("CreateSlot: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)
		return nil, fmt.Errorf("failed to create slot: %w", err)
	}

	s.logger.Info("CreateSlot: Success",
		logger.Field{Key: "slot_id", Value: slot.ID.String()},
		logger.Field{Key: "start_time", Value: slot.StartTime.Format(time.RFC3339)},
	)

	return slot, nil
}

// SetWorkingHours - Staff üçün iş saatları təyin et
func (s *SlotService) SetWorkingHours(
	ctx context.Context,
	businessID uuid.UUID,
	req *SetWorkingHoursRequest,
) (*WorkingHours, error) {
	// Validate request
	if err := ValidateSetWorkingHoursRequest(req); err != nil {
		s.logger.Warn("SetWorkingHours: Invalid request",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	// Create working hours entity
	wh := NewWorkingHours(
		businessID,
		req.StaffID,
		req.DayOfWeek,
		req.StartTime,
		req.EndTime,
	)
	wh.IsActive = req.IsActive

	// Save to repository
	if err := s.repo.CreateWorkingHours(ctx, wh); err != nil {
		s.logger.Error("SetWorkingHours: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("failed to set working hours: %w", err)
	}

	s.logger.Info("SetWorkingHours: Success",
		logger.Field{Key: "staff_id", Value: req.StaffID.String()},
		logger.Field{Key: "day", Value: GetDayOfWeekName(req.DayOfWeek)},
	)

	return wh, nil
}

// ============================================
// READ OPERATIONS
// ============================================

// GetSlot - ID-dən slot tap
func (s *SlotService) GetSlot(ctx context.Context, businessID, slotID uuid.UUID) (*Slot, error) {
	slot, err := s.repo.GetSlotByID(ctx, businessID, slotID)
	if err != nil {
		s.logger.Warn("GetSlot: Not found",
			logger.Field{Key: "slot_id", Value: slotID.String()},
		)
		return nil, err
	}

	return slot, nil
}

// ListSlots - Slot-ları list etmək
func (s *SlotService) ListSlots(
	ctx context.Context,
	businessID uuid.UUID,
	query *ListSlotsQuery,
) ([]*Slot, error) {
	// Default pagination
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	// Validate pagination
	pageSize, offset, err := ValidatePagination(query.Page, query.PageSize)
	if err != nil {
		s.logger.Warn("ListSlots: Invalid pagination",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	// Validate date range
	if err := ValidateTimeRange(query.StartDate, query.EndDate); err != nil {
		s.logger.Warn("ListSlots: Invalid date range",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	// Validate status
	var statusFilter *string
	if query.Status != nil && *query.Status != "" {
		if !IsValidSlotStatus(*query.Status) {
			return nil, NewSlotError("INVALID_STATUS", fmt.Sprintf("invalid status: %s", *query.Status))
		}
		statusFilter = query.Status
	}

	// Query repository
	slots, err := s.repo.ListSlots(
		ctx,
		businessID,
		query.StaffID,
		query.LocationID,
		statusFilter,
		pageSize,
		offset,
	)

	if err != nil {
		s.logger.Error("ListSlots: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("failed to list slots: %w", err)
	}

	s.logger.Info("ListSlots: Success",
		logger.Field{Key: "count", Value: len(slots)},
		logger.Field{Key: "page", Value: query.Page},
	)

	return slots, nil
}

// GetAvailableSlots - Müsait slot-ları tap (booking üçün)
func (s *SlotService) GetAvailableSlots(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	locationID uuid.UUID,
	page int,
	pageSize int,
) ([]*Slot, error) {
	// Validate pagination
	pageSize, offset, err := ValidatePagination(page, pageSize)
	if err != nil {
		return nil, err
	}

	// Get available slots
	slots, err := s.repo.GetAvailableSlots(
		ctx,
		businessID,
		staffID,
		locationID,
		pageSize,
		offset,
	)

	if err != nil {
		s.logger.Error("GetAvailableSlots: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("failed to get available slots: %w", err)
	}

	s.logger.Info("GetAvailableSlots: Success",
		logger.Field{Key: "count", Value: len(slots)},
	)

	return slots, nil
}

// GetStaffWorkingHours - Staff-ın iş saatları
func (s *SlotService) GetStaffWorkingHours(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
) ([]*WorkingHours, error) {
	whs, err := s.repo.GetWorkingHoursByStaff(ctx, businessID, staffID)
	if err != nil {
		s.logger.Error("GetStaffWorkingHours: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("failed to get working hours: %w", err)
	}

	return whs, nil
}

// ============================================
// UPDATE OPERATIONS
// ============================================

// UpdateSlot - Slot yeniləmə
func (s *SlotService) UpdateSlot(ctx context.Context, businessID uuid.UUID, slot *Slot) error {
	if slot == nil {
		return NewSlotError("INVALID_SLOT", "slot is nil")
	}

	// Verify slot belongs to business
	if slot.BusinessID != businessID {
		return NewSlotError("BUSINESS_MISMATCH", "slot does not belong to this business")
	}

	// Update timestamp
	slot.UpdatedAt = time.Now()

	// Save to repository
	if err := s.repo.UpdateSlot(ctx, slot); err != nil {
		s.logger.Error("UpdateSlot: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return fmt.Errorf("failed to update slot: %w", err)
	}

	s.logger.Info("UpdateSlot: Success",
		logger.Field{Key: "slot_id", Value: slot.ID.String()},
	)

	return nil
}

// UpdateWorkingHours - İş saatları yeniləmə
func (s *SlotService) UpdateWorkingHours(
	ctx context.Context,
	businessID uuid.UUID,
	req *SetWorkingHoursRequest,
) (*WorkingHours, error) {
	// Validate request
	if err := ValidateSetWorkingHoursRequest(req); err != nil {
		return nil, err
	}

	// Get existing working hours
	existing, err := s.repo.GetWorkingHoursByDayOfWeek(ctx, businessID, req.StaffID, req.DayOfWeek)
	if err != nil {
		s.logger.Error("UpdateWorkingHours: Get error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, err
	}

	if existing == nil {
		return nil, NewSlotError("NOT_FOUND", "working hours not found")
	}

	// Update fields
	existing.StartTime = req.StartTime
	existing.EndTime = req.EndTime
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()

	// Save
	if err := s.repo.UpdateWorkingHours(ctx, existing); err != nil {
		s.logger.Error("UpdateWorkingHours: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return nil, fmt.Errorf("failed to update working hours: %w", err)
	}

	s.logger.Info("UpdateWorkingHours: Success",
		logger.Field{Key: "staff_id", Value: req.StaffID.String()},
	)

	return existing, nil
}

// ============================================
// DELETE OPERATIONS
// ============================================

// DeleteSlot - Soft delete slot
func (s *SlotService) DeleteSlot(ctx context.Context, businessID, slotID uuid.UUID) error {
	if err := s.repo.DeleteSlot(ctx, businessID, slotID); err != nil {
		s.logger.Error("DeleteSlot: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return fmt.Errorf("failed to delete slot: %w", err)
	}

	s.logger.Info("DeleteSlot: Success",
		logger.Field{Key: "slot_id", Value: slotID.String()},
	)

	return nil
}

// DeleteWorkingHours - İş saatını sil
func (s *SlotService) DeleteWorkingHours(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	dayOfWeek int,
) error {
	if err := s.repo.DeleteWorkingHours(ctx, businessID, staffID, dayOfWeek); err != nil {
		s.logger.Error("DeleteWorkingHours: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return fmt.Errorf("failed to delete working hours: %w", err)
	}

	s.logger.Info("DeleteWorkingHours: Success",
		logger.Field{Key: "staff_id", Value: staffID.String()},
	)

	return nil
}

// ============================================
// BOOKING INTEGRATION
// ============================================

// ReserveSlot - Booking üçün slot reserve et
func (s *SlotService) ReserveSlot(ctx context.Context, slotID uuid.UUID, bookingID uuid.UUID) error {
	// Book slot (ATOMIC in repository)
	if err := s.repo.BookSlot(ctx, slotID, bookingID); err != nil {
		s.logger.Error("ReserveSlot: Database error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "slot_id", Value: slotID.String()},
		)
		return fmt.Errorf("failed to reserve slot: %w", err)
	}

	s.logger.Info("ReserveSlot: Success",
		logger.Field{Key: "slot_id", Value: slotID.String()},
		logger.Field{Key: "booking_id", Value: bookingID.String()},
	)

	return nil
}

// ReleaseSlot - Booking iptal zamanı slot boş et
func (s *SlotService) ReleaseSlot(ctx context.Context, slotID uuid.UUID) error {
	// Unbook slot (ATOMIC in repository)
	if err := s.repo.UnbookSlot(ctx, slotID); err != nil {
		s.logger.Error("ReleaseSlot: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return fmt.Errorf("failed to release slot: %w", err)
	}

	s.logger.Info("ReleaseSlot: Success",
		logger.Field{Key: "slot_id", Value: slotID.String()},
	)

	return nil
}

// ============================================
// SCHEDULING OPERATIONS
// ============================================

// GenerateSlots - Slot-lar generate et (cronjob)
func (s *SlotService) GenerateSlots(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	slotDurationMins int,
	daysAhead int,
) (int, error) {
	// Validate duration
	if err := ValidateDurationMinutes(slotDurationMins); err != nil {
		return 0, err
	}

	// Validate days ahead
	if daysAhead < 1 || daysAhead > 365 {
		return 0, NewSlotError("INVALID_DAYS", "daysAhead must be between 1 and 365")
	}

	// Call repository
	count, err := s.repo.GenerateSlots(ctx, businessID, staffID, slotDurationMins, daysAhead)
	if err != nil {
		s.logger.Error("GenerateSlots: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return 0, fmt.Errorf("failed to generate slots: %w", err)
	}

	s.logger.Info("GenerateSlots: Success",
		logger.Field{Key: "count", Value: count},
		logger.Field{Key: "duration_mins", Value: slotDurationMins},
	)

	return count, nil
}

// CleanupExpiredSlots - Keçmiş slot-ları sil
func (s *SlotService) CleanupExpiredSlots(ctx context.Context, businessID uuid.UUID) (int, error) {
	count, err := s.repo.DeleteExpiredSlots(ctx, businessID)
	if err != nil {
		s.logger.Error("CleanupExpiredSlots: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return 0, fmt.Errorf("failed to cleanup expired slots: %w", err)
	}

	s.logger.Info("CleanupExpiredSlots: Success",
		logger.Field{Key: "deleted_count", Value: count},
	)

	return count, nil
}

// ============================================
// VALIDATION OPERATIONS
// ============================================

// ValidateSlotCreation - Slot yaratmaq validation
func (s *SlotService) ValidateSlotCreation(ctx context.Context, businessID uuid.UUID, req *CreateSlotRequest) error {
	return ValidateCreateSlotRequest(req)
}

// ValidateSlotAvailability - Slot müsait?
func (s *SlotService) ValidateSlotAvailability(ctx context.Context, slotID uuid.UUID) error {
	// Note: This needs businessID to get slot from repo
	return NewSlotError("NOT_IMPLEMENTED", "use GetSlot then validate manually")
}

// CheckConflicts - Slot conflict check
func (s *SlotService) CheckConflicts(
	ctx context.Context,
	businessID uuid.UUID,
	staffID uuid.UUID,
	locationID uuid.UUID,
	startTime time.Time,
	endTime time.Time,
) (bool, error) {
	// Get overlapping slots
	slots, err := s.repo.ListSlots(
		ctx,
		businessID,
		&staffID,
		&locationID,
		nil, // all statuses
		100, // max 100
		0,
	)

	if err != nil {
		s.logger.Error("CheckConflicts: Database error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return false, fmt.Errorf("failed to check conflicts: %w", err)
	}

	// Check for overlaps
	for _, slot := range slots {
		// Time overlap check: newStart < existingEnd AND newEnd > existingStart
		if startTime.Before(slot.EndTime) && endTime.After(slot.StartTime) {
			return true, nil // Conflict found
		}
	}

	return false, nil // No conflicts
}
