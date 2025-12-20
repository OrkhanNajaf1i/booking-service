package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `json:"id"`
	BusinessID uuid.UUID `json:"business_id"`
	Name       string    `json:"name"`
	Phone      string    `json:"phone"`
	CreatedAt  time.Time `json:"created_at"`
}

func New(businessID uuid.UUID, name, phone string) *User {
	return &User{
		ID:         uuid.New(),
		BusinessID: businessID,
		Name:       name,
		Phone:      phone,
		CreatedAt:  time.Now(),
	}
}
