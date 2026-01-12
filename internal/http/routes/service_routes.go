// File: internal/http/routes/service_routes.go
package routes

import (
	"net/http"

	serviceHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/service"
)

func RegisterServiceRoutes(
	mux *http.ServeMux,
	h serviceHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}

	mux.Handle("GET /api/v1/services", protected(h.ListServices))
	mux.Handle("GET /api/v1/services/{id}", protected(h.GetService))
	mux.Handle("PUT /api/v1/services/{id}", protected(h.UpdateService))
	mux.Handle("DELETE /api/v1/services/{id}", protected(h.DeactivateService))
	mux.Handle("POST /api/v1/staff/{staff_id}/services", protected(h.AssignServicesToStaff))
	mux.Handle("GET /api/v1/staff/{staff_id}/services", protected(h.GetStaffServices))
	mux.Handle("DELETE /api/v1/staff/{staff_id}/services/{service_id}", protected(h.RemoveServiceFromStaff))
}
