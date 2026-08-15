package sample

import (
	"context"
	"fmt"

	"clean-template/internal/infrastructure/kafka"
)

// Publisher relays outbox events to Kafka.
type Publisher struct {
	producer *kafka.Producer
}

func NewPublisher(producer *kafka.Producer) *Publisher {
	return &Publisher{producer: producer}
}

func (p *Publisher) Publish(ctx context.Context, aggregateID string, payload []byte) error {
	if err := p.producer.Publish(ctx, aggregateID, payload); err != nil {
		return fmt.Errorf("publish sample event: %w", err)
	}
	return nil
}
