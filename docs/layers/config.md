# Configuration (`internal/config` and `configs/`)

---

### `internal/config/config.go`

- **Purpose:** Viper-backed configuration loading with defaults and environment variable overrides.
- **Layer / role:** Application configuration.
- **Key types / functions:**
  - `Config` — root struct: `App`, `HTTP`, `GRPC`, `Postgres`, `Redis`, `Mongo`, `Kafka`, `Outbox`, `Observability`, `External`
  - `PostgresConfig.DSN()` — Postgres connection string for sqlx/lib/pq
  - `Load(path)` — read YAML, apply defaults, bind env, unmarshal
  - `LoadDefault()` — tries `configs/config.yaml`, `config.yaml`, `.env`
  - `setDefaults(v)` — default values for all keys
  - `bindEnv(v)` — maps common env vars (`PORT`, `DB_HOST`, `REDIS_ADDR`, etc.)
- **Dependencies:** `github.com/spf13/viper`
- **Used by:** All `cmd/*` entrypoints, bootstrap
- **Patterns:** Viper file + env overlay; missing config file allowed when env/defaults suffice
- **Notes:** Env keys use `_` replacer for nested fields (e.g. `HTTP_PORT`). `app.secret` bound to `APP_SECRET`. Outbox defaults: poll 1s, batch 50, max retries 5.

#### Config struct reference

| Section | Key fields | Defaults (highlights) |
|---------|------------|----------------------|
| `app` | `name`, `env`, `secret`, `version` | `clean-template`, `development`, `0.1.0` |
| `http` | `port`, `read_timeout`, `write_timeout`, `shutdown_timeout` | `8081`, `15s`, `15s`, `10s` |
| `grpc` | `port`, `shutdown_timeout` | `7000`, `10s` |
| `postgres` | host, port, user, password, dbname, pool settings | `127.0.0.1:5432`, db `sample` |
| `redis` | `addr`, `password`, `db` | `127.0.0.1:6379` |
| `mongo` | `uri`, `database` | local Mongo, db `sample` |
| `kafka` | `brokers`, `group_id`, `topic` | `sample.events`, group `clean-template` |
| `outbox` | `poll_interval`, `batch_size`, `max_retries` | `1s`, `50`, `5` |
| `observability` | log level, Loki URL, Tempo endpoint, metrics path, service name, trace sample ratio | `/metrics`, ratio `1.0` |
| `external` | `http_url`, `grpc_addr` | jsonplaceholder, `127.0.0.1:7001` |

---

### `configs/config.yaml.example`

- **Purpose:** Committed example configuration for local development; copy to `configs/config.yaml`.
- **Layer / role:** Configuration template.
- **Key fields:**
  - `app.secret` — placeholder `CHANGE_ME` (never commit real secrets)
  - `postgres.password` — placeholder `CHANGE_ME`
  - `http.port` / `grpc.port` — API listen ports
  - `kafka.brokers`, `group_id`, `topic` — worker consumer and outbox publish target
  - `observability.*` — logging, Tempo OTLP gRPC endpoint, Prometheus metrics path
  - `external.*` — optional upstream HTTP API and gRPC health probe target
- **Dependencies:** None (YAML only)
- **Used by:** Developers copy to `configs/config.yaml` for local runs
- **Patterns:** Example-only; secrets as placeholders
- **Notes:** Does not include `outbox` section — those values come from `setDefaults` in `config.go`. Add explicit `outbox:` block in your local copy to override defaults.

---

### `configs/config.yaml` (local runtime copy)

- **Purpose:** Active configuration file used at runtime (`config.Load("configs/config.yaml")`).
- **Layer / role:** Local/runtime config (not committed).
- **Notes:** Excluded by `.gitignore` (`*.yaml` pattern with `!.yaml.example` exception). Create from `config.yaml.example`. Never commit credentials.
