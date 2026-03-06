package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/sakashita/memento-memo/internal/service"
)

// Re-export error variables for handler-level use
var (
	ErrMemoNotFound = service.ErrMemoNotFound
	ErrMemoTooLong  = service.ErrMemoTooLong
	ErrMemoEmpty    = service.ErrMemoEmpty
	ErrMemoDeleted  = service.ErrMemoDeleted
	ErrTagNotFound  = service.ErrTagNotFound
	ErrRateLimited  = service.ErrRateLimited
	ErrValidation   = service.ErrValidation
	ErrInternal     = service.ErrInternal
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, err error) {
	var appErr *service.AppError
	if errors.As(err, &appErr) {
		writeJSON(w, appErr.Status, map[string]any{"error": appErr})
		return
	}
	slog.Error("unhandled error", slog.String("error", err.Error()))
	writeJSON(w, 500, map[string]any{"error": ErrInternal})
}
