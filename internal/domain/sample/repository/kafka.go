package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/infrastructure/kafka"
)

type KafkaPublisher struct {
	producer *kafka.Producer
}

func NewKafkaPublisher(producer *kafka.Producer) *KafkaPublisher {
	return &KafkaPublisher{producer: producer}
}

type sampleEvent struct {
	Type   string         `json:"type"`
	Sample *entity.Sample `json:"sample"`
}

func (p *KafkaPublisher) PublishSampleEvent(ctx context.Context, eventType string, sample *entity.Sample) error {
	payload, err := json.Marshal(sampleEvent{Type: eventType, Sample: sample})
	if err != nil {
		return err
	}
	if err := p.producer.Publish(ctx, sample.ID, payload); err != nil {
		return fmt.Errorf("publish sample event: %w", err)
	}
	return nil
}
