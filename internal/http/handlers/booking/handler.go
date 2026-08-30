// File: internal/http/handlers/booking/handler.go
package booking

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	availabilityDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/availability"
	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/staff"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// StaffLookup – aktorun hemin biznesdeki staff profilini tapmaq ucun.
// Adi isci yalniz oz bronlarina toxuna bilsin deye lazimdir.
// staff.Repository bu imzani odeyir.
type StaffLookup interface {
	GetStaffByUserID(ctx context.Context, userID, businessID uuid.UUID) (*staff.StaffProfile, error)
}

type Handler struct {
	service domain.Service
	staff   StaffLookup
	log     logger.Logger
}

func NewHandler(service domain.Service, staffLookup StaffLookup, log logger.Logger) *Handler {
	return &Handler{service: service, staff: staffLookup, log: log}
}

// ============================================================
// CREATE
// ============================================================

// CreateBooking – POST /api/v1/bookings
//
// @Summary      Bron yarat
// @Description  Secilen vaxt availability muherrikinden kecirilir. Bos deyilse SLOT_TAKEN qaytarilir. auto_confirm sondurulubse bron pending yaranir ve isciye aninda bildiris gedir.
// @Tags         Booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.CreateBookingRequest true "Bron"
// @Success      201 {object} SuccessResponse
// @Failure      409 {object} ErrorResponse "Vaxt artiq tutulub"
// @Router       /bookings [post]
func (h *Handler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	var req domain.CreateBookingRequest
	if !decodeBody(w, r, &req) {
		return
	}

	created, err := h.service.CreateBooking(r.Context(), actor, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, "Bron yaradildi", created)
}

// ============================================================
// PROVIDER AXINI
// ============================================================

// Confirm – POST /api/v1/bookings/{id}/confirm
//
// @Summary      Bronu tesdiqle
// @Tags         Booking
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Booking ID"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id}/confirm [post]
func (h *Handler) Confirm(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	updated, err := h.service.Confirm(r.Context(), actor, bookingID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Bron tesdiqlendi", updated)
}

// ProposeReschedule – POST /api/v1/bookings/{id}/propose
//
// @Summary      Alternativ vaxt teklif et
// @Description  Provider musteriye basqa vaxt teklif edir. Teklif olunan vaxt da qrafik uzre yoxlanilir. Musteri qebul edene qeder ilkin vaxt saxlanilir.
// @Tags         Booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Booking ID"
// @Param        request body domain.ProposeRescheduleRequest true "Yeni vaxt"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id}/propose [post]
func (h *Handler) ProposeReschedule(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	var req domain.ProposeRescheduleRequest
	if !decodeBody(w, r, &req) {
		return
	}

	updated, err := h.service.ProposeReschedule(r.Context(), actor, bookingID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Teklif gonderildi", updated)
}

// Complete – POST /api/v1/bookings/{id}/complete
//
// @Summary      Randevunu tamamlanmis isarele
// @Tags         Booking
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Booking ID"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id}/complete [post]
func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	updated, err := h.service.Complete(r.Context(), actor, bookingID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Randevu tamamlandi", updated)
}

// MarkNoShow – POST /api/v1/bookings/{id}/no-show
//
// @Summary      Musteri gelmedi
// @Tags         Booking
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Booking ID"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id}/no-show [post]
func (h *Handler) MarkNoShow(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	updated, err := h.service.MarkNoShow(r.Context(), actor, bookingID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Gelmeme qeyd edildi", updated)
}

// ============================================================
// MUSTERI AXINI
// ============================================================

// RespondToProposal – POST /api/v1/bookings/{id}/respond
//
// @Summary      Teklife cavab ver
// @Description  accept=true olsa bron teklif olunan vaxta kecir ve tesdiqlenir; false olsa ilkin vaxta qayidir.
// @Tags         Booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Booking ID"
// @Param        request body domain.RespondToProposalRequest true "Cavab"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id}/respond [post]
func (h *Handler) RespondToProposal(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	var req domain.RespondToProposalRequest
	if !decodeBody(w, r, &req) {
		return
	}

	updated, err := h.service.RespondToProposal(r.Context(), actor, bookingID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	message := "Teklif redd edildi"
	if req.Accept {
		message = "Teklif qebul edildi"
	}
	writeSuccess(w, http.StatusOK, message, updated)
}

// Cancel – POST /api/v1/bookings/{id}/cancel
//
// @Summary      Bronu legv et
// @Tags         Booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Booking ID"
// @Param        request body domain.CancelBookingRequest false "Sebeb"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id}/cancel [post]
func (h *Handler) Cancel(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	// Sebeb opsionaldir – bos body de qebul edilir.
	var req domain.CancelBookingRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	updated, err := h.service.Cancel(r.Context(), actor, bookingID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Bron legv edildi", updated)
}

// ============================================================
// READ
// ============================================================

// ListBookings – GET /api/v1/bookings
//
// @Summary      Biznesin bronlari
// @Tags         Booking
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id    query string false "Isci ID"
// @Param        customer_id query string false "Musteri ID"
// @Param        status      query string false "pending|confirmed|reschedule_proposed|cancelled|completed|no_show"
// @Param        from        query string false "RFC3339 ve ya YYYY-MM-DD"
// @Param        to          query string false "RFC3339 ve ya YYYY-MM-DD"
// @Param        limit       query int    false "Default 50"
// @Param        offset      query int    false "Default 0"
// @Success      200 {object} SuccessResponse
// @Router       /bookings [get]
func (h *Handler) ListBookings(w http.ResponseWriter, r *http.Request) {
	actor, ok := h.actor(w, r)
	if !ok {
		return
	}

	filter, err := parseFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FILTER", err.Error())
		return
	}

	rows, err := h.service.ListBookings(r.Context(), actor, filter)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", rows)
}

// ListMyBookings – GET /api/v1/bookings/my
// Musteri tetbiqi ucun: istifadecinin butun bizneslerdeki bronlari.
//
// @Summary      Menim bronlarim
// @Tags         Booking
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse
// @Router       /bookings/my [get]
func (h *Handler) ListMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFrom(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_USER", "istifadeci tapilmadi")
		return
	}

	filter, err := parseFilter(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FILTER", err.Error())
		return
	}

	rows, err := h.service.ListMyBookings(r.Context(), userID, filter)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", rows)
}

// GetBooking – GET /api/v1/bookings/{id}
//
// @Summary      Bron detallari
// @Tags         Booking
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Booking ID"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id} [get]
func (h *Handler) GetBooking(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	found, err := h.service.GetBooking(r.Context(), actor, bookingID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", found)
}

// UpdateNotes – PATCH /api/v1/bookings/{id}
//
// @Summary      Bron qeydlerini yenile
// @Tags         Booking
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id      path string true "Booking ID"
// @Param        request body domain.UpdateBookingRequest true "Qeydler"
// @Success      200 {object} SuccessResponse
// @Router       /bookings/{id} [patch]
func (h *Handler) UpdateNotes(w http.ResponseWriter, r *http.Request) {
	actor, bookingID, ok := h.actorAndID(w, r)
	if !ok {
		return
	}

	var req domain.UpdateBookingRequest
	if !decodeBody(w, r, &req) {
		return
	}

	updated, err := h.service.UpdateNotes(r.Context(), actor, bookingID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Qeydler yenilendi", updated)
}

// ============================================================
// HELPERS
// ============================================================

// actor – JWT-den emeliyyati eden sexsi qurur.
// Rol "staff"-dirsa hemin iscinin profil ID-si de tapilir ki,
// basqasinin bronuna mudaxile ede bilmesin.
func (h *Handler) actor(w http.ResponseWriter, r *http.Request) (domain.Actor, bool) {
	userID, err := middleware.UserIDFrom(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_USER", "istifadeci tapilmadi")
		return domain.Actor{}, false
	}

	actor := domain.Actor{
		UserID:     userID,
		BusinessID: middleware.OptionalBusinessIDFrom(r),
		Role:       middleware.RoleFrom(r),
	}

	if actor.Role == "staff" && actor.BusinessID != uuid.Nil && h.staff != nil {
		profile, err := h.staff.GetStaffByUserID(r.Context(), userID, actor.BusinessID)
		if err == nil && profile != nil {
			staffID := profile.ID
			actor.StaffID = &staffID
		}
	}

	return actor, true
}

func (h *Handler) actorAndID(w http.ResponseWriter, r *http.Request) (domain.Actor, uuid.UUID, bool) {
	actor, ok := h.actor(w, r)
	if !ok {
		return actor, uuid.Nil, false
	}

	bookingID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "booking id duzgun deyil")
		return actor, uuid.Nil, false
	}

	return actor, bookingID, true
}

// writeDomainError – domain xeta kodlarini HTTP statuslarina baglayir.
func (h *Handler) writeDomainError(w http.ResponseWriter, err error) {
	var bookingErr *domain.BookingError
	if errors.As(err, &bookingErr) {
		writeError(w, statusForCode(bookingErr.Code), bookingErr.Code, bookingErr.Message)
		return
	}

	// Availability muherrikinin xetalari da istifadeciye aiddir:
	// "bu vaxt tutulub", "qrafike uygun deyil", "cox erkendir" ve s.
	var availabilityErr *availabilityDomain.Error
	if errors.As(err, &availabilityErr) {
		writeError(w, statusForCode(availabilityErr.Code), availabilityErr.Code, availabilityErr.Message)
		return
	}

	h.log.Error("Booking xetasi", logger.Field{Key: "error", Value: err.Error()})
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "daxili xeta")
}

func statusForCode(code string) int {
	switch code {
	case "BOOKING_NOT_FOUND", "NOT_FOUND", "SERVICE_NOT_FOUND":
		return http.StatusNotFound
	case "SLOT_TAKEN", "SLOT_BLOCKED", "INVALID_TRANSITION":
		return http.StatusConflict
	case "FORBIDDEN":
		return http.StatusForbidden
	default:
		return http.StatusBadRequest
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "JSON oxunmadi")
		return false
	}
	return true
}

// parseFilter – query parametrlerini ListFilter-e cevirir.
func parseFilter(r *http.Request) (*domain.ListFilter, error) {
	query := r.URL.Query()
	filter := &domain.ListFilter{}

	if raw := query.Get("staff_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return nil, errors.New("staff_id duzgun deyil")
		}
		filter.StaffID = &parsed
	}

	if raw := query.Get("customer_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			return nil, errors.New("customer_id duzgun deyil")
		}
		filter.CustomerID = &parsed
	}

	if raw := query.Get("status"); raw != "" {
		status := domain.BookingStatus(raw)
		if !status.IsValid() {
			return nil, errors.New("status duzgun deyil")
		}
		filter.Status = &status
	}

	if raw := query.Get("from"); raw != "" {
		parsed, err := parseFlexibleTime(raw)
		if err != nil {
			return nil, errors.New("from tarixi duzgun deyil")
		}
		filter.From = &parsed
	}

	if raw := query.Get("to"); raw != "" {
		parsed, err := parseFlexibleTime(raw)
		if err != nil {
			return nil, errors.New("to tarixi duzgun deyil")
		}
		filter.To = &parsed
	}

	if raw := query.Get("limit"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			filter.Limit = value
		}
	}
	if raw := query.Get("offset"); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			filter.Offset = value
		}
	}

	return filter, nil
}

// parseFlexibleTime – hem "2026-09-01", hem de tam RFC3339 qebul edir.
func parseFlexibleTime(raw string) (time.Time, error) {
	if parsed, err := time.Parse(time.RFC3339, raw); err == nil {
		return parsed, nil
	}
	return time.Parse("2006-01-02", raw)
}
