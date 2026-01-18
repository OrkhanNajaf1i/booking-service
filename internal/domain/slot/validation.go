// File: internal/domain/slot/validation.go
package slot

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================
// SLOT VALIDATION
// ============================================

// ValidateSlotTimes - Slot start/end time validation
func ValidateSlotTimes(startTime, endTime time.Time, durationMins int) error {
	// Start time < End time
	if startTime.Equal(endTime) || startTime.After(endTime) {
		return NewSlotError("INVALID_TIMES", "Start time must be before end time")
	}

	// Calculate actual duration
	actualDuration := endTime.Sub(startTime)
	actualMinutes := int(actualDuration.Minutes())

	// Duration must match parameter
	if actualMinutes != durationMins {
		return NewSlotError("DURATION_MISMATCH", fmt.Sprintf("expected %d minutes, got %d minutes", durationMins, actualMinutes))
	}

	// Minimum duration: 15 minutes
	if durationMins < 15 {
		return ErrInvalidDuration
	}

	// Maximum duration: 8 hours
	if durationMins > 480 {
		return ErrInvalidDuration
	}

	// Can't create slots in the past
	if startTime.Before(time.Now()) {
		return NewSlotError("PAST_SLOT", "Cannot create slot in the past")
	}

	return nil
}

// ValidateDurationMinutes - Valid durations: 15, 30, 45, 60, 90, 120
func ValidateDurationMinutes(durationMins int) error {
	validDurations := []int{15, 30, 45, 60, 90, 120}

	for _, valid := range validDurations {
		if durationMins == valid {
			return nil
		}
	}

	return NewSlotError("INVALID_DURATION", fmt.Sprintf("valid durations: 15, 30, 45, 60, 90, 120 minutes"))
}

// ValidateWorkingHoursFormat - Format validation: "HH:MM" (e.g., "09:00")
func ValidateWorkingHoursFormat(timeStr string) error {
	if len(timeStr) != 5 {
		return NewSlotError("INVALID_TIME_FORMAT", fmt.Sprintf("expected HH:MM format, got %s", timeStr))
	}

	_, err := time.Parse("15:04", timeStr)
	if err != nil {
		return NewSlotError("INVALID_TIME_FORMAT", fmt.Sprintf("invalid time format: %s", timeStr))
	}

	return nil
}

// ValidateWorkingHours - Business logic validation
func ValidateWorkingHours(startTime, endTime string, dayOfWeek int) error {
	// Day of week: 0-6
	if dayOfWeek < 0 || dayOfWeek > 6 {
		return NewSlotError("INVALID_DAY", fmt.Sprintf("day of week must be 0-6, got %d", dayOfWeek))
	}

	// Validate formats
	if err := ValidateWorkingHoursFormat(startTime); err != nil {
		return err
	}

	if err := ValidateWorkingHoursFormat(endTime); err != nil {
		return err
	}

	// Start < End
	start, _ := time.Parse("15:04", startTime)
	end, _ := time.Parse("15:04", endTime)

	if start.Equal(end) || start.After(end) {
		return NewSlotError("INVALID_HOURS", fmt.Sprintf("start time (%s) must be before end time (%s)", startTime, endTime))
	}

	return nil
}

// ============================================
// REQUEST VALIDATION
// ============================================

// ValidateCreateSlotRequest - CreateSlotRequest validation
func ValidateCreateSlotRequest(req *CreateSlotRequest) error {
	if req == nil {
		return NewSlotError("INVALID_REQUEST", "request cannot be nil")
	}

	// Staff ID required
	if req.StaffID == uuid.Nil {
		return NewSlotError("STAFF_ID_REQUIRED", "staff_id is required")
	}

	// Location ID required
	if req.LocationID == uuid.Nil {
		return NewSlotError("LOCATION_ID_REQUIRED", "location_id is required")
	}

	// Times required
	if req.StartTime.IsZero() {
		return NewSlotError("START_TIME_REQUIRED", "start_time is required")
	}

	if req.EndTime.IsZero() {
		return NewSlotError("END_TIME_REQUIRED", "end_time is required")
	}

	// Duration required
	if req.DurationMins <= 0 {
		return NewSlotError("DURATION_REQUIRED", "duration_mins must be positive")
	}

	// Validate times and duration
	if err := ValidateSlotTimes(req.StartTime, req.EndTime, req.DurationMins); err != nil {
		return err
	}

	// Validate duration is standard
	if err := ValidateDurationMinutes(req.DurationMins); err != nil {
		return err
	}

	return nil
}

// ValidateSetWorkingHoursRequest - SetWorkingHoursRequest validation
func ValidateSetWorkingHoursRequest(req *SetWorkingHoursRequest) error {
	if req == nil {
		return NewSlotError("INVALID_REQUEST", "request cannot be nil")
	}

	// Staff ID required
	if req.StaffID == uuid.Nil {
		return NewSlotError("STAFF_ID_REQUIRED", "staff_id is required")
	}

	// Validate working hours
	if err := ValidateWorkingHours(req.StartTime, req.EndTime, req.DayOfWeek); err != nil {
		return err
	}

	return nil
}

// ============================================
// BUSINESS LOGIC VALIDATION
// ============================================

// ValidateSlotAvailability - Check if slot is available for booking
func ValidateSlotAvailability(slot *Slot) error {
	if slot == nil {
		return ErrSlotNotFound
	}

	if slot.Status != SlotStatusAvailable {
		return ErrSlotUnavailable
	}

	if slot.BookingID != nil {
		return ErrSlotAlreadyBooked
	}

	if slot.IsPast() {
		return NewSlotError("PAST_SLOT", "cannot book past slot")
	}

	return nil
}

// ValidateNoConflicts - Check for time overlap
func ValidateNoConflicts(newSlot *Slot, existingSlots []*Slot) bool {
	for _, existing := range existingSlots {
		// Must be same staff and location
		if existing.StaffID != newSlot.StaffID || existing.LocationID != newSlot.LocationID {
			continue
		}

		// Check time overlap: newStart < existingEnd AND newEnd > existingStart
		if newSlot.StartTime.Before(existing.EndTime) && newSlot.EndTime.After(existing.StartTime) {
			return true // Conflict found
		}
	}

	return false // No conflicts
}

// ValidateSlotForBooking - Comprehensive validation before booking
func ValidateSlotForBooking(slot *Slot) error {
	if slot == nil {
		return ErrSlotNotFound
	}

	// Check availability
	if err := ValidateSlotAvailability(slot); err != nil {
		return err
	}

	// Check if slot is within reasonable booking window (90 days)
	maxBookingDays := 90 * 24 * time.Hour
	if slot.StartTime.After(time.Now().Add(maxBookingDays)) {
		return NewSlotError("SLOT_TOO_FAR", "slot is too far in the future")
	}

	return nil
}

// ============================================
// TIME RANGE VALIDATION
// ============================================

// ValidateTimeRange - Validate start/end date range for queries
func ValidateTimeRange(startDate, endDate *time.Time) error {
	if startDate != nil && endDate != nil {
		if startDate.After(*endDate) {
			return NewSlotError("INVALID_DATE_RANGE", "start_date cannot be after end_date")
		}

		// Max 90 days range
		maxRange := 90 * 24 * time.Hour
		if endDate.Sub(*startDate) > maxRange {
			return NewSlotError("DATE_RANGE_TOO_LARGE", "date range cannot exceed 90 days")
		}
	}

	return nil
}

// ============================================
// PAGINATION VALIDATION
// ============================================

// ValidatePagination - Validate page and page_size, return adjusted limit and offset
func ValidatePagination(page, pageSize int) (int, int, error) {
	// Default values
	if page == 0 {
		page = 1
	}
	if pageSize == 0 {
		pageSize = 20
	}

	// Constraints
	if page < 1 {
		return 0, 0, NewSlotError("INVALID_PAGE", "page must be >= 1")
	}

	if pageSize < 1 || pageSize > 100 {
		return 0, 0, NewSlotError("INVALID_PAGE_SIZE", "page_size must be between 1 and 100")
	}

	// Calculate offset
	offset := (page - 1) * pageSize

	return pageSize, offset, nil
}

// ============================================
// HELPER VALIDATORS
// ============================================

// IsValidSlotStatus - Check if status string is valid
func IsValidSlotStatus(status string) bool {
	s := SlotStatus(status)
	return s.IsValid()
}

// IsValidDayOfWeek - Validate day of week (0-6)
func IsValidDayOfWeek(dayOfWeek int) bool {
	return dayOfWeek >= 0 && dayOfWeek <= 6
}

// IsValidBusinessID - Validate UUID
func IsValidBusinessID(id uuid.UUID) bool {
	return id != uuid.Nil
}

// IsValidStaffID - Validate UUID
func IsValidStaffID(id uuid.UUID) bool {
	return id != uuid.Nil
}

// IsValidLocationID - Validate UUID
func IsValidLocationID(id uuid.UUID) bool {
	return id != uuid.Nil
}

// ParseWorkingHoursTime - Parse "HH:MM" to time.Time
func ParseWorkingHoursTime(timeStr string, baseDate time.Time) (time.Time, error) {
	parsedTime, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}

	result := time.Date(
		baseDate.Year(),
		baseDate.Month(),
		baseDate.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		0,
		0,
		baseDate.Location(),
	)

	return result, nil
}

// CompareTimeStrings - Compare "HH:MM" format strings
func CompareTimeStrings(time1, time2 string) int {
	// Returns: -1 if time1 < time2, 0 if equal, 1 if time1 > time2
	if time1 < time2 {
		return -1
	}
	if time1 > time2 {
		return 1
	}
	return 0
}
