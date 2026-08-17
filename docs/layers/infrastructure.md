# Infrastructure (`internal/infrastructure/`)

Adapters for databases, messaging, outbox relay, telemetry, external clients, WebSocket hub, and observability. Implements domain ports and supports worker processes.

Cross-link: [ADR 001 — Transactional Outbox](../adr/001-transactional-outbox.md).

---

## Core clients

### `database/postgres.go`

- **Purpose:** Opens sqlx Postgres pool with pool settings and initial ping.
- **Layer / role:** Database infrastructure.
- **Key types / functions:**
  - `NewPostgres(cfg)` — connect, configure pool, ping with 5s timeout
  - `Ping(ctx, db)` — `PingContext` wrapper for healthcheck
- **Dependencies:** `config.PostgresConfig`, `sqlx`, `lib/pq`
- **Used by:** `bootstrap.wireInfra`, `healthcheck.PostgresPing`
- **Patterns:** Connection pool tuning from config
- **Notes:** Returns error if initial ping fails; caller must close on failure paths in bootstrap.

---

### `redis/redis.go`

- **Purpose:** Creates go-redis client with connectivity check.
- **Layer / role:** Cache infrastructure.
- **Key types / functions:**
  - `New(cfg)` — client + ping with 5s timeout
  - `Ping(ctx, client)` — health probe
- **Dependencies:** `config.RedisConfig`, `go-redis/v9`
- **Used by:** `bootstrap.wireInfra`, `RedisPing`, sample Redis repository
- **Patterns:** Fail-fast on ping at startup
- **Notes:** Closes client if ping fails during `New`.

---

### `mongodb/mongodb.go`

- **Purpose:** MongoDB driver client wrapper with database handle.
- **Layer / role:** Document store infrastructure.
- **Key types / functions:**
  - `Client` — `Client`, `Database` fields
  - `New(cfg)` — connect, primary ping, 10s timeout
  - `Ping(ctx)`, `Close(ctx)`
- **Dependencies:** `mongo-driver`, `config.MongoConfig`
- **Used by:** Bootstrap, `MongoPing`, sample Mongo repository
- **Patterns:** Wrapper struct for ping and lifecycle
- **Notes:** Disconnects on ping failure during `New`.

---

### `kafka/kafka.go`

- **Purpose:** Kafka producer, DLQ producer, and consumer using segmentio/kafka-go.
- **Layer / role:** Messaging infrastructure.
- **Key types / functions:**
  - `Producer` — `NewProducer`, `Publish(ctx, key, value)`, `Close`
  - `DLQProducer` — `NewDLQProducer`, `PublishDLQ(ctx, key, value, reason)`, `Close`
  - `Consumer` — `NewConsumer`, `Start(ctx)`, `Stop`, `Close`
- **Dependencies:** `config.KafkaConfig`, `kafka-go`
- **Used by:** Bootstrap (producer), worker (consumer + DLQ), sample publisher
- **Patterns:** Sync producer (`Async: false`, `RequireOne` acks); consumer commit after successful handler; fetch backoff; poison message commit after DLQ or max retries
- **Notes:** `isPoison` callback determines immediate DLQ routing. Handler failures retry without commit until `max_handler_retries`.

---

### `grpc/client.go`

- **Purpose:** Upstream gRPC client with standard health check RPC.
- **Layer / role:** External gRPC client adapter.
- **Key types / functions:**
  - `Client` — `conn`, `health` client, `addr`
  - `New(addr)` — dial with block, 5s timeout, insecure creds
  - `Conn()`, `CheckHealth(ctx, service)`, `Close()`
- **Dependencies:** `grpc`, `grpc.health.v1`
- **Used by:** `GRPCPing`, bootstrap (optional)
- **Patterns:** Health probe via `grpc.health.v1.Health/Check`
- **Notes:** Expects `SERVING` status. Used by healthcheck gRPC transport path.

---

### `external/http.go`

- **Purpose:** Simple HTTP GET client for external REST APIs.
- **Layer / role:** External HTTP adapter.
- **Key types / functions:**
  - `HTTPClient` — `baseURL`, 10s timeout client
  - `NewHTTPClient(baseURL)`, `Get(ctx, path)`
- **Dependencies:** `net/http`
- **Used by:** `bootstrap.wireInfra` (wired but not used by sample/healthcheck in template)
- **Patterns:** Base URL + path GET helper
- **Notes:** Returns error on status ≥ 400. Default base URL: jsonplaceholder.typicode.com.

---

### `websocket/hub.go`

- **Purpose:** WebSocket connection hub: upgrade, track clients, broadcast JSON.
- **Layer / role:** WebSocket infrastructure.
- **Key types / functions:**
  - `Hub` — `Upgrade`, `Unregister`, `BroadcastJSON`
  - `WriteJSON(ctx, conn, payload)` — single-connection JSON write
- **Dependencies:** `github.com/coder/websocket`
- **Used by:** `PortsAdapter`, bootstrap
- **Patterns:** Hub tracks active connections; handlers own read loops
- **Notes:** `InsecureSkipVerify: true` on accept (dev-friendly). Unregister closes with normal closure status.

---

### `websocket/ports_adapter.go`

- **Purpose:** Adapts `Hub` to `domain/ports.WebSocketHub` for domain handlers.
- **Layer / role:** Port adapter.
- **Key types / functions:**
  - `PortsAdapter` — `Upgrade`, `Unregister`, `WriteJSON`
  - `connAdapter` — wraps `*websocket.Conn` as `ports.WebSocketConn`
- **Dependencies:** `domain/ports`, `Hub`
- **Used by:** `bootstrap.wireSample`, `bootstrap.wireHealthcheck`
- **Patterns:** Anti-corruption layer between domain and coder/websocket
- **Notes:** Keeps domain handlers free of infrastructure imports.

---

## Healthcheck adapters

### `healthcheck/ping.go`

- **Purpose:** Infrastructure ping adapters satisfying healthcheck checker ports.
- **Layer / role:** Healthcheck infrastructure adapters.
- **Key types / functions:**
  - `PostgresPing`, `RedisPing`, `MongoPing`, `GRPCPing`
- **Dependencies:** `database`, `redis`, `mongodb`, `grpc` packages
- **Used by:** `bootstrap.wireHealthcheck`
- **Patterns:** Moved from domain package to respect dependency rule
- **Notes:** `GRPCPing` returns nil when client is nil.

---

## Telemetry adapters

### `telemetry/telemetry.go`

- **Purpose:** Implements `domain/ports` telemetry interfaces using slog, OpenTelemetry, and Prometheus.
- **Layer / role:** Observability port adapters.
- **Key types / functions:**
  - `Tracer`, `SlogLogger`, `SampleMetrics`, `HealthMetrics`, `GRPCMetrics` — production adapters
  - `NoopTracer`, `NoopLogger`, `NoopSampleMetrics`, etc. — test no-ops
  - `DiscardLogger()` — slog discard handler for tests
- **Dependencies:** `apm`, `logger`, `metrics`, `domain/ports`
- **Used by:** Bootstrap wiring, unit tests
- **Patterns:** Domain depends on ports; infrastructure provides concrete metrics/tracing
- **Notes:** Keeps usecases free of direct Prometheus/OTel imports.

---

## Observability

### `logger/logger.go`

- **Purpose:** slog JSON logger with service name; context-scoped request logger.
- **Layer / role:** Logging infrastructure.
- **Key types / functions:**
  - `New(cfg)`, `NewWithWriter(cfg, w)` — JSON handler, log level from config
  - `FromContext(ctx)`, `WithContext(ctx, log)` — context logger propagation
  - `parseLevel` — debug/info/warn/error
- **Dependencies:** `config.ObservabilityConfig`, `pkg/constant`
- **Used by:** Bootstrap (`slog.SetDefault`), middleware logging, telemetry `SlogLogger`
- **Patterns:** Context-carried slog for request-scoped fields
- **Notes:** JSON stdout is Loki-friendly when scraped by Alloy/Promtail.

---

### `apm/apm.go`

- **Purpose:** OpenTelemetry tracer provider with OTLP gRPC export to Tempo.
- **Layer / role:** Tracing infrastructure.
- **Key types / functions:**
  - `Provider` — holds `TracerProvider`
  - `Init(ctx, cfg, log)` — OTLP exporter, resource, parent-based ratio sampler
  - `Shutdown(ctx)`, `Tracer(name)`, `Start(ctx, name)` — span helper
- **Dependencies:** OTel SDK, OTLP gRPC exporter, `config`
- **Used by:** Bootstrap, `telemetry.Tracer` adapter, middleware tracing
- **Patterns:** Global `otel.SetTracerProvider`; trace context propagation
- **Notes:** Init failure in bootstrap is non-fatal. Endpoint from `observability.tempo_endpoint`.

---

### `metrics/metrics.go`

- **Purpose:** Prometheus metric definitions and `/metrics` handler.
- **Layer / role:** Metrics infrastructure.
- **Key types / functions:**
  - `HTTPRequestsTotal`, `HTTPRequestDuration` — HTTP middleware (route pattern labels)
  - `GRPCRequestsTotal` — gRPC handler counters
  - `SampleOperationsTotal` — domain operation outcomes
  - `HealthcheckStatus` — dependency up/down gauge (1/0)
  - `KafkaEventsConsumed` — includes `duplicate` status label
  - `CacheInvalidateTotal`, `CacheOperationsTotal`
  - `MongoSyncFailuresTotal` — Mongo best-effort write failures
  - `OutboxEventsPending`, `OutboxEventsFailed` — outbox relay gauges
  - `Handler()` — `promhttp.Handler()`
- **Dependencies:** `prometheus/client_golang`
- **Used by:** Middleware, telemetry adapters, consumer, outbox relay, router
- **Patterns:** `promauto` registration at package init
- **Notes:** HTTP metrics use chi route patterns (not raw URLs) to avoid UUID cardinality.

---

## Sample persistence (`persistence/sample/`)

### `postgres.go`

- **Purpose:** Postgres implementation of `sample.Repository` with transactional outbox inserts and optimistic locking.
- **Layer / role:** Persistence adapter.
- **Key types / functions:**
  - `PostgresRepository` — `NewPostgresRepository(db)`
  - CRUD + `CreateWithEvent`, `UpdateWithEvent`, `DeleteWithEvent` (with `eventID`)
  - `insertOutboxEvent` — writes pending row in same TX
  - `postgresModel`, `toModel`, `fromModel` — DB mapping including `version`
- **Dependencies:** `sqlx`, `entity`, `pkg/errors`, `uuid`
- **Used by:** `bootstrap.wireSample`, outbox relay (reads `outbox_events`)
- **Patterns:** Transactional outbox; UUID ID generation on create; version check on update
- **Notes:** `ErrNotFound` on missing rows; `ErrConflict` on stale version. Outbox: `aggregate_type=sample`, status lifecycle `pending` → `processing` → `published`/`failed`. See [ADR 001](../adr/001-transactional-outbox.md).

---

### `redis.go`

- **Purpose:** Redis cache-aside adapter for `sample.CacheRepository`.
- **Layer / role:** Cache persistence adapter.
- **Key types / functions:**
  - `RedisRepository` — TTL from `constant.SampleCacheTTLSec` (300s)
  - `Set`, `Get`, `Delete` — JSON marshal/unmarshal entity
  - `key(id)` — prefix `sample:`
- **Dependencies:** `go-redis`, `entity`, `constant`, `pkg/errors`
- **Used by:** `bootstrap.wireSample`
- **Patterns:** Cache-aside with TTL
- **Notes:** `redis.Nil` → `ErrNotFound`.

---

### `mongo.go`

- **Purpose:** MongoDB document store for `sample.DocumentRepository`.
- **Layer / role:** Document persistence adapter.
- **Key types / functions:**
  - `MongoRepository` — collection `samples`
  - `Upsert` — `UpdateOne` with upsert flag
  - `GetByID`, `Delete`
  - `mongoSampleDoc` — BSON mapping with `_id` = sample ID
- **Dependencies:** `mongo-driver`, `entity`, `pkg/errors`
- **Used by:** `bootstrap.wireSample`
- **Patterns:** Direct document upsert (not via Kafka consumer)
- **Notes:** `ErrNoDocuments` → `ErrNotFound`. Demo multi-store write from usecase; failures warned in usecase.

---

## Messaging (`messaging/sample/`)

### `publisher.go`

- **Purpose:** Publishes outbox payloads to Kafka keyed by aggregate ID.
- **Layer / role:** Messaging adapter (outbox publisher).
- **Key types / functions:**
  - `Publisher` — `NewPublisher(producer)`
  - `Publish(ctx, aggregateID, payload)` — delegates to `kafka.Producer`
- **Dependencies:** `infrastructure/kafka`
- **Used by:** `outbox.Relay`, `bootstrap.WireWorker`
- **Patterns:** Outbox relay publisher adapter
- **Notes:** Message key = sample UUID for partition affinity.

---

### `idempotency.go`

- **Purpose:** Postgres-backed idempotency store for Kafka consumer.
- **Layer / role:** Messaging infrastructure.
- **Key types / functions:**
  - `IdempotencyStore` interface — `IsProcessed`, `MarkProcessed`
  - `PostgresIdempotencyStore` — uses `processed_events` table
- **Dependencies:** `sqlx`
- **Used by:** `EventHandler`, `bootstrap.WireWorker`
- **Patterns:** At-least-once consumer with dedup by `event_id`
- **Notes:** `ON CONFLICT DO NOTHING` on mark processed.

---

### `consumer_handler.go`

- **Purpose:** Kafka consumer handler: parse `event.Envelope`, idempotency check, dispatch by type, DLQ poison messages.
- **Layer / role:** Messaging adapter (consumer).
- **Key types / functions:**
  - `EventHandler` — `NewEventHandler(logger, store, dlq)`
  - `Handle(ctx, key, value)` — unmarshal, idempotency, validate, switch on type
  - `ErrPoisonMessage` — sentinel for DLQ routing
- **Dependencies:** `domain/sample/event`, `metrics`
- **Used by:** `bootstrap.WireWorker` → `kafka.Consumer`
- **Patterns:** Extension point for side effects; idempotent dispatch
- **Notes:** Unknown types logged and skipped (commit). Missing `event_id` or invalid JSON → DLQ + poison error. Duplicates increment `kafka_events_consumed{type="duplicate"}`.

---

### `errors.go`

- **Purpose:** `IsPoisonError` helper for Kafka consumer poison routing.
- **Layer / role:** Messaging utilities.
- **Used by:** `bootstrap.WireWorker`, consumer tests

---

## Outbox

### `outbox/relay.go`

- **Purpose:** Background poller: two-phase claim and publish outbox rows to Kafka.
- **Layer / role:** Outbox relay worker.
- **Key types / functions:**
  - `Relay` — `NewRelay(db, publisher, logger, cfg)`
  - `Start(ctx)` / `Stop()` — ticker loop with graceful shutdown
  - `claimBatch` — TX: select pending `FOR UPDATE SKIP LOCKED` → set `processing`
  - `publishRow` — publish outside TX → mark `published` or retry/fail
  - `resetStaleProcessing` — reclaim crashed worker claims
  - `refreshGauges` — update `outbox_events_pending` / `outbox_events_failed`
  - `ReplayFailed(ctx)` — reset failed rows to pending
  - `ProcessBatchForTest` — test exposure
- **Dependencies:** `sqlx`, `config.OutboxConfig`, `eventPublisher` interface, `metrics`
- **Used by:** `cmd/worker`, `cmd/outbox-replay`, integration/unit tests
- **Patterns:** Two-phase transactional outbox relay; SKIP LOCKED for concurrent workers
- **Notes:** On publish failure: increment `retry_count`, reset to `pending` until `max_retries` then `failed`. Success sets `published` + `published_at`. See [ADR 001](../adr/001-transactional-outbox.md).

---

## Tests (grouped)

| File | Scope |
|------|-------|
| `persistence/sample/postgres_integration_test.go` | Real Postgres CRUD and outbox insert in transaction |
| `outbox/relay_integration_test.go` | Relay batch processing against real DB |
| `outbox/relay_test.go` | sqlmock unit tests for claim/publish/retry |
| `messaging/sample/consumer_handler_test.go` | Envelope parsing, idempotency, poison errors |
| `messaging/sample/consumer_integration_test.go` | Integration tests with Kafka (tags/integration) |

Run unit: `make test-unit`. Run integration: `DATABASE_URL=... make test-integration`.

---

## Related docs

- [Sample domain](../layers/domain-sample.md)
- [Worker entrypoint](../layers/cmd.md#cmdworkermain.go)
- [Migrations](../layers/migrations.md) — `outbox_events`, `processed_events` schemas
