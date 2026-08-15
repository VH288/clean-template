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
}

// WireWorker wires outbox relay and Kafka consumer for the worker process.
func WireWorker(infra *Infra, cfg *config.Config) *Worker {
	publisher := samplemsg.NewPublisher(infra.KafkaProd)
	handler := samplemsg.NewEventHandler(infra.Logger)

	return &Worker{
		OutboxRelay: outbox.NewRelay(infra.DB, publisher, infra.Logger, cfg.Outbox),
		Consumer:    kafka.NewConsumer(cfg.Kafka, infra.Logger, handler.Handle),
	}
}
