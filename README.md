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
- Persist ke Postgres (source of truth), cache Redis, document Mongo (direct write), event Kafka (transactional outbox)

## Multi-store & messaging patterns

| Store / bus | Role |
|---|---|
| **Postgres** | Source of truth; sample + `outbox_events` in one TX on writes |
| **Mongo** | Demo direct write dari usecase setelah Postgres sukses (bukan via consumer) |
| **Redis** | Cache-aside on read; invalidate (`Delete`) on write |
| **Kafka publish** | Transactional outbox → worker outbox relay → topic `sample.events` |
| **Kafka consume** | Worker `EventHandler`: subscribe `group_id` + `topic`, parse `event.Envelope`, dispatch by type |

```mermaid
sequenceDiagram
    participant API
    participant UC as Usecase
    participant PG as Postgres
    participant MG as Mongo
    participant RD as Redis
    participant OB as Outbox relay
    participant KF as Kafka
    participant CN as Consumer

    API->>UC: POST /samples
    UC->>PG: TX sample + outbox_events
    UC->>MG: Upsert (best-effort)
    UC->>RD: Delete cache key
    OB->>PG: poll pending outbox
    OB->>KF: publish envelope
    CN->>KF: subscribe (group_id)
    CN->>CN: log + metrics
```

**Consumer group:** `kafka.group_id` (default `clean-template`). Semua worker instance dengan group yang sama share partition assignment (rebalance otomatis).

**Run worker (outbox relay + consumer):**

```bash
make run-worker
```

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
make run-worker       # outbox relay + Kafka consumer
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
internal/domain/      healthcheck, sample (entities, ports, usecase, handlers)
internal/infrastructure/  persistence, messaging, outbox, db, redis, mongo, kafka, logger, apm, metrics, ws, grpc
internal/pkg/         errors, formatter, validator, testutil
internal/proto/       protobuf sources + generated stubs
migrations/           goose
deployments/          compose, nginx, prometheus, tempo
```
