// File: internal/http/handlers/service/handler.go
package service

import (
	"encoding/json"
	"fmt"
	"net/http"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/service"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/google/uuid"
)

type Handler struct {
	service domain.ServiceUseCase
}

func NewHandler(service domain.ServiceUseCase) Handler {
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
	if !ok || businessID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid business id in context")
	}
	return businessID, nil
}

// @Summary      List All Services
// @Description  Retrieves all services offered by the authenticated business. Returns complete list of active and inactive services with pricing, duration, and staff assignment information.
// @Tags         Service
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse "Services retrieved successfully (array of ServiceHTTPResponse)"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/services [get]
func (h Handler) ListServices(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	services, err := h.service.ListServices(r.Context(), businessID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list services", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainServices(services),
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Get Service by ID
// @Description  Retrieves specific service details by service ID. Service must belong to authenticated business. Returns service name, description, duration, price, and staff assignments.
// @Tags         Service
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Service ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Service details retrieved successfully"
// @Failure      400  {object}  ErrorResponse "Invalid service ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Service not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/services/{id} [get]
func (h Handler) GetService(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	svcID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid service ID", err.Error())
		return
	}

	svc, err := h.service.GetService(r.Context(), svcID, businessID)
	if err != nil {
		if se, ok := err.(*domain.ServiceError); ok && se.Code == "NOT_FOUND" {
			writeJSONError(w, http.StatusNotFound, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to get service", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainService(svc),
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Update Service
// @Description  Updates service details for authenticated business. Supports partial updates - only provided fields are modified. Service must belong to business. Updates name, description, duration, pricing, and status.
// @Tags         Service
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Service ID (UUID format)"
// @Param        request body UpdateServiceHTTPRequest true "Service update data (all fields optional - Name, Description, Duration, Price, IsActive)"
// @Success      200  {object}  SuccessResponse "Service updated successfully"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid field values or invalid service ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Service not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/services/{id} [put]
func (h Handler) UpdateService(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	svcID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid service ID", err.Error())
		return
	}

	var req UpdateServiceHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	domainReq := ToDomainUpdateServiceRequest(req)

	if err := h.service.UpdateService(r.Context(), svcID, businessID, domainReq); err != nil {
		if se, ok := err.(*domain.ServiceError); ok {
			status := http.StatusBadRequest
			if se.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeJSONError(w, status, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to update service", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Service updated successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Deactivate Service
// @Description  Soft-deletes a service by marking it as inactive. Service is not permanently deleted - remains in database for historical records. Deactivated services cannot be booked. Service must belong to authenticated business.
// @Tags         Service
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Service ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Service deactivated successfully"
// @Failure      400  {object}  ErrorResponse "Validation error or invalid service ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Service not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/services/{id} [delete]
func (h Handler) DeactivateService(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	svcID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid service ID", err.Error())
		return
	}

	if err := h.service.DeactivateService(r.Context(), svcID, businessID); err != nil {
		if se, ok := err.(*domain.ServiceError); ok {
			status := http.StatusBadRequest
			if se.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeJSONError(w, status, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to deactivate service", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Service deactivated successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Assign Services to Staff
// @Description  Assigns multiple services to a staff member. Staff member becomes available to book appointments for assigned services. All services must exist and belong to authenticated business. Staff member must belong to business.
// @Tags         Service
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id path string true "Staff Member ID (UUID format)"
// @Param        request body AssignServicesHTTPRequest true "Service IDs to assign (ServiceIDs array of UUID strings)"
// @Success      200  {object}  SuccessResponse "Services assigned to staff successfully"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid staff ID format, invalid service IDs, or staff/service not found"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/{staff_id}/services [post]
func (h Handler) AssignServicesToStaff(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	staffStr := r.PathValue("staff_id")
	staffID, err := uuid.Parse(staffStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid staff ID", err.Error())
		return
	}

	var req AssignServicesHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	serviceIDs, err := ParseServiceIDs(req.ServiceIDs)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	if err := h.service.AssignServicesToStaff(r.Context(), businessID, staffID, serviceIDs); err != nil {
		if se, ok := err.(*domain.ServiceError); ok {
			writeJSONError(w, http.StatusBadRequest, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to assign services to staff", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Services assigned to staff successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Get Staff Services
// @Description  Retrieves all services assigned to a specific staff member. Returns services the staff member is qualified and available to provide. Staff member must belong to authenticated business.
// @Tags         Service
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id path string true "Staff Member ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Staff services retrieved successfully (array of ServiceHTTPResponse)"
// @Failure      400  {object}  ErrorResponse "Invalid staff ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/{staff_id}/services [get]
func (h Handler) GetStaffServices(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	staffStr := r.PathValue("staff_id")
	staffID, err := uuid.Parse(staffStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid staff ID", err.Error())
		return
	}

	services, err := h.service.GetStaffServices(r.Context(), businessID, staffID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to get staff services", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainServices(services),
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Remove Service from Staff
// @Description  Removes a specific service from a staff member's service list. Staff member becomes unavailable to book appointments for this service. Service assignment removed but service and staff member remain active.
// @Tags         Service
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        staff_id path string true "Staff Member ID (UUID format)"
// @Param        service_id path string true "Service ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Service removed from staff successfully"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid staff/service ID format or service not assigned to staff"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/{staff_id}/services/{service_id} [delete]
func (h Handler) RemoveServiceFromStaff(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	staffStr := r.PathValue("staff_id")
	serviceStr := r.PathValue("service_id")

	staffID, err := uuid.Parse(staffStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid staff ID", err.Error())
		return
	}

	serviceID, err := uuid.Parse(serviceStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid service ID", err.Error())
		return
	}

	if err := h.service.RemoveServiceFromStaff(r.Context(), businessID, staffID, serviceID); err != nil {
		if se, ok := err.(*domain.ServiceError); ok {
			writeJSONError(w, http.StatusBadRequest, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to remove service from staff", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Service removed from staff successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}
