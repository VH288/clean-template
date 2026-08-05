package bootstrap

import (
	samplehandler "clean-template/internal/domain/sample/handler"
	samplerepo "clean-template/internal/domain/sample/repository"
	sampleusecase "clean-template/internal/domain/sample/usecase"
)

// Sample is the manually wired sample domain module.
type Sample struct {
	HTTP *samplehandler.HTTPHandler
	GRPC *samplehandler.GRPCHandler
	WS   *samplehandler.WSHandler
}

// wireSample wires sample: repository → usecase → handlers.
func wireSample(infra *Infra) *Sample {
	repo := samplerepo.NewPostgresRepository(infra.DB)
	cache := samplerepo.NewRedisRepository(infra.Redis)
	document := samplerepo.NewMongoRepository(infra.Mongo.Database)
	publisher := samplerepo.NewKafkaPublisher(infra.KafkaProd)

	uc := sampleusecase.New(repo, cache, document, publisher)

	return &Sample{
		HTTP: samplehandler.NewHTTPHandler(uc),
		GRPC: samplehandler.NewGRPCHandler(uc),
		WS:   samplehandler.NewWSHandler(uc, infra.WSHub),
	}
}
