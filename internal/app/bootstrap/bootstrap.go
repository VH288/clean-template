package bootstrap

import (
	"context"
	"net/http"

	"clean-template/internal/app/middleware"
	"clean-template/internal/app/router"
	"clean-template/internal/config"
)

func Build(ctx context.Context, cfg *config.Config) (*Container, error) {
	infra, err := wireInfra(ctx, cfg)
	if err != nil {
		return nil, err
	}

	return &Container{
		Config:      cfg,
		Logger:      infra.Logger,
		Infra:       infra,
		Sample:      wireSample(infra),
		Healthcheck: wireHealthcheck(infra),
		Lifecycle:   wireLifecycle(infra),
	}, nil
}

func (c *Container) HTTPRouter() http.Handler {
	auth := middleware.SkipPaths(
		[]string{"/api/v1/health", "/api/v1/ws/health"},
		middleware.APIKeyAuth(c.Config.App.Secret),
	)
	_ = auth

	return router.New(router.Dependencies{
		SampleHTTP:  c.Sample.HTTP,
		SampleWS:    c.Sample.WS,
		HealthHTTP:  c.Healthcheck.HTTP,
		HealthWS:    c.Healthcheck.WS,
		Middlewares: c.middlewares(),
		MetricsPath: c.Config.Observability.MetricsPath,
		MetricsAuth: middleware.MetricsAuth(c.Config.App.Secret),
		// APIAuth:     auth,
	})
}

func (c *Container) middlewares() []func(http.Handler) http.Handler {
	return []func(http.Handler) http.Handler{
		middleware.Recoverer(c.Logger),
		middleware.RequestID,
		middleware.Logging(c.Logger),
		middleware.Metrics,
		middleware.Tracing(c.Config.Observability.ServiceName),
	}
}
