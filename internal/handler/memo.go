package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sakashita/memento-memo/internal/repository"
	"github.com/sakashita/memento-memo/internal/service"
)

type MemoHandler struct {
	memoService *service.MemoService
	queries     *repository.Queries
}

func NewMemoHandler(memoService *service.MemoService) *MemoHandler {
	return &MemoHandler{memoService: memoService}
}

func NewMemoHandlerWithQueries(memoService *service.MemoService, queries *repository.Queries) *MemoHandler {
	return &MemoHandler{memoService: memoService, queries: queries}
}

type createMemoRequest struct {
	Content string `json:"content"`
}

func (h *MemoHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createMemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrValidation)
		return
	}

	memo, err := h.memoService.Create(r.Context(), req.Content)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, memoResponse(memo))
}

func (h *MemoHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}

	// Handle "since" parameter for WebSocket reconnection diff fetch
	if since := r.URL.Query().Get("since"); since != "" {
		sinceTime, err := time.Parse(time.RFC3339, since)
		if err != nil {
			writeError(w, ErrValidation)
			return
		}
		memos, err := h.queries.ListMemosSince(r.Context(), repository.ListMemosSinceParams{
			Since:     service.ToPgTimestamptz(sinceTime),
			PageLimit: limit,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, paginatedResponse(memos, limit))
		return
	}

	// Cursor-based pagination
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr != "" {
		cursor, err := service.DecodeCursor(cursorStr)
		if err != nil {
			writeError(w, ErrValidation)
			return
		}
		memos, err := h.queries.ListMemos(r.Context(), repository.ListMemosParams{
			CursorCreatedAt: service.ToPgTimestamptz(cursor.CreatedAt),
			CursorID:        cursor.ID,
			PageLimit:       limit,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, paginatedResponse(memos, limit))
		return
	}

	memos, err := h.queries.ListMemosFirst(r.Context(), limit)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginatedResponse(memos, limit))
}

func (h *MemoHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	memo, err := h.memoService.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	if memo.DeletedAt.Valid {
		writeError(w, ErrMemoDeleted)
		return
	}

	writeJSON(w, http.StatusOK, memoResponse(memo))
}

type updateMemoRequest struct {
	Content string `json:"content"`
}

func (h *MemoHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	var req updateMemoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, ErrValidation)
		return
	}

	memo, err := h.memoService.Update(r.Context(), id, req.Content)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, memoResponse(memo))
}

func (h *MemoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	if err := h.memoService.Delete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MemoHandler) Restore(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	if err := h.memoService.Restore(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MemoHandler) PermanentDelete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	if err := h.memoService.PermanentDelete(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MemoHandler) TogglePin(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, ErrValidation)
		return
	}

	memo, err := h.memoService.TogglePin(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, memoResponse(memo))
}

func memoResponse(m *service.MemoWithTags) map[string]any {
	tags := make([]map[string]any, 0, len(m.Tags))
	for _, t := range m.Tags {
		tags = append(tags, map[string]any{
			"id":   t.ID,
			"name": t.Name,
		})
	}
	resp := map[string]any{
		"id":           m.ID,
		"session_id":   m.SessionID,
		"content":      m.Content,
		"content_html": m.ContentHtml,
		"tags":         tags,
		"pinned":       m.Pinned,
		"created_at":   m.CreatedAt.Time,
		"updated_at":   m.UpdatedAt.Time,
	}
	if m.DeletedAt.Valid {
		resp["deleted_at"] = m.DeletedAt.Time
	}
	return resp
}

func paginatedResponse(memos []repository.Memo, limit int32) map[string]any {
	items := make([]map[string]any, 0, len(memos))
	for _, m := range memos {
		item := map[string]any{
			"id":           m.ID,
			"session_id":   m.SessionID,
			"content":      m.Content,
			"content_html": m.ContentHtml,
			"pinned":       m.Pinned,
			"created_at":   m.CreatedAt.Time,
			"updated_at":   m.UpdatedAt.Time,
		}
		if m.DeletedAt.Valid {
			item["deleted_at"] = m.DeletedAt.Time
		}
		items = append(items, item)
	}

	hasMore := int32(len(memos)) == limit
	var nextCursor string
	if hasMore && len(memos) > 0 {
		last := memos[len(memos)-1]
		nextCursor = service.EncodeCursor(last.CreatedAt.Time, last.ID)
	}

	return map[string]any{
		"items":       items,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	}
}
