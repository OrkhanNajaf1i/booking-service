// File: internal/domain/location/entity.go
package location

import (
	"time"

	"github.com/google/uuid"
)

type Location struct {
	ID         uuid.UUID `db:"id" json:"id"`
	BusinessID uuid.UUID `db:"business_id" json:"business_id"`
	Name       string    `db:"name" json:"name"`
	Address    *string   `db:"address" json:"address,omitempty"`
	City       *string   `db:"city" json:"city,omitempty"`
	Phone      *string   `db:"phone" json:"phone,omitempty"`
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

func NewLocation(businessID uuid.UUID, name string) *Location {
	now := time.Now()
	return &Location{
		ID:         uuid.New(),
		BusinessID: businessID,
		Name:       name,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

type CreateLocationRequest struct {
	Name    string  `json:"name"`
	Address *string `json:"address,omitempty"`
	City    *string `json:"city,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

type UpdateLocationRequest struct {
	Name    string  `json:"name"`
	Address *string `json:"address,omitempty"`
	City    *string `json:"city,omitempty"`
	Phone   *string `json:"phone,omitempty"`
}

type LocationError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *LocationError) Error() string {
	return e.Message
}
