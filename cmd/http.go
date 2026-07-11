package cmd

import (
	"log"

	"clean-template/helpers"

	"github.com/gin-gonic/gin"
)

func bootstrapEngine() {
}

func ServeHTTP() {
	dependency := dependencyInject()

	r := gin.Default()

	route(r, dependency)

	err := r.Run(":" + helpers.GetEnv("PORT", "8080"))
	if err != nil {
		log.Fatal(err)
	}
}
