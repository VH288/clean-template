//go:build integration

package sample_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"testing"
	"time"

	"clean-template/internal/config"
	"clean-template/internal/domain/sample/entity"
	"clean-template/internal/domain/sample/event"
	samplemsg "clean-template/internal/infrastructure/messaging/sample"
	"clean-template/internal/infrastructure/kafka"

	"github.com/stretchr/testify/require"
)

func TestConsumerHandler_KafkaRoundTrip_Integration(t *testing.T) {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		t.Skip("KAFKA_BROKERS not set")
	}

	cfg := config.KafkaConfig{
		Brokers: []string{brokers},
		GroupID: "clean-template-test",
		Topic:   "sample.events.test",
	}

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := samplemsg.NewEventHandler(logger)
	consumer := kafka.NewConsumer(cfg, logger, handler.Handle)
	defer func() { _ = consumer.Close() }()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	consumer.Start(ctx)

	producer := kafka.NewProducer(cfg, logger)
	defer func() { _ = producer.Close() }()

	s := &entity.Sample{ID: "test-id", Name: "kafka-roundtrip", Status: entity.StatusActive}
	payload, err := json.Marshal(event.Envelope{Type: event.TypeCreated, Sample: s})
	require.NoError(t, err)

	require.NoError(t, producer.Publish(ctx, s.ID, payload))

	// Smoke: publish succeeds; consumer processes asynchronously.
	time.Sleep(2 * time.Second)
}
