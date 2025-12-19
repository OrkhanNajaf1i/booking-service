package business

import (
	"time"

	"github.com/google/uuid"
)

type Business struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
}

func New(name, phone string) *Business {
	return &Business{
		ID:        uuid.New(),
		Name:      name,
		Phone:     phone,
		CreatedAt: time.Now(),
	}
}
