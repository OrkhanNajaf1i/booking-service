// File: internal/domain/slot/entity.go
package slot

import (
	"time"

	"github.com/google/uuid"
)

// ============================================
// SLOT STATUS ENUM
// ============================================

// SlotStatus - Slot-un status-u
type SlotStatus string

const (
	SlotStatusAvailable   SlotStatus = "available"   // Mövcuddur
	SlotStatusBooked      SlotStatus = "booked"      // Bron edilib
	SlotStatusBlocked     SlotStatus = "blocked"     // Blok edilib (dərs, break)
	SlotStatusUnavailable SlotStatus = "unavailable" // Deaktiv
)

// IsValid - Status validdir?
func (s SlotStatus) IsValid() bool {
	return s == SlotStatusAvailable ||
		s == SlotStatusBooked ||
		s == SlotStatusBlocked ||
		s == SlotStatusUnavailable
}

// ============================================
// SLOT ENTITY
// ============================================

// Slot - Zaman slotu (15/30/60 dəqiqə)
type Slot struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	BusinessID   uuid.UUID  `db:"business_id" json:"business_id"`         // Multi-tenant
	StaffID      uuid.UUID  `db:"staff_id" json:"staff_id"`               // Hansı işçinin slot-u
	LocationID   uuid.UUID  `db:"location_id" json:"location_id"`         // Harada
	StartTime    time.Time  `db:"start_time" json:"start_time"`           // Başlama vaxtı
	EndTime      time.Time  `db:"end_time" json:"end_time"`               // Bitiş vaxtı
	Status       SlotStatus `db:"status" json:"status"`                   // available/booked/blocked/unavailable
	BookingID    *uuid.UUID `db:"booking_id" json:"booking_id,omitempty"` // Booking-in ID
	DurationMins int        `db:"duration_mins" json:"duration_mins"`     // Slot müddəti
	Notes        *string    `db:"notes" json:"notes,omitempty"`           // Optional notes
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

// NewSlot - Slot instance yaratır
func NewSlot(
	businessID uuid.UUID,
	staffID uuid.UUID,
	locationID uuid.UUID,
	startTime time.Time,
	endTime time.Time,
	durationMins int,
) *Slot {
	now := time.Now()
	return &Slot{
		ID:           uuid.New(),
		BusinessID:   businessID,
		StaffID:      staffID,
		LocationID:   locationID,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       SlotStatusAvailable,
		BookingID:    nil,
		DurationMins: durationMins,
		Notes:        nil,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// ============================================
// WORKING HOURS ENTITY
// ============================================

// WorkingHours - Staff-ın hər gün iş saatları
type WorkingHours struct {
	ID         uuid.UUID `db:"id" json:"id"`
	BusinessID uuid.UUID `db:"business_id" json:"business_id"` // Multi-tenant
	StaffID    uuid.UUID `db:"staff_id" json:"staff_id"`
	DayOfWeek  int       `db:"day_of_week" json:"day_of_week"` // 0=Sunday, 1=Monday, ..., 6=Saturday
	StartTime  string    `db:"start_time" json:"start_time"`   // "09:00" format
	EndTime    string    `db:"end_time" json:"end_time"`       // "17:00" format
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

// NewWorkingHours - WorkingHours instance yaratır
func NewWorkingHours(
	businessID uuid.UUID,
	staffID uuid.UUID,
	dayOfWeek int,
	startTime string,
	endTime string,
) *WorkingHours {
	now := time.Now()
	return &WorkingHours{
		ID:         uuid.New(),
		BusinessID: businessID,
		StaffID:    staffID,
		DayOfWeek:  dayOfWeek,
		StartTime:  startTime,
		EndTime:    endTime,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// ============================================
// REQUEST DTOS
// ============================================

// CreateSlotRequest - Slot yaratmaq üçün
type CreateSlotRequest struct {
	StaffID      uuid.UUID `json:"staff_id"`      // Required
	LocationID   uuid.UUID `json:"location_id"`   // Required
	StartTime    time.Time `json:"start_time"`    // Required: 2026-01-20T14:30:00Z
	EndTime      time.Time `json:"end_time"`      // Required: 2026-01-20T15:00:00Z
	DurationMins int       `json:"duration_mins"` // Required: 30
	Notes        *string   `json:"notes"`         // Optional
}

// SetWorkingHoursRequest - İş saatları təyin etmək üçün
type SetWorkingHoursRequest struct {
	StaffID   uuid.UUID `json:"staff_id"`    // Required
	DayOfWeek int       `json:"day_of_week"` // Required: 0-6
	StartTime string    `json:"start_time"`  // Required: "09:00"
	EndTime   string    `json:"end_time"`    // Required: "17:00"
	IsActive  bool      `json:"is_active"`
}

// ListSlotsQuery - Slot-ları filter etmək
type ListSlotsQuery struct {
	StaffID    *uuid.UUID `query:"staff_id"`    // Optional
	LocationID *uuid.UUID `query:"location_id"` // Optional
	Status     *string    `query:"status"`      // Optional: available/booked
	StartDate  *time.Time `query:"start_date"`  // Optional
	EndDate    *time.Time `query:"end_date"`    // Optional
	Page       int        `query:"page"`
	PageSize   int        `query:"page_size"`
}

// ============================================
// RESPONSE DTOS
// ============================================

// SlotResponse - Slot API response
type SlotResponse struct {
	ID           uuid.UUID  `json:"id"`
	StaffID      uuid.UUID  `json:"staff_id"`
	LocationID   uuid.UUID  `json:"location_id"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	Status       string     `json:"status"`
	BookingID    *uuid.UUID `json:"booking_id,omitempty"`
	DurationMins int        `json:"duration_mins"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// WorkingHoursResponse - WorkingHours API response
type WorkingHoursResponse struct {
	ID        uuid.UUID `json:"id"`
	StaffID   uuid.UUID `json:"staff_id"`
	DayOfWeek int       `json:"day_of_week"`
	StartTime string    `json:"start_time"`
	EndTime   string    `json:"end_time"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SuccessResponse - Uğurlu cavab
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// ErrorResponse - Xəta cavabı
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ============================================
// ERROR TYPES
// ============================================

// SlotError - Custom slot error
type SlotError struct {
	Code    string
	Message string
}

func (e *SlotError) Error() string {
	return e.Message
}

// NewSlotError - SlotError instance yaratır
func NewSlotError(code, message string) *SlotError {
	return &SlotError{
		Code:    code,
		Message: message,
	}
}

// Predefined errors
var (
	ErrSlotNotFound      = NewSlotError("SLOT_NOT_FOUND", "Slot not found")
	ErrSlotAlreadyBooked = NewSlotError("SLOT_ALREADY_BOOKED", "Slot is already booked")
	ErrSlotUnavailable   = NewSlotError("SLOT_UNAVAILABLE", "Slot is unavailable")
	ErrInvalidDuration   = NewSlotError("INVALID_DURATION", "Invalid slot duration")
	ErrInvalidWorkingHrs = NewSlotError("INVALID_WORKING_HOURS", "Invalid working hours")
	ErrConflictingSlots  = NewSlotError("CONFLICTING_SLOTS", "Conflicting slots exist")
)

// ============================================
// CONVERSION METHODS
// ============================================

// ToSlotResponse - Slot-u response-a çevir
func (s *Slot) ToSlotResponse() *SlotResponse {
	if s == nil {
		return nil
	}
	return &SlotResponse{
		ID:           s.ID,
		StaffID:      s.StaffID,
		LocationID:   s.LocationID,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
		Status:       string(s.Status),
		BookingID:    s.BookingID,
		DurationMins: s.DurationMins,
		Notes:        s.Notes,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

// ToWorkingHoursResponse - WorkingHours-u response-a çevir
func (wh *WorkingHours) ToWorkingHoursResponse() *WorkingHoursResponse {
	if wh == nil {
		return nil
	}
	return &WorkingHoursResponse{
		ID:        wh.ID,
		StaffID:   wh.StaffID,
		DayOfWeek: wh.DayOfWeek,
		StartTime: wh.StartTime,
		EndTime:   wh.EndTime,
		IsActive:  wh.IsActive,
		CreatedAt: wh.CreatedAt,
		UpdatedAt: wh.UpdatedAt,
	}
}

// ============================================
// HELPER METHODS
// ============================================

// IsAvailable - Slot müsait?
func (s *Slot) IsAvailable() bool {
	return s.Status == SlotStatusAvailable
}

// IsPast - Keçmiş slot?
func (s *Slot) IsPast() bool {
	return s.StartTime.Before(time.Now())
}

// Duration - Slot müddəti (time.Duration)
func (s *Slot) Duration() time.Duration {
	return s.EndTime.Sub(s.StartTime)
}

// GetDayOfWeekName - Gün adını tap
func GetDayOfWeekName(dayOfWeek int) string {
	days := []string{
		"Sunday",
		"Monday",
		"Tuesday",
		"Wednesday",
		"Thursday",
		"Friday",
		"Saturday",
	}
	if dayOfWeek >= 0 && dayOfWeek < len(days) {
		return days[dayOfWeek]
	}
	return "Unknown"
}
