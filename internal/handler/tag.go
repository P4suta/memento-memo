package handler

import (
	"net/http"
	"strconv"

	"github.com/sakashita/memento-memo/internal/repository"
	"github.com/sakashita/memento-memo/internal/service"
)

type TagHandler struct {
	tagService *service.TagService
	queries    *repository.Queries
}

func NewTagHandler(tagService *service.TagService) *TagHandler {
	return &TagHandler{tagService: tagService}
}

func NewTagHandlerWithQueries(tagService *service.TagService, queries *repository.Queries) *TagHandler {
	return &TagHandler{tagService: tagService, queries: queries}
}

func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	tags, err := h.queries.ListTags(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}

	items := make([]map[string]any, 0, len(tags))
	for _, t := range tags {
		items = append(items, map[string]any{
			"id":         t.ID,
			"name":       t.Name,
			"memo_count": t.MemoCount,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *TagHandler) Memos(w http.ResponseWriter, r *http.Request) {
	tagName := r.PathValue("name")
	if tagName == "" {
		writeError(w, ErrTagNotFound)
		return
	}

	limit := int32(20)
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = int32(v)
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr != "" {
		cursor, err := service.DecodeCursor(cursorStr)
		if err != nil {
			writeError(w, ErrValidation)
			return
		}
		memos, err := h.queries.ListMemosByTag(r.Context(), repository.ListMemosByTagParams{
			TagName:         tagName,
			CursorCreatedAt: service.ToPgTimestamptz(cursor.CreatedAt),
			CursorID:        cursor.ID,
			PageLimit:       limit,
		})
		if err != nil {
			writeError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, paginatedMemoResponse(memos, limit))
		return
	}

	memos, err := h.queries.ListMemosByTagFirst(r.Context(), repository.ListMemosByTagFirstParams{
		TagName:   tagName,
		PageLimit: limit,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, paginatedMemoResponse(memos, limit))
}

func paginatedMemoResponse(memos []repository.Memo, limit int32) map[string]any {
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

	return map[string]any{
		"items":       items,
		"next_cursor": nextCursor,
		"has_more":    hasMore,
	}
}
