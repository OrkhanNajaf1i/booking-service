package http

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/http/handlers"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/routes"
)

type Handlers struct {
	Business *handlers.BusinessHandler
	User     *handlers.UserHandler
}

func NewRouter(h Handlers) *http.ServeMux {
	mux := http.NewServeMux()
	routes.RegisterBusinessRoutes(mux, h.Business)
	routes.RegisterUserRoutes(mux, h.User)
	return mux
}
