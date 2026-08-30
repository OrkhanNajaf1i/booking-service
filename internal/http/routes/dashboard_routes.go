// File: internal/http/routes/dashboard_routes.go
package routes

import (
	"net/http"

	dashboardHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/dashboard"
)

func RegisterDashboardRoutes(
	mux *http.ServeMux,
	h *dashboardHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}

	mux.Handle("GET /api/v1/dashboard/stats", protected(h.GetStats))
}
