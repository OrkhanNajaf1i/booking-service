// File: internal/http/routes/availability_routes.go
package routes

import (
	"net/http"

	availabilityHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/availability"
)

func RegisterAvailabilityRoutes(
	mux *http.ServeMux,
	h *availabilityHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}

	// Bos vaxtlarin hesablanmasi
	mux.Handle("GET /api/v1/availability", protected(h.GetAvailability))

	// Is saatlari + nahar fasilesi
	mux.Handle("GET /api/v1/availability/working-hours", protected(h.ListWorkingHours))
	mux.Handle("POST /api/v1/availability/working-hours", protected(h.SetWorkingHours))
	mux.Handle("PUT /api/v1/availability/working-hours", protected(h.BulkSetWorkingHours))
	mux.Handle("DELETE /api/v1/availability/working-hours", protected(h.DeleteWorkingHours))

	// Slot addimi, bufer, min xeberdarliq, auto-confirm
	mux.Handle("GET /api/v1/availability/settings", protected(h.GetSettings))
	mux.Handle("PUT /api/v1/availability/settings", protected(h.UpdateSettings))

	// Mezuniyyet / bloklanmis intervallar
	mux.Handle("GET /api/v1/availability/time-off", protected(h.ListTimeOff))
	mux.Handle("POST /api/v1/availability/time-off", protected(h.CreateTimeOff))
	mux.Handle("DELETE /api/v1/availability/time-off/{id}", protected(h.DeleteTimeOff))
}
