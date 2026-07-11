package cmd

import (
	"clean-template/external"
	"clean-template/helpers"
	"clean-template/internal/api"
	"clean-template/internal/interfaces"
	"clean-template/internal/repository"
	"clean-template/internal/services"
)

type Dependency struct {
	HealthcheckAPI interfaces.IHealthcheckHandler
	External       interfaces.IExternal
	SampleAPI      interfaces.ISampleHandler
}

func dependencyInject() Dependency {
	healthcheckSvc := &services.Healthcheck{}
	healthcheckAPI := &api.Healthcheck{
		HealthcheckServices: healthcheckSvc,
	}

	external := &external.External{}

	sampleRepo := &repository.SampleRepo{
		DB: helpers.DB,
	}

	sampleService := &services.SampleService{
		SampleRepo: sampleRepo,
	}

	sampleAPI := &api.SampleAPI{
		SampleService: sampleService,
	}

	return Dependency{
		HealthcheckAPI: healthcheckAPI,
		External:       external,
		SampleAPI:      sampleAPI,
	}
}
