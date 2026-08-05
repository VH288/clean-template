package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"clean-template/internal/config"
	"clean-template/internal/infrastructure/apm"
	"clean-template/internal/infrastructure/database"
	"clean-template/internal/infrastructure/external"
	grpcclient "clean-template/internal/infrastructure/grpc"
	"clean-template/internal/infrastructure/kafka"
	"clean-template/internal/infrastructure/logger"
	"clean-template/internal/infrastructure/mongodb"
	infraredis "clean-template/internal/infrastructure/redis"
	infraws "clean-template/internal/infrastructure/websocket"
)

// wireInfra manually wires shared infrastructure.
func wireInfra(ctx context.Context, cfg *config.Config) (*Infra, error) {
	log := logger.New(cfg.Observability)
	slog.SetDefault(log)

	apmProvider, err := apm.Init(ctx, cfg.Observability, log)
	if err != nil {
		log.Warn("apm init failed, continuing without remote traces", "error", err)
	}

	db, err := database.NewPostgres(cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("postgres: %w", err)
	}

	rdb, err := infraredis.New(cfg.Redis)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("redis: %w", err)
	}

	mongoClient, err := mongodb.New(cfg.Mongo)
	if err != nil {
		_ = db.Close()
		_ = rdb.Close()
		return nil, fmt.Errorf("mongo: %w", err)
	}

	var grpcUpstream *grpcclient.Client
	if cfg.External.GRPCAddr != "" {
		client, dialErr := grpcclient.New(cfg.External.GRPCAddr)
		if dialErr != nil {
			log.Warn("external grpc client unavailable", "addr", cfg.External.GRPCAddr, "error", dialErr)
		} else {
			grpcUpstream = client
		}
	}

	return &Infra{
		DB:         db,
		Redis:      rdb,
		Mongo:      mongoClient,
		KafkaProd:  kafka.NewProducer(cfg.Kafka, log),
		GRPCClient: grpcUpstream,
		WSHub:      infraws.NewHub(log),
		HTTPClient: external.NewHTTPClient(cfg.External.HTTPURL),
		APM:        apmProvider,
		Logger:     log,
		Config:     cfg,
	}, nil
}
