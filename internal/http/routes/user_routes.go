package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers"
)

func RegisterUserRoutes(mux *http.ServeMux, h *handlers.UserHandler) {
	mux.HandleFunc("POST /users", h.Create)
	mux.HandleFunc("GET /users/by-phone", h.GetByPhone)
}
