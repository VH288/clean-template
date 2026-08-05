package mapper

import (
	"clean-template/internal/domain/sample/dto"
	"clean-template/internal/domain/sample/entity"
)

func ToSampleResponse(s *entity.Sample) dto.SampleResponse {
	return dto.SampleResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Status:      s.Status,
		CreatedAt:   s.CreatedAt,
		UpdatedAt:   s.UpdatedAt,
	}
}

func ToSampleResponses(items []entity.Sample) []dto.SampleResponse {
	out := make([]dto.SampleResponse, 0, len(items))
	for i := range items {
		out = append(out, ToSampleResponse(&items[i]))
	}
	return out
}
