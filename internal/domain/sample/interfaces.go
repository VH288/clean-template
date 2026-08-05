package sample

import (
	"context"

	"clean-template/internal/domain/sample/entity"
)

type Repository interface {
	Create(ctx context.Context, sample *entity.Sample) error
	GetByID(ctx context.Context, id string) (*entity.Sample, error)
	List(ctx context.Context, limit, offset int) ([]entity.Sample, int64, error)
	Update(ctx context.Context, sample *entity.Sample) error
	Delete(ctx context.Context, id string) error
}

type CacheRepository interface {
	Set(ctx context.Context, sample *entity.Sample) error
	Get(ctx context.Context, id string) (*entity.Sample, error)
	Delete(ctx context.Context, id string) error
}

type DocumentRepository interface {
	Upsert(ctx context.Context, sample *entity.Sample) error
	GetByID(ctx context.Context, id string) (*entity.Sample, error)
	Delete(ctx context.Context, id string) error
}

type EventPublisher interface {
	PublishSampleEvent(ctx context.Context, eventType string, sample *entity.Sample) error
}

type Usecase interface {
	Create(ctx context.Context, name, description, status string) (*entity.Sample, error)
	GetByID(ctx context.Context, id string) (*entity.Sample, error)
	List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error)
	Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error)
	Delete(ctx context.Context, id string) error
}
