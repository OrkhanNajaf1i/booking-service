// File: internal/http/handlers/staff/dto.go
package staff

import (
	"fmt"
	"strings"
	"time"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/staff"
	"github.com/google/uuid"
)

type CreateStaffHTTPRequest struct {
	UserID     string `json:"user_id"`
	Role       string `json:"role"`
	Title      string `json:"title"`
	LocationID string `json:"location_id,omitempty"`
}

type UpdateStaffHTTPRequest struct {
	Role       string  `json:"role"`
	Title      string  `json:"title"`
	Department string  `json:"department"`
	Bio        string  `json:"bio"`
	HourlyRate float64 `json:"hourly_rate"`
	LocationID string  `json:"location_id,omitempty"`
}

type InviteStaffHTTPRequest struct {
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	Role       string `json:"role"`
	LocationID string `json:"location_id,omitempty"`
}

type ValidateInviteHTTPRequest struct {
	Token string `json:"token"`
}

type AcceptInviteHTTPRequest struct {
	Token    string `json:"token"`
	Password string `json:"password"`
}

type StaffProfileResponse struct {
	ID         uuid.UUID          `json:"id"`
	UserID     uuid.UUID          `json:"user_id"`
	BusinessID uuid.UUID          `json:"business_id"`
	LocationID *uuid.UUID         `json:"location_id,omitempty"`
	Role       domain.StaffRole   `json:"role"`
	Title      string             `json:"title"`
	Department string             `json:"department"`
	Bio        string             `json:"bio"`
	HourlyRate float64            `json:"hourly_rate"`
	Status     domain.StaffStatus `json:"status"`
	JoinedAt   time.Time          `json:"joined_at"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type StaffWithUserResponse struct {
	ID         uuid.UUID          `json:"id"`
	UserID     uuid.UUID          `json:"user_id"`
	FullName   string             `json:"full_name"`
	Email      string             `json:"email"`
	Phone      string             `json:"phone"`
	Avatar     *string            `json:"avatar,omitempty"`
	Role       domain.StaffRole   `json:"role"`
	Title      string             `json:"title"`
	Department string             `json:"department"`
	LocationID *uuid.UUID         `json:"location_id,omitempty"`
	Status     domain.StaffStatus `json:"status"`
	JoinedAt   time.Time          `json:"joined_at"`
}

type InviteResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type InviteDetailsResponse struct {
	BusinessID uuid.UUID        `json:"business_id"`
	Email      string           `json:"email"`
	Phone      string           `json:"phone"`
	Role       domain.StaffRole `json:"role"`
	LocationID *uuid.UUID       `json:"location_id,omitempty"`
	ExpiresAt  time.Time        `json:"expires_at"`
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

func ToDomainCreateStaffRequest(req CreateStaffHTTPRequest) (*domain.CreateStaffRequest, error) {
	userID, err := uuid.Parse(strings.TrimSpace(req.UserID))
	if err != nil {
		return nil, fmt.Errorf("invalid user_id: %w", err)
	}

	role, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}

	locID, err := parseOptionalUUID(req.LocationID)
	if err != nil {
		return nil, err
	}

	return &domain.CreateStaffRequest{
		UserID:     userID,
		Role:       role,
		Title:      strings.TrimSpace(req.Title),
		LocationID: locID,
	}, nil
}

func ToDomainUpdateStaffRequest(req UpdateStaffHTTPRequest) (*domain.UpdateStaffRequest, error) {
	role, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}

	locID, err := parseOptionalUUID(req.LocationID)
	if err != nil {
		return nil, err
	}

	return &domain.UpdateStaffRequest{
		Role:       role,
		Title:      strings.TrimSpace(req.Title),
		Department: strings.TrimSpace(req.Department),
		Bio:        strings.TrimSpace(req.Bio),
		HourlyRate: req.HourlyRate,
		LocationID: locID,
	}, nil
}

func ToDomainInviteStaffRequest(req InviteStaffHTTPRequest) (*domain.InviteStaffRequest, error) {
	role, err := parseRole(req.Role)
	if err != nil {
		return nil, err
	}

	locID, err := parseOptionalUUID(req.LocationID)
	if err != nil {
		return nil, err
	}

	return &domain.InviteStaffRequest{
		Email:      strings.TrimSpace(req.Email),
		Phone:      strings.TrimSpace(req.Phone),
		Role:       role,
		LocationID: locID,
	}, nil
}

func FromDomainStaffProfile(p *domain.StaffProfile) StaffProfileResponse {
	return StaffProfileResponse{
		ID:         p.ID,
		UserID:     p.UserID,
		BusinessID: p.BusinessID,
		LocationID: p.LocationID,
		Role:       p.Role,
		Title:      p.Title,
		Department: p.Department,
		Bio:        p.Bio,
		HourlyRate: p.HourlyRate,
		Status:     p.Status,
		JoinedAt:   p.JoinedAt,
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}
}

func FromDomainStaffWithUser(list []*domain.StaffWithUser) []StaffWithUserResponse {
	res := make([]StaffWithUserResponse, 0, len(list))
	for _, s := range list {
		res = append(res, StaffWithUserResponse{
			ID:         s.ID,
			UserID:     s.UserID,
			FullName:   s.FullName,
			Email:      s.Email,
			Phone:      s.Phone,
			Avatar:     s.Avatar,
			Role:       s.Role,
			Title:      s.Title,
			Department: s.Department,
			LocationID: s.LocationID,
			Status:     s.Status,
			JoinedAt:   s.JoinedAt,
		})
	}
	return res
}

func FromDomainInviteDetails(inv *domain.BusinessInvite) InviteDetailsResponse {
	return InviteDetailsResponse{
		BusinessID: inv.BusinessID,
		Email:      inv.InvitedEmail,
		Phone:      inv.InvitedPhone,
		Role:       inv.Role,
		LocationID: inv.LocationID,
		ExpiresAt:  inv.ExpiresAt,
	}
}

func parseRole(roleStr string) (domain.StaffRole, error) {
	switch strings.ToLower(strings.TrimSpace(roleStr)) {
	case "admin":
		return domain.StaffRoleAdmin, nil
	case "manager":
		return domain.StaffRoleManager, nil
	case "staff":
		return domain.StaffRoleStaff, nil
	default:
		return "", fmt.Errorf("invalid role %q, valid values: admin, manager, staff", roleStr)
	}
}

func parseOptionalUUID(idStr string) (*uuid.UUID, error) {
	clean := strings.TrimSpace(idStr)
	if clean == "" {
		return nil, nil
	}
	parsed, err := uuid.Parse(clean)
	if err != nil {
		return nil, fmt.Errorf("invalid uuid %q: %w", idStr, err)
	}
	return &parsed, nil
}
