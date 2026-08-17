package constant

const (
	HeaderRequestID     = "X-Request-ID"
	HeaderTraceID       = "X-Trace-ID"
	ContextKeyRequestID = "request_id"
	ContextKeyLogger    = "logger"

	SampleCacheKeyPrefix = "sample:"
	SampleCacheTTLSec    = 300

	KafkaTopicSampleEvents    = "sample.events"
	KafkaTopicSampleEventsDLQ = "sample.events.dlq"

	HealthStatusUP   = "UP"
	HealthStatusDOWN = "DOWN"
)
