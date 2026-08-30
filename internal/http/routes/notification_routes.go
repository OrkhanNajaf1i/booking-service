// File: internal/http/routes/notification_routes.go
package routes

import (
	"net/http"

	notificationHandler "github.com/OrkhanNajaf1i/booking-service/internal/http/handlers/notification"
)

func RegisterNotificationRoutes(
	mux *http.ServeMux,
	h *notificationHandler.Handler,
	authMiddleware func(http.Handler) http.Handler,
) {
	protected := func(handlerFunc http.HandlerFunc) http.Handler {
		return authMiddleware(http.HandlerFunc(handlerFunc))
	}

	mux.Handle("GET /api/v1/notifications", protected(h.List))
	mux.Handle("GET /api/v1/notifications/unread-count", protected(h.CountUnread))
	mux.Handle("POST /api/v1/notifications/read-all", protected(h.MarkAllRead))
	mux.Handle("POST /api/v1/notifications/{id}/read", protected(h.MarkRead))

	mux.Handle("POST /api/v1/notifications/devices", protected(h.RegisterDevice))
	mux.Handle("DELETE /api/v1/notifications/devices", protected(h.UnregisterDevice))
}
