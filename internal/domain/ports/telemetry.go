package ports

import "context"

// Tracer starts observability spans without coupling domain to OpenTelemetry.
type Tracer interface {
	Start(ctx context.Context, name string) (context.Context, func())
}

// Logger provides structured logging from request context.
type Logger interface {
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
}

// SampleMetrics records sample-domain counters.
type SampleMetrics interface {
	IncOperation(operation, status string)
	IncCacheInvalidate(status string)
	IncCacheOperation(operation, status string)
	IncMongoSyncFailure(operation string)
}

// HealthMetrics records healthcheck gauges.
type HealthMetrics interface {
	SetDependencyStatus(name string, up bool)
}

// GRPCMetrics records gRPC handler counters.
type GRPCMetrics interface {
	IncRequest(method, result string)
}
