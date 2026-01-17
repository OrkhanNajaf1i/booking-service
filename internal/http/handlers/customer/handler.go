// File: internal/http/handlers/customer/handler.go
package customer

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/customer"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
)

type Handler struct {
	service customer.Service
	logger  logger.Logger
}

// NewHandler - Handler instance yaratır
func NewHandler(service customer.Service, logger logger.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// sendJSON - JSON response göndərmə
func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// sendError - Error response göndərmə (auth-dən kimi)
func (h *Handler) sendError(w http.ResponseWriter, status int, code string) {
	errorDTO := GetErrorResponse(code)
	h.sendJSON(w, status, errorDTO)
}

// sendSuccess - Success response göndərmə (auth-dən kimi)
func (h *Handler) sendSuccess(w http.ResponseWriter, status int, message string, data interface{}) {
	successResp := SuccessResponseDTO{
		Success: true,
		Message: message,
		Data:    data,
	}
	h.sendJSON(w, status, successResp)
}

// GetBusinessIDFromContext - Middleware-dən business_id əldə et (auth-dən kimi)
func GetBusinessIDFromContext(r *http.Request) (uuid.UUID, error) {
	val := r.Context().Value("business_id")
	if val == nil {
		return uuid.Nil, errors.New("unauthorized - business_id not found in context")
	}

	businessID, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("unauthorized - business_id type assertion failed")
	}

	return businessID, nil
}

// @Summary      Create Customer
// @Description  Creates a new customer in the business. Only business owner and staff can create customers. Account-first: Customer can be created with or without user account.
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Param        request body CreateCustomerRequest true "Customer data (FullName, Email, Phone, Notes optional)"
// @Success      201  {object}  CustomerResponse "Customer created successfully"
// @Failure      400  {object}  ErrorResponseDTO "Validation error - missing fields, invalid format"
// @Failure      401  {object}  ErrorResponseDTO "Unauthorized - business_id not found"
// @Failure      409  {object}  ErrorResponseDTO "Email already exists in this business"
// @Failure      500  {object}  ErrorResponseDTO "Internal server error"
// @Router       /api/v1/customers [post]
func (h *Handler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("CreateCustomer request received",
		logger.Field{Key: "method", Value: r.Method},
		logger.Field{Key: "path", Value: r.URL.Path},
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	// Get business_id from context (middleware tərəfindən təyin olunur)
	businessID, err := GetBusinessIDFromContext(r)
	if err != nil {
		h.logger.Error("CreateCustomer: business_id not found in context",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
		)
		h.sendError(w, http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	// Parse request body
	var req CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("CreateCustomer: JSON parse failed",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	// Validate required fields
	if req.FullName == "" || req.Email == "" || req.Phone == "" {
		h.logger.Warn("CreateCustomer: Missing required fields",
			logger.Field{Key: "full_name_empty", Value: req.FullName == ""},
			logger.Field{Key: "email_empty", Value: req.Email == ""},
			logger.Field{Key: "phone_empty", Value: req.Phone == ""},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	// Convert HTTP DTO to Domain DTO
	domainReq := req.ToDomain()

	// Call service
	cust, err := h.service.CreateCustomer(ctx, businessID, domainReq)
	if err != nil {
		h.logger.Error("CreateCustomer: Service error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "email", Value: req.Email},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)

		if err == customer.ErrEmailAlreadyExists {
			h.sendError(w, http.StatusConflict, "EMAIL_EXISTS")
			return
		}
		if err == customer.ErrInvalidCustomerData {
			h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
			return
		}

		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.logger.Info("CreateCustomer: Customer created successfully",
		logger.Field{Key: "customer_id", Value: cust.ID.String()},
		logger.Field{Key: "email", Value: cust.Email},
		logger.Field{Key: "business_id", Value: businessID.String()},
	)

	// Convert Domain to HTTP response
	resp := NewCustomerResponse(cust)
	h.sendJSON(w, http.StatusCreated, resp)
}

// @Summary      Get Customer
// @Description  Retrieves a specific customer by ID. Only accessible within the same business (multi-tenant isolation). Returns customer details including booking statistics.
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Param        id query string true "Customer ID (UUID format)"
// @Success      200  {object}  CustomerResponse "Customer details retrieved successfully"
// @Failure      400  {object}  ErrorResponseDTO "Invalid customer ID format"
// @Failure      401  {object}  ErrorResponseDTO "Unauthorized - business_id not found"
// @Failure      404  {object}  ErrorResponseDTO "Customer not found"
// @Failure      403  {object}  ErrorResponseDTO "Access denied - customer not in this business"
// @Failure      500  {object}  ErrorResponseDTO "Internal server error"
// @Router       /api/v1/customers [get]
func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("GetCustomer request received",
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	// Get business_id from context
	businessID, err := GetBusinessIDFromContext(r)
	if err != nil {
		h.logger.Error("GetCustomer: business_id not found in context",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	// Parse customer_id from query parameter
	customerIDStr := r.URL.Query().Get("id")
	if customerIDStr == "" {
		h.logger.Warn("GetCustomer: customer_id query param missing")
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		h.logger.Error("GetCustomer: Invalid customer_id format",
			logger.Field{Key: "customer_id", Value: customerIDStr},
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	// Call service
	cust, err := h.service.GetCustomer(ctx, businessID, customerID)
	if err != nil {
		h.logger.Error("GetCustomer: Service error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: customerID.String()},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)

		if err == customer.ErrCustomerNotFound {
			h.sendError(w, http.StatusNotFound, "CUSTOMER_NOT_FOUND")
			return
		}
		if err == customer.ErrAccessDenied {
			h.sendError(w, http.StatusForbidden, "ACCESS_DENIED")
			return
		}

		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.logger.Info("GetCustomer: Customer retrieved successfully",
		logger.Field{Key: "customer_id", Value: customerID.String()},
	)

	resp := NewCustomerResponse(cust)
	h.sendJSON(w, http.StatusOK, resp)
}

// @Summary      List Customers
// @Description  Retrieves paginated list of customers for the business. Supports pagination with page and page_size query parameters. Filters applied per business_id for multi-tenant isolation.
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Param        page query int false "Page number (default: 1, must be >= 1)"
// @Param        page_size query int false "Page size (default: 20, max: 100)"
// @Success      200  {object}  ListCustomersResponse "Paginated customer list retrieved successfully"
// @Failure      400  {object}  ErrorResponseDTO "Invalid query parameters"
// @Failure      401  {object}  ErrorResponseDTO "Unauthorized - business_id not found"
// @Failure      500  {object}  ErrorResponseDTO "Internal server error"
// @Router       /api/v1/customers [get]
func (h *Handler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("ListCustomers request received",
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	// Get business_id from context
	businessID, err := GetBusinessIDFromContext(r)
	if err != nil {
		h.logger.Error("ListCustomers: business_id not found in context",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	// Parse query parameters with defaults
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	h.logger.Info("ListCustomers parameters",
		logger.Field{Key: "business_id", Value: businessID.String()},
		logger.Field{Key: "page", Value: page},
		logger.Field{Key: "page_size", Value: pageSize},
	)

	// Call service
	listResp, err := h.service.ListCustomers(ctx, businessID, page, pageSize)
	if err != nil {
		h.logger.Error("ListCustomers: Service error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "business_id", Value: businessID.String()},
		)
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.logger.Info("ListCustomers: Retrieved successfully",
		logger.Field{Key: "total", Value: listResp.Total},
		logger.Field{Key: "page", Value: page},
		logger.Field{Key: "returned_count", Value: len(listResp.Data)},
	)

	// Convert to HTTP response
	resp := NewListCustomersResponse(listResp)
	h.sendJSON(w, http.StatusOK, resp)
}

// @Summary      Update Customer
// @Description  Updates customer information (partial update supported). Only accessible within the same business. Email must be unique within business if changed.
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Param        id query string true "Customer ID (UUID format)"
// @Param        request body UpdateCustomerRequest true "Customer update data (all fields optional)"
// @Success      200  {object}  CustomerResponse "Customer updated successfully"
// @Failure      400  {object}  ErrorResponseDTO "Validation error or invalid customer ID"
// @Failure      401  {object}  ErrorResponseDTO "Unauthorized - business_id not found"
// @Failure      404  {object}  ErrorResponseDTO "Customer not found"
// @Failure      409  {object}  ErrorResponseDTO "Email already exists in this business"
// @Failure      403  {object}  ErrorResponseDTO "Access denied - customer not in this business"
// @Failure      500  {object}  ErrorResponseDTO "Internal server error"
// @Router       /api/v1/customers [put]
func (h *Handler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("UpdateCustomer request received",
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	// Get business_id from context
	businessID, err := GetBusinessIDFromContext(r)
	if err != nil {
		h.logger.Error("UpdateCustomer: business_id not found in context",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	// Parse customer_id from query parameter
	customerIDStr := r.URL.Query().Get("id")
	if customerIDStr == "" {
		h.logger.Warn("UpdateCustomer: customer_id query param missing")
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		h.logger.Error("UpdateCustomer: Invalid customer_id format",
			logger.Field{Key: "customer_id", Value: customerIDStr},
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	// Parse request body
	var req UpdateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("UpdateCustomer: JSON parse failed",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	h.logger.Info("UpdateCustomer: Processing update",
		logger.Field{Key: "customer_id", Value: customerID.String()},
		logger.Field{Key: "business_id", Value: businessID.String()},
	)

	// Convert to domain DTO
	domainReq := req.ToDomain()

	// Call service
	cust, err := h.service.UpdateCustomer(ctx, businessID, customerID, domainReq)
	if err != nil {
		h.logger.Error("UpdateCustomer: Service error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: customerID.String()},
		)

		if err == customer.ErrCustomerNotFound {
			h.sendError(w, http.StatusNotFound, "CUSTOMER_NOT_FOUND")
			return
		}
		if err == customer.ErrAccessDenied {
			h.sendError(w, http.StatusForbidden, "ACCESS_DENIED")
			return
		}
		if err == customer.ErrEmailAlreadyExists {
			h.sendError(w, http.StatusConflict, "EMAIL_EXISTS")
			return
		}
		if err == customer.ErrInvalidCustomerData {
			h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
			return
		}
		if err == customer.ErrInvalidStatus {
			h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
			return
		}

		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.logger.Info("UpdateCustomer: Customer updated successfully",
		logger.Field{Key: "customer_id", Value: customerID.String()},
	)

	resp := NewCustomerResponse(cust)
	h.sendJSON(w, http.StatusOK, resp)
}

// @Summary      Delete Customer
// @Description  Soft deletes a customer (status changed to 'inactive'). Customer data is retained for historical records and booking history. Hard deletion is not performed to maintain referential integrity.
// @Tags         Customer
// @Accept       json
// @Produce      json
// @Param        id query string true "Customer ID (UUID format)"
// @Success      204  {object}  nil "Customer deleted successfully (no content)"
// @Failure      400  {object}  ErrorResponseDTO "Invalid customer ID format"
// @Failure      401  {object}  ErrorResponseDTO "Unauthorized - business_id not found"
// @Failure      404  {object}  ErrorResponseDTO "Customer not found"
// @Failure      403  {object}  ErrorResponseDTO "Access denied - customer not in this business"
// @Failure      500  {object}  ErrorResponseDTO "Internal server error"
// @Router       /api/v1/customers [delete]
func (h *Handler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	h.logger.Info("DeleteCustomer request received",
		logger.Field{Key: "remote_addr", Value: r.RemoteAddr},
	)

	// Get business_id from context
	businessID, err := GetBusinessIDFromContext(r)
	if err != nil {
		h.logger.Error("DeleteCustomer: business_id not found in context",
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusUnauthorized, "UNAUTHORIZED")
		return
	}

	// Parse customer_id from query parameter
	customerIDStr := r.URL.Query().Get("id")
	if customerIDStr == "" {
		h.logger.Warn("DeleteCustomer: customer_id query param missing")
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		h.logger.Error("DeleteCustomer: Invalid customer_id format",
			logger.Field{Key: "customer_id", Value: customerIDStr},
			logger.Field{Key: "error", Value: err.Error()},
		)
		h.sendError(w, http.StatusBadRequest, "VALIDATION_ERROR")
		return
	}

	h.logger.Info("DeleteCustomer: Processing soft delete",
		logger.Field{Key: "customer_id", Value: customerID.String()},
		logger.Field{Key: "business_id", Value: businessID.String()},
	)

	// Call service
	if err := h.service.DeleteCustomer(ctx, businessID, customerID); err != nil {
		h.logger.Error("DeleteCustomer: Service error",
			logger.Field{Key: "error", Value: err.Error()},
			logger.Field{Key: "customer_id", Value: customerID.String()},
		)

		if err == customer.ErrCustomerNotFound {
			h.sendError(w, http.StatusNotFound, "CUSTOMER_NOT_FOUND")
			return
		}
		if err == customer.ErrAccessDenied {
			h.sendError(w, http.StatusForbidden, "ACCESS_DENIED")
			return
		}

		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR")
		return
	}

	h.logger.Info("DeleteCustomer: Customer soft-deleted successfully",
		logger.Field{Key: "customer_id", Value: customerID.String()},
	)

	w.WriteHeader(http.StatusNoContent)
}
