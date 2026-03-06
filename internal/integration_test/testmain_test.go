//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/db"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("memento_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate container: %v", err)
		}
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	// Run migrations
	if err := db.RunMigrations(connStr); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Create pool
	testPool, err = pgxpool.New(ctx, connStr)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer testPool.Close()

	os.Exit(m.Run())
}

// cleanDB truncates all tables between tests.
func cleanDB(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testPool.Exec(ctx, `
		TRUNCATE memo_tags, tags, memos, sessions, daily_reports RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("failed to clean database: %v", err)
	}
}

// createTestSession inserts a session at the given time and returns its ID.
func createTestSession(t *testing.T, startedAt time.Time) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	err := testPool.QueryRow(ctx,
		`INSERT INTO sessions (started_at, ended_at, date_label, memo_count, total_chars)
		 VALUES ($1, $1, $2, 0, 0) RETURNING id`,
		startedAt, startedAt,
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create test session: %v", err)
	}
	return id
}

// createTestMemo inserts a memo with given content and returns its ID.
func createTestMemo(t *testing.T, sessionID int64, content string) int64 {
	t.Helper()
	ctx := context.Background()
	var id int64
	err := testPool.QueryRow(ctx,
		`INSERT INTO memos (session_id, content, content_html) VALUES ($1, $2, $3) RETURNING id`,
		sessionID, content, fmt.Sprintf("<p>%s</p>", content),
	).Scan(&id)
	if err != nil {
		t.Fatalf("failed to create test memo: %v", err)
	}
	return id
}
