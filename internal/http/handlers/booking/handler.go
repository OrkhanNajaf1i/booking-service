// File: internal/http/handlers/booking/handler.go
package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/booking"
	"github.com/google/uuid"
)

type BookingHandler struct {
	service booking.Service
}

func NewBookingHandler(service booking.Service) *BookingHandler {
	return &BookingHandler{
		service: service,
	}
}

// CreateBooking - POST /api/v1/bookings
func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := r.Context()
	businessID, err := h.extractBusinessID(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	var reqBody CreateBookingHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_BODY", "Invalid request body")
		return
	}
	defer r.Body.Close()

	domainReq := reqBody.ToDomain(businessID)
	createdBooking, err := h.service.CreateBooking(ctx, domainReq)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, ToHTTPResponse(createdBooking))
}

// GetBooking - GET /api/v1/bookings/{id}
func (h *BookingHandler) GetBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := r.Context()
	businessID, err := h.extractBusinessID(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	bookingID, err := h.extractIDFromPath(r.URL.Path, "/api/v1/bookings/")
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid booking ID")
		return
	}

	b, err := h.service.GetBookingByID(ctx, businessID, bookingID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, ToHTTPResponse(b))
}

// CancelBooking - POST /api/v1/bookings/{id}/cancel
func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := r.Context()
	businessID, err := h.extractBusinessID(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	// Path parsing logic needs to handle /cancel suffix
	// Assuming path is /api/v1/bookings/{id}/cancel
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 2 {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_URL", "Invalid URL format")
		return
	}
	// ID is likely at index len-2 (before "cancel")
	idStr := pathParts[len(pathParts)-2]
	bookingID, err := uuid.Parse(idStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "INVALID_ID", "Invalid booking ID format")
		return
	}

	if err := h.service.CancelBooking(ctx, businessID, bookingID); err != nil {
		h.handleDomainError(w, err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: "Booking cancelled successfully",
	})
}

// ListBookings - GET /api/v1/bookings (Context-aware: Staff vs Customer vs Business)
// Bu sadələşdirilmiş versiyadır. Real layihədə query param-lar (date range, status) lazımdır.
func (h *BookingHandler) ListBookings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.respondWithError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Method not allowed")
		return
	}

	ctx := r.Context()
	businessID, err := h.extractBusinessID(ctx)
	if err != nil {
		h.respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", err.Error())
		return
	}

	// User ID-ni context-dən alırıq
	userIDVal := ctx.Value("user_id")
	if userIDVal == nil {
		h.respondWithError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User ID missing")
		return
	}
	// userID, _ := uuid.Parse(userIDVal.(string))

	// User Role-u context-dən alırıq (Auth middleware qoymalıdır)
	// Default olaraq Staff və ya Owner kimi yanaşırıqsa:
	// Burada sadəcə Business-in bütün bookinglərini qaytarırıq (Admin view)
	// Daha dərin logic üçün `user_role` yoxlanmalıdır.

	// Nümunə: Əgər sadəcə business booking-ləri götürürüksə:
	bookings, err := h.service.GetBusinessBookings(ctx, businessID)
	if err != nil {
		h.handleDomainError(w, err)
		return
	}

	// QEYD: Əgər user Customerdirsə `GetCustomerBookings(ctx, businessID, userID)` çağırılmalıdır.
	// Bu məntiq middleware-də set olunan "role" əsasında qurulur.

	h.respondWithJSON(w, http.StatusOK, ToHTTPResponseList(bookings))
}

// --- Helpers ---

func (h *BookingHandler) extractBusinessID(ctx context.Context) (uuid.UUID, error) {
	val := ctx.Value("business_id")
	if val == nil {
		return uuid.Nil, fmt.Errorf("business_id not found in context")
	}
	idStr, ok := val.(string)
	if !ok || idStr == "" {
		return uuid.Nil, fmt.Errorf("invalid business_id in context")
	}
	return uuid.Parse(idStr)
}

func (h *BookingHandler) extractIDFromPath(path, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(path, prefix) {
		return uuid.Nil, fmt.Errorf("path does not match prefix")
	}
	idStr := strings.TrimPrefix(path, prefix)
	// handle trailing slash or extra paths
	parts := strings.Split(idStr, "/")
	return uuid.Parse(parts[0])
}

func (h *BookingHandler) respondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		json.NewEncoder(w).Encode(payload)
	}
}

func (h *BookingHandler) respondWithError(w http.ResponseWriter, status int, code, message string) {
	h.respondWithJSON(w, status, ErrorResponse{
		Error:   "error",
		Code:    code,
		Message: message,
	})
}

func (h *BookingHandler) handleDomainError(w http.ResponseWriter, err error) {
	if bErr, ok := err.(*booking.BookingError); ok {
		status := http.StatusInternalServerError
		switch bErr.Code {
		case "INVALID_REQUEST", "INVALID_DATA", "INVALID_TIME", "NOTES_TOO_LONG", "STATUS_REQUIRED":
			status = http.StatusBadRequest
		case "BOOKING_NOT_FOUND", "CUSTOMER_NOT_FOUND", "STAFF_NOT_FOUND":
			status = http.StatusNotFound
		case "SLOT_UNAVAILABLE", "INVALID_TRANSITION":
			status = http.StatusConflict
		case "UNAUTHORIZED":
			status = http.StatusUnauthorized
		}
		h.respondWithError(w, status, bErr.Code, bErr.Message)
		return
	}
	// Unknown error
	fmt.Printf("Internal Server Error: %v\n", err)
	h.respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred")
}
