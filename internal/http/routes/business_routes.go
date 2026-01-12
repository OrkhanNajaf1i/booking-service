// File: internal/http/routes/business_routes.go
package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
)

func RegisterBusinessRoutes(
	mux *http.ServeMux,
	handler *business.BusinessHandler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}
	mux.Handle("POST /api/v1/businesses/solo", protected(handler.CreateSoloBusiness))
	mux.Handle("POST /api/v1/businesses/multi", protected(handler.CreateMultiBusiness))
	mux.Handle("GET /api/v1/business", protected(handler.GetBusiness))
	mux.Handle("GET /api/v1/businesses/{id}", protected(handler.GetBusinessByID))
	mux.Handle("PUT /api/v1/business", protected(handler.UpdateBusiness))
}
