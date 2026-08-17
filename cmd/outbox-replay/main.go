package main

import (
	"context"
	"fmt"
	"os"

	"clean-template/internal/app/bootstrap"
	"clean-template/internal/config"
	samplemsg "clean-template/internal/infrastructure/messaging/sample"
	"clean-template/internal/infrastructure/outbox"
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

	ctx := context.Background()
	container, err := bootstrap.Build(ctx, cfg)
	if err != nil {
		return err
	}

	publisher := samplemsg.NewPublisher(container.Infra.KafkaProd)
	relay := outbox.NewRelay(container.Infra.DB, publisher, container.Logger, cfg.Outbox)

	count, err := relay.ReplayFailed(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("replayed %d failed outbox events\n", count)
	return nil
}
