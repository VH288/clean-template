//go:build integration

package outbox_test

import (
	"context"
	"os"
	"testing"
	"time"

	"clean-template/internal/config"
	"clean-template/internal/infrastructure/outbox"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type fakePublisher struct {
	published []string
}

func (f *fakePublisher) Publish(ctx context.Context, aggregateID string, payload []byte) error {
	f.published = append(f.published, aggregateID)
	return nil
}

func TestRelay_MarksPublished_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	ctx := context.Background()
	eventID := "00000000-0000-4000-8000-000000000001"
	aggregateID := "00000000-0000-4000-8000-000000000002"
	_, err = db.Exec(`
		INSERT INTO outbox_events (id, aggregate_type, aggregate_id, event_type, payload, status)
		VALUES ($1, 'sample', $2, 'sample.created', '{"type":"sample.created"}', 'pending')
	`, eventID, aggregateID)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM outbox_events WHERE id = $1`, eventID)
	})

	pub := &fakePublisher{}
	relay := outbox.NewRelay(db, pub, nil, config.OutboxConfig{
		PollInterval: time.Second,
		BatchSize:    10,
		MaxRetries:   5,
	})

	require.NoError(t, relay.ProcessBatchForTest(ctx))

	var status string
	require.NoError(t, db.Get(&status, `SELECT status FROM outbox_events WHERE id = $1`, eventID))
	require.Equal(t, "published", status)
	require.Equal(t, []string{aggregateID}, pub.published)
}
