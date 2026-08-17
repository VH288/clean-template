# Healthcheck Domain (`internal/domain/healthcheck/`)

Aggregates dependency health checks across HTTP (Postgres + Redis), gRPC (upstream gRPC health), and WebSocket (MongoDB) transports.

---

### `interfaces.go`

- **Purpose:** Health report types and checker/usecase ports.
- **Layer / role:** Domain ports and value types.
- **Key types / functions:**
  - `DependencyStatus` — `name`, `status`, optional `message`
  - `Report` — overall `status` + `dependencies` list
  - `PostgresChecker`, `RedisChecker`, `MongoChecker`, `GRPCChecker` — small ping ports
  - `Usecase` — `CheckHTTP`, `CheckGRPC`, `CheckWebSocket`
- **Dependencies:** None
- **Used by:** Usecase, adapters, handlers
- **Patterns:** Interface segregation per dependency type
- **Notes:** Status values use `constant.HealthStatusUP` / `HealthStatusDOWN`.

---

### `adapter/ping.go`

- **Purpose:** Thin adapters wrapping infrastructure clients to satisfy checker ports.
- **Layer / role:** Domain adapter (infrastructure bridge).
- **Key types / functions:**
  - `PostgresPing` — `Ping(ctx)` via `database.Ping`
  - `RedisPing` — `Ping(ctx)` via `redis.Ping`
  - `MongoPing` — `Ping(ctx)` via `mongodb.Client.Ping`
  - `GRPCPing` — `CheckHealth(ctx, service)` via gRPC health client; nil client → success
- **Dependencies:** `infrastructure/database`, `redis`, `mongodb`, `grpc` packages
- **Used by:** `bootstrap.wireHealthcheck`
- **Patterns:** Adapter structs with embedded infra clients
- **Notes:** `GRPCPing` returns nil when client is nil (upstream optional). Uses standard `grpc.health.v1` probe.

---

### `usecase/usecase.go`

- **Purpose:** Runs dependency checks per transport context and assembles health reports with metrics.
- **Layer / role:** Domain usecase.
- **Key types / functions:**
  - `HealthUsecase` — holds four checker ports
  - `New(postgres, redis, mongo, grpc)` — constructor
  - `CheckHTTP` — Postgres + Redis
  - `CheckGRPC` — upstream gRPC health (empty service name)
  - `CheckWebSocket` — MongoDB ping
  - `check`, `assemble` — per-dependency check + aggregate UP/DOWN
- **Dependencies:** `apm`, `metrics`, `pkg/constant`
- **Used by:** HTTP, gRPC, WS handlers
- **Patterns:** Transport-specific dependency subsets; gauge metrics per dependency
- **Notes:** Overall status DOWN if any dependency in the report is DOWN.

---

### `handler/handler.go`

- **Purpose:** HTTP, gRPC, and WebSocket handlers for health endpoints.
- **Layer / role:** Multi-transport handlers.
- **Key types / functions:**
  - `HTTPHandler.Check` — returns 200 or 503 with `formatter.Success`
  - `GRPCHandler.Check` — maps `Report` to proto `CheckResponse`
  - `WSHandler.ServeHTTP` — read loop; responds to any message (default action `ping`) with Mongo-backed check
- **Dependencies:** `formatter`, `constant`, `websocket`, generated healthcheck proto
- **Used by:** `transport/transport.go`
- **Patterns:** Shared usecase across transports; HTTP maps DOWN → 503
- **Notes:** gRPC healthcheck domain service is separate from `grpc.health.v1` server registered in `app/grpc/server.go`.

---

### `transport/transport.go`

- **Purpose:** Route registration for healthcheck on HTTP, gRPC, and WebSocket.
- **Layer / role:** Transport registration.
- **Key types / functions:**
  - `RegisterHTTP` — `GET /health`, `GET /healthz`
  - `RegisterGRPC` — `healthv1.RegisterHealthcheckServiceServer`
  - `RegisterWebSocket` — `GET /ws/health`
- **Dependencies:** `handler`, generated proto, chi, grpc
- **Used by:** `router/router.go`, `app/grpc/server.go`
- **Patterns:** Single transport file per domain
- **Notes:** Full HTTP paths: `/api/v1/health`, `/api/v1/healthz`. WS: `/api/v1/ws/health`.

---

## Tests (grouped)

| File | Coverage |
|------|----------|
| `usecase/usecase_test.go` | Mock checkers; HTTP/gRPC/WS report assembly; DOWN propagation |

Run: `make test-unit`.
