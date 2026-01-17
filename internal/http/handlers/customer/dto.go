// File: internal/http/handlers/customer/dto.go
package customer

import (
	custdomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/customer"
	"github.com/google/uuid"
)

// HTTP Layer - Request/Response DTOs

// CreateCustomerRequest - POST /api/v1/customers
type CreateCustomerRequest struct {
	FullName string `json:"full_name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Phone    string `json:"phone" validate:"required"`
	Notes    string `json:"notes"`
}

// UpdateCustomerRequest - PUT /api/v1/customers/{id}
type UpdateCustomerRequest struct {
	FullName *string `json:"full_name"`
	Email    *string `json:"email"`
	Phone    *string `json:"phone"`
	Notes    *string `json:"notes"`
	Status   *string `json:"status"`
}

// CustomerResponse - Müştəri response
type CustomerResponse struct {
	ID            uuid.UUID `json:"id"`
	FullName      string    `json:"full_name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	Notes         string    `json:"notes"`
	Status        string    `json:"status"`
	TotalBookings int       `json:"total_bookings"`
	LastBookingAt *string   `json:"last_booking_at"`
	CreatedAt     string    `json:"created_at"`
	UpdatedAt     string    `json:"updated_at"`
}

// ListCustomersResponse - Pagination response
type ListCustomersResponse struct {
	Data       []CustomerResponse `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// ErrorResponseDTO - Error response format (Auth-dən kimi)
type ErrorResponseDTO struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

// SuccessResponseDTO - Success response format (Auth-dən kimi)
type SuccessResponseDTO struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// GetErrorResponse - Error code-dən HTTP response yaratmaq (Auth kimi)
func GetErrorResponse(code string) *ErrorResponseDTO {
	errorMap := map[string]*ErrorResponseDTO{
		"VALIDATION_ERROR": {
			Error:   "VALIDATION_ERROR",
			Message: "Validation error - missing fields or invalid format",
			Status:  400,
		},
		"UNAUTHORIZED": {
			Error:   "UNAUTHORIZED",
			Message: "Unauthorized - business_id not found",
			Status:  401,
		},
		"CUSTOMER_NOT_FOUND": {
			Error:   "CUSTOMER_NOT_FOUND",
			Message: "Customer not found",
			Status:  404,
		},
		"ACCESS_DENIED": {
			Error:   "ACCESS_DENIED",
			Message: "Access denied - customer not in this business",
			Status:  403,
		},
		"EMAIL_EXISTS": {
			Error:   "EMAIL_EXISTS",
			Message: "Email already exists in this business",
			Status:  409,
		},
		"INTERNAL_ERROR": {
			Error:   "INTERNAL_ERROR",
			Message: "Internal server error",
			Status:  500,
		},
	}

	if resp, ok := errorMap[code]; ok {
		return resp
	}

	return &ErrorResponseDTO{
		Error:   "UNKNOWN_ERROR",
		Message: "An unknown error occurred",
		Status:  500,
	}
}

// Helper: Domain Request → HTTP DTO çevrilməsi
func (r *CreateCustomerRequest) ToDomain() *custdomain.CreateCustomerRequest {
	return &custdomain.CreateCustomerRequest{
		FullName: r.FullName,
		Email:    r.Email,
		Phone:    r.Phone,
		Notes:    r.Notes,
	}
}

func (r *UpdateCustomerRequest) ToDomain() *custdomain.UpdateCustomerRequest {
	var status *custdomain.CustomerStatus
	if r.Status != nil {
		s := custdomain.CustomerStatus(*r.Status)
		status = &s
	}

	return &custdomain.UpdateCustomerRequest{
		FullName: r.FullName,
		Email:    r.Email,
		Phone:    r.Phone,
		Notes:    r.Notes,
		Status:   status,
	}
}

// NewCustomerResponse - Domain Entity-dən HTTP Response yaratmaq
func NewCustomerResponse(customer *custdomain.Customer) *CustomerResponse {
	resp := &CustomerResponse{
		ID:            customer.ID,
		FullName:      customer.FullName,
		Email:         customer.Email,
		Phone:         customer.Phone,
		Notes:         customer.Notes,
		Status:        string(customer.Status),
		TotalBookings: customer.TotalBookings,
		CreatedAt:     customer.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     customer.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	if customer.LastBookingAt != nil {
		t := customer.LastBookingAt.Format("2006-01-02T15:04:05Z")
		resp.LastBookingAt = &t
	}

	return resp
}

// NewListCustomersResponse - List response yaratmaq
func NewListCustomersResponse(listResp *custdomain.CustomersListResponse) *ListCustomersResponse {
	data := make([]CustomerResponse, len(listResp.Data))
	for i, c := range listResp.Data {
		data[i] = CustomerResponse{
			ID:            c.ID,
			FullName:      c.FullName,
			Email:         c.Email,
			Phone:         c.Phone,
			Notes:         c.Notes,
			Status:        string(c.Status),
			TotalBookings: c.TotalBookings,
			CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		}

		if c.LastBookingAt != nil {
			t := c.LastBookingAt.Format("2006-01-02T15:04:05Z")
			data[i].LastBookingAt = &t
		}
	}

	return &ListCustomersResponse{
		Data:       data,
		Total:      listResp.Total,
		Page:       listResp.Page,
		PageSize:   listResp.PageSize,
		TotalPages: listResp.TotalPages,
	}
}
