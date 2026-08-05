package handler_test

import (
	"context"
	"net/http"
	"testing"

	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/handler"
	"clean-template/internal/pkg/formatter"
	"clean-template/internal/pkg/testutil"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

type stubUsecase struct{}

func (stubUsecase) Create(ctx context.Context, name, description, status string) (*entity.Sample, error) {
	return &entity.Sample{ID: "1", Name: name, Description: description, Status: status}, nil
}
func (stubUsecase) GetByID(ctx context.Context, id string) (*entity.Sample, error) {
	return &entity.Sample{ID: id, Name: "n"}, nil
}
func (stubUsecase) List(ctx context.Context, page, perPage int) ([]entity.Sample, int64, error) {
	return []entity.Sample{{ID: "1", Name: "n"}}, 1, nil
}
func (stubUsecase) Update(ctx context.Context, id, name, description, status string) (*entity.Sample, error) {
	return &entity.Sample{ID: id, Name: name, Description: description, Status: status}, nil
}
func (stubUsecase) Delete(ctx context.Context, id string) error { return nil }

func TestSampleHTTPHandler_CreateList(t *testing.T) {
	h := handler.NewHTTPHandler(stubUsecase{})
	r := chi.NewRouter()
	r.Post("/samples", h.Create)
	r.Get("/samples", h.List)

	rr := testutil.PerformRequest(t, r, http.MethodPost, "/samples", map[string]string{
		"name":        "demo",
		"description": "d",
		"status":      "active",
	})
	testutil.AssertStatus(t, rr, http.StatusCreated)

	var resp formatter.Response
	testutil.DecodeJSON(t, rr, &resp)
	require.True(t, resp.Success)

	rr = testutil.PerformRequest(t, r, http.MethodGet, "/samples", nil)
	testutil.AssertStatus(t, rr, http.StatusOK)
}
