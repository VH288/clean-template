package usecase

import (
	"context"
	"encoding/json"
	"log/slog"

	"clean-template/internal/domain/sample"
	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/event"
	"clean-template/internal/infrastructure/apm"
	"clean-template/internal/infrastructure/logger"
	"clean-template/internal/infrastructure/metrics"
	apperrors "clean-template/internal/pkg/errors"
	"clean-template/internal/pkg/utils"
)

type SampleUsecase struct {
	repo     sample.Repository
	cache    sample.CacheRepository
	document sample.DocumentRepository
}

func New(
	repo sample.Repository,
	cache sample.CacheRepository,
	document sample.DocumentRepository,
) *SampleUsecase {
	return &SampleUsecase{
		repo:     repo,
		cache:    cache,
		document: document,
	}
}

func (u *SampleUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	ctx, span := apm.Start(ctx, "sample.usecase.Create")
	defer span.End()

	log := logger.FromContext(ctx)
	if status == "" {
		status = entity.StatusActive
	}

	s := &entity.Sample{
		Name:        name,
		Description: description,
		Status:      status,
	}

	payload, err := marshalEvent(event.TypeCreated, s)
	if err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("create", "error").Inc()
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to build event")
	}

	if err := u.repo.CreateWithEvent(ctx, s, event.TypeCreated, payload); err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("create", "error").Inc()
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to create sample")
	}

	// demo: direct multi-store write; not transactional with Postgres
	if err := u.document.Upsert(ctx, s); err != nil {
		log.Warn("mongo upsert failed after create", slog.String("id", s.ID), slog.String("error", err.Error()))
	}
	u.invalidateCache(ctx, s.ID)

	metrics.SampleOperationsTotal.WithLabelValues("create", "success").Inc()
	log.Info("sample created", slog.String("id", s.ID))
	return s, nil
}

func (u *SampleUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	ctx, span := apm.Start(ctx, "sample.usecase.GetByID")
	defer span.End()

	if cached, err := u.cache.Get(ctx, id); err == nil {
		metrics.SampleOperationsTotal.WithLabelValues("get", "cache_hit").Inc()
		return cached, nil
	}

	s, err := u.repo.GetByID(ctx, id)
	if err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("get", "error").Inc()
		return nil, err
	}

	_ = u.cache.Set(ctx, s)
	metrics.SampleOperationsTotal.WithLabelValues("get", "success").Inc()
	return s, nil
}

func (u *SampleUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	ctx, span := apm.Start(ctx, "sample.usecase.List")
	defer span.End()

	page, perPage, offset := utils.NormalizePagination(page, perPage)
	items, total, err := u.repo.List(ctx, perPage, offset)
	if err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("list", "error").Inc()
		return nil, 0, apperrors.Wrap(err, apperrors.CodeInternal, "failed to list samples")
	}

	metrics.SampleOperationsTotal.WithLabelValues("list", "success").Inc()
	return items, total, nil
}

func (u *SampleUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	ctx, span := apm.Start(ctx, "sample.usecase.Update")
	defer span.End()

	log := logger.FromContext(ctx)

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("update", "error").Inc()
		return nil, err
	}

	existing.Name = name
	existing.Description = description
	existing.Status = status

	payload, err := marshalEvent(event.TypeUpdated, existing)
	if err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("update", "error").Inc()
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to build event")
	}

	if err := u.repo.UpdateWithEvent(ctx, existing, event.TypeUpdated, payload); err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("update", "error").Inc()
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to update sample")
	}

	// demo: direct multi-store write; not transactional with Postgres
	if err := u.document.Upsert(ctx, existing); err != nil {
		log.Warn("mongo upsert failed after update", slog.String("id", existing.ID), slog.String("error", err.Error()))
	}
	u.invalidateCache(ctx, existing.ID)

	metrics.SampleOperationsTotal.WithLabelValues("update", "success").Inc()
	log.Info("sample updated", slog.String("id", existing.ID))
	return existing, nil
}

func (u *SampleUsecase) Delete(ctx context.Context, id string) error {
	ctx, span := apm.Start(ctx, "sample.usecase.Delete")
	defer span.End()

	log := logger.FromContext(ctx)

	existing, err := u.repo.GetByID(ctx, id)
	if err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("delete", "error").Inc()
		return err
	}

	payload, err := marshalEvent(event.TypeDeleted, existing)
	if err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("delete", "error").Inc()
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to build event")
	}

	if err := u.repo.DeleteWithEvent(ctx, id, event.TypeDeleted, payload); err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("delete", "error").Inc()
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete sample")
	}

	// demo: direct multi-store write; not transactional with Postgres
	if err := u.document.Delete(ctx, id); err != nil {
		log.Warn("mongo delete failed after delete", slog.String("id", id), slog.String("error", err.Error()))
	}
	u.invalidateCache(ctx, id)

	metrics.SampleOperationsTotal.WithLabelValues("delete", "success").Inc()
	log.Info("sample deleted", slog.String("id", id))
	return nil
}

func marshalEvent(eventType string, s *entity.Sample) ([]byte, error) {
	return json.Marshal(event.Envelope{Type: eventType, Sample: s})
}

func (u *SampleUsecase) invalidateCache(ctx context.Context, id string) {
	if err := u.cache.Delete(ctx, id); err != nil {
		metrics.CacheInvalidateTotal.WithLabelValues("error").Inc()
		return
	}
	metrics.CacheInvalidateTotal.WithLabelValues("success").Inc()
}
