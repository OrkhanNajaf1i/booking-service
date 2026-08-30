// File: internal/http/handlers/business/dto.go
package business

import (
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
)

type CreateSoloBusinessHTTPRequest struct {
	Name            string `json:"name"`
	ServiceCategory string `json:"service_category"`
	Phone           string `json:"phone"`
}

type CreateMultiBusinessHTTPRequest struct {
	Name     string `json:"name"`
	Industry string `json:"industry"`
	Phone    string `json:"phone"`
}

type UpdateBusinessHTTPRequest struct {
	BusinessID uuid.UUID `json:"business_id"`
	Name       string    `json:"name"`
	Industry   string    `json:"industry"`
	Phone      string    `json:"phone"`
	// OwnerID  uuid.UUID `json:"owner_id"`
}

type BusinessHTTPResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	OwnerID         uuid.UUID `json:"owner_id"`
	Industry        string    `json:"industry"`
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

	return &BusinessHTTPResponse{
		ID:              business.ID,
		Name:            business.Name,
		OwnerID:         business.OwnerID,
		Industry:        business.Industry,
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
		Phone:           request.Phone,
		BusinessType:    business.BusinessTypeSolo,
		Industry:        "",
	}
}

func (request *CreateMultiBusinessHTTPRequest) ToCreateBusinessRequest() *business.CreateBusinessRequest {
	return &business.CreateBusinessRequest{
		Name:            request.Name,
		Industry:        request.Industry,
		Phone:           request.Phone,
		BusinessType:    business.BusinessTypeMulti,
		ServiceCategory: "",
	}
}

func (request *UpdateBusinessHTTPRequest) ToUpdateBusinessRequest() *business.UpdateBusinessRequest {
	return &business.UpdateBusinessRequest{
		Name:     request.Name,
		Industry: request.Industry,
		Phone:    request.Phone,
		// OwnerID:  request.OwnerID,
	}
}
