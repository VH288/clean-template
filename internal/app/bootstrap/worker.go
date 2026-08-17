package bootstrap

import (
	"clean-template/internal/config"
	"clean-template/internal/infrastructure/kafka"
	samplemsg "clean-template/internal/infrastructure/messaging/sample"
	"clean-template/internal/infrastructure/outbox"
)

// Worker bundles background workers: outbox relay and Kafka consumer.
type Worker struct {
	OutboxRelay *outbox.Relay
	Consumer    *kafka.Consumer
	DLQProducer *kafka.DLQProducer
}

// WireWorker wires outbox relay and Kafka consumer for the worker process.
func WireWorker(infra *Infra, cfg *config.Config) *Worker {
	publisher := samplemsg.NewPublisher(infra.KafkaProd)
	dlq := kafka.NewDLQProducer(cfg.Kafka)
	store := samplemsg.NewPostgresIdempotencyStore(infra.DB)
	handler := samplemsg.NewEventHandler(infra.Logger, store, dlq)

	return &Worker{
		OutboxRelay: outbox.NewRelay(infra.DB, publisher, infra.Logger, cfg.Outbox),
		Consumer: kafka.NewConsumer(
			cfg.Kafka,
			infra.Logger,
			handler.Handle,
			samplemsg.IsPoisonError,
		),
		DLQProducer: dlq,
	}
}
