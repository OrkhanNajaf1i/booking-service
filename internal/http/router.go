package http

import (
	"net/http"

	authDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	businessHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
	customerHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/customer"
	locationHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/location"
	serviceHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/service"
	staffHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/staff"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/routes"
	httpSwagger "github.com/swaggo/http-swagger"
)

type Handlers struct {
	Business *businessHandler.BusinessHandler
	Auth     *authHandler.Handler
	Location locationHandler.Handler
	Staff    staffHandler.Handler
	Service  serviceHandler.Handler
	Customer customerHandler.Handler
}

func NewRouter(h Handlers, tokenManager authDomain.TokenManager) *http.ServeMux {
	mux := http.NewServeMux()
	routes.RegisterAuthRoutes(mux, h.Auth)
	authMiddleware := middleware.AuthMiddleware(tokenManager)
	routes.RegisterBusinessRoutes(mux, h.Business, authMiddleware)
	routes.RegisterLocationRoutes(mux, h.Location, authMiddleware)
	routes.RegisterStaffRoutes(mux, h.Staff, authMiddleware)
	routes.RegisterServiceRoutes(mux, h.Service, authMiddleware)
	routes.RegisterCustomerRoutes(mux, h.Customer, authMiddleware)
	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	return mux
}
