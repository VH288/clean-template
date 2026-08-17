# Application Layer (`internal/app/`)

Bootstrap (manual DI), HTTP router, server wrappers, gRPC server assembly, middleware (including auth), and graceful shutdown lifecycle.

---

## Bootstrap (`internal/app/bootstrap/`)

### `bootstrap.go`

- **Purpose:** Root `Build()` and HTTP router assembly with middleware and auth stack.
- **Layer / role:** Composition root.
- **Key types / functions:**
  - `Build(ctx, cfg)` — wires infra, sample, healthcheck, lifecycle; returns `*Container`
  - `(c *Container) HTTPRouter()` — builds chi router with domain transports, auth, protected metrics
  - `(c *Container) middlewares()` — ordered middleware list
- **Dependencies:** `internal/app/router`, `internal/app/middleware`, `internal/config`
- **Used by:** All `cmd/*` entrypoints except migrate (indirectly via shared config)
- **Patterns:** Manual DI composition root; `SkipPaths` for public health routes
- **Notes:** Middleware order: Recoverer → RequestID → Logging → Metrics → Tracing. Sample routes wrapped with `APIKeyAuth` when secret is set.

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
  - `wireSample(infra)` — constructs persistence adapters, telemetry ports, and handlers
- **Dependencies:** `domain/sample/handler`, `domain/sample/usecase`, `infrastructure/persistence/sample`, `infrastructure/telemetry`, `infrastructure/websocket`
- **Used by:** `bootstrap.Build`
- **Patterns:** Repository → usecase (with ports) → handler wiring per domain
- **Notes:** Does not wire Kafka publisher or outbox (worker-only via `WireWorker`). WS hub exposed via `websocket.PortsAdapter`.

### `healthcheck.go`

- **Purpose:** Wires healthcheck ping adapters → usecase → HTTP/gRPC/WS handlers.
- **Layer / role:** Domain module wiring.
- **Key types / functions:**
  - `Healthcheck` — `HTTP`, `GRPC`, `WS` handlers
  - `wireHealthcheck(infra)` — injects `infrahealth.PostgresPing`, `RedisPing`, `MongoPing`, `GRPCPing`
- **Dependencies:** `infrastructure/healthcheck`, `domain/healthcheck/handler`, `domain/healthcheck/usecase`, `infrastructure/telemetry`
- **Used by:** `bootstrap.Build`
- **Patterns:** Infrastructure ping adapters satisfy checker ports
- **Notes:** `GRPCPing` receives optional upstream client from infra. Ping adapters moved out of domain package.

### `worker.go`

- **Purpose:** Wires outbox relay, idempotent Kafka consumer, and DLQ producer for the worker process.
- **Layer / role:** Worker composition.
- **Key types / functions:**
  - `Worker` — `OutboxRelay`, `Consumer`, `DLQProducer`
  - `WireWorker(infra, cfg)` — publisher + idempotency store + event handler + relay + consumer
- **Dependencies:** `infrastructure/outbox`, `infrastructure/messaging/sample`, `infrastructure/kafka`
- **Used by:** `cmd/worker/main.go`
- **Patterns:** Separate worker wiring from API graph
- **Notes:** Consumer uses `samplemsg.IsPoisonError` for DLQ routing.

### `lifecycle.go`

- **Purpose:** Registers infrastructure shutdown hooks on the lifecycle manager (LIFO order).
- **Layer / role:** Shutdown registration.
- **Key types / functions:**
  - `wireLifecycle(infra)` — registers Kafka producer, Mongo, Redis, Postgres, optional gRPC client, optional APM shutdown
- **Dependencies:** `internal/app/lifecycle`
- **Used by:** `bootstrap.Build`
- **Patterns:** LIFO shutdown hook registration
- **Notes:** HTTP/gRPC shutdown hooks are added in `cmd/api` / `cmd/grpc` after `Build`, so they run before these infra hooks. Worker adds relay/consumer/DLQ hooks before `Wait`.

---

## App runtime

### `router/router.go`

- **Purpose:** Chi HTTP router: root health JSON, protected Prometheus metrics, `/api/v1` domain routes with auth.
- **Layer / role:** HTTP routing assembly.
- **Key types / functions:**
  - `Dependencies` — sample/health handlers, middlewares, metrics path, `MetricsAuth`, `APIAuth`
  - `New(deps)` — chi router with RealIP, custom middlewares, public health routes, protected sample routes
- **Dependencies:** Domain transport packages, `infrastructure/metrics`, chi
- **Used by:** `Container.HTTPRouter()`
- **Patterns:** Transport packages register routes; router aggregates auth boundaries
- **Notes:** Does **not** use chi's built-in `RequestID` (custom middleware only). Default metrics path `/metrics` if `MetricsPath` empty. Root `GET /` returns `{"service":"clean-template"}`.

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

- **Purpose:** HTTP middleware: request ID, logging, recovery, Prometheus metrics (route patterns), OpenTelemetry tracing.
- **Layer / role:** HTTP cross-cutting concerns.
- **Key types / functions:**
  - `RequestID` — reads or generates `X-Request-ID`, stores in context
  - `Logging(base)` — per-request slog with latency and status
  - `Recoverer(base)` — panic recovery → 500
  - `Metrics` — Prometheus HTTP counters and histograms labeled by chi route pattern
  - `Tracing(serviceName)` — OTel span per request
  - `routePattern(r)` — resolves chi `RoutePattern()` to avoid UUID cardinality
  - `statusWriter` — captures response status for metrics/logging
- **Dependencies:** `pkg/helper`, `pkg/constant`, `infrastructure/logger`, `infrastructure/metrics`, OpenTelemetry, chi
- **Used by:** `bootstrap.middlewares()`
- **Patterns:** Chi-compatible `func(http.Handler) http.Handler` middleware
- **Notes:** Logging middleware attaches request-scoped logger to context via `logger.WithContext`.

### `middleware/auth.go`

- **Purpose:** API key authentication for protected routes and metrics.
- **Layer / role:** HTTP security middleware.
- **Key types / functions:**
  - `APIKeyAuth(secret)` — validates `X-API-Key` header; bypassed when secret empty or `CHANGE_ME`
  - `MetricsAuth(secret)` — alias for metrics endpoint protection
  - `SkipPaths(prefixes, auth)` — bypass auth for health probe paths
- **Dependencies:** `net/http`
- **Used by:** `bootstrap.HTTPRouter()`
- **Patterns:** Dev-friendly bypass for placeholder secrets
- **Notes:** Returns 401 without valid key when auth is active.

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
