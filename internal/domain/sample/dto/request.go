package dto

type CreateSampleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Status      string `json:"status" validate:"omitempty,oneof=active inactive"`
}

type UpdateSampleRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"omitempty,max=500"`
	Status      string `json:"status" validate:"required,oneof=active inactive"`
}

type ListSampleQuery struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}
