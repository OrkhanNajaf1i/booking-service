package business

import (
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
)

type CreateBusinessRequest struct {
	Name  string `json:"name" binding:"required"`
	Phone string `json:"phone" binding:"required"`
}

type BusinessResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Phone        string `json:"phone,omitempty"`
	OwnerID      string `json:"owner_id"`
	BusinessType string `json:"business_type"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

func ToResponse(b *business.Business) *BusinessResponse {
	return &BusinessResponse{
		ID:           b.ID.String(),
		Name:         b.Name,
		Phone:        b.Phone,
		OwnerID:      b.OwnerID.String(),
		BusinessType: string(b.BusinessType),
		IsActive:     b.IsActive,
		CreatedAt:    b.CreatedAt.Format(time.RFC3339),
	}
}
