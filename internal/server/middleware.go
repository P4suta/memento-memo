package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type contextKey string

const requestIDKey contextKey = "request_id"

// Chain applies middlewares in order (outermost first).
func Chain(middlewares ...func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(final http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			final = middlewares[i](final)
		}
		return final
	}
}

// Recovery recovers from panics and logs the stack trace.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				http.Error(w, `{"error":{"code":"INTERNAL_ERROR","message":"Internal server error","status":500}}`, http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// RequestID attaches a unique request ID to the context.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		rand.Read(b)
		id := hex.EncodeToString(b)
		w.Header().Set("X-Request-ID", id)
		ctx := context.WithValue(r.Context(), requestIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Logger logs each request with method, path, status, and latency.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		slog.Info("request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status", sw.status),
			slog.Duration("latency", time.Since(start)),
		)
	})
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

// SecurityHeaders adds standard security headers.
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy",
			"default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:;")
		next.ServeHTTP(w, r)
	})
}

// MaxBodySize limits the request body size.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit implements a simple token bucket rate limiter.
func RateLimit(next http.Handler) http.Handler {
	var (
		mu     sync.Mutex
		tokens float64 = 100
		last           = time.Now()
		rate           = 100.0 // tokens per second
		max            = 100.0
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip rate limit for WebSocket upgrades
		if r.Header.Get("Upgrade") == "websocket" {
			next.ServeHTTP(w, r)
			return
		}
		mu.Lock()
		now := time.Now()
		tokens += now.Sub(last).Seconds() * rate
		last = now
		if tokens > max {
			tokens = max
		}
		if tokens < 1 {
			mu.Unlock()
			http.Error(w, `{"error":{"code":"RATE_LIMITED","message":"Too many requests","status":429}}`, http.StatusTooManyRequests)
			return
		}
		tokens--
		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
