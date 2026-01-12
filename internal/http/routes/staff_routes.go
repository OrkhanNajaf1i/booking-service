// File: internal/http/routes/staff_routes.go
package routes

import (
	"net/http"

	staffHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/staff"
)

func RegisterStaffRoutes(
	mux *http.ServeMux,
	h staffHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}

	mux.Handle("POST /api/v1/staff", protected(h.CreateStaffProfile))
	mux.Handle("GET /api/v1/staff", protected(h.ListStaff))
	mux.Handle("GET /api/v1/staff/{id}", protected(h.GetStaff))
	mux.Handle("PUT /api/v1/staff/{id}", protected(h.UpdateStaff))
	mux.Handle("DELETE /api/v1/staff/{id}", protected(h.DeactivateStaff))
	mux.Handle("POST /api/v1/staff/invites", protected(h.InviteStaff))
	mux.Handle("POST /api/v1/staff/invites/accept", protected(h.AcceptInvite))
	mux.Handle("POST /api/v1/staff/invites/validate", http.HandlerFunc(h.ValidateInviteToken))
}
