package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
)

func RegisterBusinessRoutes(mux *http.ServeMux, h *business.Handler) {
	mux.HandleFunc("POST /api/v1/businesses", h.CreateBusiness)
	mux.HandleFunc("GET api/v1/businesses/{id}", h.GetBusinessByID)
	mux.HandleFunc("POST api/v1/businesses/{id}/locations/default", h.CreateDefaultLocation)
}
