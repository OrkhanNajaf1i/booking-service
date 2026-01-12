// File: internal/http/handlers/staff/handler.go
package staff

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/staff"
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
	if !ok || businessID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid business id in context")
	}
	return businessID, nil
}

func getUserIDFromContext(r *http.Request) (uuid.UUID, error) {
	v := r.Context().Value(middleware.UserIDKey)
	if v == nil {
		return uuid.Nil, fmt.Errorf("user id missing in context")
	}
	userID, ok := v.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		return uuid.Nil, fmt.Errorf("invalid user id in context")
	}
	return userID, nil
}

// @Summary      Create Staff Profile
// @Description  Creates a new staff member profile for multi-staff business. Used when directly adding staff without invitation flow (admin creates profile). Staff member becomes available for service assignment. Only available for multi-staff businesses.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateStaffHTTPRequest true "Staff profile data (FirstName, LastName, Email, Phone, Specializations - optional)"
// @Success      201  {object}  SuccessResponse "Staff profile created successfully with generated UUID"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid or missing required fields"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff [post]
func (h Handler) CreateStaffProfile(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var req CreateStaffHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	domainReq, err := ToDomainCreateStaffRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	profile, err := h.service.CreateStaffProfile(r.Context(), businessID, domainReq)
	if err != nil {
		if se, ok := err.(*domain.StaffError); ok {
			writeJSONError(w, http.StatusBadRequest, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to create staff profile", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainStaffProfile(profile),
		Message: "Staff profile created successfully",
	}
	writeJSON(w, http.StatusCreated, resp)
}

// @Summary      List All Staff Members
// @Description  Retrieves all staff members for the authenticated business. Returns complete list of active and inactive staff with user associations, specializations, and availability status.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  SuccessResponse "Staff list retrieved successfully (array of StaffWithUserHTTPResponse)"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff [get]
func (h Handler) ListStaff(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	list, err := h.service.ListStaff(r.Context(), businessID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to list staff", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainStaffWithUser(list),
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Get Staff Member Details
// @Description  Retrieves specific staff member details by staff ID. Includes services assigned, specializations, user account info, and availability status. Staff member must belong to authenticated business.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Staff ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Staff details retrieved successfully"
// @Failure      400  {object}  ErrorResponse "Invalid staff ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Staff member not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/{id} [get]
func (h Handler) GetStaff(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	staffID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid staff ID", err.Error())
		return
	}

	profile, err := h.service.GetStaff(r.Context(), staffID, businessID)
	if err != nil {
		if se, ok := err.(*domain.StaffError); ok && se.Code == "NOT_FOUND" {
			writeJSONError(w, http.StatusNotFound, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to get staff", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainStaffProfile(profile),
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Update Staff Member
// @Description  Updates staff member details for authenticated business. Supports partial updates - only provided fields are modified. Updates name, email, phone, specializations, and availability status. Staff member must belong to business.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Staff ID (UUID format)"
// @Param        request body UpdateStaffHTTPRequest true "Staff update data (all fields optional - FirstName, LastName, Email, Phone, Specializations, IsActive)"
// @Success      200  {object}  SuccessResponse "Staff member updated successfully"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid field values or invalid staff ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Staff member not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/{id} [put]
func (h Handler) UpdateStaff(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	staffID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid staff ID", err.Error())
		return
	}

	var req UpdateStaffHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	domainReq, err := ToDomainUpdateStaffRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	if err := h.service.UpdateStaff(r.Context(), staffID, businessID, domainReq); err != nil {
		if se, ok := err.(*domain.StaffError); ok {
			status := http.StatusBadRequest
			if se.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeJSONError(w, status, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to update staff", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Staff updated successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Deactivate Staff Member
// @Description  Soft-deletes a staff member by marking as inactive. Staff member is not permanently deleted - remains in database for historical booking records. Deactivated staff cannot book new appointments. Staff member must belong to authenticated business.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Staff ID (UUID format)"
// @Success      200  {object}  SuccessResponse "Staff member deactivated successfully"
// @Failure      400  {object}  ErrorResponse "Validation error or invalid staff ID format"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      404  {object}  ErrorResponse "Staff member not found"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/{id} [delete]
func (h Handler) DeactivateStaff(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	idStr := r.PathValue("id")
	staffID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid staff ID", err.Error())
		return
	}

	if err := h.service.DeactivateStaff(r.Context(), staffID, businessID); err != nil {
		if se, ok := err.(*domain.StaffError); ok {
			status := http.StatusBadRequest
			if se.Code == "NOT_FOUND" {
				status = http.StatusNotFound
			}
			writeJSONError(w, status, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to deactivate staff", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Staff deactivated successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Invite Staff Member
// @Description  Sends invitation to join business as staff member (zero-knowledge invitation flow). Generates unique invitation token valid for 7 days. Invitee receives token and can accept by creating password and joining business. Used for multi-staff businesses to onboard team members.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body InviteStaffHTTPRequest true "Invitation details (FirstName, LastName, Email, Phone, Role - provider_owner, staff, customer)"
// @Success      201  {object}  SuccessResponse "Invitation created successfully with token and expiration"
// @Failure      400  {object}  ErrorResponse "Validation error - invalid email format or staff already invited"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or business_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/invites [post]
func (h Handler) InviteStaff(w http.ResponseWriter, r *http.Request) {
	businessID, err := getBusinessIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var req InviteStaffHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	domainReq, err := ToDomainInviteStaffRequest(req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	token, err := h.service.InviteStaff(r.Context(), businessID, domainReq)
	if err != nil {
		if se, ok := err.(*domain.StaffError); ok {
			writeJSONError(w, http.StatusBadRequest, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to create invite", err.Error())
		return
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	resp := SuccessResponse{
		Success: true,
		Data: InviteResponse{
			Token:     token,
			ExpiresAt: expiresAt,
		},
		Message: "Invite created successfully",
	}
	writeJSON(w, http.StatusCreated, resp)
}

// @Summary      Validate Invitation Token
// @Description  Validates staff invitation token before acceptance. Returns invitation details if token is valid and not expired. No authentication required - used during invitation onboarding flow. Token must be valid and within 7-day expiration window.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Param        request body ValidateInviteHTTPRequest true "Invitation token to validate"
// @Success      200  {object}  SuccessResponse "Token is valid with invitation details (FirstName, LastName, Email, BusinessName)"
// @Failure      400  {object}  ErrorResponse "Invalid, expired, or already-used invitation token"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/invites/validate [post]
func (h Handler) ValidateInviteToken(w http.ResponseWriter, r *http.Request) {
	var req ValidateInviteHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	invite, err := h.service.ValidateInviteToken(r.Context(), req.Token)
	if err != nil {
		if se, ok := err.(*domain.StaffError); ok {
			writeJSONError(w, http.StatusBadRequest, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to validate invite", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Data:    FromDomainInviteDetails(invite),
	}
	writeJSON(w, http.StatusOK, resp)
}

// @Summary      Accept Invitation (Staff Onboarding)
// @Description  Completes staff member onboarding by accepting invitation and setting password. User must be authenticated (user_id from context). Uses zero-knowledge invitation flow - staff sets their own password without server storing temporary passwords. Staff becomes active member after acceptance. Token is marked as used and cannot be reused.
// @Tags         Staff
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body AcceptInviteHTTPRequest true "Invitation token and password for account setup"
// @Success      200  {object}  SuccessResponse "Invitation accepted successfully, staff member activated"
// @Failure      400  {object}  ErrorResponse "Invalid or expired token, password validation failure"
// @Failure      401  {object}  ErrorResponse "Unauthorized - user not authenticated or user_id missing"
// @Failure      500  {object}  ErrorResponse "Internal server error"
// @Router       /api/v1/staff/invites/accept [post]
func (h Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	userID, err := getUserIDFromContext(r)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "Unauthorized", err.Error())
		return
	}

	var req AcceptInviteHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	if err := h.service.AcceptInvite(r.Context(), userID, req.Token, req.Password); err != nil {
		if se, ok := err.(*domain.StaffError); ok {
			writeJSONError(w, http.StatusBadRequest, se.Message, se.Code)
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Failed to accept invite", err.Error())
		return
	}

	resp := SuccessResponse{
		Success: true,
		Message: "Invite accepted successfully",
	}
	writeJSON(w, http.StatusOK, resp)
}
