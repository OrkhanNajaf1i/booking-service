// File: internal/http/handlers/business/dto.go
package business

import (
	"fmt"
	"strings"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
)

type CreateSoloBusinessHTTPRequest struct {
	Name            string `json:"name" binding:"required"`
	Phone           string `json:"phone" binding:"required"`
	ServiceCategory string `json:"service_category" binding:"required"`
	Industry        string `json:"industry,omitempty"`
}

type CreateMultiBusinessHTTPRequest struct {
	Name            string `json:"name" binding:"required"`
	Phone           string `json:"phone" binding:"required"`
	Industry        string `json:"industry" binding:"required"`
	ServiceCategory string `json:"service_category,omitempty"`
}

type InviteStaffHTTPRequest struct {
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Role       string `json:"role" binding:"required"`
	LocationID string `json:"location_id,omitempty"`
	Title      string `json:"title,omitempty"`
}

type JoinWithInviteHTTPRequest struct {
	Token    string `json:"token" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type BusinessResponse struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Phone           string `json:"phone,omitempty"`
	OwnerID         string `json:"owner_id"`
	Industry        string `json:"industry,omitempty"`
	ServiceCategory string `json:"service_category,omitempty"`
	BusinessType    string `json:"business_type"`
	IsActive        bool   `json:"is_active"`
	CreatedAt       string `json:"created_at"`
}

type InviteResponse struct {
	Token     string `json:"token"`
	ExpiresAt string `json:"expires_at"`
	Message   string `json:"message"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   string      `json:"error"`
	Details interface{} `json:"details,omitempty"`
}

func ToDomainCreateBusinessRequest(
	name, phone, industry, serviceCategory string,
	businessType business.BusinessType,
) *business.CreateBusinessRequest {
	return &business.CreateBusinessRequest{
		Name:            strings.TrimSpace(name),
		Phone:           strings.TrimSpace(phone),
		Industry:        strings.TrimSpace(industry),
		ServiceCategory: strings.TrimSpace(serviceCategory),
		BusinessType:    businessType,
	}
}

func ToDomainInviteStaffRequest(req *InviteStaffHTTPRequest) (*business.InviteStaffRequest, error) {
	var role business.StaffRole
	switch strings.ToLower(strings.TrimSpace(req.Role)) {
	case "admin":
		role = business.StaffRoleAdmin
	case "manager":
		role = business.StaffRoleManager
	case "staff":
		role = business.StaffRoleStaff
	default:
		return nil, fmt.Errorf("invalid role: %s (valid: admin, manager, staff)", req.Role)
	}
	var locationID *uuid.UUID
	if req.LocationID != "" {
		parsed, err := uuid.Parse(req.LocationID)
		if err != nil {
			return nil, fmt.Errorf("invalid location_id format: %w", err)
		}
		locationID = &parsed
	}

	return &business.InviteStaffRequest{
		Email:      strings.TrimSpace(req.Email),
		Phone:      strings.TrimSpace(req.Phone),
		Role:       role,
		LocationID: locationID,
		Title:      strings.TrimSpace(req.Title),
	}, nil
}

func ToBusinessResponse(b *business.Business) *BusinessResponse {
	return &BusinessResponse{
		ID:              b.ID.String(),
		Name:            b.Name,
		Phone:           b.Phone,
		OwnerID:         b.OwnerID.String(),
		Industry:        b.Industry,
		ServiceCategory: b.ServiceCategory,
		BusinessType:    string(b.BusinessType),
		IsActive:        b.IsActive,
		CreatedAt:       b.CreatedAt.Format(time.RFC3339),
	}
}

func ToInviteResponse(token string, expiresAt time.Time) *InviteResponse {
	return &InviteResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		Message:   "Invite created successfully. Send this token to the invitee.",
	}
}

// var (
// 	ErrInvalidRole       = &ErrorResponse{Success: false, Error: "Invalid role. Valid values: admin, manager, staff"}
// 	ErrInvalidLocationID = &ErrorResponse{Success: false, Error: "Invalid location_id format"}
// )
