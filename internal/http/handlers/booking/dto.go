// File: internal/http/handlers/booking/dto.go
package booking

import (
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/google/uuid"
)

type CreateBookingHTTPRequest struct {
	CustomerID uuid.UUID `json:"customer_id"`
	StaffID    uuid.UUID `json:"staff_id"`
	ServiceID  uuid.UUID `json:"service_id"`
	SlotID     uuid.UUID `json:"slot_id"`
	Notes      string    `json:"notes"`
}

type UpdateBookingHTTPRequest struct {
	Status string  `json:"status"`
	Notes  *string `json:"notes,omitempty"`
}

type BookingHTTPResponse struct {
	ID         uuid.UUID `json:"id"`
	BusinessID uuid.UUID `json:"business_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	StaffID    uuid.UUID `json:"staff_id"`
	ServiceID  uuid.UUID `json:"service_id"`
	SlotID     uuid.UUID `json:"slot_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (req *CreateBookingHTTPRequest) ToDomain(businessID uuid.UUID) *booking.CreateBookingRequest {
	return &booking.CreateBookingRequest{
		BusinessID: businessID,
		CustomerID: req.CustomerID,
		StaffID:    req.StaffID,
		ServiceID:  req.ServiceID,
		SlotID:     req.SlotID,
		Notes:      req.Notes,
	}
}

func (req *UpdateBookingHTTPRequest) ToDomain() *booking.UpdateBookingRequest {
	var status booking.BookingStatus
	if req.Status != "" {
		status = booking.BookingStatus(req.Status)
	}

	return &booking.UpdateBookingRequest{
		Status: status,
		Notes:  req.Notes,
	}
}

func ToHTTPResponse(b *booking.Booking) *BookingHTTPResponse {
	if b == nil {
		return nil
	}
	return &BookingHTTPResponse{
		ID:         b.ID,
		BusinessID: b.BusinessID,
		CustomerID: b.CustomerID,
		StaffID:    b.StaffID,
		ServiceID:  b.ServiceID,
		SlotID:     b.SlotID,
		StartTime:  b.StartTime,
		EndTime:    b.EndTime,
		Status:     string(b.Status),
		Notes:      b.Notes,
		CreatedAt:  b.CreatedAt,
		UpdatedAt:  b.UpdatedAt,
	}
}

func ToHTTPResponseList(bookings []*booking.Booking) []*BookingHTTPResponse {
	list := make([]*BookingHTTPResponse, len(bookings))
	for i, b := range bookings {
		list[i] = ToHTTPResponse(b)
	}
	return list
}
