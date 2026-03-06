//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/sakashita/memento-memo/internal/service"
)

func TestSearch_FullTextHit(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sessionID := createTestSession(t, time.Now())
	createTestMemo(t, sessionID, "Golang is awesome for building servers")
	createTestMemo(t, sessionID, "Python is great for data science")

	svc := service.NewSearchService(testPool)
	results, err := svc.Search(ctx, "Golang", nil, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Content != "Golang is awesome for building servers" {
		t.Errorf("unexpected content: %s", results[0].Content)
	}
}

func TestSearch_PartialMatch(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sessionID := createTestSession(t, time.Now())
	createTestMemo(t, sessionID, "testcontainers makes integration testing easy")

	svc := service.NewSearchService(testPool)
	results, err := svc.Search(ctx, "container", nil, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result for partial match, got %d", len(results))
	}
}

func TestSearch_NoResults(t *testing.T) {
	cleanDB(t)
	ctx := context.Background()

	sessionID := createTestSession(t, time.Now())
	createTestMemo(t, sessionID, "Hello world")

	svc := service.NewSearchService(testPool)
	results, err := svc.Search(ctx, "nonexistent-xyz-query", nil, 10)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}
