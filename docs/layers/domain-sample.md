# Sample Domain (`internal/domain/sample/`)

Example CRUD domain with HTTP, gRPC, and WebSocket transports; Postgres source of truth, Redis cache-aside, Mongo direct write, transactional outbox events, and optimistic locking.

---

### `interfaces.go`

- **Purpose:** Domain ports (repository, cache, document, usecase interfaces).
- **Layer / role:** Domain ports (dependency inversion).
- **Key types / functions:**
  - `Repository` — CRUD + `CreateWithEvent`, `UpdateWithEvent`, `DeleteWithEvent` (with `eventID` param) for outbox TX
  - `CacheRepository` — `Set`, `Get`, `Delete`
  - `DocumentRepository` — `Upsert`, `GetByID`, `Delete`
  - `Usecase` — `Create`, `GetByID`, `List`, `Update`, `Delete`
- **Dependencies:** `entity.Sample` only
- **Used by:** Usecase, handlers, bootstrap wiring, persistence adapters
- **Patterns:** Port/adapter; segregated interfaces for cache and document stores
- **Notes:** `*WithEvent` methods write sample + outbox row in one Postgres transaction (implemented in infrastructure). `eventID` is generated in usecase and stored in both outbox row and envelope JSON.

---

### `entity/sample.go`

- **Purpose:** Core domain entity and status constants.
- **Layer / role:** Domain entity.
- **Key types / functions:**
  - `Sample` — `ID`, `Name`, `Description`, `Status`, `Version`, `CreatedAt`, `UpdatedAt`
  - `StatusActive`, `StatusInactive` — allowed status values
- **Dependencies:** None
- **Used by:** Usecase, mappers, handlers, events, persistence models
- **Patterns:** Plain struct entity; no ORM tags
- **Notes:** ID, version, and timestamps are assigned/updated by `PostgresRepository` on create/update.

---

### `event/event.go`

- **Purpose:** Domain event type constants and shared Kafka/outbox JSON envelope.
- **Layer / role:** Domain events contract.
- **Key types / functions:**
  - `TypeCreated`, `TypeUpdated`, `TypeDeleted` — event type strings
  - `Envelope` — `{ event_id, type, sample }` JSON shape for publish and consume
- **Dependencies:** `entity.Sample`
- **Used by:** Usecase (`marshalEvent`), Postgres outbox payload, consumer handler, outbox relay
- **Patterns:** Shared envelope contract between producer and consumer; `event_id` enables idempotency
- **Notes:** See [ADR 001](../adr/001-transactional-outbox.md).

---

### `usecase/usecase.go`

- **Purpose:** Sample business logic: CRUD, cache-aside reads, multi-store writes, outbox event payloads.
- **Layer / role:** Domain usecase.
- **Key types / functions:**
  - `SampleUsecase` — implements `sample.Usecase`
  - `New(repo, cache, document, log, trace, metrics)` — constructor with domain ports
  - `Create`, `GetByID`, `List`, `Update`, `Delete` — CRUD with telemetry via ports
  - `marshalEvent`, `invalidateCache` — private helpers
- **Dependencies:** Port interfaces, `domain/ports` (Tracer, Logger, SampleMetrics), `pkg/errors`, `pkg/utils`, `google/uuid`
- **Used by:** HTTP, gRPC, WS handlers via `sample.Usecase` interface
- **Patterns:** Cache-aside read; write-through invalidate; transactional outbox via `*WithEvent` repo methods; best-effort Mongo after Postgres commit; optimistic locking on update
- **Notes:** Default status `active` on create if empty. Mongo failures log warn + `mongo_sync_failures_total` but do not fail the API response. Cache get/set errors logged; non-`NotFound` cache errors fall through to Postgres. Update returns `ErrConflict` on stale version. Does not publish to Kafka directly.

---

### `dto/request.go`

- **Purpose:** HTTP/WS input shapes with validation tags.
- **Layer / role:** Domain DTO (input).
- **Key types / functions:**
  - `CreateSampleRequest` — `name` (required, 2–100), `description` (max 500), `status` (optional, active/inactive)
  - `UpdateSampleRequest` — same fields; `status` required on update
  - `ListSampleQuery` — `page`, `per_page`
- **Dependencies:** None
- **Used by:** HTTP handler, WS handler
- **Patterns:** Validation tags for `go-playground/validator`
- **Notes:** gRPC uses proto messages instead of these DTOs.

---

### `dto/response.go`

- **Purpose:** HTTP/WS output shapes for sample data.
- **Layer / role:** Domain DTO (output).
- **Key types / functions:**
  - `SampleResponse` — entity fields as JSON
  - `ListSampleResponse` — `items` array wrapper (used indirectly via mapper lists)
- **Dependencies:** None
- **Used by:** Mapper, HTTP/WS responses
- **Patterns:** API response DTOs separate from entity
- **Notes:** Timestamps serialized as RFC3339 JSON times.

---

### `mapper/mapper.go`

- **Purpose:** Converts `entity.Sample` to response DTOs.
- **Layer / role:** Domain mapper.
- **Key types / functions:**
  - `ToSampleResponse(s)` — single entity → DTO
  - `ToSampleResponses(items)` — slice conversion
- **Dependencies:** `entity`, `dto`
- **Used by:** HTTP handler, WS handler
- **Patterns:** Entity ↔ DTO mapping (gRPC handler maps to proto inline)
- **Notes:** No reverse mapping (handlers pass primitives to usecase).

---

### `handler/http.go`

- **Purpose:** REST handlers for sample CRUD with JSON decode, validation, and standard response envelope.
- **Layer / role:** HTTP adapter (handler).
- **Key types / functions:**
  - `HTTPHandler` — holds `sample.Usecase`
  - `Create`, `GetByID`, `List`, `Update`, `Delete` — chi handler methods
- **Dependencies:** `dto`, `mapper`, `formatter`, `validator`, `utils`, chi
- **Used by:** `transport/http.go` route registration
- **Patterns:** Thin handler → usecase; errors via `formatter.Fail`
- **Notes:** List uses query params `page`, `per_page`. Create returns 201. Protected by API key when configured.

---

### `handler/grpc.go`

- **Purpose:** gRPC service implementation for `sample.v1.SampleService`.
- **Layer / role:** gRPC adapter (handler).
- **Key types / functions:**
  - `GRPCHandler` — embeds `UnimplementedSampleServiceServer`; holds `ports.GRPCMetrics`
  - `CreateSample`, `GetSample`, `ListSamples`, `UpdateSample`, `DeleteSample`
  - `toProtoSample`, `toGRPCError` — mapping and error translation
- **Dependencies:** `sample.Usecase`, `domain/ports`, `entity`, `pkg/errors`, generated proto types
- **Used by:** `transport/grpc.go`
- **Patterns:** AppError → gRPC status code mapping; metrics via injected port
- **Notes:** `NOT_FOUND` → `codes.NotFound`, `CONFLICT` → `codes.AlreadyExists`, etc.

---

### `handler/websocket.go`

- **Purpose:** WebSocket message loop with action-based dispatch for sample CRUD.
- **Layer / role:** WebSocket adapter (handler).
- **Key types / functions:**
  - `WSHandler` — usecase + `ports.WebSocketHub`
  - `ServeHTTP` — upgrade, read loop, dispatch, JSON responses
  - `dispatch(ctx, env)` — handles `create`, `get`, `list`, `update`, `delete` actions
  - `wsEnvelope` — `{ action, payload }` wire format
- **Dependencies:** `dto`, `mapper`, `validator`, `domain/ports`
- **Used by:** `transport/websocket.go`
- **Patterns:** Action dispatch pattern over WebSocket JSON
- **Notes:** Invalid JSON returns error object on socket; does not close connection. Unknown action returns error message. Protected by API key when configured.

---

### `transport/http.go`

- **Purpose:** Registers chi routes under `/samples` for REST CRUD.
- **Layer / role:** HTTP transport registration.
- **Key types / functions:**
  - `RegisterHTTP(r, h)` — `GET/POST /`, `GET/PUT/DELETE /{id}`
- **Dependencies:** `handler.HTTPHandler`, chi
- **Used by:** `router/router.go` under `/api/v1` (protected group)
- **Patterns:** Transport layer only registers routes
- **Notes:** Full paths: `/api/v1/samples`, `/api/v1/samples/{id}`.

---

### `transport/grpc.go`

- **Purpose:** Registers gRPC `SampleService` on a server.
- **Layer / role:** gRPC transport registration.
- **Key types / functions:**
  - `RegisterGRPC(server, h)` — `samplev1.RegisterSampleServiceServer`
- **Dependencies:** Generated proto registration, `handler.GRPCHandler`
- **Used by:** `internal/app/grpc/server.go`
- **Patterns:** Thin registration wrapper
- **Notes:** Service full name: `sample.v1.SampleService`.

---

### `transport/websocket.go`

- **Purpose:** Registers WebSocket route for sample actions.
- **Layer / role:** WebSocket transport registration.
- **Key types / functions:**
  - `RegisterWebSocket(r, h)` — `GET /ws/samples`
- **Dependencies:** `handler.WSHandler`, chi
- **Used by:** `router/router.go` (protected group)
- **Patterns:** Transport registration
- **Notes:** Full path: `/api/v1/ws/samples`.

---

## Domain ports (`internal/domain/ports/`)

Cross-cutting interfaces used by domain code without importing infrastructure:

| File | Interfaces |
|------|------------|
| `telemetry.go` | `Tracer`, `Logger`, `SampleMetrics`, `HealthMetrics`, `GRPCMetrics` |
| `websocket.go` | `WebSocketConn`, `WebSocketHub` |

Implemented by `internal/infrastructure/telemetry` and `internal/infrastructure/websocket/ports_adapter.go`.

---

## Tests (grouped)

| File | Coverage |
|------|----------|
| `usecase/usecase_test.go` | Unit tests with mock/in-memory repositories; CRUD, cache behavior, event envelope, Mongo failure tolerance |
| `handler/http_test.go` | HTTP handler tests via chi router + `testutil.PerformRequest`; status codes and JSON bodies |

Run unit tests: `make test-unit` (includes domain packages).

---

## Related infrastructure

- [Postgres repository](../layers/infrastructure.md#internalinfrastructurepersistencesamplepostgresgo) — transactional outbox writes + optimistic locking
- [Redis cache](../layers/infrastructure.md#internalinfrastructurepersistencesampleredisgo) — cache-aside adapter
- [Mongo repository](../layers/infrastructure.md#internalinfrastructurepersistencesamplemongo.go) — document store
- [Outbox relay](../layers/infrastructure.md#internalinfrastructureoutboxrelaygo) — two-phase publish to Kafka
- [Consumer handler](../layers/infrastructure.md#internalinfrastructuremessagingsampleconsumer_handlergo) — idempotent `Envelope` processing
