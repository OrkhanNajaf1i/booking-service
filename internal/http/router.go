package http

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers"
	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/routes"
)

type Handlers struct {
	Business *handlers.BusinessHandler
	User     *handlers.UserHandler
	Auth     *authHandler.Handler
}

func NewRouter(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	routes.RegisterBusinessRoutes(mux, h.Business)
	routes.RegisterUserRoutes(mux, h.User)
	routes.RegisterAuthRoutes(mux, h.Auth)
	return mux
}
