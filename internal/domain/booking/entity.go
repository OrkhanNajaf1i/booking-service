// File: internal/domain/booking/entity.go
//
// Booking artiq onceden yaradilmis slot setirine baglı deyil.
// Musteri sadece "hansi isci, hansi xidmet, hansi an" gonderir;
// availability muherriki hemin anin qrafike uygun ve bos oldugunu
// tesdiqleyir, sonra booking yazilir.
package booking

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================
// STATUS
// ============================================================

type BookingStatus string

const (
	// Musteri bron etdi, provider hele baxmayib.
	BookingStatusPending BookingStatus = "pending"
	// Provider qebul etdi (ve ya auto_confirm aktiv idi).
	BookingStatusConfirmed BookingStatus = "confirmed"
	// Provider basqa vaxt teklif etdi, musterinin cavabi gozlenilir.
	BookingStatusRescheduleProposed BookingStatus = "reschedule_proposed"
	BookingStatusCancelled          BookingStatus = "cancelled"
	BookingStatusCompleted          BookingStatus = "completed"
	BookingStatusNoShow             BookingStatus = "no_show"
)

func (s BookingStatus) IsValid() bool {
	switch s {
	case BookingStatusPending, BookingStatusConfirmed, BookingStatusRescheduleProposed,
		BookingStatusCancelled, BookingStatusCompleted, BookingStatusNoShow:
		return true
	}
	return false
}

// IsActive – vaxti hele de tutan statuslar.
func (s BookingStatus) IsActive() bool {
	return s == BookingStatusPending ||
		s == BookingStatusConfirmed ||
		s == BookingStatusRescheduleProposed
}

// IsTerminal – bundan sonra kecid yoxdur.
func (s BookingStatus) IsTerminal() bool {
	return s == BookingStatusCancelled ||
		s == BookingStatusCompleted ||
		s == BookingStatusNoShow
}

// allowedTransitions – icaze verilen status kecidleri.
var allowedTransitions = map[BookingStatus][]BookingStatus{
	BookingStatusPending: {
		BookingStatusConfirmed,
		BookingStatusRescheduleProposed,
		BookingStatusCancelled,
	},
	BookingStatusRescheduleProposed: {
		BookingStatusConfirmed,           // musteri teklifi qebul etdi
		BookingStatusRescheduleProposed,  // provider yeniden teklif etdi
		BookingStatusPending,             // musteri redd etdi, ilkin vaxta qayitdi
		BookingStatusCancelled,
	},
	BookingStatusConfirmed: {
		BookingStatusRescheduleProposed,
		BookingStatusCancelled,
		BookingStatusCompleted,
		BookingStatusNoShow,
	},
	BookingStatusCancelled: {},
	BookingStatusCompleted: {},
	BookingStatusNoShow:    {},
}

// CanTransitionTo – kecid qanunidirmi?
func (s BookingStatus) CanTransitionTo(target BookingStatus) bool {
	for _, allowed := range allowedTransitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}

// ============================================================
// ENTITY
// ============================================================

type Booking struct {
	ID         uuid.UUID  `db:"id"          json:"id"`
	BusinessID uuid.UUID  `db:"business_id" json:"business_id"`
	CustomerID uuid.UUID  `db:"customer_id" json:"customer_id"`
	StaffID    uuid.UUID  `db:"staff_id"    json:"staff_id"`
	ServiceID  *uuid.UUID `db:"service_id"  json:"service_id,omitempty"`
	LocationID *uuid.UUID `db:"location_id" json:"location_id,omitempty"`

	// SlotID – kohne slot cedveli ile uygunluq ucun saxlanilir.
	// Yeni axinda hemise NULL-dur.
	SlotID *uuid.UUID `db:"slot_id" json:"slot_id,omitempty"`

	StartTime    time.Time     `db:"start_time"    json:"start_time"`
	EndTime      time.Time     `db:"end_time"      json:"end_time"`
	DurationMins int           `db:"duration_mins" json:"duration_mins"`
	Status       BookingStatus `db:"status"        json:"status"`
	Notes        string        `db:"notes"         json:"notes"`

	// Provider alternativ vaxt teklif edende doldurulur.
	ProposedStartTime *time.Time `db:"proposed_start_time" json:"proposed_start_time,omitempty"`
	ProposedEndTime   *time.Time `db:"proposed_end_time"   json:"proposed_end_time,omitempty"`
	ProposedBy        *uuid.UUID `db:"proposed_by"         json:"proposed_by,omitempty"`
	ProposalNote      *string    `db:"proposal_note"       json:"proposal_note,omitempty"`
	ProposedAt        *time.Time `db:"proposed_at"         json:"proposed_at,omitempty"`

	CancelReason *string    `db:"cancel_reason" json:"cancel_reason,omitempty"`
	CancelledBy  *uuid.UUID `db:"cancelled_by"  json:"cancelled_by,omitempty"`
	ConfirmedAt  *time.Time `db:"confirmed_at"  json:"confirmed_at,omitempty"`

	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

// HasPendingProposal – musterinin cavab vermeli oldugu teklif varmi?
func (b *Booking) HasPendingProposal() bool {
	return b.Status == BookingStatusRescheduleProposed &&
		b.ProposedStartTime != nil &&
		b.ProposedEndTime != nil
}

// ClearProposal – teklif saheleri temizlenir (qebul/red sonrasi).
func (b *Booking) ClearProposal() {
	b.ProposedStartTime = nil
	b.ProposedEndTime = nil
	b.ProposedBy = nil
	b.ProposalNote = nil
	b.ProposedAt = nil
}

// NewBooking – yeni booking qurur. Status cagiran terefden verilir,
// cunki auto_confirm ayari birbasa "confirmed" teleb ede biler.
func NewBooking(
	businessID, customerID, staffID uuid.UUID,
	serviceID, locationID *uuid.UUID,
	startTime, endTime time.Time,
	durationMins int,
	status BookingStatus,
	notes string,
) *Booking {
	now := time.Now()
	booking := &Booking{
		ID:           uuid.New(),
		BusinessID:   businessID,
		CustomerID:   customerID,
		StaffID:      staffID,
		ServiceID:    serviceID,
		LocationID:   locationID,
		StartTime:    startTime,
		EndTime:      endTime,
		DurationMins: durationMins,
		Status:       status,
		Notes:        notes,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if status == BookingStatusConfirmed {
		booking.ConfirmedAt = &now
	}
	return booking
}

// ============================================================
// PARTICIPANTS
// ============================================================

// Participants – bildiris gondermek ucun lazim olan terefler.
// CustomerUserID uuid.Nil ola biler: musteri sistemde qeydiyyatdan
// kecmeyib, provider onu elle elave edib.
type Participants struct {
	BusinessName   string
	StaffUserID    uuid.UUID
	StaffName      string
	CustomerUserID uuid.UUID
	CustomerName   string
	ServiceName    string
}

// ============================================================
// EVENTS
// ============================================================

// EventType – booking axininda bas veren hadiseler.
type EventType string

const (
	EventCreated            EventType = "booking.created"
	EventConfirmed          EventType = "booking.confirmed"
	EventRescheduleProposed EventType = "booking.reschedule_proposed"
	EventRescheduleAccepted EventType = "booking.reschedule_accepted"
	EventRescheduleDeclined EventType = "booking.reschedule_declined"
	EventCancelled          EventType = "booking.cancelled"
	EventCompleted          EventType = "booking.completed"
	EventNoShow             EventType = "booking.no_show"
)

// Event – bildiris qatina otürülen hadise.
type Event struct {
	Type         EventType
	Booking      *Booking
	Participants *Participants
	// ActorUserID – hadiseni tetikleyen istifadeci; ozune bildiris getmir.
	ActorUserID uuid.UUID
	Message     string
}

// ============================================================
// REQUEST DTOS
// ============================================================

// CreateBookingRequest – musterinin bron sorgusu.
// SlotID yoxdur: vaxt birbasa StartTime ile verilir.
type CreateBookingRequest struct {
	CustomerID uuid.UUID  `json:"customer_id"`
	StaffID    uuid.UUID  `json:"staff_id"`
	ServiceID  *uuid.UUID `json:"service_id,omitempty"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	StartTime  time.Time  `json:"start_time"`
	Notes      string     `json:"notes"`
}

// ProposeRescheduleRequest – provider alternativ vaxt teklif edir.
type ProposeRescheduleRequest struct {
	NewStartTime time.Time `json:"new_start_time"`
	Note         string    `json:"note"`
}

// RespondToProposalRequest – musteri teklife cavab verir.
type RespondToProposalRequest struct {
	Accept bool   `json:"accept"`
	Note   string `json:"note"`
}

// CancelBookingRequest – legv sebebi.
type CancelBookingRequest struct {
	Reason string `json:"reason"`
}

// UpdateBookingRequest – qeydlerin duzelisi (status ucun ayrica endpointler var).
type UpdateBookingRequest struct {
	Notes *string `json:"notes,omitempty"`
}

// ListFilter – booking siyahisinin filtrleri.
type ListFilter struct {
	StaffID    *uuid.UUID
	CustomerID *uuid.UUID
	Status     *BookingStatus
	From       *time.Time
	To         *time.Time
	Limit      int
	Offset     int
}

// ============================================================
// ERRORS
// ============================================================

type BookingError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BookingError) Error() string { return e.Message }

func NewBookingError(code, message string) *BookingError {
	return &BookingError{Code: code, Message: message}
}

var (
	ErrNotFound          = NewBookingError("BOOKING_NOT_FOUND", "bron tapilmadi")
	ErrSlotTaken         = NewBookingError("SLOT_TAKEN", "bu vaxt artiq bron olunub")
	ErrInvalidTransition = NewBookingError("INVALID_TRANSITION", "bu status kecidi mumkun deyil")
	ErrNoProposal        = NewBookingError("NO_PROPOSAL", "cavablanacaq teklif yoxdur")
	ErrProposalDisabled  = NewBookingError("PROPOSAL_DISABLED", "bu biznesde vaxt teklifi baglidir")
	ErrCustomerRequired  = NewBookingError("CUSTOMER_REQUIRED", "customer_id teleb olunur")
	ErrStaffRequired     = NewBookingError("STAFF_REQUIRED", "staff_id teleb olunur")
	ErrStartRequired     = NewBookingError("START_REQUIRED", "start_time teleb olunur")
	ErrForbidden         = NewBookingError("FORBIDDEN", "bu emeliyyat ucun icazeniz yoxdur")
)
