package repository_test

import (
	"os"
	"testing"

	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/repository"
	"clean-template/internal/pkg/testutil"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Integration test against real Postgres.
// Run: DATABASE_URL='host=127.0.0.1 ...' go test ./internal/domain/sample/repository -run Integration -count=1
func TestPostgresRepository_Integration(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}

	db, err := sqlx.Connect("postgres", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	repo := repository.NewPostgresRepository(db)
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
