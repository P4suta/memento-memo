//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/sakashita/memento-memo/internal/service"
)

func TestTag_CreateMemoWithTags(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "Learning #golang and #docker")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if len(memo.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(memo.Tags))
	}

	tagNames := make(map[string]bool)
	for _, tag := range memo.Tags {
		tagNames[tag.Name] = true
	}
	if !tagNames["golang"] || !tagNames["docker"] {
		t.Errorf("expected tags [golang, docker], got %v", memo.Tags)
	}
}

func TestTag_UpdateSyncsTags(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	svc := service.NewMemoService(testPool)

	memo, err := svc.Create(ctx, "Hello #go")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(memo.Tags) != 1 {
		t.Fatalf("expected 1 tag initially, got %d", len(memo.Tags))
	}

	updated, err := svc.Update(ctx, memo.ID, "Hello #rust #python")
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if len(updated.Tags) != 2 {
		t.Fatalf("expected 2 tags after update, got %d", len(updated.Tags))
	}

	tagNames := make(map[string]bool)
	for _, tag := range updated.Tags {
		tagNames[tag.Name] = true
	}
	if tagNames["go"] {
		t.Error("old tag 'go' should have been removed")
	}
	if !tagNames["rust"] || !tagNames["python"] {
		t.Errorf("expected tags [rust, python], got %v", updated.Tags)
	}
}

func TestTag_FilterByTag(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()
	memoSvc := service.NewMemoService(testPool)

	_, err := memoSvc.Create(ctx, "First #go memo")
	if err != nil {
		t.Fatalf("Create 1: %v", err)
	}
	_, err = memoSvc.Create(ctx, "Second #go #rust memo")
	if err != nil {
		t.Fatalf("Create 2: %v", err)
	}
	_, err = memoSvc.Create(ctx, "Third #python memo")
	if err != nil {
		t.Fatalf("Create 3: %v", err)
	}

	// Query memos by tag "go" directly via SQL
	rows, err := testPool.Query(ctx, `
		SELECT m.id FROM memos m
		JOIN memo_tags mt ON mt.memo_id = m.id
		JOIN tags t ON t.id = mt.tag_id
		WHERE t.name = 'go' AND m.deleted_at IS NULL
	`)
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	defer rows.Close()

	var count int
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("scan: %v", err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("expected 2 memos with tag 'go', got %d", count)
	}
}
