package api

import (
	"net/http"

	"clean-template/constants"
	"clean-template/helpers"
	"clean-template/internal/interfaces"

	"github.com/gin-gonic/gin"
)

type Healthcheck struct {
	HealthcheckServices interfaces.IHealthcheckServices
}

func (api *Healthcheck) HealthcheckHandlerHTTP(c *gin.Context) {
	msg, err := api.HealthcheckServices.HealthcheckServices()
	if err != nil {
		helpers.SendResponseHTTP(c, http.StatusBadRequest, constants.ErrFailedBadRequest, nil)
		return
	}
	helpers.SendResponseHTTP(c, http.StatusOK, msg, nil)
}
