package http

import (
	"net/http"

	authDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	businessHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/routes"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handlers struct {
	Business *businessHandler.Handler
	Auth     *authHandler.Handler
}

func NewRouter(h Handlers, tokenManager authDomain.TokenManager) *http.ServeMux {
	mux := http.NewServeMux()
	routes.RegisterAuthRoutes(mux, h.Auth)
	authMid := middleware.AuthMiddleware(tokenManager)
	routes.RegisterBusinessRoutes(mux, h.Business, authMid)
	// mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("http://localhost:8080/swagger/doc.json"),
	))
	return mux
}
