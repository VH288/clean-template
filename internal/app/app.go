package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"clean-template/internal/config"
	"golang.org/x/sync/errgroup"
)

type App struct {
	deps Dependencies
}

func New() *App {
	cfg := config.Load()
	return &App{deps: Wire(cfg)}
}

func (a *App) Run() error {
	defer a.deps.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.deps.HTTPServer.Start(ctx)
	})

	g.Go(func() error {
		return a.deps.GRPCServer.Start(ctx)
	})

	return g.Wait()
}
