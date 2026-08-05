package usecase

import (
	"context"
	"log/slog"

	"clean-template/internal/domain/sample"
	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/infrastructure/apm"
	"clean-template/internal/infrastructure/logger"
	"clean-template/internal/infrastructure/metrics"
	apperrors "clean-template/internal/pkg/errors"
	"clean-template/internal/pkg/utils"
)

type SampleUsecase struct {
	repo      sample.Repository
	cache     sample.CacheRepository
	document  sample.DocumentRepository
	publisher sample.EventPublisher
}

func New(
	repo sample.Repository,
	cache sample.CacheRepository,
	document sample.DocumentRepository,
	publisher sample.EventPublisher,
) *SampleUsecase {
	return &SampleUsecase{
		repo:      repo,
		cache:     cache,
		document:  document,
		publisher: publisher,
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

	if err := u.repo.Create(ctx, s); err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("create", "error").Inc()
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to create sample")
	}

	_ = u.cache.Set(ctx, s)
	_ = u.document.Upsert(ctx, s)
	_ = u.publisher.PublishSampleEvent(ctx, "sample.created", s)

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

	if err := u.repo.Update(ctx, existing); err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("update", "error").Inc()
		return nil, apperrors.Wrap(err, apperrors.CodeInternal, "failed to update sample")
	}

	_ = u.cache.Set(ctx, existing)
	_ = u.document.Upsert(ctx, existing)
	_ = u.publisher.PublishSampleEvent(ctx, "sample.updated", existing)

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

	if err := u.repo.Delete(ctx, id); err != nil {
		metrics.SampleOperationsTotal.WithLabelValues("delete", "error").Inc()
		return apperrors.Wrap(err, apperrors.CodeInternal, "failed to delete sample")
	}

	_ = u.cache.Delete(ctx, id)
	_ = u.document.Delete(ctx, id)
	_ = u.publisher.PublishSampleEvent(ctx, "sample.deleted", existing)

	metrics.SampleOperationsTotal.WithLabelValues("delete", "success").Inc()
	log.Info("sample deleted", slog.String("id", id))
	return nil
}
