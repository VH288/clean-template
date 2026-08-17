# Clean Template — Documentation

English documentation for the **clean-template** Go service: a Clean Architecture scaffold with multi-transport APIs (HTTP, gRPC, WebSocket), multi-store persistence, transactional outbox messaging, and production-oriented patterns (auth, idempotency, DLQ, optimistic locking).

## What this template is

A learning and starter codebase that demonstrates:

- Layered Clean Architecture with manual dependency injection (no Wire/Fx)
- Domain ports (`interfaces.go`) and infrastructure adapters
- Cross-cutting domain ports (`internal/domain/ports`) for telemetry and WebSocket — domain never imports infrastructure
- Sample CRUD across Postgres, Redis, Mongo, and Kafka
- Healthcheck across HTTP, gRPC, and WebSocket transports
- API key auth, protected metrics, two-phase outbox relay, idempotent Kafka consumer, DLQ
- Observability: structured JSON logging, OpenTelemetry traces, Prometheus metrics

For quick start commands and stack overview, see the root [README.md](../README.md).

## How to read these docs

Start here, then follow dependency order:

1. **[Architecture overview](architecture/overview.md)** — layers, dependency rule, request lifecycle, SOLID examples
2. **[Command entrypoints](layers/cmd.md)** — `cmd/api`, `cmd/grpc`, `cmd/worker`, `cmd/migrate`, `cmd/outbox-replay`
3. **[Application layer](layers/app.md)** — bootstrap, router, servers, middleware, lifecycle, auth
4. **[Configuration](layers/config.md)** — Viper config structs and `config.yaml.example`
5. **Domains**
   - [Sample domain](layers/domain-sample.md) — CRUD, cache-aside, outbox events, optimistic locking
   - [Healthcheck domain](layers/domain-healthcheck.md) — dependency ping aggregation
6. **[Infrastructure](layers/infrastructure.md)** — databases, Kafka, outbox relay, telemetry adapters, observability
7. **[Shared packages](layers/pkg.md)** — errors, formatter, validator, test helpers
8. **[Protobuf](layers/proto.md)** — `.proto` sources and generation workflow
9. **[Migrations](layers/migrations.md)** — Goose SQL migrations
10. **[Root and tooling](root-and-tooling.md)** — Makefile, Dockerfile, CI, `go.mod`

## Architecture decision records

- [ADR 001: Transactional Outbox for Kafka Publish](adr/001-transactional-outbox.md) — two-phase outbox relay, idempotent consumer, DLQ, Mongo direct-write trade-offs

## Layer guides

| Document | Scope |
|----------|-------|
| [architecture/overview.md](architecture/overview.md) | Clean Architecture, dependency rule, diagrams |
| [layers/cmd.md](layers/cmd.md) | `cmd/*` entrypoints |
| [layers/app.md](layers/app.md) | `internal/app/*` |
| [layers/config.md](layers/config.md) | `internal/config`, `configs/` |
| [layers/domain-sample.md](layers/domain-sample.md) | `internal/domain/sample/**` |
| [layers/domain-healthcheck.md](layers/domain-healthcheck.md) | `internal/domain/healthcheck/**` |
| [layers/infrastructure.md](layers/infrastructure.md) | `internal/infrastructure/**` |
| [layers/pkg.md](layers/pkg.md) | `internal/pkg/**` |
| [layers/proto.md](layers/proto.md) | `internal/proto/**` |
| [layers/migrations.md](layers/migrations.md) | `migrations/*.sql` |
| [root-and-tooling.md](root-and-tooling.md) | Makefile, Docker, CI, scripts, module |
