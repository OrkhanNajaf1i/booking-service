// File: internal/http/routes/public_routes.go
package routes

import (
	"net/http"

	publicHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/public"
)

// RegisterPublicRoutes – auth telem etmeyen kesf endpoint-leri.
// Musteri login olmadan biznes/isci/xidmet secib bos vaxtlara baxa bilir;
// bron yaratmaq ucun ise POST /bookings (qorunur) lazimdir.
func RegisterPublicRoutes(mux *http.ServeMux, h *publicHandler.Handler) {
	mux.HandleFunc("GET /api/v1/public/businesses", h.ListBusinesses)
	mux.HandleFunc("GET /api/v1/public/businesses/{id}", h.GetBusiness)
	mux.HandleFunc("GET /api/v1/public/businesses/{id}/staff", h.ListStaff)
	mux.HandleFunc("GET /api/v1/public/businesses/{id}/services", h.ListServices)
	mux.HandleFunc("GET /api/v1/public/availability", h.GetAvailability)
}
