// File: internal/http/handlers/business/handler.go
package business

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/business"
	"github.com/google/uuid"
)

type BusinessHandler struct {
	businessService business.Service
}

func NewBusinessHandler(businessService business.Service) *BusinessHandler {
	return &BusinessHandler{
		businessService: businessService,
	}
}

// @Summary      Create Solo Business
// @Description  Creates a solo practitioner business for the authenticated user. Automatically generates business, location, and active staff records. User role set to solo_practitioner.
// @Tags         Business
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateSoloBusinessHTTPRequest true "Solo business data (BusinessName, Phone, ServiceCategory, Industry)"
// @Success      201  {object}  BusinessHTTPResponse "Solo business created successfully"
// @Failure      400  {object}  ErrorHTTPResponse "Validation error - invalid or missing required fields"
// @Failure      401  {object}  ErrorHTTPResponse "Unauthorized - user not authenticated"
// @Failure      409  {object}  ErrorHTTPResponse "Conflict - business already exists for user"
// @Failure      500  {object}  ErrorHTTPResponse "Internal server error"
// @Router       /api/v1/businesses/solo [post]
func (handler *BusinessHandler) CreateSoloBusiness(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		handler.respondWithError(writer, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := request.Context()
	userID, err := handler.extractUserIDFromContext(ctx)
	if err != nil {
		handler.respondWithError(writer, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	var httpRequest CreateSoloBusinessHTTPRequest
	if err := json.NewDecoder(request.Body).Decode(&httpRequest); err != nil {
		handler.respondWithError(writer, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}
	defer request.Body.Close()

	domainRequest := httpRequest.ToCreateBusinessRequest()

	businessEntity, err := handler.businessService.CreateBusiness(ctx, userID, domainRequest)
	if err != nil {
		handler.handleDomainError(writer, err)
		return
	}

	response := ToBusinessHTTPResponse(businessEntity)
	handler.respondWithJSON(writer, http.StatusCreated, response)
}

// @Summary      Create Multi-Staff Business
// @Description  Creates a multi-staff business for the authenticated user. Owner can invite staff members later. Staff management handled through separate endpoints.
// @Tags         Business
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateMultiBusinessHTTPRequest true "Multi-staff business data (BusinessName, Phone, ServiceCategory, Industry)"
// @Success      201  {object}  BusinessHTTPResponse "Multi-staff business created successfully"
// @Failure      400  {object}  ErrorHTTPResponse "Validation error - invalid or missing required fields"
// @Failure      401  {object}  ErrorHTTPResponse "Unauthorized - user not authenticated"
// @Failure      409  {object}  ErrorHTTPResponse "Conflict - business already exists for user"
// @Failure      500  {object}  ErrorHTTPResponse "Internal server error"
// @Router       /api/v1/businesses/multi [post]
func (handler *BusinessHandler) CreateMultiBusiness(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		handler.respondWithError(writer, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := request.Context()

	userID, err := handler.extractUserIDFromContext(ctx)
	if err != nil {
		handler.respondWithError(writer, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	var httpRequest CreateMultiBusinessHTTPRequest
	if err := json.NewDecoder(request.Body).Decode(&httpRequest); err != nil {
		handler.respondWithError(writer, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}
	defer request.Body.Close()

	domainRequest := httpRequest.ToCreateBusinessRequest()

	businessEntity, err := handler.businessService.CreateBusiness(ctx, userID, domainRequest)
	if err != nil {
		handler.handleDomainError(writer, err)
		return
	}

	response := ToBusinessHTTPResponse(businessEntity)
	handler.respondWithJSON(writer, http.StatusCreated, response)
}

// @Summary      Get My Business
// @Description  Retrieves the authenticated user's business information. User must be a business owner. Returns full business details including location, staff, and service information.
// @Tags         Business
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  BusinessHTTPResponse "Business details retrieved successfully"
// @Failure      401  {object}  ErrorHTTPResponse "Unauthorized - user not authenticated"
// @Failure      404  {object}  ErrorHTTPResponse "Business not found for user"
// @Failure      500  {object}  ErrorHTTPResponse "Internal server error"
// @Router       /api/v1/business [get]
func (handler *BusinessHandler) GetBusiness(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		handler.respondWithError(writer, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := request.Context()

	userID, err := handler.extractUserIDFromContext(ctx)
	if err != nil {
		handler.respondWithError(writer, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	businessEntity, err := handler.businessService.GetBusinessByOwner(ctx, userID)
	if err != nil {
		handler.handleDomainError(writer, err)
		return
	}

	response := ToBusinessHTTPResponse(businessEntity)
	handler.respondWithJSON(writer, http.StatusOK, response)
}

// @Summary      Get Business by ID
// @Description  Retrieves business information by business ID. Can be accessed by any authenticated user (public business view). Returns business profile, location, and active staff information.
// @Tags         Business
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Business ID (UUID format)"
// @Success      200  {object}  BusinessHTTPResponse "Business details retrieved successfully"
// @Failure      400  {object}  ErrorHTTPResponse "Invalid business ID format"
// @Failure      401  {object}  ErrorHTTPResponse "Unauthorized - user not authenticated"
// @Failure      404  {object}  ErrorHTTPResponse "Business not found"
// @Failure      500  {object}  ErrorHTTPResponse "Internal server error"
// @Router       /api/v1/businesses/{id} [get]
func (handler *BusinessHandler) GetBusinessByID(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodGet {
		handler.respondWithError(writer, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := request.Context()

	businessIDParam := handler.extractIDFromPath(request.URL.Path, "/api/v1/businesses/")
	if businessIDParam == "" {
		handler.respondWithError(writer, http.StatusBadRequest, "INVALID_BUSINESS_ID", "Business ID is required")
		return
	}

	businessID, err := uuid.Parse(businessIDParam)
	if err != nil {
		handler.respondWithError(writer, http.StatusBadRequest, "INVALID_BUSINESS_ID", "Invalid business ID format")
		return
	}

	businessEntity, err := handler.businessService.GetBusinessByID(ctx, businessID)
	if err != nil {
		handler.handleDomainError(writer, err)
		return
	}

	response := ToBusinessHTTPResponse(businessEntity)
	handler.respondWithJSON(writer, http.StatusOK, response)
}

// @Summary      Update Business
// @Description  Updates the authenticated user's business information. Only business owner can update. Updates business name, phone, service category, and industry fields. Validation applied to all fields.
// @Tags         Business
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body UpdateBusinessHTTPRequest true "Business update data (BusinessName, Phone, ServiceCategory, Industry - all optional)"
// @Success      200  {object}  SuccessHTTPResponse "Business updated successfully"
// @Failure      400  {object}  ErrorHTTPResponse "Validation error - invalid field values"
// @Failure      401  {object}  ErrorHTTPResponse "Unauthorized - user not authenticated or not owner"
// @Failure      404  {object}  ErrorHTTPResponse "Business not found"
// @Failure      500  {object}  ErrorHTTPResponse "Internal server error"
// @Router       /api/v1/business [put]
func (handler *BusinessHandler) UpdateBusiness(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPut {
		handler.respondWithError(writer, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := request.Context()

	businessID, err := handler.extractBusinessIDFromContext(ctx)
	if err != nil {
		handler.respondWithError(writer, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	var httpRequest UpdateBusinessHTTPRequest
	if err := json.NewDecoder(request.Body).Decode(&httpRequest); err != nil {
		handler.respondWithError(writer, http.StatusBadRequest, "INVALID_REQUEST_BODY", "Invalid request body")
		return
	}
	defer request.Body.Close()

	domainRequest := httpRequest.ToUpdateBusinessRequest()

	if err := handler.businessService.UpdateBusiness(ctx, businessID, domainRequest); err != nil {
		handler.handleDomainError(writer, err)
		return
	}

	handler.respondWithJSON(writer, http.StatusOK, SuccessHTTPResponse{
		Success: true,
		Message: "Business updated successfully",
	})
}

func (handler *BusinessHandler) extractIDFromPath(path, prefix string) string {
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	id := strings.TrimPrefix(path, prefix)
	id = strings.Split(id, "/")[0]
	return id
}

func (handler *BusinessHandler) extractUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userIDValue := ctx.Value("user_id")
	if userIDValue == nil {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}

	userIDString, ok := userIDValue.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID has invalid type")
	}

	userID, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID format: %w", err)
	}

	return userID, nil
}

func (handler *BusinessHandler) extractBusinessIDFromContext(ctx context.Context) (uuid.UUID, error) {
	businessIDValue := ctx.Value("business_id")
	if businessIDValue == nil {
		return uuid.Nil, fmt.Errorf("business ID not found in context")
	}

	businessIDString, ok := businessIDValue.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("business ID has invalid type")
	}

	if businessIDString == "" {
		return uuid.Nil, fmt.Errorf("business ID is empty")
	}

	businessID, err := uuid.Parse(businessIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid business ID format: %w", err)
	}

	return businessID, nil
}

func (handler *BusinessHandler) handleDomainError(writer http.ResponseWriter, err error) {
	if businessError, ok := err.(*business.BusinessError); ok {
		statusCode := handler.mapErrorCodeToHTTPStatus(businessError.Code)
		handler.respondWithError(writer, statusCode, businessError.Code, businessError.Message)
		return
	}

	handler.respondWithError(writer, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
}

func (handler *BusinessHandler) mapErrorCodeToHTTPStatus(errorCode string) int {
	errorStatusMap := map[string]int{
		"INVALID_OWNER_ID":           http.StatusBadRequest,
		"INVALID_BUSINESS_ID":        http.StatusBadRequest,
		"INVALID_REQUEST":            http.StatusBadRequest,
		"INVALID_DATA":               http.StatusBadRequest,
		"BUSINESS_NAME_REQUIRED":     http.StatusBadRequest,
		"BUSINESS_NAME_TOO_SHORT":    http.StatusBadRequest,
		"BUSINESS_NAME_TOO_LONG":     http.StatusBadRequest,
		"PHONE_REQUIRED":             http.StatusBadRequest,
		"PHONE_INVALID":              http.StatusBadRequest,
		"SERVICE_CATEGORY_REQUIRED":  http.StatusBadRequest,
		"SERVICE_CATEGORY_TOO_SHORT": http.StatusBadRequest,
		"SERVICE_CATEGORY_TOO_LONG":  http.StatusBadRequest,
		"INDUSTRY_REQUIRED":          http.StatusBadRequest,
		"INDUSTRY_TOO_SHORT":         http.StatusBadRequest,
		"INDUSTRY_TOO_LONG":          http.StatusBadRequest,
		"INVALID_BUSINESS_TYPE":      http.StatusBadRequest,
		"BUSINESS_NOT_FOUND":         http.StatusNotFound,
	}

	if status, exists := errorStatusMap[errorCode]; exists {
		return status
	}

	return http.StatusInternalServerError
}

func (handler *BusinessHandler) respondWithJSON(writer http.ResponseWriter, statusCode int, payload interface{}) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	if payload != nil {
		if err := json.NewEncoder(writer).Encode(payload); err != nil {
			fmt.Printf("Error encoding JSON response: %v\n", err)
		}
	}
}

func (handler *BusinessHandler) respondWithError(writer http.ResponseWriter, statusCode int, errorCode, message string) {
	errorResponse := ErrorHTTPResponse{
		Error:   "error",
		Code:    errorCode,
		Message: message,
	}
	handler.respondWithJSON(writer, statusCode, errorResponse)
}
