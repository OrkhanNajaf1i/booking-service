// File: internal/infrastructure/realtime/handler.go
package realtime

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	authDomain "github.com/OrkhanNajaf1i/booking-service/internal/domain/auth"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/gorilla/websocket"
)

// allowedOrigins – bos olarsa butun origin-lere icaze verilir.
// Prod-da APP_WS_ALLOWED_ORIGINS ile mehdudlasdirin.
type Handler struct {
	hub            *Hub
	tokens         authDomain.TokenManager
	log            logger.Logger
	allowedOrigins []string
	upgrader       websocket.Upgrader
}

// NewHandler – WebSocket giris noqtesi.
func NewHandler(
	hub *Hub,
	tokens authDomain.TokenManager,
	log logger.Logger,
	allowedOrigins []string,
) *Handler {
	handler := &Handler{
		hub:            hub,
		tokens:         tokens,
		log:            log,
		allowedOrigins: allowedOrigins,
	}

	handler.upgrader = websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 10 * time.Second,
		CheckOrigin:      handler.checkOrigin,
	}

	return handler
}

// ServeHTTP – GET /api/v1/ws?token=<access_token>
//
// Brauzerin WebSocket API-si Authorization basligi qoymaga imkan
// vermir, ona gore token query parametri ile de qebul edilir.
// Mobil client basliqdan istifade ede biler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := extractToken(r)
	if token == "" {
		writeWSError(w, http.StatusUnauthorized, "NO_TOKEN", "token teleb olunur")
		return
	}

	claims, err := h.tokens.ValidateAccessToken(token)
	if err != nil {
		writeWSError(w, http.StatusUnauthorized, "INVALID_TOKEN", "token etibarsizdir")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Upgrade ozu cavabi yazib; burada yalniz loglayiriq.
		h.log.Warn("WebSocket upgrade ugursuz",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return
	}

	client := &Client{
		hub:    h.hub,
		conn:   conn,
		send:   make(chan []byte, sendBuffer),
		userID: claims.UserID,
	}

	h.hub.register(client)

	h.log.Info("WebSocket sessiyasi acildi",
		logger.Field{Key: "user_id", Value: claims.UserID.String()},
		logger.Field{Key: "online_users", Value: h.hub.OnlineUsers()},
	)

	// Client-e hazir oldugumuzu bildiririk ki, UI "canli" gostersin.
	ready, _ := json.Marshal(map[string]interface{}{
		"type":       "connection.ready",
		"user_id":    claims.UserID.String(),
		"created_at": time.Now(),
	})
	client.send <- ready

	go client.writePump()
	go client.readPump()
}

// checkOrigin – siyahi bosdursa hamisina icaze (dev rejimi).
func (h *Handler) checkOrigin(r *http.Request) bool {
	if len(h.allowedOrigins) == 0 {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		// Mobil client-lerde Origin olmur.
		return true
	}

	for _, allowed := range h.allowedOrigins {
		if strings.EqualFold(strings.TrimSpace(allowed), origin) {
			return true
		}
	}

	h.log.Warn("WebSocket origin redd edildi",
		logger.Field{Key: "origin", Value: origin},
	)
	return false
}

// extractToken – Authorization basligi, sonra ?token=, sonra
// Sec-WebSocket-Protocol.
func extractToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if parts := strings.Fields(header); len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}

	if token := r.URL.Query().Get("token"); token != "" {
		return token
	}

	protocol := r.Header.Get("Sec-WebSocket-Protocol")
	if strings.HasPrefix(protocol, "bearer,") {
		return strings.TrimSpace(strings.TrimPrefix(protocol, "bearer,"))
	}

	return ""
}

func writeWSError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"code":    code,
		"message": message,
	})
}
