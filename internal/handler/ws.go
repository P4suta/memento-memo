package handler

import (
	"log/slog"
	"net/http"

	"github.com/sakashita/memento-memo/internal/ws"
	"nhooyr.io/websocket"
)

type WSHandler struct {
	hub *ws.Hub
}

func NewWSHandler(hub *ws.Hub) *WSHandler {
	return &WSHandler{hub: hub}
}

func (h *WSHandler) Handle(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		slog.Error("websocket accept failed", slog.String("error", err.Error()))
		return
	}

	h.hub.AddClient(r.Context(), conn)
}
