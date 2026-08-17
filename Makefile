.PHONY: help run run-api run-grpc run-worker air air-api air-grpc air-worker \
	migrate-up migrate-down migrate-status \
	proto test test-unit test-integration tidy build docker-up docker-down lint

APP_NAME ?= clean-template
CONFIG ?= configs/config.yaml
GO ?= go
PROTOC ?= protoc
AIR ?= $(shell go env GOPATH)/bin/air

help:
	@echo "Targets:"
	@echo "  run / run-api       - run HTTP+gRPC API (cmd/api)"
	@echo "  run-grpc            - run gRPC-only server"
	@echo "  run-worker          - run Kafka consumer worker"
	@echo "  air / air-api       - live reload API (air)"
	@echo "  air-grpc            - live reload gRPC server (air)"
	@echo "  air-worker          - live reload worker (air)"
	@echo "  migrate-up|down|status - goose migrations"
	@echo "  proto               - generate protobuf stubs"
	@echo "  test / test-unit / test-integration"
	@echo "  tidy / build / docker-up / docker-down"

run: run-api

run-api:
	$(GO) run ./cmd/api

run-grpc:
	$(GO) run ./cmd/grpc

run-worker:
	$(GO) run ./cmd/worker

air: air-api

air-api:
	$(AIR)

air-grpc:
	$(AIR) -c .air.grpc.toml

air-worker:
	$(AIR) -c .air.worker.toml

migrate-up:
	$(GO) run ./cmd/migrate -dir migrations up

migrate-down:
	$(GO) run ./cmd/migrate -dir migrations down

migrate-status:
	$(GO) run ./cmd/migrate -dir migrations status

proto:
	$(PROTOC) -I internal/proto \
		--go_out=internal/proto --go_opt=module=clean-template/internal/proto \
		--go-grpc_out=internal/proto --go-grpc_opt=module=clean-template/internal/proto \
		internal/proto/sample/v1/sample.proto \
		internal/proto/healthcheck/v1/healthcheck.proto

test:
	$(GO) test ./... -count=1

test-unit:
	$(GO) test ./internal/domain/... ./internal/pkg/... ./internal/infrastructure/... -count=1 -short

test-integration:
	$(GO) test ./internal/domain/... ./internal/infrastructure/... -tags=integration -count=1 -run Integration

tidy:
	$(GO) mod tidy

build:
	$(GO) build -o bin/api ./cmd/api
	$(GO) build -o bin/grpc ./cmd/grpc
	$(GO) build -o bin/worker ./cmd/worker
	$(GO) build -o bin/migrate ./cmd/migrate
	$(GO) build -o bin/outbox-replay ./cmd/outbox-replay

lint:
	$(GO) vet ./...
