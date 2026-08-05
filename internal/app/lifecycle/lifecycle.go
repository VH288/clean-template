package lifecycle

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type ShutdownFunc func(ctx context.Context) error

type Manager struct {
	logger   *slog.Logger
	hooks    []ShutdownFunc
	timeout  time.Duration
}

func New(logger *slog.Logger, timeout time.Duration) *Manager {
	return &Manager{logger: logger, timeout: timeout}
}

func (m *Manager) Add(hook ShutdownFunc) {
	m.hooks = append(m.hooks, hook)
}

func (m *Manager) Wait(ctx context.Context) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		m.logger.Info("shutdown signal received", "signal", sig.String())
	case <-ctx.Done():
		m.logger.Info("root context cancelled")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	for i := len(m.hooks) - 1; i >= 0; i-- {
		if err := m.hooks[i](shutdownCtx); err != nil {
			m.logger.Error("shutdown hook failed", "error", err)
		}
	}
	m.logger.Info("graceful shutdown completed")
}
