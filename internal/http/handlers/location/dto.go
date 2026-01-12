// File: internal/http/handlers/location/dto.go
package location

import (
	"strings"
	"time"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/location"
	"github.com/google/uuid"
)

type CreateLocationHTTPRequest struct {
	Name    string  `json:"name" binding:"required"`
	Address *string `json:"address,omitempty"`
	City    *string `json:"city,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

type UpdateLocationHTTPRequest struct {
	Name    string  `json:"name" binding:"required"`
	Address *string `json:"address,omitempty"`
	City    *string `json:"city,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

type LocationResponse struct {
	ID         uuid.UUID `json:"id"`
	BusinessID uuid.UUID `json:"business_id"`
	Name       string    `json:"name"`
	Address    *string   `json:"address,omitempty"`
	City       *string   `json:"city,omitempty"`
	Phone      *string   `json:"phone,omitempty"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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

func ToDomainCreateRequest(req CreateLocationHTTPRequest) *domain.CreateLocationRequest {
	return &domain.CreateLocationRequest{
		Name:    strings.TrimSpace(req.Name),
		Address: trimPtr(req.Address),
		City:    trimPtr(req.City),
		Phone:   trimPtr(req.Phone),
	}
}

func ToDomainUpdateRequest(req UpdateLocationHTTPRequest) *domain.UpdateLocationRequest {
	return &domain.UpdateLocationRequest{
		Name:    strings.TrimSpace(req.Name),
		Address: trimPtr(req.Address),
		City:    trimPtr(req.City),
		Phone:   trimPtr(req.Phone),
	}
}

func FromDomainLocation(loc *domain.Location) LocationResponse {
	return LocationResponse{
		ID:         loc.ID,
		BusinessID: loc.BusinessID,
		Name:       loc.Name,
		Address:    loc.Address,
		City:       loc.City,
		Phone:      loc.Phone,
		IsActive:   loc.IsActive,
		CreatedAt:  loc.CreatedAt,
		UpdatedAt:  loc.UpdatedAt,
	}
}

func FromDomainLocations(list []*domain.Location) []LocationResponse {
	res := make([]LocationResponse, 0, len(list))
	for _, l := range list {
		res = append(res, FromDomainLocation(l))
	}
	return res
}

func trimPtr(v *string) *string {
	if v == nil {
		return nil
	}
	clean := strings.TrimSpace(*v)
	if clean == "" {
		return nil
	}
	return &clean
}
