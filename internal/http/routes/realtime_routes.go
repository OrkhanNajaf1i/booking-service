// File: internal/http/routes/realtime_routes.go
package routes

import (
	"net/http"

	"github.com/OrkhanNajaf1i/booking-service/internal/infrastructure/realtime"
)

// RegisterRealtimeRoutes – WebSocket giris noqtesi.
//
// AuthMiddleware tetbiq EDILMIR: brauzerin WebSocket API-si
// Authorization basligi qoya bilmir, ona gore token handler-in
// ozunde query parametrinden de oxunur.
func RegisterRealtimeRoutes(mux *http.ServeMux, h *realtime.Handler) {
	mux.Handle("GET /api/v1/ws", h)
}
