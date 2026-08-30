// File: internal/http/handlers/public/handler.go
//
// Musteri terefinin kesf endpoint-leri.
//
// Bunlar auth telebi etmir: musteri hele qeydiyyatdan kecmemis de
// xestexana/berber secib bos vaxtlara baxa bilmelidir. Bron yaratmaq
// ucun ise artiq login lazimdir (POST /bookings qorunur).
//
// Yalniz oxuma emeliyyatlaridir ve hec bir sexsi melumat qaytarmir:
// biznes adi, isci adi/vezifesi, xidmet ve hesablanmis bos vaxtlar.
package public

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	availabilityDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/availability"
	businessDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	serviceDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/service"
	staffDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/staff"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

type Handler struct {
	businesses   businessDomain.Service
	staff        staffDomain.Service
	services     serviceDomain.ServiceUseCase
	availability availabilityDomain.Service
	log          logger.Logger
}

func NewHandler(
	businesses businessDomain.Service,
	staff staffDomain.Service,
	services serviceDomain.ServiceUseCase,
	availability availabilityDomain.Service,
	log logger.Logger,
) *Handler {
	return &Handler{
		businesses:   businesses,
		staff:        staff,
		services:     services,
		availability: availability,
		log:          log,
	}
}

// ============================================================
// BIZNESLER
// ============================================================

// BusinessCard – siyahida gosterilen minimal melumat.
type BusinessCard struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Industry        string    `json:"industry"`
	ServiceCategory string    `json:"service_category"`
	Phone           string    `json:"phone"`
	BusinessType    string    `json:"business_type"`
}

// ListBusinesses – GET /api/v1/public/businesses
//
// @Summary      Aktiv bizneslerin siyahisi
// @Description  Musteri xestexana / klinika / berberxana secmek ucun. Auth telem olunmur.
// @Tags         Public
// @Produce      json
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses [get]
func (h *Handler) ListBusinesses(w http.ResponseWriter, r *http.Request) {
	all, err := h.businesses.ListBusinesses(r.Context())
	if err != nil {
		h.writeInternal(w, err)
		return
	}

	cards := make([]BusinessCard, 0, len(all))
	for _, item := range all {
		// Deaktiv bizneslər musteriye gosterilmir.
		if !item.IsActive {
			continue
		}
		cards = append(cards, BusinessCard{
			ID:              item.ID,
			Name:            item.Name,
			Industry:        item.Industry,
			ServiceCategory: item.ServiceCategory,
			Phone:           item.Phone,
			BusinessType:    string(item.BusinessType),
		})
	}

	writeSuccess(w, http.StatusOK, "", cards)
}

// GetBusiness – GET /api/v1/public/businesses/{id}
//
// @Summary      Biznes detallari
// @Tags         Public
// @Produce      json
// @Param        id path string true "Biznes ID"
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses/{id} [get]
func (h *Handler) GetBusiness(w http.ResponseWriter, r *http.Request) {
	businessID, ok := h.pathID(w, r)
	if !ok {
		return
	}

	found, err := h.businesses.GetBusinessByID(r.Context(), businessID)
	if err != nil || found == nil || !found.IsActive {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "biznes tapilmadi")
		return
	}

	writeSuccess(w, http.StatusOK, "", BusinessCard{
		ID:              found.ID,
		Name:            found.Name,
		Industry:        found.Industry,
		ServiceCategory: found.ServiceCategory,
		Phone:           found.Phone,
		BusinessType:    string(found.BusinessType),
	})
}

// ============================================================
// ISCILER
// ============================================================

// StaffCard – musteriye gosterilen isci melumati.
// E-poct/telefon qesden daxil edilmir.
type StaffCard struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Title      string    `json:"title"`
	Department string    `json:"department"`
	Avatar     *string   `json:"avatar,omitempty"`
}

// ListStaff – GET /api/v1/public/businesses/{id}/staff
//
// @Summary      Biznesin isciləri
// @Description  Musteri hansi hekim/berber ile bron edeceyini secir.
// @Tags         Public
// @Produce      json
// @Param        id path string true "Biznes ID"
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses/{id}/staff [get]
func (h *Handler) ListStaff(w http.ResponseWriter, r *http.Request) {
	businessID, ok := h.pathID(w, r)
	if !ok {
		return
	}

	members, err := h.staff.ListStaff(r.Context(), businessID)
	if err != nil {
		h.writeInternal(w, err)
		return
	}

	cards := make([]StaffCard, 0, len(members))
	for _, member := range members {
		if string(member.Status) != "active" {
			continue
		}
		cards = append(cards, StaffCard{
			ID:         member.ID,
			FullName:   member.FullName,
			Title:      member.Title,
			Department: member.Department,
			Avatar:     member.Avatar,
		})
	}

	writeSuccess(w, http.StatusOK, "", cards)
}

// ============================================================
// XIDMETLER
// ============================================================

type ServiceCard struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	Price           float64   `json:"price"`
}

// ListServices – GET /api/v1/public/businesses/{id}/services
//
// @Summary      Biznesin xidmetleri
// @Tags         Public
// @Produce      json
// @Param        id       path  string true  "Biznes ID"
// @Param        staff_id query string false "Yalniz bu iscinin xidmetleri"
// @Success      200 {object} SuccessResponse
// @Router       /public/businesses/{id}/services [get]
func (h *Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	businessID, ok := h.pathID(w, r)
	if !ok {
		return
	}

	var (
		items []*serviceDomain.Service
		err   error
	)

	if raw := r.URL.Query().Get("staff_id"); raw != "" {
		staffID, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
			return
		}
		items, err = h.services.GetStaffServices(r.Context(), businessID, staffID)
	} else {
		items, err = h.services.ListServices(r.Context(), businessID)
	}

	if err != nil {
		h.writeInternal(w, err)
		return
	}

	cards := make([]ServiceCard, 0, len(items))
	for _, item := range items {
		if !item.IsActive {
			continue
		}
		cards = append(cards, ServiceCard{
			ID:              item.ID,
			Name:            item.Name,
			Description:     item.Description,
			DurationMinutes: item.DurationMinutes,
			Price:           item.Price,
		})
	}

	writeSuccess(w, http.StatusOK, "", cards)
}

// ============================================================
// BOS VAXTLAR
// ============================================================

// GetAvailability – GET /api/v1/public/availability
//
// Qorunan /availability endpoint-i biznes konteksti JWT-den alir;
// burada ise musterinin biznesi olmadigi ucun business_id acıq verilir.
//
// @Summary      Bos vaxtlar (auth telem olunmur)
// @Tags         Public
// @Produce      json
// @Param        business_id query string true  "Biznes ID"
// @Param        staff_id    query string true  "Isci ID"
// @Param        service_id  query string false "Xidmet ID"
// @Param        from        query string false "YYYY-MM-DD"
// @Param        to          query string false "YYYY-MM-DD"
// @Success      200 {object} SuccessResponse
// @Router       /public/availability [get]
func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	businessID, err := uuid.Parse(query.Get("business_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BUSINESS_ID", "business_id duzgun deyil")
		return
	}

	staffID, err := uuid.Parse(query.Get("staff_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
		return
	}

	var serviceID *uuid.UUID
	if raw := query.Get("service_id"); raw != "" {
		parsed, parseErr := uuid.Parse(raw)
		if parseErr != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SERVICE_ID", "service_id duzgun deyil")
			return
		}
		serviceID = &parsed
	}

	from, err := parseDate(query.Get("from"), time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FROM", "from tarixi YYYY-MM-DD olmalidir")
		return
	}

	to, err := parseDate(query.Get("to"), from.AddDate(0, 0, 6))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TO", "to tarixi YYYY-MM-DD olmalidir")
		return
	}

	result, err := h.availability.GetAvailability(r.Context(), businessID, &availabilityDomain.AvailabilityQuery{
		StaffID:   staffID,
		ServiceID: serviceID,
		FromDate:  from,
		ToDate:    to,
	})
	if err != nil {
		var domainErr *availabilityDomain.Error
		if errors.As(err, &domainErr) {
			writeError(w, http.StatusBadRequest, domainErr.Code, domainErr.Message)
			return
		}
		h.writeInternal(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", result)
}

// ============================================================
// HELPERS
// ============================================================

func (h *Handler) pathID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	businessID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "biznes id duzgun deyil")
		return uuid.Nil, false
	}
	return businessID, true
}

func (h *Handler) writeInternal(w http.ResponseWriter, err error) {
	h.log.Error("Public endpoint xetasi", logger.Field{Key: "error", Value: err.Error()})
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "daxili xeta")
}

func parseDate(raw string, fallback time.Time) (time.Time, error) {
	if raw == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", raw)
}

// ============================================================
// RESPONSE FORMATI
// ============================================================

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	writeJSON(w, status, SuccessResponse{Success: true, Message: message, Data: data})
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Success: false, Code: code, Message: message})
}
