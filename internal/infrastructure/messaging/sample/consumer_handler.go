package sample

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"clean-template/internal/domain/sample/event"
	"clean-template/internal/infrastructure/metrics"
)

var ErrPoisonMessage = errors.New("poison message")

// EventHandler processes consumed Kafka sample events.
type EventHandler struct {
	logger    *slog.Logger
	store     IdempotencyStore
	dlq       DLQPublisher
}

type DLQPublisher interface {
	PublishDLQ(ctx context.Context, key string, value []byte, reason string) error
}

func NewEventHandler(logger *slog.Logger, store IdempotencyStore, dlq DLQPublisher) *EventHandler {
	return &EventHandler{logger: logger, store: store, dlq: dlq}
}

func (h *EventHandler) Handle(ctx context.Context, key, value []byte) error {
	var env event.Envelope
	if err := json.Unmarshal(value, &env); err != nil {
		if h.dlq != nil {
			_ = h.dlq.PublishDLQ(ctx, string(key), value, "invalid json")
		}
		return fmt.Errorf("%w: %v", ErrPoisonMessage, err)
	}

	if env.EventID == "" {
		if h.dlq != nil {
			_ = h.dlq.PublishDLQ(ctx, string(key), value, "missing event_id")
		}
		return fmt.Errorf("%w: event_id is empty", ErrPoisonMessage)
	}

	if h.store != nil {
		processed, err := h.store.IsProcessed(ctx, env.EventID)
		if err != nil {
			return fmt.Errorf("idempotency check: %w", err)
		}
		if processed {
			metrics.KafkaEventsConsumed.WithLabelValues(env.Type, "duplicate").Inc()
			return nil
		}
	}

	if env.Type == "" {
		return fmt.Errorf("%w: event type is empty", ErrPoisonMessage)
	}
	if env.Sample == nil {
		return fmt.Errorf("%w: event sample is nil", ErrPoisonMessage)
	}

	switch env.Type {
	case event.TypeCreated:
		h.logger.Info("consumer: received sample.created",
			slog.String("event_id", env.EventID),
			slog.String("sample_id", env.Sample.ID),
			slog.String("name", env.Sample.Name),
		)
		metrics.KafkaEventsConsumed.WithLabelValues(env.Type, "ok").Inc()
	case event.TypeUpdated:
		h.logger.Info("consumer: received sample.updated",
			slog.String("event_id", env.EventID),
			slog.String("sample_id", env.Sample.ID),
			slog.String("status", env.Sample.Status),
		)
		metrics.KafkaEventsConsumed.WithLabelValues(env.Type, "ok").Inc()
	case event.TypeDeleted:
		h.logger.Info("consumer: received sample.deleted",
			slog.String("event_id", env.EventID),
			slog.String("sample_id", env.Sample.ID),
		)
		metrics.KafkaEventsConsumed.WithLabelValues(env.Type, "ok").Inc()
	default:
		h.logger.Warn("consumer: unknown event type, skipping",
			slog.String("type", env.Type),
			slog.String("key", string(key)),
		)
	}

	if h.store != nil {
		if err := h.store.MarkProcessed(ctx, env.EventID); err != nil {
			return fmt.Errorf("mark processed: %w", err)
		}
	}

	return nil
}
