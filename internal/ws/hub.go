package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

type Message struct {
	Type      string `json:"type"`
	Payload   any    `json:"payload,omitempty"`
	Timestamp string `json:"timestamp"`
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
	}
}

func (h *Hub) Run(ctx context.Context) {
	<-ctx.Done()
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		c.Close()
	}
}

func (h *Hub) AddClient(ctx context.Context, conn *websocket.Conn) {
	client := NewClient(h, conn)
	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	slog.Info("websocket client connected", slog.Int("total", h.ClientCount()))

	client.ReadPump(ctx)
}

func (h *Hub) RemoveClient(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	slog.Info("websocket client disconnected", slog.Int("total", h.ClientCount()))
}

func (h *Hub) Broadcast(msgType string, payload any) {
	msg := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().Format(time.RFC3339),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		slog.Error("failed to marshal ws message", slog.String("error", err.Error()))
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		c.Send(data)
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
