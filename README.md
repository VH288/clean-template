# Clean Architecture Go Template

Template service dengan Clean Architecture + multi-transport (HTTP / gRPC / WebSocket).

## Stack

| Layer | Tech |
|---|---|
| HTTP | chi |
| gRPC | google.golang.org/grpc (+ health.v1) |
| WebSocket | github.com/coder/websocket (successor nhooyr) |
| Postgres | sqlx |
| Cache | Redis |
| Document | MongoDB |
| Messaging | Kafka (segmentio) |
| Logging | slog JSON → Loki-friendly |
| Tracing | OpenTelemetry → Tempo |
| Metrics | Prometheus |
| Migration | Goose |
| Config | Viper |
| DI | Manual bootstrap (per domain) |

## Domain contoh

### 1. Healthcheck
- **HTTP** `/api/v1/health` → Postgres + Redis
- **gRPC** `healthcheck.v1.HealthcheckService/Check` → upstream gRPC via **grpc.health.v1** (recommended)
- **WebSocket** `/api/v1/ws/health` → MongoDB ping

### 2. Sample CRUD
- HTTP `/api/v1/samples`
- gRPC `sample.v1.SampleService`
- WebSocket `/api/v1/ws/samples`
- Persist ke Postgres, cache Redis, document Mongo, event Kafka

## Quick start

```bash
# infra
docker compose -f deployments/docker-compose.yml up -d postgres redis mongo kafka

# migrate + run
make migrate-up
make run-api
```

## Makefile

```bash
make run-api          # go run ./cmd/api
make run-grpc         # gRPC only
make run-worker       # Kafka consumer
make migrate-up
make proto            # generate protobuf
make test-unit
make test-integration # needs DATABASE_URL
make docker-up
```

## Test strategy

1. **Unit** — usecase dengan mock/in-memory repo (`*_test.go`, `make test-unit`)
2. **Handler** — HTTP handler + chi + testutil
3. **Integration** — real Postgres (`DATABASE_URL=... make test-integration`)
4. **testutil** — context, PerformRequest, DecodeJSON helpers

## Graceful shutdown

`internal/app/lifecycle` menunggu SIGINT/SIGTERM lalu menjalankan shutdown hooks LIFO (HTTP → gRPC → Kafka → Mongo → Redis → Postgres → APM).

## Layout

```
cmd/                  entrypoints (api, grpc, worker, migrate)
internal/app/         bootstrap (manual DI per domain), middleware, router, server, lifecycle
internal/config/      viper config
internal/domain/      healthcheck, sample
internal/infrastructure/  db, redis, mongo, kafka, logger, apm, metrics, ws, grpc
internal/pkg/         errors, formatter, validator, testutil
internal/proto/       protobuf sources + generated stubs
migrations/           goose
deployments/          compose, nginx, prometheus, tempo
```
