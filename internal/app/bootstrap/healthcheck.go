package bootstrap

import (
	healthhandler "clean-template/internal/domain/healthcheck/handler"
	healthusecase "clean-template/internal/domain/healthcheck/usecase"
	infrahealth "clean-template/internal/infrastructure/healthcheck"
	"clean-template/internal/infrastructure/telemetry"
	infraws "clean-template/internal/infrastructure/websocket"
)

// Healthcheck is the manually wired healthcheck domain module.
type Healthcheck struct {
	HTTP *healthhandler.HTTPHandler
	GRPC *healthhandler.GRPCHandler
	WS   *healthhandler.WSHandler
}

// wireHealthcheck wires healthcheck: checkers → usecase → handlers.
func wireHealthcheck(infra *Infra) *Healthcheck {
	uc := healthusecase.New(
		infrahealth.PostgresPing{DB: infra.DB},
		infrahealth.RedisPing{Client: infra.Redis},
		infrahealth.MongoPing{Client: infra.Mongo},
		infrahealth.GRPCPing{Client: infra.GRPCClient},
		telemetry.NewTracer(),
		telemetry.NewHealthMetrics(),
	)

	wsHub := infraws.NewPortsAdapter(infra.WSHub)

	return &Healthcheck{
		HTTP: healthhandler.NewHTTPHandler(uc),
		GRPC: healthhandler.NewGRPCHandler(uc),
		WS:   healthhandler.NewWSHandler(uc, wsHub),
	}
}
