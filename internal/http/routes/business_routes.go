package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers"
)

func RegisterBusinessRoutes(mux *http.ServeMux, h *handlers.BusinessHandler) {
	mux.HandleFunc("POST /businesses", h.Create)
	mux.HandleFunc("GET /businesses", h.GetBusinessByID)
}
