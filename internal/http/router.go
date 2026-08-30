package http

import (
	"net/http"

	authDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	authHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/auth"
	availabilityHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/availability"
	bookingHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/booking"
	businessHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/business"
	customerHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/customer"
	locationHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/location"
	notificationHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/notification"
	publicHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/public"
	serviceHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/service"
	staffHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/staff"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/middleware"
	"github.com/OrkhanNajaf1i/booking-service/internal/http/routes"
	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/realtime"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Handlers – router-in ehtiyac duydugu butun handler-ler.
//
// QEYD: kohne slot handler-i (onceden generasiya olunan slot setirleri)
// artiq qosulmur. Bos vaxtlar availability paketinde runtime-da
// hesablanir; is saatlari cedvelinin tek sahibi de odur.
type Handlers struct {
	Business     *businessHandler.BusinessHandler
	Auth         *authHandler.Handler
	Location     locationHandler.Handler
	Staff        staffHandler.Handler
	Service      serviceHandler.Handler
	Customer     customerHandler.Handler
	Availability *availabilityHandler.Handler
	Booking      *bookingHandler.Handler
	Notification *notificationHandler.Handler
	Public       *publicHandler.Handler
	Realtime     *realtime.Handler
}

func NewRouter(h Handlers, tokenManager authDomain.TokenManager) *http.ServeMux {
	mux := http.NewServeMux()
	authMiddleware := middleware.AuthMiddleware(tokenManager)

	routes.RegisterAuthRoutes(mux, h.Auth)
	routes.RegisterBusinessRoutes(mux, h.Business, authMiddleware)
	routes.RegisterLocationRoutes(mux, h.Location, authMiddleware)
	routes.RegisterStaffRoutes(mux, h.Staff, authMiddleware)
	routes.RegisterServiceRoutes(mux, h.Service, authMiddleware)
	routes.RegisterCustomerRoutes(mux, h.Customer, authMiddleware)
	routes.RegisterAvailabilityRoutes(mux, h.Availability, authMiddleware)
	routes.RegisterBookingRoutes(mux, h.Booking, authMiddleware)
	routes.RegisterNotificationRoutes(mux, h.Notification, authMiddleware)

	// Musterinin login olmadan gore bilecekleri.
	if h.Public != nil {
		routes.RegisterPublicRoutes(mux, h.Public)
	}

	// WebSocket ozu auth edir (brauzer basliq qoya bilmir).
	if h.Realtime != nil {
		routes.RegisterRealtimeRoutes(mux, h.Realtime)
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.Handle("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return mux
}
