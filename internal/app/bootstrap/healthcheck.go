package bootstrap

import (
	healthadapter "clean-template/internal/domain/healthcheck/adapter"
	healthhandler "clean-template/internal/domain/healthcheck/handler"
	healthusecase "clean-template/internal/domain/healthcheck/usecase"
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
		healthadapter.PostgresPing{DB: infra.DB},
		healthadapter.RedisPing{Client: infra.Redis},
		healthadapter.MongoPing{Client: infra.Mongo},
		healthadapter.GRPCPing{Client: infra.GRPCClient},
	)

	return &Healthcheck{
		HTTP: healthhandler.NewHTTPHandler(uc),
		GRPC: healthhandler.NewGRPCHandler(uc),
		WS:   healthhandler.NewWSHandler(uc, infra.WSHub),
	}
}
