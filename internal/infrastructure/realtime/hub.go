// File: internal/infrastructure/realtime/hub.go
//
// WebSocket hub – acıq sessiyalari istifadeci uzre saxlayir ve
// bildirisleri aninda catdirir.
//
// Bir istifadecinin eyni anda bir nece sessiyasi ola biler
// (telefon + brauzer); hamisina paralel gonderilir.
package realtime

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/OrkhanNajaf1i/booking-service/internal/domain/notification"
	"github.com/OrkhanNajaf1i/booking-service/internal/logger"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// writeWait – bir mesajin yazilmasi ucun verilen vaxt.
	writeWait = 10 * time.Second
	// pongWait – bu muddet erzinde pong gelmese sessiya olu sayilir.
	pongWait = 60 * time.Second
	// pingPeriod – pongWait-dan qisa olmalidir ki, vaxtinda ping gedsin.
	pingPeriod = (pongWait * 9) / 10
	// maxMessageSize – client-den gelen mesajin limiti.
	maxMessageSize = 4096
	// sendBuffer – yavas client-i bloklamamaq ucun novbe olcusu.
	sendBuffer = 32
)

// Client – bir acıq WebSocket sessiyasi.
type Client struct {
	hub    *Hub
	conn   *websocket.Conn
	send   chan []byte
	userID uuid.UUID
	closed sync.Once
}

// Hub – istifadeci -> acıq sessiyalar xeritesi.
type Hub struct {
	mu      sync.RWMutex
	clients map[uuid.UUID]map[*Client]struct{}
	log     logger.Logger

	// broadcast – coxlu instansiya halinda xaricden (LISTEN/NOTIFY)
	// gelen hadiseler bu kanaldan kecir.
	external chan *notification.Envelope
}

// NewHub – hub yaradir.
func NewHub(log logger.Logger) *Hub {
	return &Hub{
		clients:  make(map[uuid.UUID]map[*Client]struct{}),
		log:      log,
		external: make(chan *notification.Envelope, 256),
	}
}

// Run – xaricden gelen hadiseleri yerli sessiyalara paylayir.
// Coxlu instansiya isletmirsinizse de tehlukesizdir: kanal sadece bos qalir.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			h.closeAll()
			return
		case envelope := <-h.external:
			if envelope == nil {
				continue
			}
			h.deliver(envelope.UserID, envelope)
		}
	}
}

// PublishToUser – notification.RealtimePublisher realizasiyasi.
// Bagli sessiya yoxdursa xeta deyil: bildiris onsuz da bazadadir
// ve push kanali ile de gedecek.
func (h *Hub) PublishToUser(_ context.Context, userID uuid.UUID, envelope *notification.Envelope) error {
	h.deliver(userID, envelope)
	return nil
}

// DeliverExternal – LISTEN/NOTIFY korpusunun cagirdigi giris noqtesi.
func (h *Hub) DeliverExternal(envelope *notification.Envelope) {
	select {
	case h.external <- envelope:
	default:
		h.log.Warn("Realtime novbesi doludur, hadise atildi",
			logger.Field{Key: "type", Value: string(envelope.Type)},
		)
	}
}

// OnlineUsers – hazirda bagli olan istifadeci sayi (metriklər/debug ucun).
func (h *Hub) OnlineUsers() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// IsOnline – istifadecinin acıq sessiyasi varmi?
func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[userID]) > 0
}

// ============================================================
// DAXILI
// ============================================================

func (h *Hub) deliver(userID uuid.UUID, envelope *notification.Envelope) {
	payload, err := json.Marshal(envelope)
	if err != nil {
		h.log.Error("Envelope serializasiya olunmadi",
			logger.Field{Key: "error", Value: err.Error()},
		)
		return
	}

	h.mu.RLock()
	sessions := make([]*Client, 0, len(h.clients[userID]))
	for client := range h.clients[userID] {
		sessions = append(sessions, client)
	}
	h.mu.RUnlock()

	for _, client := range sessions {
		select {
		case client.send <- payload:
		default:
			// Novbe dolub – client cavab vermir. Sessiyani bagliyiriq ki,
			// yavas bir baglanti butun hub-i saxlamasin.
			h.log.Warn("Yavas WebSocket sessiyasi baglanildi",
				logger.Field{Key: "user_id", Value: userID.String()},
			)
			client.close()
		}
	}
}

func (h *Hub) register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.clients[client.userID] == nil {
		h.clients[client.userID] = make(map[*Client]struct{}, 2)
	}
	h.clients[client.userID][client] = struct{}{}
}

func (h *Hub) unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	sessions, ok := h.clients[client.userID]
	if !ok {
		return
	}
	delete(sessions, client)
	if len(sessions) == 0 {
		delete(h.clients, client.userID)
	}
}

func (h *Hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()

	for userID, sessions := range h.clients {
		for client := range sessions {
			client.close()
		}
		delete(h.clients, userID)
	}
}

// ============================================================
// CLIENT PUMP-LARI
// ============================================================

func (c *Client) close() {
	c.closed.Do(func() {
		close(c.send)
		_ = c.conn.Close()
	})
}

// readPump – client-den gelen mesajlari oxuyur.
// Bizde client -> server mesaji yoxdur; oxuma yalniz baglantinin
// canli oldugunu bilmek (pong) ve baglanmani askarlamaq ucundur.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister(c)
		c.close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		if _, _, err := c.conn.ReadMessage(); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				c.hub.log.Debug("WebSocket gozlenilmez baglandi",
					logger.Field{Key: "user_id", Value: c.userID.String()},
					logger.Field{Key: "error", Value: err.Error()},
				)
			}
			return
		}
	}
}

// writePump – novbedeki mesajlari yazir ve periodik ping atir.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
