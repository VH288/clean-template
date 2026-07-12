package dto

type SampleQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

type SampleRequest struct {
	Name string `json:"name" binding:"required"`
}

type SampleResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func ToSampleResponse(id int, name string) SampleResponse {
	return SampleResponse{
		ID:   id,
		Name: name,
	}
}
