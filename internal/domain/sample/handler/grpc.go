package handler

import (
	"context"

	"clean-template/internal/domain/ports"
	"clean-template/internal/domain/sample"
	"clean-template/internal/domain/sample/entity"
	apperrors "clean-template/internal/pkg/errors"
	samplev1 "clean-template/internal/proto/sample/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type GRPCHandler struct {
	samplev1.UnimplementedSampleServiceServer
	usecase sample.Usecase
	metrics ports.GRPCMetrics
}

func NewGRPCHandler(usecase sample.Usecase, metrics ports.GRPCMetrics) *GRPCHandler {
	return &GRPCHandler{usecase: usecase, metrics: metrics}
}

func (h *GRPCHandler) CreateSample(ctx context.Context, req *samplev1.CreateSampleRequest) (*samplev1.SampleResponse, error) {
	result, err := h.usecase.Create(ctx, req.GetName(), req.GetDescription(), req.GetStatus())
	if err != nil {
		h.metrics.IncRequest("CreateSample", "error")
		return nil, toGRPCError(err)
	}
	h.metrics.IncRequest("CreateSample", "ok")
	return toProtoSample(result), nil
}

func (h *GRPCHandler) GetSample(ctx context.Context, req *samplev1.GetSampleRequest) (*samplev1.SampleResponse, error) {
	result, err := h.usecase.GetByID(ctx, req.GetId())
	if err != nil {
		h.metrics.IncRequest("GetSample", "error")
		return nil, toGRPCError(err)
	}
	h.metrics.IncRequest("GetSample", "ok")
	return toProtoSample(result), nil
}

func (h *GRPCHandler) ListSamples(ctx context.Context, req *samplev1.ListSamplesRequest) (*samplev1.ListSamplesResponse, error) {
	items, total, err := h.usecase.List(ctx, int(req.GetPage()), int(req.GetPerPage()))
	if err != nil {
		h.metrics.IncRequest("ListSamples", "error")
		return nil, toGRPCError(err)
	}

	resp := &samplev1.ListSamplesResponse{Total: total}
	for i := range items {
		resp.Items = append(resp.Items, toProtoSample(&items[i]))
	}
	h.metrics.IncRequest("ListSamples", "ok")
	return resp, nil
}

func (h *GRPCHandler) UpdateSample(ctx context.Context, req *samplev1.UpdateSampleRequest) (*samplev1.SampleResponse, error) {
	result, err := h.usecase.Update(ctx, req.GetId(), req.GetName(), req.GetDescription(), req.GetStatus())
	if err != nil {
		h.metrics.IncRequest("UpdateSample", "error")
		return nil, toGRPCError(err)
	}
	h.metrics.IncRequest("UpdateSample", "ok")
	return toProtoSample(result), nil
}

func (h *GRPCHandler) DeleteSample(ctx context.Context, req *samplev1.DeleteSampleRequest) (*samplev1.DeleteSampleResponse, error) {
	if err := h.usecase.Delete(ctx, req.GetId()); err != nil {
		h.metrics.IncRequest("DeleteSample", "error")
		return nil, toGRPCError(err)
	}
	h.metrics.IncRequest("DeleteSample", "ok")
	return &samplev1.DeleteSampleResponse{Success: true}, nil
}

func toProtoSample(s *entity.Sample) *samplev1.SampleResponse {
	return &samplev1.SampleResponse{
		Id:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Status:      s.Status,
		CreatedAt:   timestamppb.New(s.CreatedAt),
		UpdatedAt:   timestamppb.New(s.UpdatedAt),
	}
}

func toGRPCError(err error) error {
	if appErr, ok := apperrors.As(err); ok {
		switch appErr.Code {
		case apperrors.CodeNotFound:
			return status.Error(codes.NotFound, appErr.Message)
		case apperrors.CodeValidation, apperrors.CodeBadRequest:
			return status.Error(codes.InvalidArgument, appErr.Message)
		case apperrors.CodeConflict:
			return status.Error(codes.AlreadyExists, appErr.Message)
		case apperrors.CodeUnauthorized:
			return status.Error(codes.Unauthenticated, appErr.Message)
		case apperrors.CodeForbidden:
			return status.Error(codes.PermissionDenied, appErr.Message)
		case apperrors.CodeUnavailable:
			return status.Error(codes.Unavailable, appErr.Message)
		case apperrors.CodeTimeout:
			return status.Error(codes.DeadlineExceeded, appErr.Message)
		default:
			return status.Error(codes.Internal, appErr.Message)
		}
	}
	return status.Error(codes.Internal, "internal error")
}
