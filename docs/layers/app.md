# Application Layer (`internal/app/`)

Bootstrap (manual DI), HTTP router, server wrappers, gRPC server assembly, middleware, and graceful shutdown lifecycle.

---

## Bootstrap (`internal/app/bootstrap/`)

### `bootstrap.go`

- **Purpose:** Root `Build()` and HTTP router assembly with middleware stack.
- **Layer / role:** Composition root.
- **Key types / functions:**
  - `Build(ctx, cfg)` — wires infra, sample, healthcheck, lifecycle; returns `*Container`
  - `(c *Container) HTTPRouter()` — builds chi router with domain transports
  - `(c *Container) middlewares()` — ordered middleware list
- **Dependencies:** `internal/app/router`, `internal/app/middleware`, `internal/config`
- **Used by:** All `cmd/*` entrypoints except migrate (indirectly via shared config)
- **Patterns:** Manual DI composition root
- **Notes:** Middleware order: Recoverer → RequestID → Logging → Metrics → Tracing.

### `container.go`

- **Purpose:** Root DI container structs holding wired domain modules and shared infra.
- **Layer / role:** Composition root types.
- **Key types / functions:**
  - `Container` — `Config`, `Logger`, `Infra`, `Sample`, `Healthcheck`, `Lifecycle`
  - `Infra` — shared `DB`, `Redis`, `Mongo`, `KafkaProd`, `GRPCClient`, `WSHub`, `HTTPClient`, `APM`, `Logger`, `Config`
- **Dependencies:** Infrastructure client types, `lifecycle.Manager`
- **Used by:** `bootstrap.Build`, `cmd/*` mains
- **Patterns:** Per-domain fields on `Container` for growth without bloating `bootstrap.go`
- **Notes:** Add new domain struct fields here when introducing domains.

### `infra.go`

- **Purpose:** Wires all shared infrastructure clients in dependency order with cleanup on failure.
- **Layer / role:** Infrastructure wiring.
- **Key types / functions:**
  - `wireInfra(ctx, cfg)` — creates logger, APM (warn on failure), Postgres, Redis, Mongo, optional gRPC client, Kafka producer, WS hub, HTTP client
- **Dependencies:** `database`, `redis`, `mongodb`, `kafka`, `grpc`, `external`, `websocket`, `apm`, `logger`, `config`
- **Used by:** `bootstrap.Build`
- **Patterns:** Fail-fast with partial cleanup (`Close` on earlier clients if later init fails)
- **Notes:** APM init failure is non-fatal (logs warning). External gRPC client is optional when `external.grpc_addr` is empty or dial fails.

### `sample.go`

- **Purpose:** Wires the sample domain: Postgres repo, Redis cache, Mongo document store → usecase → handlers.
- **Layer / role:** Domain module wiring.
- **Key types / functions:**
  - `Sample` — holds `HTTP`, `GRPC`, `WS` handlers
  - `wireSample(infra)` — constructs persistence adapters and handlers
- **Dependencies:** `domain/sample/handler`, `domain/sample/usecase`, `infrastructure/persistence/sample`
- **Used by:** `bootstrap.Build`
- **Patterns:** Repository → usecase → handler wiring per domain
- **Notes:** Does not wire Kafka publisher or outbox (worker-only via `WireWorker`).

### `healthcheck.go`

- **Purpose:** Wires healthcheck ping adapters → usecase → HTTP/gRPC/WS handlers.
- **Layer / role:** Domain module wiring.
- **Key types / functions:**
  - `Healthcheck` — `HTTP`, `GRPC`, `WS` handlers
  - `wireHealthcheck(infra)` — injects `PostgresPing`, `RedisPing`, `MongoPing`, `GRPCPing`
- **Dependencies:** `domain/healthcheck/adapter`, `handler`, `usecase`
- **Used by:** `bootstrap.Build`
- **Patterns:** Adapter structs satisfy checker ports
- **Notes:** `GRPCPing` receives optional upstream client from infra.

### `worker.go`

- **Purpose:** Wires outbox relay and Kafka consumer for the worker process.
- **Layer / role:** Worker composition.
- **Key types / functions:**
  - `Worker` — `OutboxRelay`, `Consumer`
  - `WireWorker(infra, cfg)` — publisher + event handler + relay + consumer
- **Dependencies:** `infrastructure/outbox`, `infrastructure/messaging/sample`, `infrastructure/kafka`
- **Used by:** `cmd/worker/main.go`
- **Patterns:** Separate worker wiring from API graph
- **Notes:** Consumer handler is `samplemsg.EventHandler.Handle`.

### `lifecycle.go`

- **Purpose:** Registers infrastructure shutdown hooks on the lifecycle manager (LIFO order).
- **Layer / role:** Shutdown registration.
- **Key types / functions:**
  - `wireLifecycle(infra)` — registers Kafka producer, Mongo, Redis, Postgres, optional gRPC client, optional APM shutdown
- **Dependencies:** `internal/app/lifecycle`
- **Used by:** `bootstrap.Build`
- **Patterns:** LIFO shutdown hook registration
- **Notes:** HTTP/gRPC shutdown hooks are added in `cmd/api` / `cmd/grpc` after `Build`, so they run before these infra hooks. Registration order here: Kafka → Mongo → Redis → Postgres → gRPC client → APM.

---

## App runtime

### `router/router.go`

- **Purpose:** Chi HTTP router: root health JSON, Prometheus metrics, `/api/v1` domain routes.
- **Layer / role:** HTTP routing assembly.
- **Key types / functions:**
  - `Dependencies` — sample/health handlers, middlewares, metrics path
  - `New(deps)` — chi router with RealIP, RequestID, custom middlewares, domain `RegisterHTTP` / `RegisterWebSocket`
- **Dependencies:** Domain transport packages, `infrastructure/metrics`, chi middleware
- **Used by:** `Container.HTTPRouter()`
- **Patterns:** Transport packages register routes; router only aggregates
- **Notes:** Default metrics path `/metrics` if `MetricsPath` empty. Root `GET /` returns `{"service":"clean-template"}`.

### `server/server.go`

- **Purpose:** HTTP and gRPC server wrappers with start/shutdown helpers.
- **Layer / role:** Transport server adapters.
- **Key types / functions:**
  - `HTTPServer` — `NewHTTP`, `Start`, `Shutdown`
  - `GRPCServer` — `NewGRPC`, `Start`, `GracefulStop`
- **Dependencies:** `net/http`, `google.golang.org/grpc`
- **Used by:** `cmd/api`, `cmd/grpc`
- **Patterns:** Thin wrapper around stdlib servers
- **Notes:** HTTP uses configured read/write timeouts. gRPC listens synchronously in `NewGRPC`.

### `grpc/server.go`

- **Purpose:** Assembles gRPC server with domain services, std health, and reflection.
- **Layer / role:** gRPC service registration.
- **Key types / functions:**
  - `Dependencies` — sample and healthcheck gRPC handlers
  - `NewServer(deps)` — registers domain services, `grpc.health.v1`, reflection
- **Dependencies:** Domain transport `RegisterGRPC`, `google.golang.org/grpc/health`
- **Used by:** `cmd/api`, `cmd/grpc`
- **Patterns:** Central gRPC registration point
- **Notes:** Sets serving status for empty service name and `sample.v1.SampleService`.

### `middleware/middleware.go`

- **Purpose:** HTTP middleware: request ID, logging, recovery, Prometheus metrics, OpenTelemetry tracing.
- **Layer / role:** HTTP cross-cutting concerns.
- **Key types / functions:**
  - `RequestID` — reads or generates `X-Request-ID`, stores in context
  - `Logging(base)` — per-request slog with latency and status
  - `Recoverer(base)` — panic recovery → 500
  - `Metrics` — Prometheus HTTP counters and histograms
  - `Tracing(serviceName)` — OTel span per request
  - `statusWriter` — captures response status for metrics/logging
- **Dependencies:** `pkg/helper`, `pkg/constant`, `infrastructure/logger`, `infrastructure/metrics`, OpenTelemetry
- **Used by:** `bootstrap.middlewares()`
- **Patterns:** Chi-compatible `func(http.Handler) http.Handler` middleware
- **Notes:** Logging middleware attaches request-scoped logger to context via `logger.WithContext`.

### `lifecycle/lifecycle.go`

- **Purpose:** SIGINT/SIGTERM handling and LIFO shutdown hook execution with timeout.
- **Layer / role:** Process lifecycle manager.
- **Key types / functions:**
  - `Manager` — `New`, `Add`, `Wait`
  - `ShutdownFunc` — `func(ctx context.Context) error`
- **Dependencies:** `os/signal`, `syscall`
- **Used by:** `bootstrap.wireLifecycle`, all long-running `cmd/*` processes
- **Patterns:** LIFO hook stack; bounded shutdown context
- **Notes:** Also exits when root `ctx` is cancelled (e.g. server error in `cmd/api`). Hook errors are logged but do not stop subsequent hooks.
