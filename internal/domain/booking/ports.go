// File: internal/domain/booking/ports.go
package booking

import (
	"context"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/availability"
	"github.com/google/uuid"
)

// Repository – booking saxlama qati.
type Repository interface {
	// Create – ust-uste dusme olarsa ErrSlotTaken qaytarmalidir
	// (DB-deki bookings_no_overlap exclusion constraint-i).
	Create(ctx context.Context, booking *Booking) error
	GetByID(ctx context.Context, businessID, bookingID uuid.UUID) (*Booking, error)
	Update(ctx context.Context, booking *Booking) error

	List(ctx context.Context, businessID uuid.UUID, filter *ListFilter) ([]*Booking, error)
	CountByStatus(ctx context.Context, businessID uuid.UUID, status BookingStatus) (int, error)

	// ListForCustomerUser – musteri tetbiqi ucun: istifadecinin butun
	// bizneslerdeki bronlari.
	ListForCustomerUser(ctx context.Context, userID uuid.UUID, filter *ListFilter) ([]*Booking, error)

	// GetByIDForUser – multi-tenant business_id olmadan, istifadecinin
	// iştirakcisi oldugu bronu tapir (musteri tetbiqi ucun).
	GetByIDForUser(ctx context.Context, userID, bookingID uuid.UUID) (*Booking, error)
}

// AvailabilityChecker – availability.Service-in booking-e lazim olan hissesi.
type AvailabilityChecker interface {
	CheckSlot(
		ctx context.Context,
		businessID, staffID uuid.UUID,
		serviceID *uuid.UUID,
		start time.Time,
	) (*availability.TimeSlot, error)

	ResolveSettings(ctx context.Context, businessID uuid.UUID, staffID uuid.UUID) (*availability.ScheduleSettings, error)
}

// ParticipantResolver – bildiris ucun terefleri (isci/musteri/biznes) toplayir.
//
// Resolve eyni zamanda tehlukesizlik yoxlamasi rolunu oynayir: sorgu
// isci ve musterini business_id uzre JOIN edir, ona gore basqa biznesin
// iscisi/musterisi ile booking yaratmaq mumkun deyil.
type ParticipantResolver interface {
	Resolve(ctx context.Context, businessID, staffID, customerID uuid.UUID, serviceID *uuid.UUID) (*Participants, error)

	// BusinessIDForStaff – iscinin aid oldugu biznes.
	//
	// Musteri rolundaki istifadecinin JWT-sinde business_id olmur
	// (o, hec bir biznesin iscisi deyil). Biznesi client-den qebul
	// etmek evezine sectiyi iscinin profilinden cixaririq.
	BusinessIDForStaff(ctx context.Context, staffID uuid.UUID) (uuid.UUID, error)
}

// EventPublisher – booking hadiselerini bildiris qatina otürür.
// Xetasi booking emeliyyatini ucurmamalidir.
type EventPublisher interface {
	Publish(ctx context.Context, event *Event) error
}

// CustomerStats – musteri sayğaclarini yenileyir (opsional).
type CustomerStats interface {
	IncrementBookingCount(ctx context.Context, customerID uuid.UUID) error
}

// Actor – emeliyyati eden istifadeci. Icaze yoxlamasi ucun lazimdir.
type Actor struct {
	UserID     uuid.UUID
	BusinessID uuid.UUID
	Role       string
	// StaffID – istifadeci hemin biznesde iscidirse dolu olur.
	StaffID *uuid.UUID
}

// IsProvider – biznes terefi (sahib / isci / solo praktik).
func (a Actor) IsProvider() bool {
	return a.Role == "provider_owner" || a.Role == "staff" || a.Role == "solo_practitioner"
}

// Service – booking use-case-leri.
type Service interface {
	// CreateBooking – vaxti yoxlayir, bronu yaradir, provider-e bildiris atir.
	CreateBooking(ctx context.Context, actor Actor, req *CreateBookingRequest) (*Booking, error)

	// Confirm – provider bronu qebul edir.
	Confirm(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error)

	// ProposeReschedule – provider basqa vaxt teklif edir.
	ProposeReschedule(ctx context.Context, actor Actor, bookingID uuid.UUID, req *ProposeRescheduleRequest) (*Booking, error)

	// RespondToProposal – musteri teklifi qebul edir ve ya redd edir.
	RespondToProposal(ctx context.Context, actor Actor, bookingID uuid.UUID, req *RespondToProposalRequest) (*Booking, error)

	Cancel(ctx context.Context, actor Actor, bookingID uuid.UUID, req *CancelBookingRequest) (*Booking, error)
	Complete(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error)
	MarkNoShow(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error)

	UpdateNotes(ctx context.Context, actor Actor, bookingID uuid.UUID, req *UpdateBookingRequest) (*Booking, error)

	GetBooking(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error)
	ListBookings(ctx context.Context, actor Actor, filter *ListFilter) ([]*Booking, error)
	ListMyBookings(ctx context.Context, userID uuid.UUID, filter *ListFilter) ([]*Booking, error)

	CountByStatus(ctx context.Context, businessID uuid.UUID, status BookingStatus) (int, error)
}
