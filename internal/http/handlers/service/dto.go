// File: internal/http/handlers/service/dto.go
package service

import (
	"fmt"
	"strings"
	"time"

	domain "github.com/OrkhanNajaf1i/booking-service/internal/domain/service"
	"github.com/google/uuid"
)

type ServiceResponse struct {
	ID              uuid.UUID `json:"id"`
	BusinessID      uuid.UUID `json:"business_id"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	DurationMinutes int       `json:"duration_minutes"`
	Price           float64   `json:"price"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type UpdateServiceHTTPRequest struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
}

type AssignServicesHTTPRequest struct {
	ServiceIDs []string `json:"service_ids"`
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

func FromDomainService(svc *domain.Service) ServiceResponse {
	return ServiceResponse{
		ID:              svc.ID,
		BusinessID:      svc.BusinessID,
		Name:            svc.Name,
		Description:     svc.Description,
		DurationMinutes: svc.DurationMinutes,
		Price:           svc.Price,
		IsActive:        svc.IsActive,
		CreatedAt:       svc.CreatedAt,
		UpdatedAt:       svc.UpdatedAt,
	}
}

func FromDomainServices(list []*domain.Service) []ServiceResponse {
	res := make([]ServiceResponse, 0, len(list))
	for _, s := range list {
		res = append(res, FromDomainService(s))
	}
	return res
}

func ToDomainUpdateServiceRequest(req UpdateServiceHTTPRequest) *domain.UpdateServiceRequest {
	return &domain.UpdateServiceRequest{
		Name:            strings.TrimSpace(req.Name),
		Description:     strings.TrimSpace(req.Description),
		DurationMinutes: req.DurationMinutes,
		Price:           req.Price,
	}
}

func ParseServiceIDs(ids []string) ([]uuid.UUID, error) {
	result := make([]uuid.UUID, 0, len(ids))
	for _, raw := range ids {
		clean := strings.TrimSpace(raw)
		if clean == "" {
			continue
		}
		id, err := uuid.Parse(clean)
		if err != nil {
			return nil, fmt.Errorf("invalid service id %q: %w", raw, err)
		}
		result = append(result, id)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no valid service ids provided")
	}
	return result, nil
}
