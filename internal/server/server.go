package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sakashita/memento-memo/internal/config"
	"github.com/sakashita/memento-memo/internal/handler"
	"github.com/sakashita/memento-memo/internal/repository"
	"github.com/sakashita/memento-memo/internal/service"
	"github.com/sakashita/memento-memo/internal/ws"
)

type Server struct {
	httpServer *http.Server
	pool       *pgxpool.Pool
	hub        *ws.Hub
	cfg        *config.Config
}

func New(cfg *config.Config, pool *pgxpool.Pool, hub *ws.Hub) *Server {
	queries := repository.New(pool)

	memoService := service.NewMemoService(pool)
	tagService := service.NewTagService(pool)
	searchService := service.NewSearchService(pool)
	sessionService := service.NewSessionService(pool)
	reportService := service.NewReportService(pool)

	memoHandler := handler.NewMemoHandlerWithQueries(memoService, queries)
	tagHandler := handler.NewTagHandlerWithQueries(tagService, queries)
	searchHandler := handler.NewSearchHandler(searchService)
	sessionHandler := handler.NewSessionHandler(sessionService)
	reportHandler := handler.NewReportHandler(reportService)
	healthHandler := handler.NewHealthHandler(pool)
	wsHandler := handler.NewWSHandler(hub)

	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("POST /api/v1/memos", memoHandler.Create)
	mux.HandleFunc("GET /api/v1/memos", memoHandler.List)
	mux.HandleFunc("GET /api/v1/memos/{id}", memoHandler.Get)
	mux.HandleFunc("PATCH /api/v1/memos/{id}", memoHandler.Update)
	mux.HandleFunc("DELETE /api/v1/memos/{id}", memoHandler.Delete)
	mux.HandleFunc("POST /api/v1/memos/{id}/restore", memoHandler.Restore)
	mux.HandleFunc("DELETE /api/v1/memos/{id}/permanent", memoHandler.PermanentDelete)
	mux.HandleFunc("POST /api/v1/memos/{id}/pin", memoHandler.TogglePin)

	mux.HandleFunc("GET /api/v1/search", searchHandler.Search)

	mux.HandleFunc("GET /api/v1/tags", tagHandler.List)
	mux.HandleFunc("GET /api/v1/tags/{name}/memos", tagHandler.Memos)

	mux.HandleFunc("GET /api/v1/sessions", sessionHandler.List)
	mux.HandleFunc("GET /api/v1/calendar/heatmap", sessionHandler.Heatmap)
	mux.HandleFunc("GET /api/v1/calendar/daily/{date}", sessionHandler.Daily)

	mux.HandleFunc("GET /api/v1/reports/daily/{date}", reportHandler.Daily)
	mux.HandleFunc("GET /api/v1/reports/stats", reportHandler.Stats)

	mux.HandleFunc("GET /api/v1/health", healthHandler.Health)
	mux.HandleFunc("GET /api/v1/ws", wsHandler.Handle)

	// SPA fallback: serve static files or index.html
	mux.Handle("/", spaHandler())

	// Apply middleware chain
	h := Chain(
		Recovery,
		RequestID,
		Logger,
		SecurityHeaders,
		MaxBodySize(1<<20),
		RateLimit,
	)(mux)

	return &Server{
		httpServer: &http.Server{
			Addr:              ":" + cfg.Port,
			Handler:           h,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      60 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxHeaderBytes:    1 << 20,
		},
		pool: pool,
		hub:  hub,
		cfg:  cfg,
	}
}

func (s *Server) Start(ctx context.Context) error {
	// Start WebSocket hub
	go s.hub.Run(ctx)

	slog.Info("server starting", slog.String("addr", s.httpServer.Addr))
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("listen: %w", err)
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}
