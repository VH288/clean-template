package bootstrap

import (
	samplehandler "clean-template/internal/domain/sample/handler"
	sampleusecase "clean-template/internal/domain/sample/usecase"
	"clean-template/internal/infrastructure/persistence/sample"
	"clean-template/internal/infrastructure/telemetry"
	infraws "clean-template/internal/infrastructure/websocket"
)

// Sample is the manually wired sample domain module.
type Sample struct {
	HTTP *samplehandler.HTTPHandler
	GRPC *samplehandler.GRPCHandler
	WS   *samplehandler.WSHandler
}

// wireSample wires sample: repository → usecase → handlers.
func wireSample(infra *Infra) *Sample {
	repo := sample.NewPostgresRepository(infra.DB)
	cache := sample.NewRedisRepository(infra.Redis)
	document := sample.NewMongoRepository(infra.Mongo.Database)

	uc := sampleusecase.New(
		repo,
		cache,
		document,
		telemetry.NewSlogLogger(),
		telemetry.NewTracer(),
		telemetry.NewSampleMetrics(),
	)

	wsHub := infraws.NewPortsAdapter(infra.WSHub)

	return &Sample{
		HTTP: samplehandler.NewHTTPHandler(uc),
		GRPC: samplehandler.NewGRPCHandler(uc, telemetry.NewGRPCMetrics()),
		WS:   samplehandler.NewWSHandler(uc, wsHub),
	}
}
