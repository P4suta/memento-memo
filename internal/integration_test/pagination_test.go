//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/sakashita/memento-memo/internal/repository"
	"github.com/sakashita/memento-memo/internal/service"
)

func TestPagination_LimitAndCursor(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sessionID := createTestSession(t, time.Now())
	for i := 0; i < 5; i++ {
		createTestMemo(t, sessionID, "memo "+string(rune('A'+i)))
	}

	queries := repository.New(testPool)

	// First page: limit 2
	first, err := queries.ListMemosFirst(ctx, 2)
	if err != nil {
		t.Fatalf("ListMemosFirst: %v", err)
	}
	if len(first) != 2 {
		t.Fatalf("expected 2 results in first page, got %d", len(first))
	}

	// Second page using cursor from last item of first page
	last := first[len(first)-1]
	cursor := service.ToPgTimestamptz(last.CreatedAt.Time)
	second, err := queries.ListMemos(ctx, repository.ListMemosParams{
		CursorCreatedAt: cursor,
		CursorID:        last.ID,
		PageLimit:       2,
	})
	if err != nil {
		t.Fatalf("ListMemos: %v", err)
	}
	if len(second) != 2 {
		t.Fatalf("expected 2 results in second page, got %d", len(second))
	}

	// Ensure no overlap
	if second[0].ID == first[0].ID || second[0].ID == first[1].ID {
		t.Error("second page overlaps with first page")
	}

	// Third page: should have 1 remaining
	lastSecond := second[len(second)-1]
	cursor2 := service.ToPgTimestamptz(lastSecond.CreatedAt.Time)
	third, err := queries.ListMemos(ctx, repository.ListMemosParams{
		CursorCreatedAt: cursor2,
		CursorID:        lastSecond.ID,
		PageLimit:       2,
	})
	if err != nil {
		t.Fatalf("ListMemos third: %v", err)
	}
	if len(third) != 1 {
		t.Fatalf("expected 1 result in third page, got %d", len(third))
	}
}

func TestPagination_SinceParameter(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sessionID := createTestSession(t, time.Now())

	// Create memos with small time gaps
	for i := 0; i < 3; i++ {
		createTestMemo(t, sessionID, "memo "+string(rune('A'+i)))
	}

	queries := repository.New(testPool)

	// Get all memos first
	all, err := queries.ListMemosFirst(ctx, 10)
	if err != nil {
		t.Fatalf("ListMemosFirst: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 memos, got %d", len(all))
	}

	// Use the oldest memo's created_at as "since"
	oldest := all[len(all)-1]
	sinceTime := service.ToPgTimestamptz(oldest.CreatedAt.Time)

	since, err := queries.ListMemosSince(ctx, repository.ListMemosSinceParams{
		Since:     sinceTime,
		PageLimit: 10,
	})
	if err != nil {
		t.Fatalf("ListMemosSince: %v", err)
	}

	// Should return memos created AFTER the oldest one
	if len(since) != 2 {
		t.Errorf("expected 2 memos since oldest, got %d", len(since))
	}
}
