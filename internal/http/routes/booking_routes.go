// File: internal/http/routes/booking_routes.go
package routes

import (
	"net/http"

	bookingHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/booking"
)

func RegisterBookingRoutes(
	mux *http.ServeMux,
	h *bookingHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}

	mux.Handle("POST /api/v1/bookings", protected(h.CreateBooking))
	mux.Handle("GET /api/v1/bookings", protected(h.ListBookings))

	// "my" konkret seqmentdir, ona gore "{id}"-den ust tutulur.
	mux.Handle("GET /api/v1/bookings/my", protected(h.ListMyBookings))

	mux.Handle("GET /api/v1/bookings/{id}", protected(h.GetBooking))
	mux.Handle("PATCH /api/v1/bookings/{id}", protected(h.UpdateNotes))

	// Provider axini
	mux.Handle("POST /api/v1/bookings/{id}/confirm", protected(h.Confirm))
	mux.Handle("POST /api/v1/bookings/{id}/propose", protected(h.ProposeReschedule))
	mux.Handle("POST /api/v1/bookings/{id}/complete", protected(h.Complete))
	mux.Handle("POST /api/v1/bookings/{id}/no-show", protected(h.MarkNoShow))

	// Musteri axini
	mux.Handle("POST /api/v1/bookings/{id}/respond", protected(h.RespondToProposal))
	mux.Handle("POST /api/v1/bookings/{id}/cancel", protected(h.Cancel))
}
