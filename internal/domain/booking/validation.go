// File: internal/domain/booking/validation.go
package booking

import (
	"strings"

	"github.com/google/uuid"
)

func (s *BookingService) validateCreateRequest(req *CreateBookingRequest) error {
	if req == nil {
		return NewBookingError("INVALID_REQUEST", "Request cannot be nil")
	}
	if req.BusinessID == uuid.Nil {
		return NewBookingError("BUSINESS_ID_REQUIRED", "Business ID is required")
	}
	if req.CustomerID == uuid.Nil {
		return NewBookingError("CUSTOMER_ID_REQUIRED", "Customer ID is required")
	}
	if req.StaffID == uuid.Nil {
		return NewBookingError("STAFF_ID_REQUIRED", "Staff ID is required")
	}
	if req.ServiceID == uuid.Nil {
		return NewBookingError("SERVICE_ID_REQUIRED", "Service ID is required")
	}
	if req.SlotID == uuid.Nil {
		return NewBookingError("SLOT_ID_REQUIRED", "Slot ID is required")
	}
	if len(req.Notes) > 500 {
		return NewBookingError("NOTES_TOO_LONG", "Notes cannot exceed 500 characters")
	}
	return nil
}

func (s *BookingService) validateBookingData(b *Booking) error {
	if b.StartTime.After(b.EndTime) {
		return NewBookingError("INVALID_TIME", "Start time cannot be after end time")
	}
	if strings.TrimSpace(string(b.Status)) == "" {
		return NewBookingError("STATUS_REQUIRED", "Booking status is required")
	}
	return nil
}
