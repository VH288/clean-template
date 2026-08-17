package router

import (
	"net/http"

	healthhandler "clean-template/internal/domain/healthcheck/handler"
	healthtransport "clean-template/internal/domain/healthcheck/transport"
	samplehandler "clean-template/internal/domain/sample/handler"
	sampletransport "clean-template/internal/domain/sample/transport"
	"clean-template/internal/infrastructure/metrics"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Dependencies struct {
	SampleHTTP  *samplehandler.HTTPHandler
	SampleWS    *samplehandler.WSHandler
	HealthHTTP  *healthhandler.HTTPHandler
	HealthWS    *healthhandler.WSHandler
	Middlewares []func(http.Handler) http.Handler
	MetricsPath string
	MetricsAuth func(http.Handler) http.Handler
	APIAuth     func(http.Handler) http.Handler
}

func New(deps Dependencies) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.RealIP)
	for _, mw := range deps.Middlewares {
		r.Use(mw)
	}

	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"service":"clean-template"}`))
	})

	if deps.MetricsPath == "" {
		deps.MetricsPath = "/metrics"
	}
	metricsHandler := metrics.Handler()
	if deps.MetricsAuth != nil {
		metricsHandler = deps.MetricsAuth(metricsHandler)
	}
	r.Handle(deps.MetricsPath, metricsHandler)

	r.Route("/api/v1", func(api chi.Router) {
		healthtransport.RegisterHTTP(api, deps.HealthHTTP)
		healthtransport.RegisterWebSocket(api, deps.HealthWS)

		api.Group(func(protected chi.Router) {
			if deps.APIAuth != nil {
				protected.Use(deps.APIAuth)
			}
			sampletransport.RegisterHTTP(protected, deps.SampleHTTP)
			sampletransport.RegisterWebSocket(protected, deps.SampleWS)
		})
	})

	return r
}
