# Database Migrations (`migrations/`)

Goose SQL migrations for Postgres. Applied via `cmd/migrate` or Makefile targets.

---

### `00001_create_samples.sql`

- **Purpose:** Creates the `samples` table — Postgres source of truth for sample entities.
- **Schema:**
  - `id` UUID PRIMARY KEY
  - `name` VARCHAR(100) NOT NULL
  - `description` TEXT NOT NULL DEFAULT `''`
  - `status` VARCHAR(20) NOT NULL DEFAULT `'active'`
  - `created_at`, `updated_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
- **Indexes:**
  - `idx_samples_status` on `status`
  - `idx_samples_created_at` on `created_at DESC`
- **Up:** `CREATE TABLE` + indexes
- **Down:** Drop indexes and table

---

### `00002_outbox_events.sql`

- **Purpose:** Transactional outbox table for reliable Kafka publish.
- **Schema:**
  - `id` UUID PRIMARY KEY (same as `event_id` in envelope)
  - `aggregate_type` VARCHAR(50) NOT NULL
  - `aggregate_id` UUID NOT NULL
  - `event_type` VARCHAR(100) NOT NULL
  - `payload` JSONB NOT NULL
  - `status` VARCHAR(20) NOT NULL DEFAULT `'pending'` (`pending` | `processing` | `published` | `failed`)
  - `retry_count` INT NOT NULL DEFAULT 0
  - `created_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
  - `published_at` TIMESTAMPTZ nullable
- **Indexes:**
  - Partial `idx_outbox_pending` on `(status, created_at)` WHERE `status = 'pending'` — optimizes relay poll
- **Up:** `CREATE TABLE` + partial index
- **Down:** Drop index and table
- **Notes:** See [ADR 001](../adr/001-transactional-outbox.md) and [outbox relay](../layers/infrastructure.md#internalinfrastructureoutboxrelaygo).

---

### `00003_version_and_outbox_claimed.sql`

- **Purpose:** Optimistic locking on samples; outbox claim tracking for two-phase relay.
- **Schema changes:**
  - `samples.version` INT NOT NULL DEFAULT 1
  - `outbox_events.claimed_at` TIMESTAMPTZ nullable
- **Indexes:**
  - Partial `idx_outbox_processing` on `(status, claimed_at)` WHERE `status = 'processing'`
- **Up:** `ALTER TABLE` + index
- **Down:** Drop index and columns
- **Notes:** `version` used in `UPDATE … WHERE id = $1 AND version = $expected`. Stale `processing` rows reclaimed when `claimed_at` older than `outbox.processing_stale_after`.

---

### `00004_processed_events.sql`

- **Purpose:** Idempotency store for Kafka consumer at-least-once delivery.
- **Schema:**
  - `event_id` UUID PRIMARY KEY
  - `processed_at` TIMESTAMPTZ NOT NULL DEFAULT NOW()
- **Indexes:**
  - `idx_processed_events_processed_at` on `processed_at`
- **Up:** `CREATE TABLE` + index
- **Down:** Drop index and table
- **Notes:** Consumer checks before dispatch; `ON CONFLICT DO NOTHING` on insert.

---

## Goose workflow

| Command | Makefile target | Description |
|---------|-----------------|-------------|
| `go run ./cmd/migrate -dir migrations up` | `make migrate-up` | Apply all pending migrations |
| `go run ./cmd/migrate -dir migrations down` | `make migrate-down` | Roll back one migration |
| `go run ./cmd/migrate -dir migrations status` | `make migrate-status` | Show migration versions |
| `reset` | — | Roll back all (use with care) |

Configuration: Postgres DSN from `configs/config.yaml` via `internal/config`. Ensure Postgres is running before `migrate-up`.

Docker image copies `migrations/` to `/app/migrations` for containerized migrate runs.

---

## Related docs

- [cmd/migrate](../layers/cmd.md#cmdmigratemaingo)
- [Postgres repository](../layers/infrastructure.md#internalinfrastructurepersistencesamplepostgresgo)
- [Consumer idempotency](../layers/infrastructure.md#internalinfrastructuremessagingsampleidempotencygo)
