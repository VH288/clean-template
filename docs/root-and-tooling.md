# Root Files and Tooling

Makefile, Docker image, code generation script, module definition, and git ignore rules.

---

## `Makefile`

- **Purpose:** Developer commands for run, test, migrate, proto generation, and build.
- **Key targets:**
  - `help` — list targets
  - `run` / `run-api` — `go run ./cmd/api` (HTTP + gRPC)
  - `run-grpc` — gRPC-only server
  - `run-worker` — outbox relay + Kafka consumer
  - `migrate-up`, `migrate-down`, `migrate-status` — Goose via `cmd/migrate`
  - `proto` — regenerate protobuf stubs
  - `test` — all packages `go test ./...`
  - `test-unit` — domain, pkg, infrastructure with `-short`
  - `test-integration` — `-tags=integration`, `Integration` test name filter
  - `tidy` — `go mod tidy`
  - `build` — binaries to `bin/` (api, grpc, worker, migrate)
  - `lint` — `go vet ./...`
  - `docker-up`, `docker-down` — referenced in help (compose in `deployments/` when present)
- **Variables:** `APP_NAME`, `CONFIG`, `GO`, `PROTOC`
- **Notes:** Integration tests expect `DATABASE_URL` for Postgres. Proto target requires `protoc` and Go grpc plugins.

---

## `Dockerfile`

- **Purpose:** Multi-stage production image for API and worker binaries.
- **Stages:**
  - `builder` — `golang:1.26-alpine`; `go mod download`; builds `api`, `worker`, `migrate` with `CGO_ENABLED=0`, stripped binaries
  - `runtime` — `alpine:3.20`; copies binaries, `configs/`, `migrations/`
- **Exposed ports:** `8081` (HTTP), `7000` (gRPC)
- **User:** `nobody`
- **Default entrypoint:** `/app/api`
- **Notes:** Does not build `grpc`-only binary. Includes `curl` and CA certificates. Worker/migrate available at `/app/worker`, `/app/migrate`.

---

## `scripts/generate.sh`

- **Purpose:** Wrapper that runs `make proto` from repo root.
- **Usage:** `./scripts/generate.sh`
- **Behavior:** `set -euo pipefail`; resolves repo root; echoes progress
- **Notes:** Single entry for codegen; extend here if adding mock generation later.

---

## `go.mod`

- **Purpose:** Go module definition and dependency lock for `clean-template`.
- **Module path:** `clean-template`
- **Go version:** `1.26.4`
- **Key direct dependencies and why:**
  | Dependency | Role |
  |------------|------|
  | `go-chi/chi` | HTTP router |
  | `jmoiron/sqlx` + `lib/pq` | Postgres access |
  | `redis/go-redis` | Redis cache |
  | `mongo-driver` | MongoDB documents |
  | `segmentio/kafka-go` | Kafka producer/consumer |
  | `google.golang.org/grpc` + `protobuf` | gRPC services |
  | `coder/websocket` | WebSocket server |
  | `spf13/viper` | Configuration |
  | `pressly/goose` | Migrations |
  | `prometheus/client_golang` | Metrics |
  | `go.opentelemetry.io/otel` (+ OTLP exporter) | Tracing to Tempo |
  | `go-playground/validator` | Request validation |
  | `stretchr/testify` | Tests |

---

## `.gitignore`

- **Purpose:** Excludes local secrets, build artifacts, and editor config from version control.
- **Excluded patterns:**
  - `.env`, `.env.*` (except `.env.example` if added)
  - `.cursor`
  - `bin/`, `*.exe`, shared libraries
  - `coverage.out`, `*.test`
  - `vendor/`
  - `*.yaml` except `!.yaml.example` — local `configs/config.yaml` is not committed
- **Notes:** Commit `configs/config.yaml.example` only; copy locally for runtime config.

---

## Related docs

- [Configuration](../layers/config.md)
- [Protobuf generation](../layers/proto.md)
- [Migrations](../layers/migrations.md)
- [Command entrypoints](../layers/cmd.md)
