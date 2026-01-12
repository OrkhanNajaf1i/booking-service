// File: internal/http/handlers/location/handler.go
package location

import (
	"encoding/json"
	"fmt"
	"net/http"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/location"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	service domain.Service
}

func NewHandler(service domain.Service) Handler {
	return Handler{service: service}
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, message string, details interface{}) {
	resp := ErrorResponse{
		Success: false,
		Error:   message,
		Details: details,
	}
	writeJSON(w, status, resp)
}

func getBusinessIDFromContext(r *http.Request) (uuid.UUID, error) {
	v := r.Context().Value(middleware.BusinessKey)
	if v == nil {
		return uuid.Nil, fmt.Errorf("business id missing in context")
	}

	businessID, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid business id in context")
	}
	if businessID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("business id is empty")
	}

	return businessID, nil
}

func (h Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var req CreateLocationHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	domainReq := ToDomainCreateRequest(req)
	loc, err := h.service.CreateLocation(r.Context(), businessID, domainReq)
	if err != nil {
		if locErr, ok := err.(*domain.LocationError); ok {
			writeJSONError(w, http.StatusBadRequest, locErr.Message, locErr.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to create location", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainLocation(loc),
		Message: "Location created successfully",
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h Handler) ListLocations(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	locs, err := h.service.ListLocations(r.Context(), businessID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list locations", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainLocations(locs),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h Handler) GetLocation(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	loc, err := h.service.GetLocation(r.Context(), locID, businessID)
	if err != nil {
		if locErr, ok := err.(*domain.LocationError); ok && locErr.Code == "NOT_FOUND" {
			writeJSONError(w, http.StatusNotFound, locErr.Message, locErr.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to get location", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainLocation(loc),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	var req UpdateLocationHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	domainReq := ToDomainUpdateRequest(req)
	if err := h.service.UpdateLocation(r.Context(), locID, businessID, domainReq); err != nil {
		if locErr, ok := err.(*domain.LocationError); ok {
			status := http.StatusBadRequest
			if locErr.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeJSONError(w, status, locErr.Message, locErr.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to update location", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Location updated successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h Handler) DeactivateLocation(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	locID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid location ID", err.Error())
		return
	}

	if err := h.service.DeactivateLocation(r.Context(), locID, businessID); err != nil {
		if locErr, ok := err.(*domain.LocationError); ok {
			status := http.StatusBadRequest
			if locErr.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeJSONError(w, status, locErr.Message, locErr.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to deactivate location", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Location deactivated successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}
