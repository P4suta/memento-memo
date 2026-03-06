package ws

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"nhooyr.io/websocket"
)

const (
	pingTimeout    = 60 * time.Second
	writeTimeout   = 10 * time.Second
	maxMessageSize = 512
)

type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

func NewClient(hub *Hub, conn *websocket.Conn) *Client {
	conn.SetReadLimit(maxMessageSize)
	return &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 64),
	}
}

func (c *Client) ReadPump(ctx context.Context) {
	defer func() {
		c.hub.RemoveClient(c)
		c.conn.Close(websocket.StatusNormalClosure, "")
	}()

	// Start write pump
	go c.writePump(ctx)

	for {
		_, data, err := c.conn.Read(ctx)
		if err != nil {
			return
		}

		var msg struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		if msg.Type == "ping" {
			pong, _ := json.Marshal(Message{
				Type:      "pong",
				Timestamp: time.Now().Format(time.RFC3339),
			})
			c.Send(pong)
		}
	}
}

func (c *Client) writePump(ctx context.Context) {
	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				return
			}
			writeCtx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(writeCtx, websocket.MessageText, msg)
			cancel()
			if err != nil {
				slog.Debug("websocket write failed", slog.String("error", err.Error()))
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

func (c *Client) Send(data []byte) {
	select {
	case c.send <- data:
	default:
		// Buffer full, drop message
		slog.Debug("websocket send buffer full, dropping message")
	}
}

func (c *Client) Close() {
	close(c.send)
}
