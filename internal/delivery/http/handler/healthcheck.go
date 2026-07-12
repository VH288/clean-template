package handler

import (
	"net/http"

	"clean-template/constants"
	"clean-template/internal/delivery/http/response"
	"clean-template/internal/service"

	"github.com/gin-gonic/gin"
)

type HealthcheckHandler struct {
	healthcheckService *service.HealthcheckService
}

func NewHealthcheckHandler(healthcheckService *service.HealthcheckService) *HealthcheckHandler {
	return &HealthcheckHandler{healthcheckService: healthcheckService}
}

func (h *HealthcheckHandler) Healthcheck(c *gin.Context) {
	msg, err := h.healthcheckService.Check(c.Request.Context())
	if err != nil {
		response.Send(c, http.StatusServiceUnavailable, constants.ErrExternalUnavailable, nil)
		return
	}
	response.Send(c, http.StatusOK, msg, nil)
}
