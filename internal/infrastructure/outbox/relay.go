package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"clean-template/internal/config"

	"github.com/jmoiron/sqlx"
)

type eventPublisher interface {
	Publish(ctx context.Context, aggregateID string, payload []byte) error
}

type outboxRow struct {
	ID          string `db:"id"`
	AggregateID string `db:"aggregate_id"`
	EventType   string `db:"event_type"`
	Payload     []byte `db:"payload"`
	RetryCount  int    `db:"retry_count"`
}

// Relay polls pending outbox events and publishes them to Kafka.
type Relay struct {
	db        *sqlx.DB
	publisher eventPublisher
	logger    *slog.Logger
	cfg       config.OutboxConfig
}

func NewRelay(db *sqlx.DB, publisher eventPublisher, logger *slog.Logger, cfg config.OutboxConfig) *Relay {
	return &Relay{
		db:        db,
		publisher: publisher,
		logger:    logger,
		cfg:       cfg,
	}
}

func (r *Relay) Start(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(r.cfg.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := r.processBatch(ctx); err != nil {
					r.logger.Error("outbox relay batch failed", "error", err)
				}
			}
		}
	}()
}

func (r *Relay) processBatch(ctx context.Context) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const selectQ = `
		SELECT id, aggregate_id, event_type, payload, retry_count
		FROM outbox_events
		WHERE status = 'pending'
		ORDER BY created_at
		LIMIT $1
		FOR UPDATE SKIP LOCKED
	`
	var rows []outboxRow
	if err := tx.SelectContext(ctx, &rows, selectQ, r.cfg.BatchSize); err != nil {
		return fmt.Errorf("select pending outbox: %w", err)
	}

	for _, row := range rows {
		if err := r.publisher.Publish(ctx, row.AggregateID, row.Payload); err != nil {
			r.logger.Error("outbox relay: publish failed",
				"event_id", row.ID,
				"event_type", row.EventType,
				"error", err,
			)
			retryCount := row.RetryCount + 1
			status := "pending"
			if retryCount >= r.cfg.MaxRetries {
				status = "failed"
			}
			const failQ = `
				UPDATE outbox_events
				SET retry_count = $1, status = $2
				WHERE id = $3
			`
			if _, uerr := tx.ExecContext(ctx, failQ, retryCount, status, row.ID); uerr != nil {
				return fmt.Errorf("update failed outbox: %w", uerr)
			}
			continue
		}

		const okQ = `
			UPDATE outbox_events
			SET status = 'published', published_at = NOW()
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, okQ, row.ID); err != nil {
			return fmt.Errorf("mark published: %w", err)
		}

		r.logger.Info("outbox relay: published event",
			"event_type", row.EventType,
			"id", row.ID,
			"aggregate_id", row.AggregateID,
		)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit outbox batch: %w", err)
	}
	return nil
}

// ProcessBatchForTest exposes batch processing for integration tests.
func (r *Relay) ProcessBatchForTest(ctx context.Context) error {
	return r.processBatch(ctx)
}
