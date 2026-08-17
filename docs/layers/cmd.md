# Command Entrypoints (`cmd/`)

Four binaries: combined API server, gRPC-only server, background worker, and database migrations.

---

### `cmd/api/main.go`

- **Purpose:** Starts HTTP and gRPC servers concurrently; primary development entrypoint.
- **Layer / role:** Application entrypoint (composition root consumer).
- **Key types / functions:**
  - `main()` — exits with code 1 on fatal error
  - `run()` — loads config, builds container, starts servers, registers shutdown hooks, waits for lifecycle
- **Dependencies:** `internal/config`, `internal/app/bootstrap`, `internal/app/server`, `internal/app/grpc`
- **Used by:** `make run-api`, Docker default `ENTRYPOINT`
- **Patterns:** Dual-server goroutine with error channel; shutdown hooks registered after `Build()` so HTTP/gRPC stop before infra cleanup
- **Notes:** On server error, cancels root context to trigger shutdown. HTTP listens on `cfg.HTTP.Port`; gRPC on `cfg.GRPC.Port`.

---

### `cmd/grpc/main.go`

- **Purpose:** gRPC-only server without HTTP router or HTTP shutdown.
- **Layer / role:** Application entrypoint for gRPC-only deployments.
- **Key types / functions:**
  - `run()` — same bootstrap as API minus HTTP server
- **Dependencies:** `internal/config`, `internal/app/bootstrap`, `internal/app/server`, `internal/app/grpc`
- **Used by:** `make run-grpc`
- **Patterns:** Single transport entrypoint
- **Notes:** Registers only gRPC graceful-stop hook on lifecycle; infra shutdown hooks still run via `wireLifecycle`.

---

### `cmd/worker/main.go`

- **Purpose:** Runs outbox relay and Kafka consumer background workers.
- **Layer / role:** Worker process entrypoint.
- **Key types / functions:**
  - `run()` — builds container, wires worker via `bootstrap.WireWorker`, starts relay and consumer
- **Dependencies:** `internal/config`, `internal/app/bootstrap`
- **Used by:** `make run-worker`
- **Patterns:** Background worker process separate from API
- **Notes:** Registers `worker.Consumer.Close` on lifecycle (runs first on shutdown, LIFO). Does not start HTTP/gRPC servers.

---

### `cmd/migrate/main.go`

- **Purpose:** Goose CLI wrapper for SQL migrations against Postgres.
- **Layer / role:** Database migration tool entrypoint.
- **Key types / functions:**
  - `main()` — parses flags and subcommand
  - `fatal()` — prints error and exits 1
- **Dependencies:** `internal/config` (for Postgres DSN), `github.com/pressly/goose/v3`, `github.com/lib/pq`
- **Used by:** `make migrate-up`, `make migrate-down`, `make migrate-status`
- **Patterns:** Thin CLI over Goose
- **Notes:** Flag `-dir` defaults to `migrations`. Commands: `up`, `down`, `status`, `reset`. Unknown commands exit with error. See [migrations.md](migrations.md).
