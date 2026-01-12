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
