package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
)

func RegisterBusinessRoutes(mux *http.ServeMux, h *business.Handler, authMiddleware func(http.Handler) http.Handler) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}

	mux.Handle("POST /api/v1/businesses/solo", protected(h.CreateSoloBusiness))
	mux.Handle("POST /api/v1/businesses/multi", protected(h.CreateMultiBusiness))
	mux.Handle("POST /api/v1/businesses/{id}/invites", protected(h.InviteStaff))
	mux.Handle("POST /api/v1/businesses/{id}/locations/default", protected(h.CreateDefaultLocation))
	mux.Handle("POST /api/v1/businesses/join-with-invite", protected(h.JoinWithInvite))
}
