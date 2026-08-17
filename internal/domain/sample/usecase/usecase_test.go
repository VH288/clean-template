package usecase_test

import (
	"context"
	"encoding/json"
	"testing"

	"clean-template/internal/domain/sample"
	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/event"
	"clean-template/internal/domain/sample/usecase"
	"clean-template/internal/infrastructure/telemetry"
	apperrors "clean-template/internal/pkg/errors"
	"clean-template/internal/pkg/testutil"

	"github.com/stretchr/testify/require"
)

type memRepo struct {
	items      map[string]*entity.Sample
	eventTypes []string
}

func newMemRepo() *memRepo {
	return &memRepo{items: map[string]*entity.Sample{}}
}

func (m *memRepo) Create(ctx context.Context, sample *entity.Sample) error {
	return m.CreateWithEvent(ctx, sample, "", "", nil)
}

func (m *memRepo) CreateWithEvent(ctx context.Context, sample *entity.Sample, eventType, eventID string, eventPayload []byte) error {
	if sample.ID == "" {
		sample.ID = "id-1"
	}
	if sample.Version == 0 {
		sample.Version = 1
	}
	cp := *sample
	m.items[sample.ID] = &cp
	if eventType != "" {
		m.eventTypes = append(m.eventTypes, eventType)
	}
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
	return m.UpdateWithEvent(ctx, sample, "", "", nil)
}

func (m *memRepo) UpdateWithEvent(ctx context.Context, sample *entity.Sample, eventType, eventID string, eventPayload []byte) error {
	if _, ok := m.items[sample.ID]; !ok {
		return apperrors.ErrNotFound
	}
	sample.Version++
	cp := *sample
	m.items[sample.ID] = &cp
	if eventType != "" {
		m.eventTypes = append(m.eventTypes, eventType)
	}
	return nil
}

func (m *memRepo) Delete(ctx context.Context, id string) error {
	return m.DeleteWithEvent(ctx, id, "", "", nil)
}

func (m *memRepo) DeleteWithEvent(ctx context.Context, id, eventType, eventID string, eventPayload []byte) error {
	if _, ok := m.items[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(m.items, id)
	if eventType != "" {
		m.eventTypes = append(m.eventTypes, eventType)
	}
	return nil
}

type trackingCache struct {
	deleted []string
}

func (c *trackingCache) Set(ctx context.Context, sample *entity.Sample) error { return nil }
func (c *trackingCache) Get(ctx context.Context, id string) (*entity.Sample, error) {
	return nil, apperrors.ErrNotFound
}
func (c *trackingCache) Delete(ctx context.Context, id string) error {
	c.deleted = append(c.deleted, id)
	return nil
}

type trackingDoc struct {
	upserted []string
	deleted  []string
}

func (d *trackingDoc) Upsert(ctx context.Context, sample *entity.Sample) error {
	d.upserted = append(d.upserted, sample.ID)
	return nil
}
func (d *trackingDoc) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	return nil, apperrors.ErrNotFound
}
func (d *trackingDoc) Delete(ctx context.Context, id string) error {
	d.deleted = append(d.deleted, id)
	return nil
}

func newUsecase(repo sample.Repository, cache sample.CacheRepository, doc sample.DocumentRepository) *usecase.SampleUsecase {
	return usecase.New(repo, cache, doc, telemetry.NoopLogger{}, telemetry.NoopTracer{}, telemetry.NoopSampleMetrics{})
}

func TestSampleUsecase_CRUD(t *testing.T) {
	ctx := testutil.NewContext()
	repo := newMemRepo()
	cache := &trackingCache{}
	doc := &trackingDoc{}
	uc := newUsecase(repo, cache, doc)

	created, err := uc.Create(ctx, "alpha", "desc", entity.StatusActive)
	require.NoError(t, err)
	require.Equal(t, "alpha", created.Name)
	require.Contains(t, repo.eventTypes, event.TypeCreated)
	require.Contains(t, cache.deleted, created.ID)
	require.Contains(t, doc.upserted, created.ID)

	got, err := uc.GetByID(ctx, created.ID)
	require.NoError(t, err)
	require.Equal(t, created.ID, got.ID)

	updated, err := uc.Update(ctx, created.ID, "beta", "updated", entity.StatusInactive)
	require.NoError(t, err)
	require.Equal(t, "beta", updated.Name)
	require.Contains(t, repo.eventTypes, event.TypeUpdated)
	require.Contains(t, cache.deleted, created.ID)
	require.Contains(t, doc.upserted, created.ID)

	items, total, err := uc.List(ctx, 1, 10)
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)

	require.NoError(t, uc.Delete(ctx, created.ID))
	require.Contains(t, repo.eventTypes, event.TypeDeleted)
	require.Contains(t, cache.deleted, created.ID)
	require.Contains(t, doc.deleted, created.ID)
	_, err = uc.GetByID(ctx, created.ID)
	require.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestSampleUsecase_MongoFailureDoesNotFailRequest(t *testing.T) {
	ctx := testutil.NewContext()
	failingDoc := failingDocRepo{}
	uc := newUsecase(newMemRepo(), &trackingCache{}, failingDoc)

	created, err := uc.Create(ctx, "alpha", "desc", entity.StatusActive)
	require.NoError(t, err)
	require.Equal(t, "alpha", created.Name)
}

func TestSampleUsecase_EventPayloadIsValidEnvelope(t *testing.T) {
	ctx := testutil.NewContext()
	repo := &capturingRepo{inner: newMemRepo()}
	uc := newUsecase(repo, &trackingCache{}, &trackingDoc{})

	_, err := uc.Create(ctx, "alpha", "desc", entity.StatusActive)
	require.NoError(t, err)
	require.NotEmpty(t, repo.lastPayload)

	var env event.Envelope
	require.NoError(t, json.Unmarshal(repo.lastPayload, &env))
	require.Equal(t, event.TypeCreated, env.Type)
	require.NotEmpty(t, env.EventID)
	require.Equal(t, "alpha", env.Sample.Name)
}

type capturingRepo struct {
	inner       *memRepo
	lastPayload []byte
}

func (c *capturingRepo) Create(ctx context.Context, sample *entity.Sample) error {
	return c.CreateWithEvent(ctx, sample, "", "", nil)
}
func (c *capturingRepo) CreateWithEvent(ctx context.Context, sample *entity.Sample, eventType, eventID string, eventPayload []byte) error {
	c.lastPayload = append([]byte(nil), eventPayload...)
	return c.inner.CreateWithEvent(ctx, sample, eventType, eventID, eventPayload)
}
func (c *capturingRepo) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	return c.inner.GetByID(ctx, id)
}
func (c *capturingRepo) List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error) {
	return c.inner.List(ctx, limit, offset)
}
func (c *capturingRepo) Update(ctx context.Context, sample *entity.Sample) error {
	return c.UpdateWithEvent(ctx, sample, "", "", nil)
}
func (c *capturingRepo) UpdateWithEvent(ctx context.Context, sample *entity.Sample, eventType, eventID string, eventPayload []byte) error {
	return c.inner.UpdateWithEvent(ctx, sample, eventType, eventID, eventPayload)
}
func (c *capturingRepo) Delete(ctx context.Context, id string) error {
	return c.DeleteWithEvent(ctx, id, "", "", nil)
}
func (c *capturingRepo) DeleteWithEvent(ctx context.Context, id, eventType, eventID string, eventPayload []byte) error {
	return c.inner.DeleteWithEvent(ctx, id, eventType, eventID, eventPayload)
}

type failingDocRepo struct{}

func (failingDocRepo) Upsert(ctx context.Context, sample *entity.Sample) error {
	return apperrors.New(apperrors.CodeInternal, "mongo down")
}
func (failingDocRepo) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	return nil, apperrors.ErrNotFound
}
func (failingDocRepo) Delete(ctx context.Context, id string) error {
	return apperrors.New(apperrors.CodeInternal, "mongo down")
}
