// File: internal/domain/customer/entity.go
package customer

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// CustomerStatus - Müştəri statusunun enum-u
type CustomerStatus string

const (
	StatusActive   CustomerStatus = "active"
	StatusInactive CustomerStatus = "inactive" // Soft delete
	StatusBlocked  CustomerStatus = "blocked"  // Rədd edilmiş müştəri
)

// IsValid - Status-un valid olub-olmadığını yoxla
func (s CustomerStatus) IsValid() bool {
	return s == StatusActive || s == StatusInactive || s == StatusBlocked
}

// Customer - Domain Entity (Biznes tərəfindən idarə olunan müştəri kartı)
type Customer struct {
	ID            uuid.UUID      `db:"id" json:"id"`
	BusinessID    uuid.UUID      `db:"business_id" json:"business_id"` // Multi-tenant isolasiya
	UserID        *uuid.UUID     `db:"user_id" json:"user_id"`         // NULL ola bilər (walk-in müştəri)
	FullName      string         `db:"full_name" json:"full_name"`
	Email         string         `db:"email" json:"email"`
	Phone         string         `db:"phone" json:"phone"`
	Notes         string         `db:"notes" json:"notes"` // Admin qeydləri
	Status        CustomerStatus `db:"status" json:"status"`
	TotalBookings int            `db:"total_bookings" json:"total_bookings"` // Cache: bron sayı
	LastBookingAt *time.Time     `db:"last_booking_at" json:"last_booking_at"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at" json:"updated_at"`
}

// NewCustomer - Factory function
func NewCustomer(businessID uuid.UUID, fullName, email, phone string) *Customer {
	now := time.Now()
	return &Customer{
		ID:            uuid.New(),
		BusinessID:    businessID,
		UserID:        nil,
		FullName:      fullName,
		Email:         email,
		Phone:         phone,
		Notes:         "",
		Status:        StatusActive,
		TotalBookings: 0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Domain Errors - Custom error type-ləri
var (
	ErrCustomerNotFound    = errors.New("customer not found")
	ErrEmailAlreadyExists  = errors.New("customer with this email already exists in this business")
	ErrInvalidCustomerData = errors.New("invalid customer data provided")
	ErrAccessDenied        = errors.New("access denied to this customer")
	ErrInvalidStatus       = errors.New("invalid customer status")
	ErrCannotDeleteOwner   = errors.New("cannot delete customer with pending bookings")
)

// DTOs (Data Transfer Objects)

// CreateCustomerRequest - POST /api/v1/customers üçün request body
type CreateCustomerRequest struct {
	FullName string `json:"full_name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required,min=10,max=20"`
	Notes    string `json:"notes" validate:"max=500"`
}

// UpdateCustomerRequest - PUT /api/v1/customers/{id} üçün request body
type UpdateCustomerRequest struct {
	FullName *string         `json:"full_name" validate:"omitempty,min=2,max=100"`
	Email    *string         `json:"email" validate:"omitempty,email"`
	Phone    *string         `json:"phone" validate:"omitempty,min=10,max=20"`
	Notes    *string         `json:"notes" validate:"omitempty,max=500"`
	Status   *CustomerStatus `json:"status" validate:"omitempty"`
}

// CustomerResponse - API response
type CustomerResponse struct {
	ID            uuid.UUID      `json:"id"`
	FullName      string         `json:"full_name"`
	Email         string         `json:"email"`
	Phone         string         `json:"phone"`
	Notes         string         `json:"notes"`
	Status        CustomerStatus `json:"status"`
	TotalBookings int            `json:"total_bookings"`
	LastBookingAt *time.Time     `json:"last_booking_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

// ToResponse - Customer Entity-ni Response-ə convert et
func (c *Customer) ToResponse() *CustomerResponse {
	return &CustomerResponse{
		ID:            c.ID,
		FullName:      c.FullName,
		Email:         c.Email,
		Phone:         c.Phone,
		Notes:         c.Notes,
		Status:        c.Status,
		TotalBookings: c.TotalBookings,
		LastBookingAt: c.LastBookingAt,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

// CustomersListResponse - Pagination-li response
type CustomersListResponse struct {
	Data       []*CustomerResponse `json:"data"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}
