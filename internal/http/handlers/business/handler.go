package business

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/google/uuid"
)

//	type Handler struct {
//		service *business.Service
//	}
type Handler struct {
	service business.BusinessService
}

func NewHandler(service business.BusinessService) *Handler {
	return &Handler{
		service: service,
	}
}
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "JSON encode error", http.StatusInternalServerError)
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string, details interface{}) {
	response := map[string]interface{}{
		"error": message,
	}
	if details != nil {
		response["details"] = details
	}
	writeJSON(w, status, response)
}

func getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	v := r.Context().Value(middleware.UserIDKey)

	userID, ok := v.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("user_id missing in context (JWT required)")
	}

	return userID, nil
}

func (h *Handler) CreateSoloBusiness(w http.ResponseWriter, r *http.Request) {
	ownerID, err := getUserIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	var req CreateSoloBusinessHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	domainReq := ToDomainCreateBusinessRequest(req.Name, req.Phone, req.Industry, req.ServiceCategory, business.BusinessTypeSolo)
	businessID, err := h.service.CreateSoloBusiness(r.Context(), ownerID, domainReq)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "failed to create solo business", err.Error())
	}
	response := SuccessResponse{
		Success: true,
		Data: map[string]string{
			"business_id": businessID.String(),
			"message":     "Solo business created successfully",
		},
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) CreateMultiBusiness(w http.ResponseWriter, r *http.Request) {
	ownerID, err := getUserIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}
	var req CreateMultiBusinessHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	domainReq := ToDomainCreateBusinessRequest(
		req.Name,
		req.Phone,
		req.Industry,
		req.ServiceCategory,
		business.BusinessTypeMulti,
	)
	businessID, err := h.service.CreateMultiBusiness(r.Context(), ownerID, domainReq)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to create multi business", err.Error())
		return
	}
	response := SuccessResponse{
		Success: true,
		Data: map[string]string{
			"business_id": businessID.String(),
			"message":     "Multi-staff business created successfully",
		},
	}
	writeJSON(w, http.StatusCreated, response)
}
func (h *Handler) InviteStaff(w http.ResponseWriter, r *http.Request) {
	businessIDStr := r.PathValue("id")
	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid business ID", err.Error())
	}
	var req InviteStaffHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	domainReq, err := ToDomainInviteStaffRequest(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err)
		return
	}
	token, err := h.service.InviteStaff(r.Context(), businessID, domainReq)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to create invite", err.Error())
		return
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	response := SuccessResponse{
		Success: true,
		Data:    ToInviteResponse(token, expiresAt),
	}
	writeJSON(w, http.StatusCreated, response)
}

func (h *Handler) JoinWithInvite(w http.ResponseWriter, r *http.Request) {
	var req JoinWithInviteHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
	}
	if err := h.service.JoinWithInvite(r.Context(), req.Token, req.Password); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to join business", err.Error())
	}
	response := SuccessResponse{
		Success: true,
		Message: "Successfully joined business",
	}
	writeJSON(w, http.StatusOK, response)
}
func (h *Handler) CreateDefaultLocation(w http.ResponseWriter, r *http.Request) {
	businessIDStr := r.PathValue("id")
	businessID, err := uuid.Parse(businessIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid business ID", nil)
		return
	}

	locationID, err := h.service.CreateDefaultLocation(r.Context(), businessID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to create location", err.Error())
		return
	}

	response := SuccessResponse{
		Success: true,
		Data: map[string]string{
			"location_id": locationID.String(),
			"business_id": businessID.String(),
			"message":     "Default location created",
		},
	}
	writeJSON(w, http.StatusCreated, response)
}
