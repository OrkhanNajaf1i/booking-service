// File: internal/http/handlers/availability/handler.go
package availability

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/availability"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

type Handler struct {
	service domain.Service
	log     logger.Logger
}

func NewHandler(service domain.Service, log logger.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// ============================================================
// AVAILABILITY (musteri terefi)
// ============================================================

// GetAvailability – GET /api/v1/availability
//
// Query: staff_id (mecburi), service_id, from (YYYY-MM-DD), to (YYYY-MM-DD)
// from verilmese bugun, to verilmese from + 6 gun goturulur.
//
// @Summary      Bos vaxtlari getir
// @Description  Isci qrafiki, nahar fasilesi, slot addimi ve movcud bronlar esasinda hesablanmis bos vaxtlar.
// @Tags         Availability
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id   query string true  "Isci ID"
// @Param        service_id query string false "Xidmet ID (mueddet bundan goturulur)"
// @Param        from       query string false "Baslangic tarixi YYYY-MM-DD"
// @Param        to         query string false "Bitis tarixi YYYY-MM-DD"
// @Success      200 {object} SuccessResponse
// @Failure      400 {object} ErrorResponse
// @Router       /availability [get]
func (h *Handler) GetAvailability(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	staffID, err := uuid.Parse(r.URL.Query().Get("staff_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
		return
	}

	var serviceID *uuid.UUID
	if raw := r.URL.Query().Get("service_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_SERVICE_ID", "service_id duzgun deyil")
			return
		}
		serviceID = &parsed
	}

	from, err := parseDate(r.URL.Query().Get("from"), time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FROM", "from tarixi YYYY-MM-DD olmalidir")
		return
	}

	to, err := parseDate(r.URL.Query().Get("to"), from.AddDate(0, 0, 6))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TO", "to tarixi YYYY-MM-DD olmalidir")
		return
	}

	result, err := h.service.GetAvailability(r.Context(), businessID, &domain.AvailabilityQuery{
		StaffID:   staffID,
		ServiceID: serviceID,
		FromDate:  from,
		ToDate:    to,
	})
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", result)
}

// ============================================================
// WORKING HOURS (admin terefi)
// ============================================================

// SetWorkingHours – POST /api/v1/availability/working-hours
//
// @Summary      Bir gunun is saatlarini teyin et
// @Description  Baslangic/bitis saati ve (istese) nahar fasilesi. break_enabled=false olsa fasile tetbiq edilmir.
// @Tags         Availability
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.SetWorkingHoursRequest true "Is saatlari"
// @Success      200 {object} SuccessResponse
// @Router       /availability/working-hours [post]
func (h *Handler) SetWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	var req domain.SetWorkingHoursRequest
	if !decodeBody(w, r, &req) {
		return
	}

	saved, err := h.service.SetWorkingHours(r.Context(), businessID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Is saatlari yadda saxlanildi", saved)
}

// BulkSetWorkingHours – PUT /api/v1/availability/working-hours
// Butun hefteni bir sorguda yazir.
//
// @Summary      Heftelik qrafiki yaz
// @Tags         Availability
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.BulkWorkingHoursRequest true "Heftelik qrafik"
// @Success      200 {object} SuccessResponse
// @Router       /availability/working-hours [put]
func (h *Handler) BulkSetWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	var req domain.BulkWorkingHoursRequest
	if !decodeBody(w, r, &req) {
		return
	}

	saved, err := h.service.BulkSetWorkingHours(r.Context(), businessID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Heftelik qrafik yadda saxlanildi", saved)
}

// ListWorkingHours – GET /api/v1/availability/working-hours?staff_id=...
//
// @Summary      Iscinin heftelik qrafiki
// @Tags         Availability
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id query string true "Isci ID"
// @Success      200 {object} SuccessResponse
// @Router       /availability/working-hours [get]
func (h *Handler) ListWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	staffID, err := uuid.Parse(r.URL.Query().Get("staff_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
		return
	}

	rows, err := h.service.ListWorkingHours(r.Context(), businessID, staffID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", rows)
}

// DeleteWorkingHours – DELETE /api/v1/availability/working-hours?staff_id=...&day_of_week=...
//
// @Summary      Bir gunun qrafikini sil
// @Tags         Availability
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id    query string true "Isci ID"
// @Param        day_of_week query int    true "0=Bazar ... 6=Senbe"
// @Success      200 {object} SuccessResponse
// @Router       /availability/working-hours [delete]
func (h *Handler) DeleteWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	staffID, err := uuid.Parse(r.URL.Query().Get("staff_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
		return
	}

	dayOfWeek, err := strconv.Atoi(r.URL.Query().Get("day_of_week"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_DAY", "day_of_week reqem olmalidir")
		return
	}

	if err := h.service.DeleteWorkingHours(r.Context(), businessID, staffID, dayOfWeek); err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Qrafik silindi", nil)
}

// ============================================================
// SCHEDULE SETTINGS
// ============================================================

// GetSettings – GET /api/v1/availability/settings?staff_id=...
// staff_id verilmese biznesin default ayari qaytarilir.
//
// @Summary      Qrafik ayarlarini getir
// @Tags         Availability
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id query string false "Isci ID (bos olsa biznes default-u)"
// @Success      200 {object} SuccessResponse
// @Router       /availability/settings [get]
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	var staffID *uuid.UUID
	if raw := r.URL.Query().Get("staff_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
			return
		}
		staffID = &parsed
	}

	settings, err := h.service.GetSettings(r.Context(), businessID, staffID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", settings)
}

// UpdateSettings – PUT /api/v1/availability/settings
//
// @Summary      Qrafik ayarlarini yenile
// @Description  slot_step_mins secim addimini teyin edir (mes. 16 deqiqe). auto_confirm true olsa bron tesdiq gozlemeden confirmed olur.
// @Tags         Availability
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.UpdateScheduleSettingsRequest true "Ayarlar"
// @Success      200 {object} SuccessResponse
// @Router       /availability/settings [put]
func (h *Handler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	var req domain.UpdateScheduleSettingsRequest
	if !decodeBody(w, r, &req) {
		return
	}

	settings, err := h.service.UpdateSettings(r.Context(), businessID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Ayarlar yenilendi", settings)
}

// ============================================================
// TIME OFF
// ============================================================

// CreateTimeOff – POST /api/v1/availability/time-off
//
// @Summary      Bloklanmis interval elave et
// @Tags         Availability
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.CreateTimeOffRequest true "Interval"
// @Success      201 {object} SuccessResponse
// @Router       /availability/time-off [post]
func (h *Handler) CreateTimeOff(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	var req domain.CreateTimeOffRequest
	if !decodeBody(w, r, &req) {
		return
	}

	created, err := h.service.CreateTimeOff(r.Context(), businessID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, "Interval bloklandi", created)
}

// ListTimeOff – GET /api/v1/availability/time-off?staff_id=&from=&to=
//
// @Summary      Bloklanmis intervallar
// @Tags         Availability
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id query string true  "Isci ID"
// @Param        from     query string false "YYYY-MM-DD"
// @Param        to       query string false "YYYY-MM-DD"
// @Success      200 {object} SuccessResponse
// @Router       /availability/time-off [get]
func (h *Handler) ListTimeOff(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	staffID, err := uuid.Parse(r.URL.Query().Get("staff_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "staff_id duzgun deyil")
		return
	}

	from, err := parseDate(r.URL.Query().Get("from"), time.Now())
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_FROM", "from tarixi YYYY-MM-DD olmalidir")
		return
	}
	to, err := parseDate(r.URL.Query().Get("to"), from.AddDate(0, 0, 30))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_TO", "to tarixi YYYY-MM-DD olmalidir")
		return
	}

	rows, err := h.service.ListTimeOff(r.Context(), businessID, staffID, from, to)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", rows)
}

// DeleteTimeOff – DELETE /api/v1/availability/time-off/{id}
//
// @Summary      Bloklanmis intervali sil
// @Tags         Availability
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Time off ID"
// @Success      200 {object} SuccessResponse
// @Router       /availability/time-off/{id} [delete]
func (h *Handler) DeleteTimeOff(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.businessID(w, r)
	if err != nil {
		return
	}

	timeOffID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "id duzgun deyil")
		return
	}

	if err := h.service.DeleteTimeOff(r.Context(), businessID, timeOffID); err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Interval silindi", nil)
}

// ============================================================
// HELPERS
// ============================================================

func (h *Handler) businessID(w http.ResponseWriter, r *http.Request) (uuid.UUID, error) {
	businessID, err := middleware.BusinessIDFrom(r)
	if err != nil {
		writeError(w, http.StatusForbidden, "NO_BUSINESS", "business konteksti tapilmadi")
		return uuid.Nil, err
	}
	return businessID, nil
}

// writeDomainError – domain xetalarini uygun HTTP statusuna cevirir.
func (h *Handler) writeDomainError(w http.ResponseWriter, err error) {
	var domainErr *domain.Error
	if errors.As(err, &domainErr) {
		status := http.StatusBadRequest
		if domainErr.Code == "NOT_FOUND" {
			status = http.StatusNotFound
		}
		writeError(w, status, domainErr.Code, domainErr.Message)
		return
	}

	h.log.Error("Availability xetasi", logger.Field{Key: "error", Value: err.Error()})
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "daxili xeta")
}

func decodeBody(w http.ResponseWriter, r *http.Request, target interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "JSON oxunmadi")
		return false
	}
	return true
}

// parseDate – bos olarsa fallback qaytarir.
func parseDate(raw string, fallback time.Time) (time.Time, error) {
	if raw == "" {
		return fallback, nil
	}
	return time.Parse("2006-01-02", raw)
}
