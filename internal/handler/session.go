package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/sakashita/memento-memo/internal/service"
)

type SessionHandler struct {
	sessionService *service.SessionService
}

func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

func (h *SessionHandler) List(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	fromTime, err := time.Parse("2006-01-02", from)
	if err != nil {
		fromTime = time.Now().AddDate(0, -1, 0)
	}
	toTime, err := time.Parse("2006-01-02", to)
	if err != nil {
		toTime = time.Now()
	}

	sessions, err := h.sessionService.List(r.Context(), fromTime, toTime)
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(sessions))
	for _, s := range sessions {
		items = append(items, map[string]any{
			"id":          s.ID,
			"started_at":  s.StartedAt,
			"ended_at":    s.EndedAt,
			"date_label":  s.DateLabel.Time.Format("2006-01-02"),
			"memo_count":  s.MemoCount,
			"total_chars": s.TotalChars,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *SessionHandler) Heatmap(w http.ResponseWriter, r *http.Request) {
	year := time.Now().Year()
	if y := r.URL.Query().Get("year"); y != "" {
		if v, err := strconv.Atoi(y); err == nil {
			year = v
		}
	}

	from := time.Date(year, 1, 1, 0, 0, 0, 0, time.Local)
	to := time.Date(year, 12, 31, 0, 0, 0, 0, time.Local)

	entries, err := h.sessionService.Heatmap(r.Context(), from, to)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": entries})
}

func (h *SessionHandler) Daily(w http.ResponseWriter, r *http.Request) {
	dateStr := r.PathValue("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	summary, err := h.sessionService.Daily(r.Context(), date)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}
