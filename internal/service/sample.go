package service

import (
	"context"

	"clean-template/internal/entity"
)

type SampleRepository interface {
	ListSample(ctx context.Context, page, limit int) ([]entity.Sample, error)
	GetSampleByID(ctx context.Context, id int) (entity.Sample, error)
	CreateSample(ctx context.Context, sample entity.Sample) (entity.Sample, error)
	UpdateSample(ctx context.Context, sample entity.Sample, id int) error
	DeleteSample(ctx context.Context, id int) error
}

type SampleService struct {
	repo          SampleRepository
	externalHTTP  ExternalHTTPClient
	externalGRPC  ExternalGRPCClient
}

func NewSampleService(repo SampleRepository, externalHTTP ExternalHTTPClient, externalGRPC ExternalGRPCClient) *SampleService {
	return &SampleService{
		repo:         repo,
		externalHTTP: externalHTTP,
		externalGRPC: externalGRPC,
	}
}

func (s *SampleService) ListSample(ctx context.Context, page, limit int) ([]entity.Sample, error) {
	return s.repo.ListSample(ctx, page, limit)
}

func (s *SampleService) GetSample(ctx context.Context, id int) (entity.Sample, error) {
	sample, err := s.repo.GetSampleByID(ctx, id)
	if err != nil {
		return entity.Sample{}, err
	}
	if sample.ID == 0 {
		return entity.Sample{}, ErrNotFound
	}

	// example: enrich from external gRPC service
	if ref, err := s.externalGRPC.GetReference(ctx, int32(id)); err == nil && ref != "" {
		sample.Name = sample.Name + " (" + ref + ")"
	}

	return sample, nil
}

func (s *SampleService) CreateSample(ctx context.Context, sample entity.Sample) (entity.Sample, error) {
	if err := sample.Validate(); err != nil {
		return entity.Sample{}, err
	}

	if _, err := s.externalHTTP.FetchExample(ctx); err != nil {
		return entity.Sample{}, err
	}

	return s.repo.CreateSample(ctx, sample)
}

func (s *SampleService) UpdateSample(ctx context.Context, sample entity.Sample, id int) error {
	if err := sample.Validate(); err != nil {
		return err
	}

	existing, err := s.GetSample(ctx, id)
	if err != nil {
		return err
	}
	if existing.ID == 0 {
		return ErrNotFound
	}

	return s.repo.UpdateSample(ctx, sample, id)
}

func (s *SampleService) DeleteSample(ctx context.Context, id int) error {
	existing, err := s.GetSample(ctx, id)
	if err != nil {
		return err
	}
	if existing.ID == 0 {
		return ErrNotFound
	}

	return s.repo.DeleteSample(ctx, id)
}
