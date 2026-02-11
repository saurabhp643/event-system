package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"event-ingestion-system/internal/config"
	"event-ingestion-system/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// Client represents a WebSocket client
type Client struct {
	conn     *websocket.Conn
	send     chan []byte
	tenantID string
}

// Hub manages WebSocket connections
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	config     *config.WebSocketConfig
}

// NewHub creates a new WebSocket hub
func NewHub(cfg *config.WebSocketConfig) *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		config:     cfg,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastToTenant sends a message to all clients of a specific tenant
func (h *Hub) BroadcastToTenant(tenantID string, event *models.Event) error {
	data, err := json.Marshal(event.ToEventResponse())
	if err != nil {
		return err
	}

	h.mu.RLock()
	for client := range h.clients {
		if client.tenantID == tenantID {
			select {
			case client.send <- data:
			default:
				close(client.send)
				delete(h.clients, client)
			}
		}
	}
	h.mu.RUnlock()

	return nil
}

// HandleWebSocket handles WebSocket connections
func (h *Hub) HandleWebSocket(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &Client{
		conn:     conn,
		send:     make(chan []byte, 256),
		tenantID: tenantID,
	}

	h.register <- client

	go client.writePump(h.config)
	go client.readPump(h, h.config)
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump(cfg *config.WebSocketConfig) {
	ticker := time.NewTicker(cfg.PingInterval)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(cfg.WriteTimeout))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(cfg.WriteTimeout))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump(h *Hub, cfg *config.WebSocketConfig) {
	defer func() {
		h.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(cfg.PongTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(cfg.PongTimeout))
		return nil
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error if needed
			}
			break
		}
	}
}

// WebSocketMessage represents a message sent to WebSocket clients
type WebSocketMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}
