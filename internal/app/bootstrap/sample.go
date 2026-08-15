package bootstrap

import (
	samplehandler "clean-template/internal/domain/sample/handler"
	sampleusecase "clean-template/internal/domain/sample/usecase"
	samplepersist "clean-template/internal/infrastructure/persistence/sample"
)

// Sample is the manually wired sample domain module.
type Sample struct {
	HTTP *samplehandler.HTTPHandler
	GRPC *samplehandler.GRPCHandler
	WS   *samplehandler.WSHandler
}

// wireSample wires sample: repository → usecase → handlers.
func wireSample(infra *Infra) *Sample {
	repo := samplepersist.NewPostgresRepository(infra.DB)
	cache := samplepersist.NewRedisRepository(infra.Redis)
	document := samplepersist.NewMongoRepository(infra.Mongo.Database)

	uc := sampleusecase.New(repo, cache, document)

	return &Sample{
		HTTP: samplehandler.NewHTTPHandler(uc),
		GRPC: samplehandler.NewGRPCHandler(uc),
		WS:   samplehandler.NewWSHandler(uc, infra.WSHub),
	}
}
