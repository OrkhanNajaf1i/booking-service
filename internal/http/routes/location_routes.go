// File: internal/http/routes/location_routes.go
package routes

import (
	"net/http"

	locationHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/location"
)

func RegisterLocationRoutes(
	mux *http.ServeMux,
	handler locationHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}
	mux.Handle("POST /api/v1/locations", protected(handler.CreateLocation))
	mux.Handle("GET /api/v1/locations", protected(handler.ListLocations))
	mux.Handle("GET /api/v1/locations/{id}", protected(handler.GetLocation))
	mux.Handle("PUT /api/v1/locations/{id}", protected(handler.UpdateLocation))
	mux.Handle("DELETE /api/v1/locations/{id}", protected(handler.DeactivateLocation))
	mux.Handle("POST /api/v1/locations/{id}/activate", protected(handler.ActivateLocation))
	// Tam silme ayri yoldadir: DELETE /locations/{id} yumsaq silmedir
	// (deaktiv), bu ise setri hemiselik goturur.
	mux.Handle("DELETE /api/v1/locations/{id}/permanent", protected(handler.DeleteLocation))
}
