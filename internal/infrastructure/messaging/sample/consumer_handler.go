package sample

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"clean-template/internal/domain/sample/event"
	"clean-template/internal/infrastructure/metrics"
)

// EventHandler processes consumed Kafka sample events.
// Replace the log/metric body with your business logic (email, analytics, etc.).
type EventHandler struct {
	logger *slog.Logger
}

func NewEventHandler(logger *slog.Logger) *EventHandler {
	return &EventHandler{logger: logger}
}

func (h *EventHandler) Handle(ctx context.Context, key, value []byte) error {
	var env event.Envelope
	if err := json.Unmarshal(value, &env); err != nil {
		return fmt.Errorf("unmarshal event envelope: %w", err)
	}

	if env.Type == "" {
		return fmt.Errorf("event type is empty")
	}
	if env.Sample == nil {
		return fmt.Errorf("event sample is nil")
	}

	switch env.Type {
	case event.TypeCreated:
		h.logger.Info("consumer: received sample.created",
			slog.String("sample_id", env.Sample.ID),
			slog.String("name", env.Sample.Name),
		)
		metrics.KafkaEventsConsumed.WithLabelValues(env.Type, "ok").Inc()
	case event.TypeUpdated:
		h.logger.Info("consumer: received sample.updated",
			slog.String("sample_id", env.Sample.ID),
			slog.String("status", env.Sample.Status),
		)
		metrics.KafkaEventsConsumed.WithLabelValues(env.Type, "ok").Inc()
	case event.TypeDeleted:
		h.logger.Info("consumer: received sample.deleted",
			slog.String("sample_id", env.Sample.ID),
		)
		metrics.KafkaEventsConsumed.WithLabelValues(env.Type, "ok").Inc()
	default:
		h.logger.Warn("consumer: unknown event type, skipping",
			slog.String("type", env.Type),
			slog.String("key", string(key)),
		)
	}

	return nil
}
