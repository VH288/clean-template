package http

import (
	"clean-template/internal/delivery/http/handler"
	"clean-template/internal/delivery/http/middleware"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handlers struct {
	Healthcheck *handler.HealthcheckHandler
	Sample      *handler.SampleHandler
}

func NewRouter(log *logrus.Logger, handlers Handlers) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger(log))

	r.GET("/health", handlers.Healthcheck.Healthcheck)

	sampleV1 := r.Group("/sample/v1")
	sampleV1.GET("/list", handlers.Sample.ListSample)
	sampleV1.GET("/get/:id", handlers.Sample.GetSample)
	sampleV1.POST("/create", handlers.Sample.CreateSample)
	sampleV1.PUT("/update/:id", handlers.Sample.UpdateSample)
	sampleV1.DELETE("/delete/:id", handlers.Sample.DeleteSample)

	return r
}
