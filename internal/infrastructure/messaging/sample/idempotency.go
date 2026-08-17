package sample

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// IdempotencyStore tracks processed event IDs for at-least-once consumers.
type IdempotencyStore interface {
	IsProcessed(ctx context.Context, eventID string) (bool, error)
	MarkProcessed(ctx context.Context, eventID string) error
}

type PostgresIdempotencyStore struct {
	db *sqlx.DB
}

func NewPostgresIdempotencyStore(db *sqlx.DB) *PostgresIdempotencyStore {
	return &PostgresIdempotencyStore{db: db}
}

func (s *PostgresIdempotencyStore) IsProcessed(ctx context.Context, eventID string) (bool, error) {
	var exists bool
	err := s.db.GetContext(ctx, &exists, `SELECT EXISTS(SELECT 1 FROM processed_events WHERE event_id = $1)`, eventID)
	return exists, err
}

func (s *PostgresIdempotencyStore) MarkProcessed(ctx context.Context, eventID string) error {
	const q = `INSERT INTO processed_events (event_id) VALUES ($1) ON CONFLICT DO NOTHING`
	_, err := s.db.ExecContext(ctx, q, eventID)
	if err != nil {
		return fmt.Errorf("mark processed event: %w", err)
	}
	return nil
}
