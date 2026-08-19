package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"clean-template/internal/domain/ports"
	"clean-template/internal/domain/sample"
	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/event"
	apperrors "clean-template/internal/pkg/errors"
	"clean-template/internal/pkg/utils"

	"github.com/google/uuid"
)

type SampleUsecase struct {
	repo     sample.Repository
	cache    sample.CacheRepository
	document sample.DocumentRepository
	log      ports.Logger
	trace    ports.Tracer
	metrics  ports.SampleMetrics
}

func New(
	repo sample.Repository,
	cache sample.CacheRepository,
	document sample.DocumentRepository,
	log ports.Logger,
	trace ports.Tracer,
	metrics ports.SampleMetrics,
) *SampleUsecase {
	return &SampleUsecase{
		repo:     repo,
		cache:    cache,
		document: document,
		log:      log,
		trace:    trace,
		metrics:  metrics,
	}
}

func (u *SampleUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	ctx, end := u.trace.Start(ctx, "sample.usecase.Create")
	defer end()

	if status == "" {
		status = entity.StatusActive
	}

	s := &entity.Sample{
		Name:        name,
		Description: description,
		Status:      status,
	}

	eventID := uuid.NewString()
	payload, err := marshalEvent(eventID, event.TypeCreated, s)
	if err != nil {
		u.metrics.IncOperation("create", "error")
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to build event")
	}

	if err := u.repo.CreateWithEvent(ctx, s, event.TypeCreated, eventID, payload); err != nil {
		u.metrics.IncOperation("create", "error")
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to create sample")
	}

	if err := u.document.Upsert(ctx, s); err != nil {
		u.log.Warn(ctx, "mongo upsert failed after create", "id", s.ID, "error", err.Error())
		u.metrics.IncMongoSyncFailure("create")
	}
	u.invalidateCache(ctx, s.ID)

	u.metrics.IncOperation("create", "success")
	u.log.Info(ctx, "sample created", "id", s.ID)
	return s, nil
}

func (u *SampleUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	ctx, end := u.trace.Start(ctx, "sample.usecase.GetByID")
	defer end()

	if cached, err := u.cache.Get(ctx, id); err == nil {
		u.metrics.IncOperation("get", "cache_hit")
		return cached, nil
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		u.log.Warn(ctx, "cache get failed", "id", id, "error", err.Error())
		u.metrics.IncCacheOperation("get", "error")
	}

	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.metrics.IncOperation("get", "error")
		return nil, err
	}

	if err := u.cache.Set(ctx, s); err != nil {
		u.log.Warn(ctx, "cache set failed", "id", id, "error", err.Error())
		u.metrics.IncCacheOperation("set", "error")
	}
	u.metrics.IncOperation("get", "success")
	return s, nil
}

func (u *SampleUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	ctx, end := u.trace.Start(ctx, "sample.usecase.List")
	defer end()

	page, perPage, offset := utils.NormalizePagination(page, perPage)
	items, total, err := u.repo.List(ctx, perPage, offset)
	if err != nil {
		u.log.Error(ctx, err.Error(), err)
		u.metrics.IncOperation("list", "error")
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list samples")
	}

	u.metrics.IncOperation("list", "success")
	return items, total, nil
}

func (u *SampleUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	ctx, end := u.trace.Start(ctx, "sample.usecase.Update")
	defer end()

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.metrics.IncOperation("update", "error")
		return nil, err
	}

	existing.Name = name
	existing.Description = description
	existing.Status = status

	eventID := uuid.NewString()
	payload, err := marshalEvent(eventID, event.TypeUpdated, existing)
	if err != nil {
		u.metrics.IncOperation("update", "error")
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to build event")
	}

	if err := u.repo.UpdateWithEvent(ctx, existing, event.TypeUpdated, eventID, payload); err != nil {
		u.metrics.IncOperation("update", "error")
		if errors.Is(err, apperrors.ErrConflict) {
			return nil, err
		}
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to update sample")
	}

	if err := u.document.Upsert(ctx, existing); err != nil {
		u.log.Warn(ctx, "mongo upsert failed after update", "id", existing.ID, "error", err.Error())
		u.metrics.IncMongoSyncFailure("update")
	}
	u.invalidateCache(ctx, existing.ID)

	u.metrics.IncOperation("update", "success")
	u.log.Info(ctx, "sample updated", "id", existing.ID)
	return existing, nil
}

func (u *SampleUsecase) Delete(ctx context.Context, id string) error {
	ctx, end := u.trace.Start(ctx, "sample.usecase.Delete")
	defer end()

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		u.metrics.IncOperation("delete", "error")
		return err
	}

	eventID := uuid.NewString()
	payload, err := marshalEvent(eventID, event.TypeDeleted, existing)
	if err != nil {
		u.metrics.IncOperation("delete", "error")
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to build event")
	}

	if err := u.repo.DeleteWithEvent(ctx, id, event.TypeDeleted, eventID, payload); err != nil {
		u.metrics.IncOperation("delete", "error")
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete sample")
	}

	if err := u.document.Delete(ctx, id); err != nil {
		u.log.Warn(ctx, "mongo delete failed after delete", "id", id, "error", err.Error())
		u.metrics.IncMongoSyncFailure("delete")
	}
	u.invalidateCache(ctx, id)

	u.metrics.IncOperation("delete", "success")
	u.log.Info(ctx, "sample deleted", "id", id)
	return nil
}

func marshalEvent(eventID, eventType string, s *entity.Sample) ([]byte, error) {
	return json.Marshal(event.Envelope{EventID: eventID, Type: eventType, Sample: s})
}

func (u *SampleUsecase) invalidateCache(ctx context.Context, id string) {
	if err := u.cache.Delete(ctx, id); err != nil {
		u.metrics.IncCacheInvalidate("error")
		return
	}
	u.metrics.IncCacheInvalidate("success")
}
