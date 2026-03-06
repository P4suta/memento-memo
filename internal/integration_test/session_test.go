//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/sakashita/memento-memo/internal/service"
)

func TestSession_FirstMemoCreatesSession(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "first memo")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if memo.SessionID == 0 {
		t.Error("expected session to be created")
	}
}

func TestSession_ConsecutiveMemosUseSameSession(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo1, err := svc.Create(ctx, "memo 1")
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}

	memo2, err := svc.Create(ctx, "memo 2")
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}

	if memo1.SessionID != memo2.SessionID {
		t.Errorf("consecutive memos should use same session: got %d and %d",
			memo1.SessionID, memo2.SessionID)
	}
}

func TestSession_GapExceeded_CreatesNewSession(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create a session that ended 5 hours ago (exceeds 4h gap)
	fiveHoursAgo := time.Now().Add(-5 * time.Hour)
	oldSessionID := createTestSession(t, fiveHoursAgo)

	// Update the session's ended_at to 5 hours ago
	_, err := testPool.Exec(ctx,
		`UPDATE sessions SET ended_at = $1, memo_count = 1 WHERE id = $2`,
		fiveHoursAgo, oldSessionID)
	if err != nil {
		t.Fatalf("update session: %v", err)
	}

	svc := service.NewMemoService(testPool)
	memo, err := svc.Create(ctx, "new session memo")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if memo.SessionID == oldSessionID {
		t.Error("expected new session after 4h gap, but got same session")
	}
}

func TestSession_MaxDuration_CreatesNewSession(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	// Create a session that started 25 hours ago (exceeds 24h max)
	twentyFiveHoursAgo := time.Now().Add(-25 * time.Hour)
	oldSessionID := createTestSession(t, twentyFiveHoursAgo)

	// Update ended_at to recently (within 4h gap) but started > 24h ago
	_, err := testPool.Exec(ctx,
		`UPDATE sessions SET ended_at = $1, memo_count = 1 WHERE id = $2`,
		time.Now().Add(-1*time.Hour), oldSessionID)
	if err != nil {
		t.Fatalf("update session: %v", err)
	}

	svc := service.NewMemoService(testPool)
	memo, err := svc.Create(ctx, "new session memo")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if memo.SessionID == oldSessionID {
		t.Error("expected new session after 24h max duration, but got same session")
	}
}

func TestSession_StatsUpdated(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "hello")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var memoCount int32
	var totalChars int32
	err = testPool.QueryRow(ctx,
		`SELECT memo_count, total_chars FROM sessions WHERE id = $1`,
		memo.SessionID,
	).Scan(&memoCount, &totalChars)
	if err != nil {
		t.Fatalf("query session: %v", err)
	}

	if memoCount != 1 {
		t.Errorf("memo_count = %d, want 1", memoCount)
	}
	if totalChars != 5 { // "hello" = 5 runes
		t.Errorf("total_chars = %d, want 5", totalChars)
	}
}
