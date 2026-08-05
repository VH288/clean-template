package usecase_test

import (
	"context"
	"errors"
	"testing"

	"clean-template/internal/domain/healthcheck"
	"clean-template/internal/domain/healthcheck/usecase"
	"clean-template/internal/pkg/constant"
	"clean-template/internal/pkg/testutil"

	"github.com/stretchr/testify/require"
)

type pingFunc func(ctx context.Context) error

func (f pingFunc) Ping(ctx context.Context) error { return f(ctx) }

type grpcFunc func(ctx context.Context, service string) error

func (f grpcFunc) CheckHealth(ctx context.Context, service string) error {
	return f(ctx, service)
}

func TestHealthUsecase_CheckHTTP(t *testing.T) {
	ctx := testutil.NewContext()
	uc := usecase.New(
		pingFunc(func(ctx context.Context) error { return nil }),
		pingFunc(func(ctx context.Context) error { return nil }),
		pingFunc(func(ctx context.Context) error { return nil }),
		grpcFunc(func(ctx context.Context, service string) error { return nil }),
	)

	report := uc.CheckHTTP(ctx)
	require.Equal(t, constant.HealthStatusUP, report.Status)
	require.Len(t, report.Dependencies, 2)
}

func TestHealthUsecase_CheckHTTP_Down(t *testing.T) {
	ctx := testutil.NewContext()
	uc := usecase.New(
		pingFunc(func(ctx context.Context) error { return errors.New("pg down") }),
		pingFunc(func(ctx context.Context) error { return nil }),
		pingFunc(func(ctx context.Context) error { return nil }),
		grpcFunc(func(ctx context.Context, service string) error { return nil }),
	)

	report := uc.CheckHTTP(ctx)
	require.Equal(t, constant.HealthStatusDOWN, report.Status)
}

func TestHealthUsecase_CheckGRPCAndWS(t *testing.T) {
	ctx := testutil.NewContext()
	uc := usecase.New(
		pingFunc(func(ctx context.Context) error { return nil }),
		pingFunc(func(ctx context.Context) error { return nil }),
		pingFunc(func(ctx context.Context) error { return nil }),
		grpcFunc(func(ctx context.Context, service string) error { return nil }),
	)

	require.Equal(t, constant.HealthStatusUP, uc.CheckGRPC(ctx).Status)
	require.Equal(t, constant.HealthStatusUP, uc.CheckWebSocket(ctx).Status)
	_ = healthcheck.Report{}
}
