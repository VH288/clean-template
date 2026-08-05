package grpcapp

import (
	healthhandler "clean-template/internal/domain/healthcheck/handler"
	healthtransport "clean-template/internal/domain/healthcheck/transport"
	samplehandler "clean-template/internal/domain/sample/handler"
	sampletransport "clean-template/internal/domain/sample/transport"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Dependencies struct {
	Sample *samplehandler.GRPCHandler
	Health *healthhandler.GRPCHandler
}

func NewServer(deps Dependencies) *grpc.Server {
	server := grpc.NewServer()

	sampletransport.RegisterGRPC(server, deps.Sample)
	healthtransport.RegisterGRPC(server, deps.Health)

	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("sample.v1.SampleService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(server)
	return server
}
