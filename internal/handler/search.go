package handler

import (
	"net/http"
	"strconv"

	"github.com/sakashita/memento-memo/internal/service"
)

type SearchHandler struct {
	searchService *service.SearchService
}

func NewSearchHandler(searchService *service.SearchService) *SearchHandler {
	return &SearchHandler{searchService: searchService}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, ErrValidation)
		return
	}

	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}

	var cursor *service.Cursor
	if cursorStr := r.URL.Query().Get("cursor"); cursorStr != "" {
		var err error
		cursor, err = service.DecodeCursor(cursorStr)
		if err != nil {
			writeError(w, ErrValidation)
			return
		}
	}

	memos, err := h.searchService.Search(r.Context(), query, cursor, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(memos))
	for _, m := range memos {
		items = append(items, map[string]any{
			"id":           m.ID,
			"session_id":   m.SessionID,
			"content":      m.Content,
			"content_html": m.ContentHtml,
			"pinned":       m.Pinned,
			"created_at":   m.CreatedAt.Time,
			"updated_at":   m.UpdatedAt.Time,
		})
	}

	hasMore := int32(len(memos)) == limit
	var nextCursor string
	if hasMore && len(memos) > 0 {
		last := memos[len(memos)-1]
		nextCursor = service.EncodeCursor(last.CreatedAt.Time, last.ID)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items":       items,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	})
}
