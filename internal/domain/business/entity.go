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
	ID              uuid.UUID `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	OwnerID         uuid.UUID `db:"owner_id" json:"owner_id"`
	Industry        string    `db:"industry" json:"industry"`
	ServiceCategory string    `db:"service_category" json:"service_category"`
	// CategorySlug – kesf ekranindaki sabit kateqoriya (bax:
	// domain/catalog). Bos ola bilir: kohne setirlerde serbest metn
	// acar sozlerle tanilir.
	CategorySlug string       `db:"category_slug" json:"category_slug"`
	Phone        string       `db:"phone" json:"phone"`
	BusinessType BusinessType `db:"business_type" json:"business_type"`
	IsActive     bool         `db:"is_active" json:"is_active"`
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
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
	Name            string `json:"name"`
	Industry        string `json:"industry"`
	ServiceCategory string `json:"service_category"`
	// CategorySlug – sahibin secdiyi sabit kateqoriya.
	CategorySlug string       `json:"category_slug"`
	Phone        string       `json:"phone"`
	BusinessType BusinessType `json:"business_type"`
}

type UpdateBusinessRequest struct {
	Name            string `json:"name"`
	Industry        string `json:"industry"`
	ServiceCategory string `json:"service_category"`
	CategorySlug    string `json:"category_slug"`
	Phone           string `json:"phone"`
	// OwnerID  uuid.UUID `json:"owner_id"`
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

// ============================================================
// KESF (musteri terefi)
// ============================================================

// LocationSummary – kesf siyahisinda gosterilen filial melumati.
// Telefon qesden daxil edilmir: siyahida lazim deyil.
type LocationSummary struct {
	Name      string   `json:"name"`
	Address   string   `json:"address,omitempty"`
	City      string   `json:"city,omitempty"`
	Latitude  *float64 `json:"latitude,omitempty"`
	Longitude *float64 `json:"longitude,omitempty"`
}

// HasCoordinates – filial xeritede yerlesdirile bilirmi.
// Koordinatlar ya birlikde var, ya hec biri (bax: miqrasiya 009).
func (l LocationSummary) HasCoordinates() bool {
	return l.Latitude != nil && l.Longitude != nil
}

// BookableBusiness – biznes + filiallari.
//
// "Yaxinliqdakilar" filtri ucun filiallar lazimdir: mesafe biznesin
// ozune yox, onun en yaxin filialina gore olcunur.
type BookableBusiness struct {
	Business
	Locations []LocationSummary `json:"locations"`
}
