package repository

import (
	"context"

	"clean-template/internal/entity"

	"gorm.io/gorm"
)

type SampleRepo struct {
	db *gorm.DB
}

func NewSampleRepo(db *gorm.DB) *SampleRepo {
	return &SampleRepo{db: db}
}

func (r *SampleRepo) ListSample(ctx context.Context, page, limit int) ([]entity.Sample, error) {
	var resp []entity.Sample

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	offset := (page - 1) * limit

	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("id ASC").
		Find(&resp).Error

	return resp, err
}

func (r *SampleRepo) GetSampleByID(ctx context.Context, id int) (entity.Sample, error) {
	var resp entity.Sample

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&resp).Error
	if err == gorm.ErrRecordNotFound {
		return entity.Sample{}, nil
	}

	return resp, err
}

func (r *SampleRepo) CreateSample(ctx context.Context, sample entity.Sample) (entity.Sample, error) {
	err := r.db.WithContext(ctx).Create(&sample).Error
	return sample, err
}

func (r *SampleRepo) UpdateSample(ctx context.Context, sample entity.Sample, id int) error {
	return r.db.WithContext(ctx).
		Model(&entity.Sample{}).
		Where("id = ?", id).
		Update("name", sample.Name).Error
}

func (r *SampleRepo) DeleteSample(ctx context.Context, id int) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entity.Sample{}).Error
}
