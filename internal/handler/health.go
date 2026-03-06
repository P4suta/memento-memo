package handler

import (
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type HealthHandler struct {
	pool      *pgxpool.Pool
	startedAt time.Time
}

func NewHealthHandler(pool *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{pool: pool, startedAt: time.Now()}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := h.pool.Ping(r.Context()); err != nil {
		dbStatus = "disconnected"
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"version":  "1.0.0",
		"uptime":   time.Since(h.startedAt).String(),
		"database": dbStatus,
	})
}
