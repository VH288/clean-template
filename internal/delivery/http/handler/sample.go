package handler

import (
	"errors"
	"net/http"
	"strconv"

	"clean-template/constants"
	"clean-template/internal/delivery/http/dto"
	"clean-template/internal/delivery/http/response"
	"clean-template/internal/entity"
	"clean-template/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type SampleHandler struct {
	sampleService *service.SampleService
	log           *logrus.Logger
}

func NewSampleHandler(sampleService *service.SampleService, log *logrus.Logger) *SampleHandler {
	return &SampleHandler{
		sampleService: sampleService,
		log:           log,
	}
}

func (h *SampleHandler) ListSample(c *gin.Context) {
	var query dto.SampleQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.log.Error("failed to parse query param: ", err)
		response.Send(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	samples, err := h.sampleService.ListSample(c.Request.Context(), query.Page, query.Limit)
	if err != nil {
		h.log.Error("failed to get list sample: ", err)
		response.Send(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	resp := make([]dto.SampleResponse, 0, len(samples))
	for _, s := range samples {
		resp = append(resp, dto.ToSampleResponse(s.ID, s.Name))
	}

	response.Send(c, http.StatusOK, constants.SuccessMessage, resp)
}

func (h *SampleHandler) GetSample(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("failed to parse id: ", err)
		response.Send(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	sample, err := h.sampleService.GetSample(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Send(c, http.StatusNotFound, constants.ErrNotFound, nil)
			return
		}
		h.log.Error("failed to get sample: ", err)
		response.Send(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	response.Send(c, http.StatusOK, constants.SuccessMessage, dto.ToSampleResponse(sample.ID, sample.Name))
}

func (h *SampleHandler) CreateSample(c *gin.Context) {
	var req dto.SampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("failed to parse request: ", err)
		response.Send(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	sample := entity.Sample{Name: req.Name}
	created, err := h.sampleService.CreateSample(c.Request.Context(), sample)
	if err != nil {
		h.log.Error("failed to create sample: ", err)
		response.Send(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	response.Send(c, http.StatusOK, constants.SuccessMessage, dto.ToSampleResponse(created.ID, created.Name))
}

func (h *SampleHandler) UpdateSample(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("failed to parse id: ", err)
		response.Send(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	var req dto.SampleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.Error("failed to parse request: ", err)
		response.Send(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	sample := entity.Sample{Name: req.Name}
	if err := h.sampleService.UpdateSample(c.Request.Context(), sample, id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Send(c, http.StatusNotFound, constants.ErrNotFound, nil)
			return
		}
		h.log.Error("failed to update sample: ", err)
		response.Send(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	response.Send(c, http.StatusOK, constants.SuccessMessage, dto.ToSampleResponse(id, sample.Name))
}

func (h *SampleHandler) DeleteSample(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		h.log.Error("failed to parse id: ", err)
		response.Send(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	if err := h.sampleService.DeleteSample(c.Request.Context(), id); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			response.Send(c, http.StatusNotFound, constants.ErrNotFound, nil)
			return
		}
		h.log.Error("failed to delete sample: ", err)
		response.Send(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	response.Send(c, http.StatusOK, constants.SuccessMessage, nil)
}
