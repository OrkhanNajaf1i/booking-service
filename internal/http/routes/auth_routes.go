package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
)

func RegisterAuthRoutes(mux *http.ServeMux, h *auth.Handler) {
	mux.HandleFunc("POST /api/v1/auth/register", h.Register)
	mux.HandleFunc("POST /api/v1/auth/login", h.Login)
	mux.HandleFunc("POST /api/v1/auth/refresh", h.RefreshAccessToken)
	mux.HandleFunc("POST /api/v1/auth/forgot-password", h.ForgotPassword)
	mux.HandleFunc("POST /api/v1/auth/reset-password", h.ResetPassword)
	mux.HandleFunc("POST /api/v1/auth/logout", h.Logout)
}
