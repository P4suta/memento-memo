//go:build integration

package integration_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sakashita/memento-memo/internal/service"
)

func TestMemo_CreateAndGet(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "Hello #world")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if memo.Content != "Hello #world" {
		t.Errorf("Content = %q, want %q", memo.Content, "Hello #world")
	}
	if memo.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if memo.SessionID == 0 {
		t.Error("expected non-zero SessionID")
	}

	got, err := svc.Get(ctx, memo.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Content != memo.Content {
		t.Errorf("Get content = %q, want %q", got.Content, memo.Content)
	}
}

func TestMemo_Update(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "original")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	updated, err := svc.Update(ctx, memo.ID, "updated content")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Content != "updated content" {
		t.Errorf("Content = %q, want %q", updated.Content, "updated content")
	}
}

func TestMemo_SoftDeleteAndRestore(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "to delete")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.Delete(ctx, memo.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	// Memo should still exist (soft deleted)
	got, err := svc.Get(ctx, memo.ID)
	if err != nil {
		t.Fatalf("Get after soft delete: %v", err)
	}
	if !got.DeletedAt.Valid {
		t.Error("expected DeletedAt to be set after soft delete")
	}

	if err := svc.Restore(ctx, memo.ID); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	got, err = svc.Get(ctx, memo.ID)
	if err != nil {
		t.Fatalf("Get after restore: %v", err)
	}
	if got.DeletedAt.Valid {
		t.Error("expected DeletedAt to be null after restore")
	}
}

func TestMemo_PermanentDelete(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "to permanently delete")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := svc.PermanentDelete(ctx, memo.ID); err != nil {
		t.Fatalf("PermanentDelete: %v", err)
	}

	_, err = svc.Get(ctx, memo.ID)
	var appErr *service.AppError
	if !errors.As(err, &appErr) || appErr.Code != "MEMO_NOT_FOUND" {
		t.Errorf("expected MEMO_NOT_FOUND, got: %v", err)
	}
}

func TestMemo_TogglePin(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "pin me")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if memo.Pinned {
		t.Error("expected Pinned=false initially")
	}

	pinned, err := svc.TogglePin(ctx, memo.ID)
	if err != nil {
		t.Fatalf("TogglePin: %v", err)
	}
	if !pinned.Pinned {
		t.Error("expected Pinned=true after toggle")
	}

	unpinned, err := svc.TogglePin(ctx, memo.ID)
	if err != nil {
		t.Fatalf("TogglePin again: %v", err)
	}
	if unpinned.Pinned {
		t.Error("expected Pinned=false after second toggle")
	}
}

func TestMemo_ValidationErrors(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	// Empty content
	_, err := svc.Create(ctx, "")
	var appErr *service.AppError
	if !errors.As(err, &appErr) || appErr.Code != "MEMO_EMPTY" {
		t.Errorf("expected MEMO_EMPTY for empty content, got: %v", err)
	}

	// Too long content (>10000 runes)
	longContent := make([]rune, 10001)
	for i := range longContent {
		longContent[i] = 'a'
	}
	_, err = svc.Create(ctx, string(longContent))
	if !errors.As(err, &appErr) || appErr.Code != "MEMO_TOO_LONG" {
		t.Errorf("expected MEMO_TOO_LONG, got: %v", err)
	}
}
