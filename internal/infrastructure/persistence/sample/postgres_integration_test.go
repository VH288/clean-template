//go:build integration

package sample_test

import (
	"os"
	"testing"

	"clean-template/internal/domain/sample/entity"
	samplepersist "clean-template/internal/infrastructure/persistence/sample"
	"clean-template/internal/pkg/testutil"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Integration test against real Postgres.
// Run: DATABASE_URL='host=127.0.0.1 ...' go test ./internal/infrastructure/persistence/sample -tags=integration -count=1
func TestPostgresRepository_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := samplepersist.NewPostgresRepository(db)
	ctx := testutil.NewContext()

	sample := &entity.Sample{
		Name:        "integration",
		Description: "from test",
		Status:      entity.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, sample))
	require.NotEmpty(t, sample.ID)

	got, err := repo.GetByID(ctx, sample.ID)
	require.NoError(t, err)
	require.Equal(t, "integration", got.Name)

	sample.Name = "updated"
	require.NoError(t, repo.Update(ctx, sample))

	items, total, err := repo.List(ctx, 10, 0)
	require.NoError(t, err)
	require.GreaterOrEqual(t, total, int64(1))
	require.NotEmpty(t, items)

	require.NoError(t, repo.Delete(ctx, sample.ID))
}

func TestPostgresRepository_CreateWithEvent_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := samplepersist.NewPostgresRepository(db)
	ctx := testutil.NewContext()

	sample := &entity.Sample{
		Name:        "outbox",
		Description: "event test",
		Status:      entity.StatusActive,
	}
	payload := []byte(`{"type":"sample.created","sample":{"id":"","name":"outbox"}}`)
	require.NoError(t, repo.CreateWithEvent(ctx, sample, "sample.created", payload))

	var count int
	require.NoError(t, db.Get(&count, `SELECT COUNT(*) FROM outbox_events WHERE aggregate_id = $1 AND status = 'pending'`, sample.ID))
	require.Equal(t, 1, count)

	require.NoError(t, repo.Delete(ctx, sample.ID))
}
