package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type MemoChangePayload struct {
	Action    string `json:"action"`
	MemoID    int64  `json:"memo_id"`
	SessionID int64  `json:"session_id"`
}

// ListenForNotifications establishes a dedicated connection for LISTEN/NOTIFY.
// This is separate from the connection pool per design spec 9.3.
func ListenForNotifications(ctx context.Context, dsn string, hub *Hub) {
	for {
		if err := listenLoop(ctx, dsn, hub); err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Error("LISTEN connection lost, reconnecting", slog.String("error", err.Error()))
			time.Sleep(5 * time.Second)
		}
	}
}

func listenLoop(ctx context.Context, dsn string, hub *Hub) error {
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return fmt.Errorf("connect for LISTEN: %w", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, "LISTEN memo_changes")
	if err != nil {
		return fmt.Errorf("LISTEN: %w", err)
	}

	slog.Info("LISTEN memo_changes started")

	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return fmt.Errorf("wait for notification: %w", err)
		}

		var payload MemoChangePayload
		if err := json.Unmarshal([]byte(notification.Payload), &payload); err != nil {
			slog.Error("failed to unmarshal notification", slog.String("error", err.Error()))
			continue
		}

		var msgType string
		switch payload.Action {
		case "INSERT":
			msgType = "memo.created"
		case "UPDATE":
			msgType = "memo.updated"
		case "DELETE":
			msgType = "memo.deleted"
		default:
			continue
		}

		hub.Broadcast(msgType, payload)
	}
}
