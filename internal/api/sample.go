package api

import (
	"net/http"
	"strconv"

	"clean-template/constants"
	"clean-template/helpers"
	"clean-template/internal/interfaces"
	"clean-template/internal/models"

	"github.com/gin-gonic/gin"
)

type SampleAPI struct {
	SampleService interfaces.ISampleService
}

func (api *SampleAPI) ListSample(c *gin.Context) {
	log := helpers.Logger

	var param models.SampleParam

	if err := c.ShouldBindQuery(&param); err != nil {
		log.Error("failed to parse query param: ", err)
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	res, err := api.SampleService.ListSample(c.Request.Context(), param)
	if err != nil {
		log.Error("failed to get list sample: ", err)
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, res)
}

func (api *SampleAPI) GetSample(c *gin.Context) {
	log := helpers.Logger

	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error("failed to get id: ", err)
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	res, err := api.SampleService.GetSample(c.Request.Context(), id)
	if err != nil {
		log.Error("failed to get sample: ", err)
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, res)
}

func (api *SampleAPI) CreateSample(c *gin.Context) {
	var req models.Sample
	log := helpers.Logger

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to parse request: ", err)
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	if err := api.SampleService.CreateSample(c.Request.Context(), req); err != nil {
		log.Error("failed to create sample: ", err)
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, req)
}

func (api *SampleAPI) UpdateSample(c *gin.Context) {
	var req models.Sample
	log := helpers.Logger

	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error("failed to get id: ", err)
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Error("failed to parse request: ", err)
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	if err := api.SampleService.UpdateSample(c.Request.Context(), req, id); err != nil {
		log.Error("failed to update sample: ", err)
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, req)
}

func (api *SampleAPI) DeleteSample(c *gin.Context) {
	log := helpers.Logger

	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		log.Error("failed to get id: ", err)
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}

	if err := api.SampleService.DeleteSample(c.Request.Context(), id); err != nil {
		log.Error("failed to delete sample: ", err)
		helpers.SendResponseHTTP(c, http.StatusInternalServerError, constants.ErrServerError, nil)
		return
	}

	helpers.SendResponseHTTP(c, http.StatusOK, constants.SuccessMessage, nil)
}
