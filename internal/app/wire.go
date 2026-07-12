package app

import (
	"clean-template/external/jsonplaceholder"
	"clean-template/external/reference"
	"clean-template/internal/config"
	deliverygrpc "clean-template/internal/delivery/grpc"
	grpchandler "clean-template/internal/delivery/grpc/handler"
	deliveryhttp "clean-template/internal/delivery/http"
	httphandler "clean-template/internal/delivery/http/handler"
	"clean-template/internal/infrastructure/database"
	"clean-template/internal/infrastructure/logger"
	"clean-template/internal/repository"
	"clean-template/internal/service"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Dependencies struct {
	Config             config.Config
	Log                *logrus.Logger
	DB                 *gorm.DB
	ReferenceGRPC      *reference.Client
	SampleService      *service.SampleService
	HealthcheckService *service.HealthcheckService
	HTTPServer         *deliveryhttp.Server
	GRPCServer         *deliverygrpc.Server
}

func Wire(cfg config.Config) Dependencies {
	log := logger.New()
	db := database.New(cfg, log)

	externalHTTP := jsonplaceholder.NewClient(cfg)

	externalGRPC, err := reference.NewClient(cfg)
	if err != nil {
		log.Fatal("failed to init external grpc client: ", err)
	}

	sampleRepo := repository.NewSampleRepo(db)
	sampleService := service.NewSampleService(sampleRepo, externalHTTP, externalGRPC)
	healthcheckService := service.NewHealthcheckService(externalHTTP, externalGRPC)

	httpHandlers := deliveryhttp.Handlers{
		Healthcheck: httphandler.NewHealthcheckHandler(healthcheckService),
		Sample:      httphandler.NewSampleHandler(sampleService, log),
	}

	router := deliveryhttp.NewRouter(log, httpHandlers)
	httpServer := deliveryhttp.NewServer(router, cfg.Port, log)

	grpcSampleHandler := grpchandler.NewSampleHandler(sampleService)
	grpcServer := deliverygrpc.NewServer(cfg.GRPCPort, log, grpcSampleHandler)

	return Dependencies{
		Config:             cfg,
		Log:                log,
		DB:                 db,
		ReferenceGRPC:      externalGRPC,
		SampleService:      sampleService,
		HealthcheckService: healthcheckService,
		HTTPServer:         httpServer,
		GRPCServer:         grpcServer,
	}
}

func (d Dependencies) Close() {
	if d.ReferenceGRPC != nil {
		_ = d.ReferenceGRPC.Close()
	}
}
