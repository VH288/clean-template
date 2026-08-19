package telemetry

import (
	"context"
	"log/slog"

	"clean-template/internal/domain/ports"
	"clean-template/internal/infrastructure/apm"
	"clean-template/internal/infrastructure/logger"
	"clean-template/internal/infrastructure/metrics"
)

type Tracer struct{}

func NewTracer() *Tracer { return &Tracer{} }

func (Tracer) Start(ctx context.Context, name string) (context.Context, func()) {
	ctx, span := apm.Start(ctx, name)
	return ctx, func() { span.End() }
}

type SlogLogger struct{}

func NewSlogLogger() *SlogLogger { return &SlogLogger{} }

func (SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	logger.FromContext(ctx).Info(msg, args...)
}

func (SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	logger.FromContext(ctx).Warn(msg, args...)
}

func (SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	logger.FromContext(ctx).Error(msg, args...)
}

type SampleMetrics struct{}

func NewSampleMetrics() *SampleMetrics { return &SampleMetrics{} }

func (SampleMetrics) IncOperation(operation, status string) {
	metrics.SampleOperationsTotal.WithLabelValues(operation, status).Inc()
}

func (SampleMetrics) IncCacheInvalidate(status string) {
	metrics.CacheInvalidateTotal.WithLabelValues(status).Inc()
}

func (SampleMetrics) IncCacheOperation(operation, status string) {
	metrics.CacheOperationsTotal.WithLabelValues(operation, status).Inc()
}

func (SampleMetrics) IncMongoSyncFailure(operation string) {
	metrics.MongoSyncFailuresTotal.WithLabelValues(operation).Inc()
}

type HealthMetrics struct{}

func NewHealthMetrics() *HealthMetrics { return &HealthMetrics{} }

func (HealthMetrics) SetDependencyStatus(name string, up bool) {
	val := 0.0
	if up {
		val = 1
	}
	metrics.HealthcheckStatus.WithLabelValues(name).Set(val)
}

type GRPCMetrics struct{}

func NewGRPCMetrics() *GRPCMetrics { return &GRPCMetrics{} }

func (GRPCMetrics) IncRequest(method, result string) {
	metrics.GRPCRequestsTotal.WithLabelValues(method, result).Inc()
}

// Noop implementations for tests.

type NoopTracer struct{}

func (NoopTracer) Start(ctx context.Context, _ string) (context.Context, func()) {
	return ctx, func() {}
}

type NoopLogger struct{}

func (NoopLogger) Info(context.Context, string, ...any)  {}
func (NoopLogger) Warn(context.Context, string, ...any)  {}
func (NoopLogger) Error(context.Context, string, ...any) {}

type NoopSampleMetrics struct{}

func (NoopSampleMetrics) IncOperation(string, string)      {}
func (NoopSampleMetrics) IncCacheInvalidate(string)        {}
func (NoopSampleMetrics) IncCacheOperation(string, string) {}
func (NoopSampleMetrics) IncMongoSyncFailure(string)       {}

type NoopHealthMetrics struct{}

func (NoopHealthMetrics) SetDependencyStatus(string, bool) {}

type NoopGRPCMetrics struct{}

func (NoopGRPCMetrics) IncRequest(string, string) {}

var (
	_ ports.Tracer        = Tracer{}
	_ ports.Logger        = SlogLogger{}
	_ ports.SampleMetrics = SampleMetrics{}
	_ ports.HealthMetrics = HealthMetrics{}
	_ ports.GRPCMetrics   = GRPCMetrics{}
)

// DiscardLogger writes to io.Discard for tests.
func DiscardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}
