package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"clean-template/internal/app/bootstrap"
	"clean-template/internal/config"
	"clean-template/internal/infrastructure/kafka"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	container, err := bootstrap.Build(ctx, cfg)
	if err != nil {
		return err
	}

	consumer := kafka.NewConsumer(cfg.Kafka, container.Logger, func(ctx context.Context, key, value []byte) error {
		var payload map[string]any
		if err := json.Unmarshal(value, &payload); err != nil {
			return err
		}
		container.Logger.Info("sample event consumed",
			"key", string(key),
			"type", payload["type"],
		)
		return nil
	})
	consumer.Start(ctx)

	container.Lifecycle.Add(func(ctx context.Context) error {
		return consumer.Close()
	})
	container.Lifecycle.Wait(ctx)
	return nil
}
