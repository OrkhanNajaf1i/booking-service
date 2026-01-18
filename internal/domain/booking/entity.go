// File: internal/domain/booking/entity.go
package booking

import (
	"time"

	"github.com/google/uuid"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"   // created, waiting for confirmation
	BookingStatusConfirmed BookingStatus = "confirmed" // approved by provider
	BookingStatusCancelled BookingStatus = "cancelled" // cancelled by customer or staff
	BookingStatusCompleted BookingStatus = "completed" // service delivered
)

func (s BookingStatus) IsValid() bool {
	return s == BookingStatusPending ||
		s == BookingStatusConfirmed ||
		s == BookingStatusCancelled ||
		s == BookingStatusCompleted
}

type Booking struct {
	ID         uuid.UUID `db:"id" json:"id"`
	BusinessID uuid.UUID `db:"business_id" json:"business_id"`
	CustomerID uuid.UUID `db:"customer_id" json:"customer_id"`
	StaffID    uuid.UUID `db:"staff_id" json:"staff_id"`
	ServiceID  uuid.UUID `db:"service_id" json:"service_id"`
	SlotID     uuid.UUID `db:"slot_id" json:"slot_id"`

	StartTime time.Time     `db:"start_time" json:"start_time"`
	EndTime   time.Time     `db:"end_time" json:"end_time"`
	Status    BookingStatus `db:"status" json:"status"`
	Notes     string        `db:"notes" json:"notes"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

func NewBooking(
	businessID, customerID, staffID, serviceID, slotID uuid.UUID,
	startTime, endTime time.Time,
	notes string,
) *Booking {
	now := time.Now()
	return &Booking{
		ID:         uuid.New(),
		BusinessID: businessID,
		CustomerID: customerID,
		StaffID:    staffID,
		ServiceID:  serviceID,
		SlotID:     slotID,
		StartTime:  startTime,
		EndTime:    endTime,
		Status:     BookingStatusPending,
		Notes:      notes,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

type CreateBookingRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	StaffID    uuid.UUID `json:"staff_id"`
	ServiceID  uuid.UUID `json:"service_id"`
	SlotID     uuid.UUID `json:"slot_id"` // Artıq SlotID mütləqdir
	Notes      string    `json:"notes"`
	// StartTime/EndTime Slot-dan götürüləcək, amma request-də validation üçün ola bilər
}

type UpdateBookingRequest struct {
	Status BookingStatus `json:"status"`
	Notes  *string       `json:"notes,omitempty"`
}

type BookingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BookingError) Error() string { return e.Message }

func NewBookingError(code, message string) *BookingError {
	return &BookingError{Code: code, Message: message}
}
