// File: internal/http/handlers/dashboard/handler.go
//
// Admin panelinin ana ekrani ucun statistika.
package dashboard

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/postgres"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// StatsReader – dashboard gostericilerini oxuyan port.
type StatsReader interface {
	GetStats(
		ctx context.Context,
		businessID uuid.UUID,
		dayStart, dayEnd, monthStart time.Time,
	) (*postgres.DashboardStats, error)
}

type Handler struct {
	stats StatsReader
	log   logger.Logger
	// defaultTZ – "bu gun"un serhedlerini hesablamaq ucun.
	defaultTZ *time.Location
}

func NewHandler(stats StatsReader, log logger.Logger, timezone string) *Handler {
	location, err := time.LoadLocation(timezone)
	if err != nil || location == nil {
		location = time.UTC
	}
	return &Handler{stats: stats, log: log, defaultTZ: location}
}

// GetStats – GET /api/v1/dashboard/stats
//
// @Summary      Dashboard gostericileri
// @Description  Bu gunun bronlari, cavab gozleyenler, cari ayin geliri ve musteri sayi.
// @Tags         Dashboard
// @Produce      json
// @Security     BearerAuth
// @Param        tz query string false "IANA timezone (default Asia/Baku)"
// @Success      200 {object} SuccessResponse
// @Failure      403 {object} ErrorResponse
// @Router       /dashboard/stats [get]
func (h *Handler) GetStats(w http.ResponseWriter, r *http.Request) {
	businessID, err := middleware.BusinessIDFrom(r)
	if err != nil {
		writeError(w, http.StatusForbidden, "NO_BUSINESS", "business konteksti tapilmadi")
		return
	}

	// "Bu gun" istifadecinin gunu olmalidir, serverin yox.
	location := h.defaultTZ
	if raw := r.URL.Query().Get("tz"); raw != "" {
		if parsed, err := time.LoadLocation(raw); err == nil {
			location = parsed
		}
	}

	now := time.Now().In(location)
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
	dayEnd := dayStart.AddDate(0, 0, 1)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)

	stats, err := h.stats.GetStats(r.Context(), businessID, dayStart, dayEnd, monthStart)
	if err != nil {
		h.log.Error("Dashboard statistikasi alinmadi",
			logger.Field{Key: "business_id", Value: businessID.String()},
			logger.Field{Key: "error", Value: err.Error()},
		)
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "daxili xeta")
		return
	}

	writeSuccess(w, http.StatusOK, "", stats)
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
