package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	"clean-template/internal/config"
	"clean-template/internal/pkg/constant"
)

func New(cfg config.ObservabilityConfig) *slog.Logger {
	level := parseLevel(cfg.LogLevel)
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	// JSON stdout is Loki-friendly when scraped by Alloy/Promtail.
	logger := slog.New(handler).With(
		"service", cfg.ServiceName,
	)
	return logger
}

func NewWithWriter(cfg config.ObservabilityConfig, w io.Writer) *slog.Logger {
	handler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: parseLevel(cfg.LogLevel),
	})
	return slog.New(handler).With("service", cfg.ServiceName)
}

func FromContext(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(constant.ContextKeyLogger).(*slog.Logger); ok && v != nil {
		return v
	}
	return slog.Default()
}

func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, constant.ContextKeyLogger, log)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
