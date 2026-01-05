package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
)

func RegisterBusinessRoutes(mux *http.ServeMux, h *business.Handler) {
	mux.HandleFunc("POST /api/v1/businesses/solo", h.CreateSoloBusiness)
	mux.HandleFunc("POST /api/v1/businesses/multi", h.CreateMultiBusiness)
	mux.HandleFunc("POST /api/v1/businesses/{id}/invites", h.InviteStaff)
	mux.HandleFunc("POST /api/v1/businesses/join-with-invite", h.JoinWithInvite)
	mux.HandleFunc("POST /api/v1/businesses/{id}/locations/default", h.CreateDefaultLocation)
}
