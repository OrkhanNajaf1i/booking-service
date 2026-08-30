// File: internal/domain/booking/service.go
package booking

import (
	"context"
	"fmt"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

type bookingService struct {
	repo         Repository
	availability AvailabilityChecker
	participants ParticipantResolver
	events       EventPublisher
	stats        CustomerStats
	log          logger.Logger
	now          func() time.Time
}

// NewService – booking servisi. stats ve events nil ola biler.
func NewService(
	repo Repository,
	availabilityChecker AvailabilityChecker,
	participants ParticipantResolver,
	events EventPublisher,
	stats CustomerStats,
	log logger.Logger,
) Service {
	return &bookingService{
		repo:         repo,
		availability: availabilityChecker,
		participants: participants,
		events:       events,
		stats:        stats,
		log:          log,
		now:          time.Now,
	}
}

// ============================================================
// CREATE
// ============================================================

// CreateBooking – musteri vaxt secir.
//
// Iki qat mudafie var:
//  1. availability.CheckSlot – vaxt qrafike uygundur ve gorunusde bosdur
//  2. DB-deki exclusion constraint – eyni anda gelen iki sorgudan
//     yalnizca biri kecir, ikincisi ErrSlotTaken alir
func (s *bookingService) CreateBooking(
	ctx context.Context,
	actor Actor,
	req *CreateBookingRequest,
) (*Booking, error) {
	if err := validateCreateRequest(req); err != nil {
		return nil, err
	}

	// Musterinin JWT-sinde business_id olmur — o, hec bir biznesin
	// iscisi deyil. Bele halda biznesi sectiyi iscinin profilinden
	// cixaririq; client-in gonderdiyi deyere guvenmirik.
	businessID := actor.BusinessID
	if businessID == uuid.Nil {
		if s.participants == nil {
			return nil, NewBookingError("BUSINESS_REQUIRED", "business konteksti tapilmadi")
		}

		resolved, err := s.participants.BusinessIDForStaff(ctx, req.StaffID)
		if err != nil || resolved == uuid.Nil {
			return nil, NewBookingError("STAFF_NOT_FOUND", "isci tapilmadi")
		}
		businessID = resolved
	}

	// Terefleri burada, yazilisdan EVVEL cozuruk: sorgu isci ve musterini
	// business_id uzre JOIN edir, ona gore basqa biznesin musterisi adina
	// bron yaratmaq cehdi burada dayanir.
	var parties *Participants
	if s.participants != nil {
		resolved, err := s.participants.Resolve(ctx, businessID, req.StaffID, req.CustomerID, req.ServiceID)
		if err != nil {
			return nil, NewBookingError("INVALID_PARTICIPANTS", "isci ve ya musteri bu biznese aid deyil")
		}
		parties = resolved
	}

	slot, err := s.availability.CheckSlot(ctx, businessID, req.StaffID, req.ServiceID, req.StartTime)
	if err != nil {
		return nil, err
	}

	settings, err := s.availability.ResolveSettings(ctx, businessID, req.StaffID)
	if err != nil {
		return nil, err
	}

	// auto_confirm aktivdirse provider tesdiqi gozlenmir.
	status := BookingStatusPending
	if settings.AutoConfirm {
		status = BookingStatusConfirmed
	}

	newBooking := NewBooking(
		businessID,
		req.CustomerID,
		req.StaffID,
		req.ServiceID,
		req.LocationID,
		slot.Start,
		slot.End,
		slot.DurationMins,
		status,
		req.Notes,
	)

	if err := s.repo.Create(ctx, newBooking); err != nil {
		return nil, err
	}

	if s.stats != nil {
		if err := s.stats.IncrementBookingCount(ctx, req.CustomerID); err != nil {
			s.log.Warn("Musteri sayğaci yenilenmedi",
				logger.Field{Key: "customer_id", Value: req.CustomerID.String()},
				logger.Field{Key: "error", Value: err.Error()},
			)
		}
	}

	eventType := EventCreated
	if status == BookingStatusConfirmed {
		eventType = EventConfirmed
	}
	s.emitWith(ctx, eventType, newBooking, parties, actor.UserID, "")

	s.log.Info("Bron yaradildi",
		logger.Field{Key: "booking_id", Value: newBooking.ID.String()},
		logger.Field{Key: "staff_id", Value: req.StaffID.String()},
		logger.Field{Key: "start_time", Value: newBooking.StartTime.Format(time.RFC3339)},
		logger.Field{Key: "status", Value: string(status)},
	)

	return newBooking, nil
}

// ============================================================
// PROVIDER: CONFIRM
// ============================================================

// Confirm – provider gelen bronu qebul edir.
func (s *bookingService) Confirm(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error) {
	existing, err := s.loadForProvider(ctx, actor, bookingID)
	if err != nil {
		return nil, err
	}

	if !existing.Status.CanTransitionTo(BookingStatusConfirmed) {
		return nil, ErrInvalidTransition
	}

	now := s.now()
	existing.Status = BookingStatusConfirmed
	existing.ConfirmedAt = &now
	existing.ClearProposal()
	existing.UpdatedAt = now

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("bron tesdiqlenmedi: %w", err)
	}

	s.emit(ctx, EventConfirmed, existing, actor.UserID, "")
	return existing, nil
}

// ============================================================
// PROVIDER: PROPOSE RESCHEDULE
// ============================================================

// ProposeReschedule – provider "bu vaxt olmur, bunu teklif edirem" deyir.
//
// Teklif olunan vaxt da availability muherrikinden kecir; provider
// ozunun bagli oldugu ve ya dolu olan bir ani teklif ede bilmez.
// Ilkin vaxt bronun uzerinde qalir – musteri redd etse ora qayidilir.
func (s *bookingService) ProposeReschedule(
	ctx context.Context,
	actor Actor,
	bookingID uuid.UUID,
	req *ProposeRescheduleRequest,
) (*Booking, error) {
	if req == nil || req.NewStartTime.IsZero() {
		return nil, NewBookingError("START_REQUIRED", "new_start_time teleb olunur")
	}

	existing, err := s.loadForProvider(ctx, actor, bookingID)
	if err != nil {
		return nil, err
	}

	settings, err := s.availability.ResolveSettings(ctx, existing.BusinessID, existing.StaffID)
	if err != nil {
		return nil, err
	}
	if !settings.AllowRescheduleProposal {
		return nil, ErrProposalDisabled
	}

	if !existing.Status.CanTransitionTo(BookingStatusRescheduleProposed) {
		return nil, ErrInvalidTransition
	}

	slot, err := s.availability.CheckSlot(
		ctx, existing.BusinessID, existing.StaffID, existing.ServiceID, req.NewStartTime,
	)
	if err != nil {
		return nil, err
	}

	now := s.now()
	proposedStart := slot.Start
	proposedEnd := slot.End
	note := req.Note

	existing.Status = BookingStatusRescheduleProposed
	existing.ProposedStartTime = &proposedStart
	existing.ProposedEndTime = &proposedEnd
	existing.ProposedBy = &actor.UserID
	existing.ProposedAt = &now
	existing.ProposalNote = &note
	existing.UpdatedAt = now

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("teklif yazilmadi: %w", err)
	}

	s.emit(ctx, EventRescheduleProposed, existing, actor.UserID, note)

	s.log.Info("Alternativ vaxt teklif edildi",
		logger.Field{Key: "booking_id", Value: existing.ID.String()},
		logger.Field{Key: "proposed_start", Value: proposedStart.Format(time.RFC3339)},
	)

	return existing, nil
}

// ============================================================
// CUSTOMER: RESPOND TO PROPOSAL
// ============================================================

// RespondToProposal – musteri teklifi qebul edir ve ya redd edir.
//
// Qebul: bron teklif olunan vaxta kocurulur ve confirmed olur.
// Red:   ilkin vaxt qalir, bron yeniden pending olur – provider
//
//	ya tesdiqleyir, ya legv edir.
func (s *bookingService) RespondToProposal(
	ctx context.Context,
	actor Actor,
	bookingID uuid.UUID,
	req *RespondToProposalRequest,
) (*Booking, error) {
	if req == nil {
		return nil, NewBookingError("INVALID_REQUEST", "request bosdur")
	}

	existing, err := s.loadForParticipant(ctx, actor, bookingID)
	if err != nil {
		return nil, err
	}

	if !existing.HasPendingProposal() {
		return nil, ErrNoProposal
	}

	now := s.now()

	if req.Accept {
		if !existing.Status.CanTransitionTo(BookingStatusConfirmed) {
			return nil, ErrInvalidTransition
		}

		existing.StartTime = *existing.ProposedStartTime
		existing.EndTime = *existing.ProposedEndTime
		existing.DurationMins = int(existing.EndTime.Sub(existing.StartTime).Minutes())
		existing.Status = BookingStatusConfirmed
		existing.ConfirmedAt = &now
		existing.ClearProposal()
		existing.UpdatedAt = now

		if err := s.repo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("teklif qebul edilmedi: %w", err)
		}

		s.emit(ctx, EventRescheduleAccepted, existing, actor.UserID, req.Note)
		return existing, nil
	}

	if !existing.Status.CanTransitionTo(BookingStatusPending) {
		return nil, ErrInvalidTransition
	}

	existing.Status = BookingStatusPending
	existing.ClearProposal()
	existing.UpdatedAt = now

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("teklif reddi yazilmadi: %w", err)
	}

	s.emit(ctx, EventRescheduleDeclined, existing, actor.UserID, req.Note)
	return existing, nil
}

// ============================================================
// CANCEL / COMPLETE / NO-SHOW
// ============================================================

// Cancel – hem musteri, hem provider legv ede biler.
func (s *bookingService) Cancel(
	ctx context.Context,
	actor Actor,
	bookingID uuid.UUID,
	req *CancelBookingRequest,
) (*Booking, error) {
	existing, err := s.loadForParticipant(ctx, actor, bookingID)
	if err != nil {
		return nil, err
	}

	if !existing.Status.CanTransitionTo(BookingStatusCancelled) {
		return nil, ErrInvalidTransition
	}

	now := s.now()
	reason := ""
	if req != nil {
		reason = req.Reason
	}

	existing.Status = BookingStatusCancelled
	existing.CancelReason = &reason
	existing.CancelledBy = &actor.UserID
	existing.ClearProposal()
	existing.UpdatedAt = now

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("bron legv edilmedi: %w", err)
	}

	s.emit(ctx, EventCancelled, existing, actor.UserID, reason)
	return existing, nil
}

func (s *bookingService) Complete(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error) {
	return s.providerStatusChange(ctx, actor, bookingID, BookingStatusCompleted, EventCompleted)
}

func (s *bookingService) MarkNoShow(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error) {
	return s.providerStatusChange(ctx, actor, bookingID, BookingStatusNoShow, EventNoShow)
}

func (s *bookingService) providerStatusChange(
	ctx context.Context,
	actor Actor,
	bookingID uuid.UUID,
	target BookingStatus,
	eventType EventType,
) (*Booking, error) {
	existing, err := s.loadForProvider(ctx, actor, bookingID)
	if err != nil {
		return nil, err
	}

	if !existing.Status.CanTransitionTo(target) {
		return nil, ErrInvalidTransition
	}

	existing.Status = target
	existing.ClearProposal()
	existing.UpdatedAt = s.now()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("status deyisdirilmedi: %w", err)
	}

	s.emit(ctx, eventType, existing, actor.UserID, "")
	return existing, nil
}

// ============================================================
// READ / UPDATE
// ============================================================

func (s *bookingService) UpdateNotes(
	ctx context.Context,
	actor Actor,
	bookingID uuid.UUID,
	req *UpdateBookingRequest,
) (*Booking, error) {
	existing, err := s.loadForProvider(ctx, actor, bookingID)
	if err != nil {
		return nil, err
	}

	if req != nil && req.Notes != nil {
		existing.Notes = *req.Notes
	}
	existing.UpdatedAt = s.now()

	if err := s.repo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("qeydler yazilmadi: %w", err)
	}
	return existing, nil
}

func (s *bookingService) GetBooking(ctx context.Context, actor Actor, bookingID uuid.UUID) (*Booking, error) {
	return s.loadForParticipant(ctx, actor, bookingID)
}

func (s *bookingService) ListBookings(
	ctx context.Context,
	actor Actor,
	filter *ListFilter,
) ([]*Booking, error) {
	if actor.BusinessID == uuid.Nil {
		return nil, NewBookingError("BUSINESS_REQUIRED", "business konteksti tapilmadi")
	}
	if filter == nil {
		filter = &ListFilter{}
	}
	normalizeFilter(filter)

	// Adi isci yalniz oz bronlarini gorur; sahib butun biznesi gorur.
	if actor.Role == "staff" && actor.StaffID != nil {
		filter.StaffID = actor.StaffID
	}

	return s.repo.List(ctx, actor.BusinessID, filter)
}

func (s *bookingService) ListMyBookings(
	ctx context.Context,
	userID uuid.UUID,
	filter *ListFilter,
) ([]*Booking, error) {
	if filter == nil {
		filter = &ListFilter{}
	}
	normalizeFilter(filter)
	return s.repo.ListForCustomerUser(ctx, userID, filter)
}

func (s *bookingService) CountByStatus(
	ctx context.Context,
	businessID uuid.UUID,
	status BookingStatus,
) (int, error) {
	return s.repo.CountByStatus(ctx, businessID, status)
}

// ============================================================
// HELPERS
// ============================================================

// loadForProvider – yalniz biznes terefinin ede bileceyi emeliyyatlar ucun.
func (s *bookingService) loadForProvider(
	ctx context.Context,
	actor Actor,
	bookingID uuid.UUID,
) (*Booking, error) {
	if !actor.IsProvider() {
		return nil, ErrForbidden
	}
	if actor.BusinessID == uuid.Nil {
		return nil, NewBookingError("BUSINESS_REQUIRED", "business konteksti tapilmadi")
	}

	existing, err := s.repo.GetByID(ctx, actor.BusinessID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("bron oxunmadi: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	// Adi isci basqasinin bronuna toxuna bilmez.
	if actor.Role == "staff" && actor.StaffID != nil && existing.StaffID != *actor.StaffID {
		return nil, ErrForbidden
	}

	return existing, nil
}

// loadForParticipant – hem provider, hem de bronun sahibi olan musteri
// ucun isleyir. Provider business konteksti ile, musteri user_id ile tapir.
func (s *bookingService) loadForParticipant(
	ctx context.Context,
	actor Actor,
	bookingID uuid.UUID,
) (*Booking, error) {
	if actor.IsProvider() && actor.BusinessID != uuid.Nil {
		return s.loadForProvider(ctx, actor, bookingID)
	}

	existing, err := s.repo.GetByIDForUser(ctx, actor.UserID, bookingID)
	if err != nil {
		return nil, fmt.Errorf("bron oxunmadi: %w", err)
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	return existing, nil
}

// emit – hadiseni bildiris qatina otürür.
// Bildiris xetasi booking emeliyyatini legv etmir; sadece loglanir.
func (s *bookingService) emit(
	ctx context.Context,
	eventType EventType,
	booking *Booking,
	actorUserID uuid.UUID,
	message string,
) {
	if s.events == nil {
		return
	}

	var parties *Participants
	if s.participants != nil {
		resolved, err := s.participants.Resolve(
			ctx, booking.BusinessID, booking.StaffID, booking.CustomerID, booking.ServiceID,
		)
		if err != nil {
			s.log.Warn("Bildiris terefleri tapilmadi",
				logger.Field{Key: "booking_id", Value: booking.ID.String()},
				logger.Field{Key: "error", Value: err.Error()},
			)
		} else {
			parties = resolved
		}
	}

	s.emitWith(ctx, eventType, booking, parties, actorUserID, message)
}

// emitWith – terefler artiq melum olanda tekrar sorgu atmir.
func (s *bookingService) emitWith(
	ctx context.Context,
	eventType EventType,
	booking *Booking,
	parties *Participants,
	actorUserID uuid.UUID,
	message string,
) {
	if s.events == nil {
		return
	}

	event := &Event{
		Type:         eventType,
		Booking:      booking,
		Participants: parties,
		ActorUserID:  actorUserID,
		Message:      message,
	}

	if err := s.events.Publish(ctx, event); err != nil {
		s.log.Warn("Hadise yayimlanmadi",
			logger.Field{Key: "booking_id", Value: booking.ID.String()},
			logger.Field{Key: "type", Value: string(eventType)},
			logger.Field{Key: "error", Value: err.Error()},
		)
	}
}

func normalizeFilter(filter *ListFilter) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
}
