package usecase_test

import (
	"context"
	"testing"

	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/usecase"
	apperrors "clean-template/internal/pkg/errors"
	"clean-template/internal/pkg/testutil"

	"github.com/stretchr/testify/require"
)

type memRepo struct {
	items map[string]*entity.Sample
}

func newMemRepo() *memRepo {
	return &memRepo{items: map[string]*entity.Sample{}}
}

func (m *memRepo) Create(ctx context.Context, sample *entity.Sample) error {
	if sample.ID == "" {
		sample.ID = "id-1"
	}
	cp := *sample
	m.items[sample.ID] = &cp
	return nil
}

func (m *memRepo) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	s, ok := m.items[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	cp := *s
	return &cp, nil
}

func (m *memRepo) List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error) {
	out := make([]entity.Sample, 0, len(m.items))
	for _, s := range m.items {
		out = append(out, *s)
	}
	return out, int64(len(out)), nil
}

func (m *memRepo) Update(ctx context.Context, sample *entity.Sample) error {
	if _, ok := m.items[sample.ID]; !ok {
		return apperrors.ErrNotFound
	}
	cp := *sample
	m.items[sample.ID] = &cp
	return nil
}

func (m *memRepo) Delete(ctx context.Context, id string) error {
	if _, ok := m.items[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.items, id)
	return nil
}

type nopCache struct{}

func (nopCache) Set(ctx context.Context, sample *entity.Sample) error  { return nil }
func (nopCache) Get(ctx context.Context, id string) (*entity.Sample, error) {
	return nil, apperrors.ErrNotFound
}
func (nopCache) Delete(ctx context.Context, id string) error { return nil }

type nopDoc struct{}

func (nopDoc) Upsert(ctx context.Context, sample *entity.Sample) error { return nil }
func (nopDoc) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	return nil, apperrors.ErrNotFound
}
func (nopDoc) Delete(ctx context.Context, id string) error { return nil }

type nopPub struct{}

func (nopPub) PublishSampleEvent(ctx context.Context, eventType string, sample *entity.Sample) error {
	return nil
}

func TestSampleUsecase_CRUD(t *testing.T) {
	ctx := testutil.NewContext()
	uc := usecase.New(newMemRepo(), nopCache{}, nopDoc{}, nopPub{})

	created, err := uc.Create(ctx, "alpha", "desc", entity.StatusActive)
	require.NoError(t, err)
	require.Equal(t, "alpha", created.Name)

	got, err := uc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	updated, err := uc.Update(ctx, created.ID, "beta", "updated", entity.StatusInactive)
	require.NoError(t, err)
	require.Equal(t, "beta", updated.Name)

	items, total, err := uc.List(ctx, 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)

	require.NoError(t, uc.Delete(ctx, created.ID))
	_, err = uc.GetByID(ctx, created.ID)
	require.ErrorIs(t, err, apperrors.ErrNotFound)
}
