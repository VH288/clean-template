package outbox_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"clean-template/internal/config"
	"clean-template/internal/infrastructure/outbox"
	"clean-template/internal/infrastructure/telemetry"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

type fakePublisher struct {
	err       error
	published []string
}

func (f *fakePublisher) Publish(ctx context.Context, aggregateID string, payload []byte) error {
	if f.err != nil {
		return f.err
	}
	f.published = append(f.published, aggregateID)
	return nil
}

func TestRelay_ClaimAndPublish(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, aggregate_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "aggregate_id", "event_type", "payload", "retry_count"}).
			AddRow("evt-1", "agg-1", "sample.created", []byte(`{}`), 0))
	mock.ExpectExec(`UPDATE outbox_events\s+SET status = 'processing'`).
		WithArgs("evt-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	mock.ExpectExec(`UPDATE outbox_events\s+SET status = 'published'`).
		WithArgs("evt-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	pub := &fakePublisher{}
	relay := outbox.NewRelay(sqlxDB, pub, telemetry.DiscardLogger(), config.OutboxConfig{BatchSize: 10, MaxRetries: 5})

	require.NoError(t, relay.ProcessBatchForTest(context.Background()))
	require.Equal(t, []string{"agg-1"}, pub.published)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRelay_PublishFailureIncrementsRetry(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	sqlxDB := sqlx.NewDb(db, "sqlmock")

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id, aggregate_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "aggregate_id", "event_type", "payload", "retry_count"}).
			AddRow("evt-1", "agg-1", "sample.created", []byte(`{}`), 0))
	mock.ExpectExec(`UPDATE outbox_events\s+SET status = 'processing'`).
		WithArgs("evt-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	pub := &fakePublisher{err: errors.New("kafka down")}
	mock.ExpectExec(`UPDATE outbox_events\s+SET retry_count`).
		WithArgs(1, "pending", "evt-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	relay := outbox.NewRelay(sqlxDB, pub, telemetry.DiscardLogger(), config.OutboxConfig{BatchSize: 10, MaxRetries: 5, ProcessingStaleAfter: time.Minute})
	require.NoError(t, relay.ProcessBatchForTest(context.Background()))
	require.NoError(t, mock.ExpectationsWereMet())
}
