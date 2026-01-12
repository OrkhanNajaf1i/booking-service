// File: internal/domain/business/entity.go
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

func NewBusiness(name, industry, serviceCategory, phone string, businessType BusinessType) *Business {
	now := time.Now()
	return &Business{
		ID:              uuid.New(),
		Name:            name,
		OwnerID:         uuid.Nil,
		Industry:        industry,
		ServiceCategory: serviceCategory,
		Phone:           phone,
		BusinessType:    businessType,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

type CreateBusinessRequest struct {
	Name            string       `json:"name"`
	Industry        string       `json:"industry"`
	ServiceCategory string       `json:"service_category"`
	Phone           string       `json:"phone"`
	BusinessType    BusinessType `json:"business_type"`
}

type UpdateBusinessRequest struct {
	Name     string `json:"name"`
	Industry string `json:"industry"`
	Phone    string `json:"phone"`
}

type BusinessError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *BusinessError) Error() string {
	return e.Message
}

func NewBusinessError(code, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}
