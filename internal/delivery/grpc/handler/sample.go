package handler

import (
	"context"
	"errors"

	"clean-template/internal/entity"
	"clean-template/internal/service"
	samplev1 "clean-template/proto/sample/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SampleHandler struct {
	sampleService *service.SampleService
	samplev1.UnimplementedSampleServiceServer
}

func NewSampleHandler(sampleService *service.SampleService) *SampleHandler {
	return &SampleHandler{sampleService: sampleService}
}

func (h *SampleHandler) ListSample(ctx context.Context, req *samplev1.ListSampleRequest) (*samplev1.ListSampleResponse, error) {
	samples, err := h.sampleService.ListSample(ctx, int(req.GetPage()), int(req.GetLimit()))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	resp := make([]*samplev1.Sample, 0, len(samples))
	for _, s := range samples {
		resp = append(resp, toProtoSample(s))
	}

	return &samplev1.ListSampleResponse{Samples: resp}, nil
}

func (h *SampleHandler) GetSample(ctx context.Context, req *samplev1.GetSampleRequest) (*samplev1.GetSampleResponse, error) {
	sample, err := h.sampleService.GetSample(ctx, int(req.GetId()))
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &samplev1.GetSampleResponse{Sample: toProtoSample(sample)}, nil
}

func (h *SampleHandler) CreateSample(ctx context.Context, req *samplev1.CreateSampleRequest) (*samplev1.CreateSampleResponse, error) {
	sample := entity.Sample{Name: req.GetName()}
	created, err := h.sampleService.CreateSample(ctx, sample)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &samplev1.CreateSampleResponse{Sample: toProtoSample(created)}, nil
}

func (h *SampleHandler) UpdateSample(ctx context.Context, req *samplev1.UpdateSampleRequest) (*samplev1.UpdateSampleResponse, error) {
	sample := entity.Sample{Name: req.GetName()}
	if err := h.sampleService.UpdateSample(ctx, sample, int(req.GetId())); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	sample.ID = int(req.GetId())
	return &samplev1.UpdateSampleResponse{Sample: toProtoSample(sample)}, nil
}

func (h *SampleHandler) DeleteSample(ctx context.Context, req *samplev1.DeleteSampleRequest) (*samplev1.DeleteSampleResponse, error) {
	if err := h.sampleService.DeleteSample(ctx, int(req.GetId())); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &samplev1.DeleteSampleResponse{Success: true}, nil
}

func toProtoSample(sample entity.Sample) *samplev1.Sample {
	return &samplev1.Sample{
		Id:   int32(sample.ID),
		Name: sample.Name,
	}
}
