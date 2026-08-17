package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
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

type DLQProducer struct {
	writer *kafkago.Writer
}

func NewDLQProducer(cfg config.KafkaConfig) *DLQProducer {
	topic := cfg.DLQTopic
	if topic == "" {
		topic = "sample.events.dlq"
	}
	return &DLQProducer{
		writer: &kafkago.Writer{
			Addr:         kafkago.TCP(cfg.Brokers...),
			Topic:        topic,
			Balancer:     &kafkago.LeastBytes{},
			RequiredAcks: kafkago.RequireOne,
			Async:        false,
		},
	}
}

func (p *DLQProducer) PublishDLQ(ctx context.Context, key string, value []byte, reason string) error {
	return p.writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: value,
		Headers: []kafkago.Header{
			{Key: "dlq-reason", Value: []byte(reason)},
		},
		Time: time.Now().UTC(),
	})
}

func (p *DLQProducer) Close() error {
	if p == nil || p.writer == nil {
		return nil
	}
	return p.writer.Close()
}

type Consumer struct {
	reader           *kafkago.Reader
	logger           *slog.Logger
	handler          func(ctx context.Context, key, value []byte) error
	maxHandlerRetries int
	isPoison         func(error) bool

	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewConsumer(
	cfg config.KafkaConfig,
	logger *slog.Logger,
	handler func(ctx context.Context, key, value []byte) error,
	isPoison func(error) bool,
) *Consumer {
	maxRetries := cfg.MaxHandlerRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}
	reader := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:  cfg.Brokers,
		GroupID:  cfg.GroupID,
		Topic:    cfg.Topic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Consumer{
		reader:            reader,
		logger:            logger,
		handler:           handler,
		maxHandlerRetries: maxRetries,
		isPoison:          isPoison,
	}
}

func (c *Consumer) Start(ctx context.Context) {
	childCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		backoff := time.Second
		for {
			msg, err := c.reader.FetchMessage(childCtx)
			if err != nil {
				if childCtx.Err() != nil {
					return
				}
				c.logger.Error("kafka fetch failed", "error", err)
				time.Sleep(backoff)
				if backoff < 30*time.Second {
					backoff *= 2
				}
				continue
			}
			backoff = time.Second

			retries := retryCount(msg.Headers)
			err = c.handler(childCtx, msg.Key, msg.Value)
			if err != nil {
				if c.isPoison != nil && c.isPoison(err) {
					c.logger.Error("kafka poison message, committing", "error", err)
					_ = c.reader.CommitMessages(childCtx, msg)
					continue
				}
				if retries+1 >= c.maxHandlerRetries {
					c.logger.Error("kafka handler exceeded retries, committing", "error", err, "retries", retries)
					_ = c.reader.CommitMessages(childCtx, msg)
					continue
				}
				c.logger.Error("kafka handler failed", "error", err, "retries", retries)
				time.Sleep(backoff)
				continue
			}

			if err := c.reader.CommitMessages(childCtx, msg); err != nil {
				c.logger.Error("kafka commit failed", "error", err)
			}
		}
	}()
}

func (c *Consumer) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *Consumer) Close() error {
	c.Stop()
	return c.reader.Close()
}

func retryCount(headers []kafkago.Header) int {
	for _, h := range headers {
		if h.Key == "x-retry-count" {
			var n int
			_, _ = fmt.Sscanf(string(h.Value), "%d", &n)
			return n
		}
	}
	return 0
}
