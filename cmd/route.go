package cmd

import "github.com/gin-gonic/gin"

func route(r *gin.Engine, dependency Dependency) {
	r.GET("/health", dependency.HealthcheckAPI.HealthcheckHandlerHTTP)

	sampleV1 := r.Group("/sample/v1")
	sampleV1.GET("/list", dependency.SampleAPI.ListSample)
	sampleV1.GET("/get/:id", dependency.SampleAPI.GetSample)
	sampleV1.POST("/create", dependency.SampleAPI.CreateSample)
	sampleV1.PUT("/update/:id", dependency.SampleAPI.UpdateSample)
	sampleV1.DELETE("/delete/:id", dependency.SampleAPI.DeleteSample)
}
