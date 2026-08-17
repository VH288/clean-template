package usecase

import (
	"context"

	"clean-template/internal/domain/healthcheck"
	"clean-template/internal/domain/ports"
	"clean-template/internal/pkg/constant"
)

type HealthUsecase struct {
	postgres healthcheck.PostgresChecker
	redis    healthcheck.RedisChecker
	mongo    healthcheck.MongoChecker
	grpc     healthcheck.GRPCChecker
	trace    ports.Tracer
	metrics  ports.HealthMetrics
}

func New(
	postgres healthcheck.PostgresChecker,
	redis healthcheck.RedisChecker,
	mongo healthcheck.MongoChecker,
	grpc healthcheck.GRPCChecker,
	trace ports.Tracer,
	metrics ports.HealthMetrics,
) *HealthUsecase {
	return &HealthUsecase{
		postgres: postgres,
		redis:    redis,
		mongo:    mongo,
		grpc:     grpc,
		trace:    trace,
		metrics:  metrics,
	}
}

// CheckHTTP verifies Postgres + Redis (typical synchronous HTTP dependencies).
func (u *HealthUsecase) CheckHTTP(ctx context.Context) *healthcheck.Report {
	ctx, end := u.trace.Start(ctx, "healthcheck.CheckHTTP")
	defer end()

	deps := []healthcheck.DependencyStatus{
		u.check("postgres", func() error { return u.postgres.Ping(ctx) }),
		u.check("redis", func() error { return u.redis.Ping(ctx) }),
	}
	return assemble(deps)
}

// CheckGRPC verifies upstream gRPC via standard grpc.health.v1.
func (u *HealthUsecase) CheckGRPC(ctx context.Context) *healthcheck.Report {
	ctx, end := u.trace.Start(ctx, "healthcheck.CheckGRPC")
	defer end()

	deps := []healthcheck.DependencyStatus{
		u.check("grpc_upstream", func() error {
			if u.grpc == nil {
				return nil
			}
			return u.grpc.CheckHealth(ctx, "")
		}),
	}
	return assemble(deps)
}

// CheckWebSocket verifies MongoDB — used by realtime/document flows behind WS.
func (u *HealthUsecase) CheckWebSocket(ctx context.Context) *healthcheck.Report {
	ctx, end := u.trace.Start(ctx, "healthcheck.CheckWebSocket")
	defer end()

	deps := []healthcheck.DependencyStatus{
		u.check("mongodb", func() error { return u.mongo.Ping(ctx) }),
	}
	return assemble(deps)
}

func (u *HealthUsecase) check(name string, fn func() error) healthcheck.DependencyStatus {
	status := healthcheck.DependencyStatus{Name: name, Status: constant.HealthStatusUP}
	if err := fn(); err != nil {
		status.Status = constant.HealthStatusDOWN
		status.Message = err.Error()
		u.metrics.SetDependencyStatus(name, false)
		return status
	}
	u.metrics.SetDependencyStatus(name, true)
	return status
}

func assemble(deps []healthcheck.DependencyStatus) *healthcheck.Report {
	report := &healthcheck.Report{
		Status:       constant.HealthStatusUP,
		Dependencies: deps,
	}
	for _, d := range deps {
		if d.Status != constant.HealthStatusUP {
			report.Status = constant.HealthStatusDOWN
			break
		}
	}
	return report
}
