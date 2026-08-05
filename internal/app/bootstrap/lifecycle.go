package bootstrap

import (
	"context"

	"clean-template/internal/app/lifecycle"
)

// wireLifecycle registers graceful shutdown hooks (LIFO).
func wireLifecycle(infra *Infra) *lifecycle.Manager {
	lc := lifecycle.New(infra.Logger, infra.Config.HTTP.ShutdownTimeout)

	lc.Add(func(ctx context.Context) error { return infra.KafkaProd.Close() })
	lc.Add(func(ctx context.Context) error { return infra.Mongo.Close(ctx) })
	lc.Add(func(ctx context.Context) error { return infra.Redis.Close() })
	lc.Add(func(ctx context.Context) error { return infra.DB.Close() })

	if infra.GRPCClient != nil {
		lc.Add(func(ctx context.Context) error { return infra.GRPCClient.Close() })
	}
	if infra.APM != nil {
		lc.Add(infra.APM.Shutdown)
	}

	return lc
}
