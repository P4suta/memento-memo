package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/db"
	"github.com/sakashita/memento-memo/internal/config"
	"github.com/sakashita/memento-memo/internal/server"
	"github.com/sakashita/memento-memo/internal/worker"
	"github.com/sakashita/memento-memo/internal/ws"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	cfg := config.MustLoad()

	// Setup structured logging
	logHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	})
	slog.SetDefault(slog.New(logHandler))

	slog.Info("starting memento-memo",
		slog.String("version", version),
		slog.String("commit", commit),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. DB connection with retry (design spec 4.6)
	pool, err := connectWithRetry(ctx, cfg.DSN)
	if err != nil {
		slog.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// 2. Run migrations (design spec 6.5 step 2)
	if err := db.RunMigrations(cfg.DSN); err != nil {
		slog.Error("migration failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// 3. Start background workers (design spec 6.5 step 3)
	reportGen := worker.NewReportGenerator(pool)
	sessionDet := worker.NewSessionDetector(pool)
	go reportGen.Run(ctx)
	go sessionDet.Run(ctx)

	// 4. Start LISTEN/NOTIFY listener
	hub := ws.NewHub()
	go ws.ListenForNotifications(ctx, cfg.DSN, hub)

	// 5. Start HTTP server (design spec 6.5 step 4)
	srv := server.New(cfg, pool, hub)
	go func() {
		if err := srv.Start(ctx); err != nil {
			slog.Error("server error", slog.String("error", err.Error()))
			cancel()
		}
	}()

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	slog.Info("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	cancel() // Cancel main context to stop workers

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", slog.String("error", err.Error()))
	}

	slog.Info("shutdown complete")
}

func connectWithRetry(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse pool config: %w", err)
	}
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 2
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute
	poolConfig.HealthCheckPeriod = time.Minute

	var pool *pgxpool.Pool
	for attempt := 0; attempt < 10; attempt++ {
		pool, err = pgxpool.NewWithConfig(ctx, poolConfig)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				slog.Info("database connected", slog.Int("attempt", attempt+1))
				return pool, nil
			}
			pool.Close()
		}
		wait := time.Duration(attempt+1) * time.Second
		slog.Warn("database not ready, retrying",
			slog.Int("attempt", attempt+1),
			slog.Duration("wait", wait),
		)
		time.Sleep(wait)
	}
	return nil, fmt.Errorf("failed to connect to database after 10 attempts: %w", err)
}
