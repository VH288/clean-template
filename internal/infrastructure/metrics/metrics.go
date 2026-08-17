package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_request_duration_seconds",
		Help:    "HTTP request latency in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "path"})

	GRPCRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "grpc_requests_total",
		Help: "Total number of gRPC requests",
	}, []string{"method", "code"})

	SampleOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "sample_operations_total",
		Help: "Total sample domain operations",
	}, []string{"operation", "status"})

	HealthcheckStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "healthcheck_dependency_up",
		Help: "Health status of dependencies (1=up, 0=down)",
	}, []string{"dependency"})

	KafkaEventsConsumed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kafka_events_consumed",
		Help: "Total Kafka events consumed by type and status",
	}, []string{"type", "status"})

	CacheInvalidateTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_invalidate_total",
		Help: "Total cache invalidation attempts",
	}, []string{"status"})

	CacheOperationsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_operations_total",
		Help: "Total cache read/write operations",
	}, []string{"operation", "status"})

	MongoSyncFailuresTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "mongo_sync_failures_total",
		Help: "Total Mongo best-effort sync failures after Postgres commit",
	}, []string{"operation"})

	OutboxEventsPending = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_events_pending",
		Help: "Number of pending outbox events",
	})

	OutboxEventsFailed = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "outbox_events_failed",
		Help: "Number of failed outbox events",
	})
)

func Handler() http.Handler {
	return promhttp.Handler()
}
