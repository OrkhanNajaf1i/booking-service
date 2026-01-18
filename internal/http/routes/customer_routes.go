// File: internal/http/routes/customer_routes.go
package routes

import (
	"net/http"

	customerHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/customer"
)

func RegisterCustomerRoutes(
	mux *http.ServeMux,
	h customerHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}
	mux.Handle("POST /api/v1/customers", protected(h.CreateCustomer))
	mux.Handle("GET /api/v1/customers", protected(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id") != "" {
			h.GetCustomer(w, r)
		}
		h.ListCustomers(w, r)
	}))
	mux.Handle("PUT /api/v1/customers", protected(h.UpdateCustomer))
	mux.Handle("DELETE /api/v1/customers", protected(h.DeleteCustomer))

}
