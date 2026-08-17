# Shared Packages (`internal/pkg/`)

Cross-cutting utilities used by domain, app, and infrastructure layers. Safe for domain to import (no infrastructure dependencies).

---

### `errors/errors.go`

- **Purpose:** Application error type with codes, HTTP status mapping, and wrap helpers.
- **Layer / role:** Shared error model.
- **Key types / functions:**
  - `AppError` — `Code`, `Message`, `HTTPStatus`, optional `Details`, wrapped `Err`
  - `Code` — `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`, `INTERNAL_ERROR`, etc.
  - Predefined: `ErrNotFound`, `ErrConflict`, `ErrValidation`, `ErrUnauthorized`, ...
  - `New`, `Wrap`, `WithDetails`, `WithMessage`, `As`, `Is`
  - `httpStatusFromCode` — maps code to HTTP status
- **Dependencies:** `net/http`, std `errors`
- **Used by:** Repositories, usecases, handlers, `formatter.Fail`
- **Patterns:** Typed application errors; gRPC handler maps via `As`
- **Notes:** `Wrap` preserves underlying error for logging.

---

### `errors/errors_test.go`

- **Purpose:** Unit tests for error construction, `As`, HTTP status mapping.
- **Layer / role:** Test.
- **Notes:** Run with `make test-unit`.

---

### `formatter/response.go`

- **Purpose:** Standard JSON API response envelope for HTTP success and failure.
- **Layer / role:** HTTP response formatter.
- **Key types / functions:**
  - `Response` — `success`, `message`, `data`, `meta`, `error`
  - `Meta` — pagination: `page`, `per_page`, `total`, `total_pages`
  - `ErrorBody` — `code`, `message`, `details`
  - `Success`, `SuccessWithMeta`, `Fail`, `JSON`
- **Dependencies:** `pkg/errors`
- **Used by:** Sample and healthcheck HTTP handlers
- **Patterns:** Uniform JSON API shape
- **Notes:** `Fail` uses `AppError.HTTPStatus` when error is typed.

---

### `validator/validator.go`

- **Purpose:** go-playground validator setup with field-level detail mapping to `ErrValidation`.
- **Layer / role:** Input validation helper.
- **Key types / functions:**
  - `Struct(s)` — validate struct tags; returns `WithDetails(ErrValidation, ...)`
- **Dependencies:** `go-playground/validator/v10`, `pkg/errors`
- **Used by:** Sample HTTP and WS handlers
- **Patterns:** Central validator instance
- **Notes:** Field names lowercased in details map.

---

### `utils/pagination.go`

- **Purpose:** Page/perPage normalization and total pages calculation.
- **Layer / role:** Shared utility.
- **Key types / functions:**
  - `NormalizePagination(page, perPage)` — defaults page≥1, perPage 10 (max 100), returns offset
  - `TotalPages(total, perPage)`
- **Dependencies:** None
- **Used by:** Sample usecase, HTTP handler list
- **Patterns:** Consistent pagination defaults
- **Notes:** Default perPage 10 when missing or invalid.

---

### `helper/context.go`

- **Purpose:** Request ID storage in `context.Context`.
- **Layer / role:** Context helper.
- **Key types / functions:**
  - `RequestIDFromContext(ctx)`, `WithRequestID(ctx, id)`
- **Dependencies:** `pkg/constant`
- **Used by:** Middleware, `testutil`
- **Patterns:** Context value for request correlation
- **Notes:** Key: `constant.ContextKeyRequestID`.

---

### `pointer/pointer.go`

- **Purpose:** Generic pointer helpers for optional values.
- **Layer / role:** Utility.
- **Key types / functions:**
  - `Of[T](v)` — `*T` from value
  - `Deref[T](v)` — safe dereference with zero value if nil
- **Dependencies:** None
- **Used by:** Available for handlers/adapters (minimal use in template)
- **Patterns:** Generic helper
- **Notes:** Small utility package for API boundary conversions.

---

### `constant/constant.go`

- **Purpose:** Shared string constants for headers, context keys, cache, Kafka topic, health status.
- **Layer / role:** Constants.
- **Key symbols:**
  - `HeaderRequestID`, `HeaderTraceID`
  - `ContextKeyRequestID`, `ContextKeyLogger`
  - `SampleCacheKeyPrefix`, `SampleCacheTTLSec`
  - `KafkaTopicSampleEvents`
  - `HealthStatusUP`, `HealthStatusDOWN`
- **Dependencies:** None
- **Used by:** Middleware, logger, Redis repo, healthcheck usecase
- **Patterns:** Single constants package
- **Notes:** Kafka topic in config may differ; constant documents canonical name.

---

### `testutil/testutil.go`

- **Purpose:** HTTP test helpers and context utilities for handler tests.
- **Layer / role:** Test support.
- **Key types / functions:**
  - `NewContext()` — background + random request ID
  - `NewContextWithTimeout(t, d)`
  - `PerformRequest(t, handler, method, path, body)` — httptest recorder
  - `DecodeJSON(t, rr, dest)`, `AssertStatus(t, rr, want)`
- **Dependencies:** `testify/require`, `helper`, `uuid`
- **Used by:** `handler/http_test.go`, domain/infrastructure tests
- **Patterns:** Chi handler integration-style tests without full server
- **Notes:** Sets `Content-Type: application/json` when body provided.
