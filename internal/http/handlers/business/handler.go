package business

import (
	"encoding/json"
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
)

type Handler struct {
	service *business.Service
}

func NewHandler(service *business.Service) *Handler {
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

func (h *Handler) CreateBusiness(w http.ResponseWriter, r *http.Request) {
	var request CreateBusinessRequest
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
	}
	result, err := h.service.CreateBusiness(r.Context(), request.Name, request.Phone)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Failed to create business", err.Error())
	}
	writeJSON(w, http.StatusCreated, ToResponse(result))
}

func (h *Handler) GetBusinessByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid UUID format", nil)
	}
	result, err := h.service.GetBusinessByID(r.Context(), id)
	if err != nil {
		if err.Error() == "business not found" {
			writeJSONError(w, http.StatusNotFound, "Business not found", err.Error())
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Internal server error", nil)
		return
	}
	writeJSON(w, http.StatusOK, ToResponse(result))
}
func (h *Handler) CreateDefaultLocation(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	businessID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid Business ID", nil)
		return
	}
	locationID, err := h.service.CreateDefaultLocation(r.Context(), businessID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to create Location", err.Error())
		return
	}
	response := map[string]string{
		"message":     "Default Location created",
		"location_id": locationID,
		"business_id": businessID.String(),
	}
	writeJSON(w, http.StatusCreated, response)
}
