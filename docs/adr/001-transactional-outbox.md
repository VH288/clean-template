# ADR 001: Transactional Outbox for Kafka Publish

## Status

Accepted (updated)

## Context

Sample CRUD needs to publish domain events to Kafka (`sample.created`, `sample.updated`, `sample.deleted`). Publishing directly from the usecase after a Postgres write is unreliable: if the process crashes between the DB commit and the Kafka publish, events are lost. Conversely, publishing before commit can emit events for transactions that roll back.

This template is a learning scaffold. We want patterns that resemble production systems without full distributed-transaction complexity.

## Decision

### Kafka publish: transactional outbox (two-phase relay)

1. Postgres repository methods `CreateWithEvent`, `UpdateWithEvent`, and `DeleteWithEvent` write the sample row and an `outbox_events` row in a **single transaction**. Each event has a stable `event_id` (UUID) embedded in the `event.Envelope` JSON payload.
2. A background **outbox relay** (worker goroutine) uses a **two-phase** pattern:
   - **Phase 1 (TX):** `SELECT … FOR UPDATE SKIP LOCKED` pending rows → set `status = processing`, `claimed_at = NOW()` → commit
   - **Phase 2 (outside TX):** publish to Kafka → `UPDATE status = published` or increment retry on failure
3. Stale `processing` rows (worker crash) are reset to `pending` after `outbox.processing_stale_after` (default 5m).
4. Failed publishes increment `retry_count`; after `max_retries` (default 5) the row is marked `failed`.
5. `cmd/outbox-replay` resets `failed` rows to `pending` for manual recovery.
6. Prometheus gauges: `outbox_events_pending`, `outbox_events_failed`.

Message key = `aggregate_id` (sample UUID) for partition affinity.

### Mongo: direct write from usecase (learning only)

Mongo upsert/delete runs **after** the Postgres transaction succeeds. This demonstrates multi-store writes but is **not** transactional with Postgres. Failures are **logged as warnings** and increment `mongo_sync_failures_total`; the API still returns success based on Postgres.

Mongo is **not** updated by the Kafka consumer in this scaffold.

### Kafka consume: idempotent handler with DLQ

The worker runs a second goroutine: a Kafka consumer subscribed via `group_id` + `topic`.

- `EventHandler.Handle` unmarshals `event.Envelope` (requires `event_id`), checks `processed_events` for idempotency, dispatches by type, records metrics, marks processed.
- Unknown event types are skipped (commit offset).
- Poison messages (invalid JSON, missing `event_id`) are published to DLQ topic (`kafka.dlq_topic`, default `sample.events.dlq`) and committed.
- Transient handler errors retry without commit; after `kafka.max_handler_retries` the offset is committed to avoid infinite loops.
- Fetch errors use exponential backoff.

Replace the handler body with real side effects (email, analytics, projections) in your own services.

### Optimistic locking

Sample updates use a `version` column. Concurrent updates with a stale version return `409 Conflict` (`ErrConflict`).

## Consequences

**Positive**

- Reliable publish path aligned with common production practice
- Two-phase relay avoids holding DB locks during Kafka publish
- Idempotent consumer safe under at-least-once delivery
- DLQ and outbox replay for operability
- Clear separation: outbox relay = publish, consumer handler = subscribe/consume
- Shared `event.Envelope` contract between producer and consumer

**Negative / trade-offs**

- Eventual consistency between Postgres and Kafka (relay poll interval, default 1s)
- Duplicate publishes still possible if publish succeeds but mark-published fails; mitigated by consumer idempotency on `event_id`
- Mongo + Postgres can diverge on Mongo failure (by design in this scaffold)

## Implemented enhancements (formerly future)

- Dead-letter queue for poison Kafka messages (`sample.events.dlq`)
- Idempotency store in consumer (`processed_events` table)
- Failed outbox replay (`cmd/outbox-replay`)
- Two-phase outbox relay with stale processing recovery

## Future enhancements

- Mongo projection via consumer (out of scope for this template)
- Automated alerting on `outbox_events_failed` gauge
- gRPC/API admin endpoint for outbox replay
