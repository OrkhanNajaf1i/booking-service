// File: internal/http/routes/booking_routes.go
package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/booking"
)

func RegisterBookingRoutes(
	mux *http.ServeMux,
	handler *booking.BookingHandler,
	authMiddleware func(http.Handler) http.Handler,
) {

	protected := func(h http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(h))
	}

	mux.Handle("POST /api/v1/bookings", protected(handler.CreateBooking))
	mux.Handle("GET /api/v1/bookings", protected(handler.ListBookings))
	mux.Handle("GET /api/v1/bookings/", protected(handler.GetBooking))
}
