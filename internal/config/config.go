package config

import (
	"log/slog"
	"os"
)

type Config struct {
	DSN      string
	Port     string
	LogLevel slog.Level
}

func MustLoad() *Config {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		slog.Error("DATABASE_URL is required but not set")
		os.Exit(1)
	}
	return &Config{
		DSN:      dsn,
		Port:     getEnvOr("PORT", "3000"),
		LogLevel: parseLogLevel(os.Getenv("LOG_LEVEL")),
	}
}

func getEnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		if s != "" && s != "info" {
			slog.Warn("unknown LOG_LEVEL, falling back to info", slog.String("value", s))
		}
		return slog.LevelInfo
	}
}
