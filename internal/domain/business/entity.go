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

type Business struct {
	ID              uuid.UUID    `db:"id" json:"id"`
	Name            string       `db:"name" json:"name"`
	OwnerID         uuid.UUID    `db:"owner_id" json:"owner_id"`
	Industry        string       `db:"industry" json:"industry"`
	ServiceCategory string       `db:"service_category" json:"service_category"`
	Phone           string       `db:"phone" json:"phone"`
	BusinessType    BusinessType `db:"business_type" json:"business_type"`
	IsActive        bool         `db:"is_active" json:"is_active"`
	CreatedAt       time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time    `db:"updated_at" json:"updated_at"`
}

func NewBusiness(name, industry, serviceCategory, phone string, bType BusinessType) *Business {
	now := time.Now()
	return &Business{
		ID:              uuid.New(),
		Name:            name,
		OwnerID:         uuid.Nil,
		Industry:        industry,
		ServiceCategory: serviceCategory,
		Phone:           phone,
		BusinessType:    bType,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

type Location struct {
	ID         uuid.UUID `db:"id" json:"id"`
	BusinessID uuid.UUID `db:"business_id" json:"business_id"`
	Name       string    `db:"name" json:"name"`
	Address    *string   `db:"address" json:"address,omitempty"`
	City       *string   `db:"city" json:"city,omitempty"`
	IsActive   bool      `db:"is_active" json:"is_active"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type StaffRole string

const (
	StaffRoleAdmin   StaffRole = "admin"
	StaffRoleManager StaffRole = "manager"
	StaffRoleStaff   StaffRole = "staff"
)

type StaffProfile struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	UserID     uuid.UUID  `db:"user_id" json:"user_id"`
	BusinessID uuid.UUID  `db:"business_id" json:"business_id"`
	LocationID *uuid.UUID `db:"location_id" json:"location_id,omitempty"`
	Role       StaffRole  `db:"role" json:"role"`
	Title      string     `db:"title" json:"title"`
	Department string     `db:"department" json:"department"`
	Bio        string     `db:"bio" json:"bio"`
	HourlyRate float64    `db:"hourly_rate" json:"hourly_rate"`
	Status     string     `db:"status" json:"status"`
	JoinedAt   time.Time  `db:"joined_at" json:"joined_at"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updated_at"`
}
type BusinessInvite struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	BusinessID   uuid.UUID  `db:"business_id" json:"business_id"`
	InvitedEmail string     `db:"invited_email" json:"invited_email"`
	InvitedPhone string     `db:"invited_phone" json:"invited_phone"`
	Role         StaffRole  `db:"role" json:"role"`
	LocationID   *uuid.UUID `db:"location_id" json:"location_id"`
	Token        string     `db:"token" json:"token"`
	ExpiresAt    time.Time  `db:"expires_at" json:"expires_at"`
	Used         bool       `db:"used" json:"used"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
}

type CreateBusinessRequest struct {
	Name            string       `json:"name"`
	Industry        string       `json:"industry"`
	ServiceCategory string       `json:"service_category"`
	Phone           string       `json:"phone"`
	BusinessType    BusinessType `json:"business_type"`
}

type InviteStaffRequest struct {
	Email      string     `json:"email"`
	Phone      string     `json:"phone"`
	Role       StaffRole  `json:"role"`
	LocationID *uuid.UUID `json:"location_id,omitempty"`
	Title      string     `json:"title,omitempty"`
}
type RegistrationError struct {
	Code    string `db:"code" json:"code"`
	Message string `db:"message" json:"message"`
}

func (e *RegistrationError) Error() string {
	return e.Message
}
