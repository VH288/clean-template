package outbox

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"clean-template/internal/config"
	"clean-template/internal/infrastructure/metrics"

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

	cancel context.CancelFunc
	wg     sync.WaitGroup
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
	childCtx, cancel := context.WithCancel(ctx)
	r.cancel = cancel

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		if err := r.resetStaleProcessing(childCtx); err != nil {
			r.logger.Error("outbox relay: reset stale processing failed", "error", err)
		}

		ticker := time.NewTicker(r.cfg.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-childCtx.Done():
				return
			case <-ticker.C:
				if err := r.processBatch(childCtx); err != nil {
					r.logger.Error("outbox relay batch failed", "error", err)
				}
				r.refreshGauges(childCtx)
			}
		}
	}()
}

func (r *Relay) Stop() {
	if r.cancel != nil {
		r.cancel()
	}
	r.wg.Wait()
}

func (r *Relay) resetStaleProcessing(ctx context.Context) error {
	const q = `
		UPDATE outbox_events
		SET status = 'pending', claimed_at = NULL
		WHERE status = 'processing'
		  AND claimed_at IS NOT NULL
		  AND claimed_at < NOW() - $1::interval
	`
	_, err := r.db.ExecContext(ctx, q, r.cfg.ProcessingStaleAfter.String())
	return err
}

func (r *Relay) processBatch(ctx context.Context) error {
	rows, err := r.claimBatch(ctx)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}

	for _, row := range rows {
		if err := r.publishRow(ctx, row); err != nil {
			r.logger.Error("outbox relay: publish row failed", "event_id", row.ID, "error", err)
		}
	}
	return nil
}

func (r *Relay) claimBatch(ctx context.Context) ([]outboxRow, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
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
		return nil, fmt.Errorf("select pending outbox: %w", err)
	}
	if len(rows) == 0 {
		return nil, tx.Commit()
	}

	for _, row := range rows {
		const claimQ = `
			UPDATE outbox_events
			SET status = 'processing', claimed_at = NOW()
			WHERE id = $1
		`
		if _, err := tx.ExecContext(ctx, claimQ, row.ID); err != nil {
			return nil, fmt.Errorf("claim outbox row: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claim batch: %w", err)
	}
	return rows, nil
}

func (r *Relay) publishRow(ctx context.Context, row outboxRow) error {
	if err := r.publisher.Publish(ctx, row.AggregateID, row.Payload); err != nil {
		return r.markPublishFailed(ctx, row, err)
	}

	const okQ = `
		UPDATE outbox_events
		SET status = 'published', published_at = NOW(), claimed_at = NULL
		WHERE id = $1
	`
	if _, err := r.db.ExecContext(ctx, okQ, row.ID); err != nil {
		return fmt.Errorf("mark published: %w", err)
	}

	r.logger.Info("outbox relay: published event",
		"event_type", row.EventType,
		"id", row.ID,
		"aggregate_id", row.AggregateID,
	)
	return nil
}

func (r *Relay) markPublishFailed(ctx context.Context, row outboxRow, publishErr error) error {
	retryCount := row.RetryCount + 1
	status := "pending"
	if retryCount >= r.cfg.MaxRetries {
		status = "failed"
	}

	const failQ = `
		UPDATE outbox_events
		SET retry_count = $1, status = $2, claimed_at = NULL
		WHERE id = $3
	`
	if _, err := r.db.ExecContext(ctx, failQ, retryCount, status, row.ID); err != nil {
		return fmt.Errorf("update failed outbox: %w", err)
	}

	r.logger.Error("outbox relay: publish failed",
		"event_id", row.ID,
		"event_type", row.EventType,
		"retry_count", retryCount,
		"status", status,
		"error", publishErr,
	)
	return publishErr
}

func (r *Relay) refreshGauges(ctx context.Context) {
	var pending, failed int
	_ = r.db.GetContext(ctx, &pending, `SELECT COUNT(*) FROM outbox_events WHERE status = 'pending'`)
	_ = r.db.GetContext(ctx, &failed, `SELECT COUNT(*) FROM outbox_events WHERE status = 'failed'`)
	metrics.OutboxEventsPending.Set(float64(pending))
	metrics.OutboxEventsFailed.Set(float64(failed))
}

// ReplayFailed resets failed rows to pending for manual recovery.
func (r *Relay) ReplayFailed(ctx context.Context) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE outbox_events
		SET status = 'pending', retry_count = 0, claimed_at = NULL
		WHERE status = 'failed'
	`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

// ProcessBatchForTest exposes batch processing for integration tests.
func (r *Relay) ProcessBatchForTest(ctx context.Context) error {
	return r.processBatch(ctx)
}
