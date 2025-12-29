package http

import (
	"net/http"

	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	businessHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/routes"
)

type Handlers struct {
	Business *businessHandler.Handler
	Auth     *authHandler.Handler
}

func NewRouter(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	routes.RegisterBusinessRoutes(mux, h.Business)
	routes.RegisterAuthRoutes(mux, h.Auth)
	return mux
}
