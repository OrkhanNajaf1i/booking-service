// File: internal/http/handlers/slot_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/slot"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
)

// ============================================
// SLOT HANDLER
// ============================================

// SlotHandler - Slot HTTP handlers (Standard net/http compatible)
type SlotHandler struct {
	service slot.Service
	logger  logger.Logger
}

// NewSlotHandler - Handler instance yaratır
func NewSlotHandler(service slot.Service, logger logger.Logger) *SlotHandler {
	return &SlotHandler{
		service: service,
		logger:  logger,
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// getBusinessID - Context-dən business_id çıxarması
func (h *SlotHandler) getBusinessID(r *http.Request) (uuid.UUID, error) {
	businessIDStr := r.Header.Get("X-Business-ID")
	if businessIDStr == "" {
		return uuid.Nil, fmt.Errorf("missing X-Business-ID header")
	}

	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid business_id format: %w", err)
	}

	return businessID, nil
}

// writeJSON - Response JSON yazması
func (h *SlotHandler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("writeJSON: Encoding error",
			logger.Field{Key: "error", Value: err.Error()},
		)
	}
}

// writeError - Error response yazması
func (h *SlotHandler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

// ============================================
// CREATE ENDPOINTS
// ============================================

// CreateSlot - POST /slots
// Handler: Yeni slot yaratmaq
func (h *SlotHandler) CreateSlot(w http.ResponseWriter, r *http.Request) {
	// Get business ID
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.logger.Warn("CreateSlot: Missing business_id",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Parse request body
	var req CreateSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("CreateSlot: Invalid request body",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Convert to domain request
	domainReq := req.ToEntity()

	// Call service
	createdSlot, err := h.service.CreateSlot(r.Context(), businessID, domainReq)
	if err != nil {
		h.logger.Error("CreateSlot: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "CREATE_SLOT_ERROR", err.Error())
		return
	}

	// Convert to response
	response := FromSlotEntity(createdSlot)
	h.writeJSON(w, http.StatusCreated, response)
}

// SetWorkingHours - POST /staff/{staff_id}/working-hours
// Handler: İş saatları təyin etmək
func (h *SlotHandler) SetWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Parse request body
	var req SetWorkingHoursRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("SetWorkingHours: Invalid request body",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Convert to domain request
	domainReq := req.ToEntity()

	// Call service
	wh, err := h.service.SetWorkingHours(r.Context(), businessID, domainReq)
	if err != nil {
		h.logger.Error("SetWorkingHours: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "SET_WORKING_HOURS_ERROR", err.Error())
		return
	}

	response := FromWorkingHoursEntity(wh)
	h.writeJSON(w, http.StatusCreated, response)
}

// ============================================
// READ ENDPOINTS
// ============================================

// GetSlot - GET /slots/{id}
// Handler: ID-dən slot tap
func (h *SlotHandler) GetSlot(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Get slot ID from URL path
	// Note: Implementation depends on router. Example with net/http:
	// slotID := r.PathValue("id") - Go 1.22+
	// OR use chi/gorilla/mux router
	slotIDStr := r.URL.Query().Get("id")
	if slotIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_SLOT_ID", "slot id is required")
		return
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_SLOT_ID", "invalid slot_id format")
		return
	}

	// Call service
	slotObj, err := h.service.GetSlot(r.Context(), businessID, slotID)
	if err != nil {
		h.logger.Warn("GetSlot: Not found",
			logger.Field{Key: "slot_id", Value: slotIDStr},
		)
		h.writeError(w, http.StatusNotFound, "SLOT_NOT_FOUND", err.Error())
		return
	}

	response := FromSlotEntity(slotObj)
	h.writeJSON(w, http.StatusOK, response)
}

// ListSlots - GET /slots
// Handler: Slot-ları list etmək
func (h *SlotHandler) ListSlots(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Parse query parameters
	query := &ListSlotsQuery{
		Page:     1,
		PageSize: 20,
	}

	// Parse pagination
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			query.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 {
			query.PageSize = pageSize
		}
	}

	// Parse filters
	if staffIDStr := r.URL.Query().Get("staff_id"); staffIDStr != "" {
		if staffID, err := uuid.Parse(staffIDStr); err == nil {
			query.StaffID = &staffID
		}
	}

	if locationIDStr := r.URL.Query().Get("location_id"); locationIDStr != "" {
		if locationID, err := uuid.Parse(locationIDStr); err == nil {
			query.LocationID = &locationID
		}
	}

	if status := r.URL.Query().Get("status"); status != "" {
		query.Status = &status
	}

	// Convert to domain query
	domainQuery := query.ToEntity()

	// Call service
	slots, err := h.service.ListSlots(r.Context(), businessID, domainQuery)
	if err != nil {
		h.logger.Error("ListSlots: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "LIST_SLOTS_ERROR", err.Error())
		return
	}

	// Convert to responses
	responses := FromSlotEntities(slots)

	listResponse := ListSlotsResponse{
		Slots:      responses,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalCount: len(slots),
	}

	h.writeJSON(w, http.StatusOK, listResponse)
}

// GetAvailableSlots - GET /available-slots
// Handler: Müsait slot-ları tap
func (h *SlotHandler) GetAvailableSlots(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Parse required parameters
	staffIDStr := r.URL.Query().Get("staff_id")
	locationIDStr := r.URL.Query().Get("location_id")

	if staffIDStr == "" || locationIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_PARAMS", "staff_id and location_id are required")
		return
	}

	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "invalid staff_id format")
		return
	}

	locationID, err := uuid.Parse(locationIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_LOCATION_ID", "invalid location_id format")
		return
	}

	// Parse pagination
	page := 1
	pageSize := 20

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	// Call service
	slots, err := h.service.GetAvailableSlots(
		r.Context(),
		businessID,
		staffID,
		locationID,
		page,
		pageSize,
	)
	if err != nil {
		h.logger.Error("GetAvailableSlots: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "GET_AVAILABLE_SLOTS_ERROR", err.Error())
		return
	}

	// Convert to responses
	responses := FromSlotEntities(slots)

	availResponse := AvailableSlotsResponse{
		Slots:      responses,
		Page:       page,
		PageSize:   pageSize,
		TotalCount: len(slots),
	}

	h.writeJSON(w, http.StatusOK, availResponse)
}

// GetStaffWorkingHours - GET /staff/{staff_id}/working-hours
// Handler: Staff iş saatları tap
func (h *SlotHandler) GetStaffWorkingHours(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Get staff ID from URL
	staffIDStr := r.URL.Query().Get("staff_id")
	if staffIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_STAFF_ID", "staff_id is required")
		return
	}

	staffID, err := uuid.Parse(staffIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_STAFF_ID", "invalid staff_id format")
		return
	}

	// Call service
	whs, err := h.service.GetStaffWorkingHours(r.Context(), businessID, staffID)
	if err != nil {
		h.logger.Error("GetStaffWorkingHours: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "GET_WORKING_HOURS_ERROR", err.Error())
		return
	}

	// Convert to responses
	responses := FromWorkingHoursEntities(whs)

	h.writeJSON(w, http.StatusOK, responses)
}

// ============================================
// UPDATE ENDPOINTS
// ============================================

// UpdateSlot - PATCH /slots/{id}
// Handler: Slot yeniləmə
func (h *SlotHandler) UpdateSlot(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Get slot ID
	slotIDStr := r.URL.Query().Get("id")
	if slotIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_SLOT_ID", "slot id is required")
		return
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_SLOT_ID", "invalid slot_id format")
		return
	}

	// Parse request
	var req UpdateSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Get existing slot
	existingSlot, err := h.service.GetSlot(r.Context(), businessID, slotID)
	if err != nil {
		h.writeError(w, http.StatusNotFound, "SLOT_NOT_FOUND", "slot not found")
		return
	}

	// Update fields
	if req.Status != nil {
		existingSlot.Status = slot.SlotStatus(*req.Status)
	}
	if req.Notes != nil {
		existingSlot.Notes = req.Notes
	}

	// Call service to update
	if err := h.service.UpdateSlot(r.Context(), businessID, existingSlot); err != nil {
		h.logger.Error("UpdateSlot: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "UPDATE_SLOT_ERROR", err.Error())
		return
	}

	response := FromSlotEntity(existingSlot)
	h.writeJSON(w, http.StatusOK, response)
}

// ============================================
// DELETE ENDPOINTS
// ============================================

// DeleteSlot - DELETE /slots/{id}
// Handler: Slot silmə
func (h *SlotHandler) DeleteSlot(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Get slot ID
	slotIDStr := r.URL.Query().Get("id")
	if slotIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_SLOT_ID", "slot id is required")
		return
	}

	slotID, err := uuid.Parse(slotIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_SLOT_ID", "invalid slot_id format")
		return
	}

	// Call service
	if err := h.service.DeleteSlot(r.Context(), businessID, slotID); err != nil {
		h.logger.Error("DeleteSlot: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "DELETE_SLOT_ERROR", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ============================================
// ADMIN/BATCH ENDPOINTS
// ============================================

// GenerateSlots - POST /admin/slots/generate
// Handler: Bulk slot generation
func (h *SlotHandler) GenerateSlots(w http.ResponseWriter, r *http.Request) {
	businessID, err := h.getBusinessID(r)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "MISSING_BUSINESS_ID", err.Error())
		return
	}

	// Parse request
	var req GenerateSlotsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// Call service
	count, err := h.service.GenerateSlots(
		r.Context(),
		businessID,
		req.StaffID,
		req.SlotDurationMs,
		req.DaysAhead,
	)
	if err != nil {
		h.logger.Error("GenerateSlots: Service error",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.writeError(w, http.StatusInternalServerError, "GENERATE_SLOTS_ERROR", err.Error())
		return
	}

	response := GenerateSlotsResponse{
		GeneratedCount: count,
		Message:        "Slots generated successfully",
	}

	h.writeJSON(w, http.StatusOK, response)
}
