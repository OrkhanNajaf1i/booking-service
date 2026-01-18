// File: internal/domain/booking/service.go
package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Slot interface-i üçün sadələşdirilmiş struktur (adapter-də map olunacaq)
type SlotData struct {
	ID        uuid.UUID
	StartTime time.Time
	EndTime   time.Time
	IsBooked  bool
}

type BookingService struct {
	repo            Repository
	slotService     SlotService
	customerService CustomerService
	staffService    StaffService
	catalogService  ServiceCatalogService
}

func NewService(
	repo Repository,
	slotService SlotService,
	customerService CustomerService,
	staffService StaffService,
	catalogService ServiceCatalogService,
) *BookingService {
	return &BookingService{
		repo:            repo,
		slotService:     slotService,
		customerService: customerService,
		staffService:    staffService,
		catalogService:  catalogService,
	}
}

// Status keçid qaydaları
func (s *BookingService) isValidStatusTransition(from, to BookingStatus) bool {
	transitions := map[BookingStatus][]BookingStatus{
		BookingStatusPending:   {BookingStatusConfirmed, BookingStatusCancelled},
		BookingStatusConfirmed: {BookingStatusCancelled, BookingStatusCompleted},
		BookingStatusCancelled: {},
		BookingStatusCompleted: {},
	}
	for _, allowed := range transitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

func (s *BookingService) CreateBooking(ctx context.Context, req *CreateBookingRequest) (*Booking, error) {
	if err := s.validateCreateRequest(req); err != nil {
		return nil, err
	}

	// 1. Slotun mövcudluğunu yoxla
	if err := s.slotService.ValidateSlotAvailability(ctx, req.SlotID); err != nil {
		return nil, NewBookingError("SLOT_UNAVAILABLE", "Selected slot is not available")
	}

	// 2. Slotun detallarını al (start/end time üçün)
	// QEYD: Burada real implementasiyada SlotService-dən dönən data cast olunmalıdır.
	// Sadəlik üçün slotun valid olduğunu fərz edib, vaxtları başqa yolla alırıq və ya
	// SlotService.GetSlot() metodunun qaytardığı strukturdan istifadə edirik.
	// Təxmini: slot := s.slotService.GetSlot(...)

	// 3. Digər entity-lərin mövcudluğunu yoxla (sadə ID check)
	if _, err := s.customerService.GetCustomer(ctx, req.BusinessID, req.CustomerID); err != nil {
		return nil, NewBookingError("CUSTOMER_NOT_FOUND", "Customer not found")
	}
	if _, err := s.staffService.GetStaff(ctx, req.StaffID, req.BusinessID); err != nil {
		return nil, NewBookingError("STAFF_NOT_FOUND", "Staff not found")
	}

	// 4. Booking yarat
	// Qeyd: Real layihədə StartTime/EndTime slot-dan gəlməlidir.
	// İndilik request-dən gələn vaxtın slotla uyğunluğunu yoxladığımızı fərz edirik.
	// Tutaq ki, slot service bizə vaxtları verir. (Burada simulyasiya edirik)
	slot, err := s.slotService.GetSlot(ctx, req.BusinessID, req.SlotID)
	if err != nil {
		return nil, fmt.Errorf("failed to GetSlot: %w", err)
	}
	booking := NewBooking(
		req.BusinessID,
		req.CustomerID,
		req.StaffID,
		req.ServiceID,
		req.SlotID,
		slot.StartTime,
		slot.EndTime,
		req.Notes,
	)

	// 5. Booking-i DB-də saxla
	if err := s.repo.Create(ctx, booking); err != nil {
		return nil, fmt.Errorf("failed to create booking record: %w", err)
	}

	// 6. Slot-u "Reserve" et (ATOMICITY vacibdir, amma burada sadə çağırırıq)
	if err := s.slotService.ReserveSlot(ctx, req.SlotID, booking.ID); err != nil {
		// Rollback (booking-i silmək lazımdır)
		// s.repo.Delete(ctx, booking.ID)
		return nil, fmt.Errorf("failed to reserve slot: %w", err)
	}

	// 7. Müştəri statistikasını yenilə
	_ = s.customerService.IncrementBookingCount(ctx, req.CustomerID)

	return booking, nil
}

func (s *BookingService) CancelBooking(ctx context.Context, businessID, bookingID uuid.UUID) error {
	booking, err := s.repo.GetByID(ctx, businessID, bookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return NewBookingError("BOOKING_NOT_FOUND", "Booking not found")
	}

	if !s.isValidStatusTransition(booking.Status, BookingStatusCancelled) {
		return NewBookingError("INVALID_TRANSITION", "Cannot cancel this booking")
	}

	booking.Status = BookingStatusCancelled
	booking.UpdatedAt = time.Now()

	// 1. Booking statusunu yenilə
	if err := s.repo.Update(ctx, booking); err != nil {
		return fmt.Errorf("failed to update booking status: %w", err)
	}

	// 2. Slotu azad et
	if err := s.slotService.ReleaseSlot(ctx, booking.SlotID); err != nil {
		// Log error (kritik deyil, amma düzəldilməlidir)
		fmt.Printf("Failed to release slot %v: %v\n", booking.SlotID, err)
	}

	return nil
}

func (s *BookingService) UpdateBooking(ctx context.Context, businessID, bookingID uuid.UUID, req *UpdateBookingRequest) error {
	booking, err := s.repo.GetByID(ctx, businessID, bookingID)
	if err != nil {
		return err
	}
	if booking == nil {
		return NewBookingError("BOOKING_NOT_FOUND", "Booking not found")
	}

	if req.Status != "" && req.Status != booking.Status {
		if !s.isValidStatusTransition(booking.Status, req.Status) {
			return NewBookingError("INVALID_TRANSITION", "Invalid status transition")
		}
		booking.Status = req.Status
	}

	if req.Notes != nil {
		booking.Notes = *req.Notes
	}
	booking.UpdatedAt = time.Now()

	return s.repo.Update(ctx, booking)
}

// Get metodları standartdır (Repo-ya yönləndirir)
func (s *BookingService) GetBookingByID(ctx context.Context, bid, id uuid.UUID) (*Booking, error) {
	return s.repo.GetByID(ctx, bid, id)
}
func (s *BookingService) GetCustomerBookings(ctx context.Context, bid, cid uuid.UUID) ([]*Booking, error) {
	customer, err := s.repo.GetByCustomer(ctx, bid, cid)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer booking: %w", err)
	}
	return customer, nil
}
func (s *BookingService) GetStaffBookings(ctx context.Context, bid, sid uuid.UUID) ([]*Booking, error) {
	staff, err := s.repo.GetByStaff(ctx, bid, sid)
	if err != nil {
		return nil, fmt.Errorf("failed to get staff: %w", err)
	}
	return staff, err
}
func (s *BookingService) GetBusinessBookings(ctx context.Context, bid uuid.UUID) ([]*Booking, error) {
	business, err := s.repo.GetByBusiness(ctx, bid)
	if err != nil {
		return nil, fmt.Errorf("failed to get business: %w", err)
	}
	return business, nil
}
