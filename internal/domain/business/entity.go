package business

import (
	"time"

	"github.com/google/uuid"
)

type BusinessType string

const (
	BusinessTypeSolo  BusinessType = "solo"
	BusinessTypeMulti BusinessType = "multi"
)

func (bt BusinessType) IsValid() bool {
	return bt == BusinessTypeMulti || bt == BusinessTypeSolo
}

type UserRole string

const (
	RoleSoloPractitioner UserRole = "solo_practitioner"
	RoleProviderOwner    UserRole = "provider_owner"
	RoleStaff            UserRole = "staff"
	RoleCustomer         UserRole = "customer"
)

func (ur UserRole) IsValid() bool {
	switch ur {
	case RoleSoloPractitioner, RoleProviderOwner, RoleStaff, RoleCustomer:
		return true
	default:
		return false
	}
}

type Business struct {
	ID           uuid.UUID    `db:"id" json:"id"`
	Name         string       `db:"name" json:"name"`
	OwnerID      uuid.UUID    `db:"owner_id" json:"owner_id"`
	Phone        string       `db:"phone" json:"phone"`
	BusinessType BusinessType `db:"business_type" json:"business_type"`
	IsActive     bool         `db:"is_active" json:"is_active"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
}

func New(name, phone string) *Business {
	return &Business{
		ID:        uuid.New(),
		Name:      name,
		Phone:     phone,
		CreatedAt: time.Now(),
	}
}

// type Business struct {
// 	ID           uuid.UUID    `db:"id" json:"id"`
// 	Name         string       `db:"name" json:"name"`
// 	OwnerID      uuid.UUID    `db:"owner_id" json:"owner_id"`
// 	BusinessType BusinessType `db:"business_type" json:"business_type"`
// 	IsActive     bool         `db:"is_active" json:"is_active"`
// 	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
// 	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
// }

// type Location struct {
// 	ID         uuid.UUID `db:"id" json:"id"`
// 	BusinessID uuid.UUID `db:"business_id" json:"business_id"`
// 	Name       string    `db:"name" json:"name"`
// 	Address    *string   `db:"address" json:"address"`
// 	City       *string   `db:"city" json:"city"`
// 	IsActive   bool      `db:"is_active" json:"is_active"`
// 	CreatedAt  time.Time `db:"created_at" json:"created_at"`
// }

// type Staff struct {
// 	ID         uuid.UUID `db:"id" json:"id"`
// 	BusinessID uuid.UUID `db:"business_id" json:"business_id"`
// 	UserID     uuid.UUID `db:"user_id" json:"user_id"`
// 	Position   string    `db:"position" josn:"position"`
// 	IsActive   bool      `db:"is_active" json:"is_active"`
// 	CreatedAt  time.Time `db:"created_at" json:"created_at"`
// }
