// File: internal/http/handlers/business/dto.go
package business

import (
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/OrkhanNajaf1i/booking-service/internal/domain/catalog"
	"github.com/google/uuid"
)

// LocationHTTPRequest – biznesle birlikde yaradilan ilk filial.
type LocationHTTPRequest struct {
	Name      string   `json:"name"`
	Address   string   `json:"address"`
	City      string   `json:"city"`
	Latitude  *float64 `json:"latitude"`
	Longitude *float64 `json:"longitude"`
}

func (request *LocationHTTPRequest) ToDraft() *business.LocationDraft {
	if request == nil {
		return nil
	}
	return &business.LocationDraft{
		Name:      request.Name,
		Address:   request.Address,
		City:      request.City,
		Latitude:  request.Latitude,
		Longitude: request.Longitude,
	}
}

// SwitchModeHTTPRequest – tek isci ↔ komanda kecidi.
type SwitchModeHTTPRequest struct {
	BusinessType string `json:"business_type"`
}

type CreateSoloBusinessHTTPRequest struct {
	Name string `json:"name"`
	// CategorySlug – kesf ekranindaki sabit kateqoriya (mes. "dentist").
	CategorySlug string `json:"category_slug"`
	// ServiceCategory – sahibin oz sozu ("Kardioloq"); kartda alt basliq.
	ServiceCategory string               `json:"service_category"`
	Phone           string               `json:"phone"`
	Location        *LocationHTTPRequest `json:"location"`
}

type CreateMultiBusinessHTTPRequest struct {
	Name            string               `json:"name"`
	Industry        string               `json:"industry"`
	CategorySlug    string               `json:"category_slug"`
	ServiceCategory string               `json:"service_category"`
	Phone           string               `json:"phone"`
	Location        *LocationHTTPRequest `json:"location"`
}

type UpdateBusinessHTTPRequest struct {
	BusinessID      uuid.UUID `json:"business_id"`
	Name            string    `json:"name"`
	Industry        string    `json:"industry"`
	CategorySlug    string    `json:"category_slug"`
	ServiceCategory string    `json:"service_category"`
	Phone           string    `json:"phone"`
	// OwnerID  uuid.UUID `json:"owner_id"`
}

type BusinessHTTPResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	OwnerID         uuid.UUID `json:"owner_id"`
	Industry        string    `json:"industry"`
	CategorySlug    string    `json:"category_slug"`
	CategoryName    string    `json:"category_name"`
	ServiceCategory string    `json:"service_category"`
	Phone           string    `json:"phone"`
	BusinessType    string    `json:"business_type"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ErrorHTTPResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type SuccessHTTPResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

func ToBusinessesHTTPResponse(businesses []*business.Business) []*BusinessHTTPResponse {
	responses := make([]*BusinessHTTPResponse, 0, len(businesses))
	for _, b := range businesses {
		responses = append(responses, ToBusinessHTTPResponse(b))
	}
	return responses
}

type ListBusinessesHTTPResponse struct {
	Success bool                    `json:"success"`
	Data    []*BusinessHTTPResponse `json:"data"`
	Message string                  `json:"message,omitempty"`
}

func ToBusinessHTTPResponse(business *business.Business) *BusinessHTTPResponse {
	if business == nil {
		return nil
	}

	resolved := catalog.ResolveWith(
		business.CategorySlug, business.ServiceCategory, business.Industry,
	)

	return &BusinessHTTPResponse{
		ID:              business.ID,
		Name:            business.Name,
		OwnerID:         business.OwnerID,
		Industry:        business.Industry,
		CategorySlug:    resolved.Slug,
		CategoryName:    resolved.Name,
		ServiceCategory: business.ServiceCategory,
		Phone:           business.Phone,
		BusinessType:    string(business.BusinessType),
		IsActive:        business.IsActive,
		CreatedAt:       business.CreatedAt,
		UpdatedAt:       business.UpdatedAt,
	}
}

func (request *CreateSoloBusinessHTTPRequest) ToCreateBusinessRequest() *business.CreateBusinessRequest {
	return &business.CreateBusinessRequest{
		Name:            request.Name,
		ServiceCategory: request.ServiceCategory,
		CategorySlug:    request.CategorySlug,
		Phone:           request.Phone,
		BusinessType:    business.BusinessTypeSolo,
		Industry:        "",
		Location:        request.Location.ToDraft(),
	}
}

func (request *CreateMultiBusinessHTTPRequest) ToCreateBusinessRequest() *business.CreateBusinessRequest {
	return &business.CreateBusinessRequest{
		Name:            request.Name,
		Industry:        request.Industry,
		ServiceCategory: request.ServiceCategory,
		CategorySlug:    request.CategorySlug,
		Phone:           request.Phone,
		Location:        request.Location.ToDraft(),
		BusinessType:    business.BusinessTypeMulti,
	}
}

func (request *UpdateBusinessHTTPRequest) ToUpdateBusinessRequest() *business.UpdateBusinessRequest {
	return &business.UpdateBusinessRequest{
		Name:            request.Name,
		Industry:        request.Industry,
		ServiceCategory: request.ServiceCategory,
		CategorySlug:    request.CategorySlug,
		Phone:           request.Phone,
		// OwnerID:  request.OwnerID,
	}
}
