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

// @Summary      Create Location
// @Description  Creates a new location for the authenticated business. Location is address where services are provided. Each business can have multiple locations (main office, branch, etc).
// @Tags         Location
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateLocationHTTPRequest true "Location data (Name, Address, City, State, Country, PostalCode, Latitude, Longitude, PhoneNumber)"
// @Success      201  {object}  SuccessResponse "Location created successfully with generated UUID"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid or missing required fields"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/locations [post]
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

// @Summary      List All Locations
// @Description  Retrieves all locations for the authenticated business. Returns complete list of active and inactive locations with full details.
// @Tags         Location
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse "Locations retrieved successfully (array of LocationHTTPResponse)"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/locations [get]
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

// @Summary      Get Location by ID
// @Description  Retrieves specific location details by location ID. Location must belong to authenticated business (multi-tenancy filtering applied).
// @Tags         Location
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Location ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Location details retrieved successfully"
// @Failure      400  {object}  ErrorResponse "Invalid location ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Location not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/locations/{id} [get]
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

// @Summary      Update Location
// @Description  Updates location details for authenticated business. Supports partial updates - only provided fields are updated. Location must belong to business (multi-tenancy filtering applied).
// @Tags         Location
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Location ID (UUID format)"
// @Param        request body UpdateLocationHTTPRequest true "Location update data (all fields optional - Name, Address, City, State, Country, PostalCode, Latitude, Longitude, PhoneNumber)"
// @Success      200  {object}  SuccessResponse "Location updated successfully"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid field values or invalid location ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Location not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/locations/{id} [put]
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

// @Summary      Deactivate Location
// @Description  Soft-deletes a location by marking it as inactive. Location is not permanently deleted - it remains in database for historical records. Deactivated locations cannot be used for new bookings. Location must belong to authenticated business.
// @Tags         Location
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Location ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Location deactivated successfully"
// @Failure      400  {object}  ErrorResponse "Validation error or invalid location ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Location not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/locations/{id} [delete]
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
