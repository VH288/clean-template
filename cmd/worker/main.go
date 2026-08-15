package main

import (
	"context"
	"fmt"
	"os"

	"clean-template/internal/app/bootstrap"
	"clean-template/internal/config"
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

	worker := bootstrap.WireWorker(container.Infra, cfg)
	worker.OutboxRelay.Start(ctx)
	worker.Consumer.Start(ctx)

	container.Lifecycle.Add(func(ctx context.Context) error {
		return worker.Consumer.Close()
	})
	container.Lifecycle.Wait(ctx)
	return nil
}
