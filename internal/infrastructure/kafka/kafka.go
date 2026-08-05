package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"clean-template/internal/config"

	kafkago "github.com/segmentio/kafka-go"
)

type Producer struct {
	writer *kafkago.Writer
	logger *slog.Logger
}

func NewProducer(cfg config.KafkaConfig, logger *slog.Logger) *Producer {
	writer := &kafkago.Writer{
		Addr:         kafkago.TCP(cfg.Brokers...),
		Topic:        cfg.Topic,
		Balancer:     &kafkago.LeastBytes{},
		RequiredAcks: kafkago.RequireOne,
		Async:        false,
	}
	return &Producer{writer: writer, logger: logger}
}

func (p *Producer) Publish(ctx context.Context, key string, value []byte) error {
	err := p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("publish kafka message: %w", err)
	}
	return nil
}

func (p *Producer) Close() error {
	return p.writer.Close()
}

type Consumer struct {
	reader  *kafkago.Reader
	logger  *slog.Logger
	handler func(ctx context.Context, key, value []byte) error
}

func NewConsumer(cfg config.KafkaConfig, logger *slog.Logger, handler func(ctx context.Context, key, value []byte) error) *Consumer {
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Consumer{reader: reader, logger: logger, handler: handler}
}

func (c *Consumer) Start(ctx context.Context) {
	go func() {
		for {
			msg, err := c.reader.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return
				}
				c.logger.Error("kafka fetch failed", "error", err)
				continue
			}

			if err := c.handler(ctx, msg.Key, msg.Value); err != nil {
				c.logger.Error("kafka handler failed", "error", err)
				continue
			}

			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Error("kafka commit failed", "error", err)
			}
		}
	}()
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
