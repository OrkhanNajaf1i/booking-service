// File: internal/domain/service/entity.go
package service

import (
	"time"

	"github.com/google/uuid"
)

type Service struct {
	ID              uuid.UUID `db:"id" json:"id"`
	BusinessID      uuid.UUID `db:"business_id" json:"business_id"`
	Name            string    `db:"name" json:"name"`
	Description     string    `db:"description" json:"description"`
	DurationMinutes int       `db:"duration_minutes" json:"duration_minutes"`
	Price           float64   `db:"price" json:"price"`
	IsActive        bool      `db:"is_active" json:"is_active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time `db:"updated_at" json:"updated_at"`
}

type StaffService struct {
	StaffID   uuid.UUID `db:"staff_id" json:"staff_id"`
	ServiceID uuid.UUID `db:"service_id" json:"service_id"`
}

func NewService(businessID uuid.UUID, name, description string, durationMinutes int, price float64) *Service {
	now := time.Now()
	return &Service{
		ID:              uuid.New(),
		BusinessID:      businessID,
		Name:            name,
		Description:     description,
		DurationMinutes: durationMinutes,
		Price:           price,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

type CreateServiceRequest struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
}

type UpdateServiceRequest struct {
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	DurationMinutes int     `json:"duration_minutes"`
	Price           float64 `json:"price"`
}

type ServiceError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *ServiceError) Error() string {
	return e.Message
}
