package business

import (
	"time"

	"github.com/google/uuid"
)

type BusinessType string

const (
	BusinessTypeSolo  BusinessType = "solo_practitioner"
	BusinessTypeMulti BusinessType = "multi_staff_business"
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

func New(name string, businessType BusinessType) *Business {
	now := time.Now()
	return &Business{
		ID:           uuid.New(),
		Name:         name,
		OwnerID:      uuid.Nil,
		BusinessType: businessType,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

type Location struct {
	ID         uuid.UUID `db:"id" json:"id"`
	BusinessID uuid.UUID `db:"business_id" json:"business_id"`
	Name       string    `db:"name" json:"name"`
	Address    *string   `db:"address" json:"address"`
	City       *string   `db:"city" json:"city"`
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}
