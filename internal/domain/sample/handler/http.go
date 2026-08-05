package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"clean-template/internal/domain/sample"
	"clean-template/internal/domain/sample/dto"
	"clean-template/internal/domain/sample/mapper"
	"clean-template/internal/pkg/formatter"
	"clean-template/internal/pkg/utils"
	"clean-template/internal/pkg/validator"

	"github.com/go-chi/chi/v5"
)

type HTTPHandler struct {
	usecase sample.Usecase
}

func NewHTTPHandler(usecase sample.Usecase) *HTTPHandler {
	return &HTTPHandler{usecase: usecase}
}

func (h *HTTPHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		formatter.Fail(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		formatter.Fail(w, err)
		return
	}

	result, err := h.usecase.Create(r.Context(), req.Name, req.Description, req.Status)
	if err != nil {
		formatter.Fail(w, err)
		return
	}
	formatter.Success(w, http.StatusCreated, "sample created", mapper.ToSampleResponse(result))
}

func (h *HTTPHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	result, err := h.usecase.GetByID(r.Context(), id)
	if err != nil {
		formatter.Fail(w, err)
		return
	}
	formatter.Success(w, http.StatusOK, "ok", mapper.ToSampleResponse(result))
}

func (h *HTTPHandler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	items, total, err := h.usecase.List(r.Context(), page, perPage)
	if err != nil {
		formatter.Fail(w, err)
		return
	}

	page, perPage, _ = utils.NormalizePagination(page, perPage)
	formatter.SuccessWithMeta(w, http.StatusOK, "ok", mapper.ToSampleResponses(items), &formatter.Meta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: utils.TotalPages(total, perPage),
	})
}

func (h *HTTPHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req dto.UpdateSampleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		formatter.Fail(w, err)
		return
	}
	if err := validator.Struct(req); err != nil {
		formatter.Fail(w, err)
		return
	}

	result, err := h.usecase.Update(r.Context(), id, req.Name, req.Description, req.Status)
	if err != nil {
		formatter.Fail(w, err)
		return
	}
	formatter.Success(w, http.StatusOK, "sample updated", mapper.ToSampleResponse(result))
}

func (h *HTTPHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.usecase.Delete(r.Context(), id); err != nil {
		formatter.Fail(w, err)
		return
	}
	formatter.Success(w, http.StatusOK, "sample deleted", nil)
}
