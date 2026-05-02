package daemon

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"go.uber.org/zap"
)

// Event represents an event to be broadcasted to WebSocket clients
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Hub maintains the set of active clients and broadcasts events
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan Event
	register   chan *Client
	unregister chan *Client
	mu         sync.Mutex
	logger     *zap.Logger
}

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan Event
}

func NewHub(logger *zap.Logger) *Hub {
	return &Hub{
		broadcast:  make(chan Event),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		logger:     logger,
	}
}

func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
		case event := <-h.broadcast:
			h.mu.Lock()
			for client := range h.clients {
				select {
				case client.send <- event:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// Broadcast sends an event to all connected clients
func (h *Hub) Broadcast(eventType string, data interface{}) {
	h.broadcast <- Event{Type: eventType, Data: data}
}

func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow all origins for the dashboard
	})
	if err != nil {
		h.logger.Error("websocket accept error", zap.Error(err))
		return
	}

	client := &Client{hub: h, conn: conn, send: make(chan Event, 256)}
	client.hub.register <- client

	// Start writer
	go client.writePump()
	// Keep connection alive
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		_, _, err := c.conn.Read(context.Background())
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		select {
		case event, ok := <-c.send:
			if !ok {
				c.conn.Close(websocket.StatusNormalClosure, "")
				return
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			err := wsjson.Write(ctx, c.conn, event)
			cancel()
			if err != nil {
				return
			}
		}
	}
}
