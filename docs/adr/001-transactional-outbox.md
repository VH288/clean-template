# ADR 001: Transactional Outbox for Kafka Publish

## Status

Accepted

## Context

Sample CRUD needs to publish domain events to Kafka (`sample.created`, `sample.updated`, `sample.deleted`). Publishing directly from the usecase after a Postgres write is unreliable: if the process crashes between the DB commit and the Kafka publish, events are lost. Conversely, publishing before commit can emit events for transactions that roll back.

This template is a learning scaffold. We want patterns that resemble production systems without full distributed-transaction complexity.

## Decision

### Kafka publish: transactional outbox

1. Postgres repository methods `CreateWithEvent`, `UpdateWithEvent`, and `DeleteWithEvent` write the sample row and an `outbox_events` row in a **single transaction**.
2. A background **outbox relay** (worker goroutine) polls pending rows (`FOR UPDATE SKIP LOCKED`), publishes to Kafka, and marks rows `published`.
3. Failed publishes increment `retry_count`; after `max_retries` (default 5) the row is marked `failed`.
4. The usecase builds a shared `event.Envelope` JSON payload; it no longer depends on `EventPublisher`.

Message key = `aggregate_id` (sample UUID) for partition affinity.

### Mongo: direct write from usecase (learning only)

Mongo upsert/delete runs **after** the Postgres transaction succeeds. This demonstrates multi-store writes but is **not** transactional with Postgres. Failures are logged and metered; the API still returns success based on Postgres.

Mongo is **not** updated by the Kafka consumer in this scaffold.

### Kafka consume: structured handler as extension point

The worker runs a second goroutine: a Kafka consumer subscribed via `group_id` + `topic`. `EventHandler.Handle` unmarshals `event.Envelope`, dispatches by type, logs, and records metrics. Unknown event types are skipped (commit offset). Invalid JSON returns an error (no commit → retry).

Replace the handler body with real side effects (email, analytics, projections) in your own services.

## Consequences

**Positive**

- Reliable publish path aligned with common production practice
- Clear separation: outbox relay = publish, consumer handler = subscribe/consume
- Shared `event.Envelope` contract between producer and consumer

**Negative / trade-offs**

- Eventual consistency between Postgres and Kafka (relay poll interval, default 1s)
- Duplicate publishes possible on relay retry; consumers should be idempotent
- No DLQ in this scaffold (future enhancement)
- Mongo + Postgres can diverge on Mongo failure

## Future enhancements

- Dead-letter queue for `failed` outbox rows and poison Kafka messages
- Idempotency store in consumer
- Mongo projection via consumer (out of scope for this template)
