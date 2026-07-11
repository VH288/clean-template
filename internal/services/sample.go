package services

import (
	"context"

	"clean-template/internal/interfaces"
	"clean-template/internal/models"
)

type SampleService struct {
	SampleRepo interfaces.ISampleRepo
}

func (s *SampleService) ListSample(ctx context.Context, param models.SampleParam) ([]models.Sample, error) {
	return s.SampleRepo.ListSample(ctx, param)
}

func (s *SampleService) GetSample(ctx context.Context, ID int) (models.Sample, error) {
	return s.SampleRepo.GetSampleByID(ctx, ID)
}

func (s *SampleService) CreateSample(ctx context.Context, sample models.Sample) error {
	return s.SampleRepo.CreateSample(ctx, sample)
}

func (s *SampleService) UpdateSample(ctx context.Context, sample models.Sample, id int) error {
	return s.SampleRepo.UpdateSample(ctx, sample, id)
}

func (s *SampleService) DeleteSample(ctx context.Context, id int) error {
	return s.SampleRepo.DeleteSample(ctx, id)
}
