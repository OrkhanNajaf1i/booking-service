// File: internal/http/handlers/notification/handler.go
package notification

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
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

// List – GET /api/v1/notifications?unread=true&limit=20&offset=0
//
// @Summary      Bildirisler
// @Tags         Notification
// @Produce      json
// @Security     BearerAuth
// @Param        unread query bool false "Yalniz oxunmamislar"
// @Param        limit  query int  false "Default 20"
// @Param        offset query int  false "Default 0"
// @Success      200 {object} SuccessResponse
// @Router       /notifications [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}

	unreadOnly := r.URL.Query().Get("unread") == "true"
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	rows, err := h.service.List(r.Context(), userID, unreadOnly, limit, offset)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	unread, err := h.service.CountUnread(r.Context(), userID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", map[string]interface{}{
		"items":        rows,
		"unread_count": unread,
	})
}

// CountUnread – GET /api/v1/notifications/unread-count
//
// @Summary      Oxunmamis bildiris sayi
// @Tags         Notification
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse
// @Router       /notifications/unread-count [get]
func (h *Handler) CountUnread(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}

	count, err := h.service.CountUnread(r.Context(), userID)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "", map[string]int{"unread_count": count})
}

// MarkRead – POST /api/v1/notifications/{id}/read
//
// @Summary      Bildirisi oxunmus isarele
// @Tags         Notification
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Notification ID"
// @Success      200 {object} SuccessResponse
// @Router       /notifications/{id}/read [post]
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}

	notificationID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "id duzgun deyil")
		return
	}

	if err := h.service.MarkRead(r.Context(), userID, notificationID); err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Oxunmus isarelendi", nil)
}

// MarkAllRead – POST /api/v1/notifications/read-all
//
// @Summary      Hamisini oxunmus isarele
// @Tags         Notification
// @Produce      json
// @Security     BearerAuth
// @Success      200 {object} SuccessResponse
// @Router       /notifications/read-all [post]
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}

	if err := h.service.MarkAllRead(r.Context(), userID); err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Hamisi oxunmus isarelendi", nil)
}

// RegisterDevice – POST /api/v1/notifications/devices
//
// @Summary      Push ucun cihaz token-i qeyd et
// @Tags         Notification
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body domain.RegisterDeviceRequest true "FCM token"
// @Success      201 {object} SuccessResponse
// @Router       /notifications/devices [post]
func (h *Handler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	userID, ok := h.userID(w, r)
	if !ok {
		return
	}

	var req domain.RegisterDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_BODY", "JSON oxunmadi")
		return
	}

	saved, err := h.service.RegisterDevice(r.Context(), userID, &req)
	if err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, "Cihaz qeyd edildi", saved)
}

// UnregisterDevice – DELETE /api/v1/notifications/devices?token=...
//
// @Summary      Cihaz token-ini sil
// @Tags         Notification
// @Produce      json
// @Security     BearerAuth
// @Param        token query string true "FCM token"
// @Success      200 {object} SuccessResponse
// @Router       /notifications/devices [delete]
func (h *Handler) UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.userID(w, r); !ok {
		return
	}

	token := r.URL.Query().Get("token")
	if err := h.service.UnregisterDevice(r.Context(), token); err != nil {
		h.writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, "Cihaz silindi", nil)
}

// ============================================================
// HELPERS
// ============================================================

func (h *Handler) userID(w http.ResponseWriter, r *http.Request) (uuid.UUID, bool) {
	userID, err := middleware.UserIDFrom(r)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "NO_USER", "istifadeci tapilmadi")
		return uuid.Nil, false
	}
	return userID, true
}

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

	h.log.Error("Bildiris xetasi", logger.Field{Key: "error", Value: err.Error()})
	writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "daxili xeta")
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
