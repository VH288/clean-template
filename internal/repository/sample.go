package repository

import (
	"context"

	"clean-template/internal/models"

	"gorm.io/gorm"
)

type SampleRepo struct {
	DB *gorm.DB
}

func (r *SampleRepo) ListSample(ctx context.Context, param models.SampleParam) ([]models.Sample, error) {
	var resp []models.Sample

	offset := (param.Page - 1) * param.Limit

	err := r.DB.Limit(param.Limit).Offset(offset).Order("id ASC").Find(&resp).Error

	return resp, err
}

func (r *SampleRepo) GetSampleByID(ctx context.Context, id int) (models.Sample, error) {
	var resp models.Sample

	err := r.DB.Where("id = ?", id).First(&resp).Error

	if err == gorm.ErrRecordNotFound {
		return resp, nil
	}

	return resp, err
}

func (r *SampleRepo) CreateSample(ctx context.Context, sample models.Sample) error {
	return r.DB.Create(&sample).Error
}

func (r *SampleRepo) UpdateSample(ctx context.Context, sample models.Sample, id int) error {
	return r.DB.Model(&models.Sample{}).Where("id = ?", id).Update("name", &sample.Name).Error
}

func (r *SampleRepo) DeleteSample(ctx context.Context, id int) error {
	return r.DB.Where("id = ?", id).Delete(&models.Sample{}).Error
}
