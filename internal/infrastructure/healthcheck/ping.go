package healthcheck

import (
	"context"

	"clean-template/internal/domain/healthcheck"
	"clean-template/internal/infrastructure/database"
	grpcclient "clean-template/internal/infrastructure/grpc"
	"clean-template/internal/infrastructure/mongodb"
	infraredis "clean-template/internal/infrastructure/redis"

	"github.com/jmoiron/sqlx"
	goredis "github.com/redis/go-redis/v9"
)

type PostgresPing struct{ DB *sqlx.DB }

func (p PostgresPing) Ping(ctx context.Context) error {
	return database.Ping(ctx, p.DB)
}

type RedisPing struct{ Client *goredis.Client }

func (r RedisPing) Ping(ctx context.Context) error {
	return infraredis.Ping(ctx, r.Client)
}

type MongoPing struct{ Client *mongodb.Client }

func (m MongoPing) Ping(ctx context.Context) error {
	return m.Client.Ping(ctx)
}

type GRPCPing struct{ Client *grpcclient.Client }

func (g GRPCPing) CheckHealth(ctx context.Context, service string) error {
	if g.Client == nil {
		return nil
	}
	return g.Client.CheckHealth(ctx, service)
}

var (
	_ healthcheck.PostgresChecker = PostgresPing{}
	_ healthcheck.RedisChecker    = RedisPing{}
	_ healthcheck.MongoChecker    = MongoPing{}
	_ healthcheck.GRPCChecker     = GRPCPing{}
)
