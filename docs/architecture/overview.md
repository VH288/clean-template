# Architecture Overview

## Layer diagram

```
cmd/           entrypoints (api, grpc, worker, migrate, outbox-replay)
  ↓
internal/app/  bootstrap (manual DI), router, server, middleware, lifecycle
  ↓
internal/domain/   entities, ports, usecases, handlers, transport registration
  ↓
internal/infrastructure/   adapters: persistence, messaging, telemetry, clients, observability
```

Each **domain** (e.g. `sample`, `healthcheck`) owns its business logic. Infrastructure implements domain ports. Application code wires everything together at startup.

## Dependency rule

**Domain never imports infrastructure.**

- Repository and checker ports live in `internal/domain/<name>/interfaces.go`.
- Cross-cutting ports (tracer, logger, metrics, WebSocket hub) live in `internal/domain/ports/`.
- Infrastructure adapters live in `internal/infrastructure/**` (including `healthcheck/ping.go`, `telemetry/`, `websocket/ports_adapter.go`).
- `internal/app/bootstrap` is the composition root: it constructs concrete adapters and injects them into usecases and handlers.

Usecases depend on `ports.Tracer`, `ports.Logger`, and `ports.SampleMetrics` — implemented by `infrastructure/telemetry` adapters that wrap OpenTelemetry, slog, and Prometheus.

## Manual dependency injection

There is no Wire, Fx, or code-generated DI. Each domain has a `wire*` function in `internal/app/bootstrap/`:

- `wireInfra` — shared clients (Postgres, Redis, Mongo, Kafka producer, gRPC client, WebSocket hub, APM, logger)
- `wireSample` — repo → usecase (with telemetry ports) → HTTP/gRPC/WS handlers
- `wireHealthcheck` — infra ping adapters → usecase → handlers
- `WireWorker` — outbox relay + idempotent Kafka consumer + DLQ producer

The `Container` struct groups domain modules so `bootstrap.go` stays small as new domains are added.

## Multi-transport pattern

Every domain follows the same shape:

```
transport (route registration)
    → handler (protocol-specific: decode/encode, validation)
        → usecase (business logic)
            → repository / checker ports (interfaces)
                → infrastructure adapters
```

| Transport | Sample routes | Healthcheck routes |
|-----------|---------------|-------------------|
| HTTP | `/api/v1/samples` (auth when secret set) | `/api/v1/health`, `/api/v1/healthz` (public) |
| gRPC | `sample.v1.SampleService` | `healthcheck.v1.HealthcheckService` |
| WebSocket | `/api/v1/ws/samples` (auth when secret set) | `/api/v1/ws/health` (public) |

Handlers are thin; usecases are shared across transports.

## API authentication

When `app.secret` is configured to a non-placeholder value, sample routes and `/metrics` require `X-API-Key: <secret>`. Health endpoints remain public for load balancer probes.

## Sample CRUD data flow

Multi-store write on create/update/delete; cache-aside on read. Kafka publish is **not** direct from the usecase — events go through the transactional outbox with two-phase relay.

```mermaid
sequenceDiagram
    participant Client
    participant HTTP as HTTP Handler
    participant UC as Sample Usecase
    participant PG as Postgres Repo
    participant MG as Mongo Repo
    participant RD as Redis Cache
    participant OB as Outbox Relay
    participant KF as Kafka
    participant CN as Consumer

    Client->>HTTP: POST /api/v1/samples
    HTTP->>UC: Create(name, description, status)
    UC->>PG: CreateWithEvent (TX: sample + outbox_events, event_id)
    PG-->>UC: committed
    UC->>MG: Upsert (best-effort; warn + metric on failure)
    UC->>RD: Delete cache key (invalidate)
    UC-->>HTTP: entity.Sample
    HTTP-->>Client: 201 JSON

    Note over OB,KF: Worker process (async)
    OB->>PG: claim pending → status processing (commit TX)
    OB->>KF: Publish(aggregate_id, envelope with event_id)
    OB->>PG: mark published / retry / failed

    CN->>KF: FetchMessage (group_id)
    CN->>PG: processed_events idempotency check
    CN->>CN: parse Envelope, dispatch by type, mark processed
```

Updates use optimistic locking (`version` column). Stale concurrent updates return `409 Conflict`.

See [ADR 001](adr/001-transactional-outbox.md) for publish reliability rationale.

## Worker process

`cmd/worker` runs two background goroutines:

1. **Outbox relay** — two-phase poll: claim `pending` rows as `processing`, publish outside DB transaction, mark `published` or retry/fail. Stale `processing` rows are reclaimed after `outbox.processing_stale_after`.
2. **Kafka consumer** — subscribes with `kafka.group_id`, idempotent via `processed_events` table, DLQ for poison messages, fetch backoff on errors.

Consumer group members share partition assignment; scale workers horizontally with the same `group_id`.

Failed outbox rows (`status = failed`) can be replayed via `cmd/outbox-replay`.

## Graceful shutdown

`internal/app/lifecycle.Manager` waits for `SIGINT` or `SIGTERM`, then runs shutdown hooks **LIFO** (last registered runs first) within `http.shutdown_timeout`.

Typical order for `cmd/api` (after hooks registered in main + bootstrap):

1. gRPC `GracefulStop`
2. HTTP `Shutdown`
3. APM tracer shutdown
4. External gRPC client close
5. Postgres pool close
6. Redis close
7. Mongo disconnect
8. Kafka producer close

Worker shutdown order:

1. Outbox relay `Stop()` (cancel goroutine, wait)
2. Kafka consumer `Close()` (stop + reader close)
3. DLQ producer `Close()`
4. Infra hooks (same as API)

## SOLID in this codebase

| Principle | Example |
|-----------|---------|
| **S** — Single responsibility | `PostgresRepository` only handles SQL; `SampleUsecase` only orchestrates business rules |
| **O** — Open/closed | New storage backend: implement `sample.Repository` without changing usecase |
| **L** — Liskov substitution | `PostgresPing`, `RedisPing`, etc. satisfy `healthcheck.PostgresChecker` / `RedisChecker` ports |
| **I** — Interface segregation | `CacheRepository` and `DocumentRepository` are separate from `Repository`; telemetry split into small ports |
| **D** — Dependency inversion | Usecase depends on `ports.Tracer` / `sample.Repository`; bootstrap injects concrete adapters |

## Cross-links

- [Sample usecase](layers/domain-sample.md#internaldomainsampleusecaseusecasego) — business logic and multi-store orchestration
- [Postgres repository](layers/infrastructure.md#internalinfrastructurepersistencesamplepostgresgo) — transactional outbox writes + optimistic locking
- [Outbox relay](layers/infrastructure.md#internalinfrastructureoutboxrelaygo) — two-phase claim and publish
- [ADR 001: Transactional Outbox](adr/001-transactional-outbox.md)
