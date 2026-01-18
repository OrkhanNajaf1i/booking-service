// File: internal/http/handlers/slot_dto.go
package handlers

import (
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/slot"
	"github.com/google/uuid"
)

// ============================================
// REQUEST DTOS
// ============================================

// CreateSlotRequest - Yeni slot yaratmaq üçün
type CreateSlotRequest struct {
	StaffID      uuid.UUID `json:"staff_id"`
	LocationID   uuid.UUID `json:"location_id"`
	StartTime    time.Time `json:"start_time"`
	EndTime      time.Time `json:"end_time"`
	DurationMins int       `json:"duration_mins"`
	Notes        *string   `json:"notes,omitempty"`
}

// SetWorkingHoursRequest - İş saatları təyin etmək
type SetWorkingHoursRequest struct {
	StaffID   uuid.UUID `json:"staff_id"`
	DayOfWeek int       `json:"day_of_week"` // 0-6 (Mon-Sun)
	StartTime string    `json:"start_time"`  // HH:MM format
	EndTime   string    `json:"end_time"`    // HH:MM format
	IsActive  bool      `json:"is_active"`
}

// UpdateSlotRequest - Slot yeniləmək
type UpdateSlotRequest struct {
	Status *string `json:"status,omitempty"` // available, booked, cancelled, unavailable
	Notes  *string `json:"notes,omitempty"`
}

// ListSlotsQuery - Slot-ları filtreləmək
type ListSlotsQuery struct {
	StaffID    *uuid.UUID `json:"staff_id,omitempty"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	Status     *string    `json:"status,omitempty"` // available, booked, cancelled, unavailable
	StartDate  *time.Time `json:"start_date,omitempty"`
	EndDate    *time.Time `json:"end_date,omitempty"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
}

// GetAvailableSlotsQuery - Müsait slot-ları tap
type GetAvailableSlotsQuery struct {
	StaffID    uuid.UUID `json:"staff_id"`
	LocationID uuid.UUID `json:"location_id"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
}

// GenerateSlotsRequest - Slot-lar generate etmək
type GenerateSlotsRequest struct {
	StaffID        uuid.UUID `json:"staff_id"`
	SlotDurationMs int       `json:"slot_duration_mins"`
	DaysAhead      int       `json:"days_ahead"`
}

// ============================================
// RESPONSE DTOS
// ============================================

// SlotResponse - Slot response model
type SlotResponse struct {
	ID           uuid.UUID  `json:"id"`
	BusinessID   uuid.UUID  `json:"business_id"`
	StaffID      uuid.UUID  `json:"staff_id"`
	LocationID   uuid.UUID  `json:"location_id"`
	StartTime    time.Time  `json:"start_time"`
	EndTime      time.Time  `json:"end_time"`
	DurationMins int        `json:"duration_mins"`
	Status       string     `json:"status"` // available, booked, cancelled, unavailable
	BookingID    *uuid.UUID `json:"booking_id,omitempty"`
	Notes        *string    `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// WorkingHoursResponse - İş saatları response
type WorkingHoursResponse struct {
	ID         uuid.UUID `json:"id"`
	BusinessID uuid.UUID `json:"business_id"`
	StaffID    uuid.UUID `json:"staff_id"`
	DayOfWeek  int       `json:"day_of_week"`
	StartTime  string    `json:"start_time"` // HH:MM
	EndTime    string    `json:"end_time"`   // HH:MM
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AvailableSlotsResponse - Müsait slot-lar response
type AvailableSlotsResponse struct {
	Slots      []*SlotResponse `json:"slots"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalCount int             `json:"total_count"`
}

// ListSlotsResponse - Slot-lar list response
type ListSlotsResponse struct {
	Slots      []*SlotResponse `json:"slots"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
	TotalCount int             `json:"total_count"`
}

// GenerateSlotsResponse - Generate slots response
type GenerateSlotsResponse struct {
	GeneratedCount int    `json:"generated_count"`
	Message        string `json:"message"`
}

// ErrorResponse - Error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// FromSlotEntity - Domain entity-dən DTO-ya çevirə
// Input: *slot.Slot entity
// Output: *SlotResponse DTO
func FromSlotEntity(s *slot.Slot) *SlotResponse {
	if s == nil {
		return nil
	}

	return &SlotResponse{
		ID:           s.ID,
		BusinessID:   s.BusinessID,
		StaffID:      s.StaffID,
		LocationID:   s.LocationID,
		StartTime:    s.StartTime,
		EndTime:      s.EndTime,
		DurationMins: s.DurationMins,
		Status:       string(s.Status),
		BookingID:    s.BookingID,
		Notes:        s.Notes,
		CreatedAt:    s.CreatedAt,
		UpdatedAt:    s.UpdatedAt,
	}
}

// FromSlotEntities - Multiple slots domain entity-dən DTO-ya çevirə
// Input: []*slot.Slot entities
// Output: []*SlotResponse DTOs
func FromSlotEntities(slots []*slot.Slot) []*SlotResponse {
	if slots == nil {
		return []*SlotResponse{}
	}

	responses := make([]*SlotResponse, 0, len(slots))
	for _, s := range slots {
		if s != nil {
			responses = append(responses, FromSlotEntity(s))
		}
	}
	return responses
}

// FromWorkingHoursEntity - Domain entity-dən DTO-ya çevirə
// Input: *slot.WorkingHours entity
// Output: *WorkingHoursResponse DTO
func FromWorkingHoursEntity(wh *slot.WorkingHours) *WorkingHoursResponse {
	if wh == nil {
		return nil
	}

	return &WorkingHoursResponse{
		ID:         wh.ID,
		BusinessID: wh.BusinessID,
		StaffID:    wh.StaffID,
		DayOfWeek:  wh.DayOfWeek,
		StartTime:  wh.StartTime,
		EndTime:    wh.EndTime,
		IsActive:   wh.IsActive,
		CreatedAt:  wh.CreatedAt,
		UpdatedAt:  wh.UpdatedAt,
	}
}

// FromWorkingHoursEntities - Multiple working hours domain entity-dən DTO-ya çevirə
// Input: []*slot.WorkingHours entities
// Output: []*WorkingHoursResponse DTOs
func FromWorkingHoursEntities(whs []*slot.WorkingHours) []*WorkingHoursResponse {
	if whs == nil {
		return []*WorkingHoursResponse{}
	}

	responses := make([]*WorkingHoursResponse, 0, len(whs))
	for _, wh := range whs {
		if wh != nil {
			responses = append(responses, FromWorkingHoursEntity(wh))
		}
	}
	return responses
}

// ToSlotRequestEntity - DTO-dan domain request-ə çevirə
// Input: *CreateSlotRequest DTO
// Output: *slot.CreateSlotRequest domain
func (req *CreateSlotRequest) ToEntity() *slot.CreateSlotRequest {
	if req == nil {
		return nil
	}

	return &slot.CreateSlotRequest{
		StaffID:      req.StaffID,
		LocationID:   req.LocationID,
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		DurationMins: req.DurationMins,
		Notes:        req.Notes,
	}
}

// ToWorkingHoursRequestEntity - DTO-dan domain request-ə çevirə
// Input: *SetWorkingHoursRequest DTO
// Output: *slot.SetWorkingHoursRequest domain
func (req *SetWorkingHoursRequest) ToEntity() *slot.SetWorkingHoursRequest {
	if req == nil {
		return nil
	}

	return &slot.SetWorkingHoursRequest{
		StaffID:   req.StaffID,
		DayOfWeek: req.DayOfWeek,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		IsActive:  req.IsActive,
	}
}

// ToListSlotsQuery - DTO-dan domain query-ə çevirə
// Input: *ListSlotsQuery DTO
// Output: *slot.ListSlotsQuery domain
func (q *ListSlotsQuery) ToEntity() *slot.ListSlotsQuery {
	if q == nil {
		return nil
	}

	return &slot.ListSlotsQuery{
		StaffID:    q.StaffID,
		LocationID: q.LocationID,
		Status:     q.Status,
		StartDate:  q.StartDate,
		EndDate:    q.EndDate,
		Page:       q.Page,
		PageSize:   q.PageSize,
	}
}
