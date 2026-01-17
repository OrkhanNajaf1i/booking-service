// File: internal/domain/booking/ports.go
package booking

import (
	"context"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/slot"
	"github.com/google/uuid"
)

// Repository – Booking məlumatlarını saxlamaq üçün
type Repository interface {
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, businessID, bookingID uuid.UUID) (*Booking, error)
	GetByCustomer(ctx context.Context, businessID, customerID uuid.UUID) ([]*Booking, error)
	GetByStaff(ctx context.Context, businessID, staffID uuid.UUID) ([]*Booking, error)
	GetByBusiness(ctx context.Context, businessID uuid.UUID) ([]*Booking, error)
	Update(ctx context.Context, booking *Booking) error
	// Reporting
	CountByStatus(ctx context.Context, businessID uuid.UUID, status BookingStatus) (int, error)
}

// Service – Booking biznes məntiqi
type Service interface {
	CreateBooking(ctx context.Context, request *CreateBookingRequest) (*Booking, error)
	GetBookingByID(ctx context.Context, businessID, bookingID uuid.UUID) (*Booking, error)
	GetCustomerBookings(ctx context.Context, businessID, customerID uuid.UUID) ([]*Booking, error)
	GetStaffBookings(ctx context.Context, businessID, staffID uuid.UUID) ([]*Booking, error)
	GetBusinessBookings(ctx context.Context, businessID uuid.UUID) ([]*Booking, error)
	UpdateBooking(ctx context.Context, businessID, bookingID uuid.UUID, request *UpdateBookingRequest) error
	CancelBooking(ctx context.Context, businessID, bookingID uuid.UUID) error
}

// Xarici Servis İnterfeysləri (External Dependencies)
// Bu interfeyslər Booking Service-də istifadə olunacaq

type SlotService interface {
	GetSlot(ctx context.Context, businessID, slotID uuid.UUID) (*slot.Slot, error) // Generic qaytarır, amma bizə vaxt və status lazımdır
	ReserveSlot(ctx context.Context, slotID uuid.UUID, bookingID uuid.UUID) error
	ReleaseSlot(ctx context.Context, slotID uuid.UUID) error
	ValidateSlotAvailability(ctx context.Context, slotID uuid.UUID) error
}

type CustomerService interface {
	GetCustomer(ctx context.Context, businessID, id uuid.UUID) (interface{}, error)
	IncrementBookingCount(ctx context.Context, customerID uuid.UUID) error
}

type StaffService interface {
	GetStaff(ctx context.Context, staffID, businessID uuid.UUID) (interface{}, error)
}

type ServiceCatalogService interface {
	GetService(ctx context.Context, id, businessID uuid.UUID) (interface{}, error)
}
