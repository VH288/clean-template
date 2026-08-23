# Microservice SDK & Architecture Guide

A comprehensive architectural analysis of the **clean-template** codebase: understanding Clean Architecture layers, evaluating what components to extract into a reusable internal SDK across microservices, and identifying what must remain local to each service.

---

## Table of Contents

1. [Understanding the Current Architecture](#1-understanding-the-current-architecture)
2. [SDK vs. Microservice Decision Framework](#2-sdk-vs-microservice-decision-framework)
3. [What to Extract into Shared SDKs (With File References)](#3-what-to-extract-into-shared-sdks-with-file-references)
   - [A. Observability & Telemetry (Logger, APM, Metrics)](#a-observability--telemetry-logger-apm-metrics)
   - [B. Standard Errors & API Response Envelopes](#b-standard-errors--api-response-envelopes)
   - [C. Transport Middlewares & Context Helpers](#c-transport-middlewares--context-helpers)
   - [D. Eventing, Transactional Outbox & Consumer Idempotency](#d-eventing-transactional-outbox--consumer-idempotency)
   - [E. Resilient Inter-Service Clients](#e-resilient-inter-service-clients)
   - [F. API Contracts (Protobuf / gRPC)](#f-api-contracts-protobuf--grpc)
4. [What Must ALWAYS Stay Inside Each Microservice](#4-what-must-always-stay-inside-each-microservice)
5. [Recommended Modular SDK Structure](#5-recommended-modular-sdk-structure)
6. [Summary Checklist](#6-summary-checklist)

---

## 1. Understanding the Current Architecture

This project follows **Clean Architecture** (Ports & Adapters / Hexagonal Architecture):

```
                   ┌────────────────────────┐
                   │  cmd/ (Entrypoints)    │
                   └───────────┬────────────┘
                               │
                   ┌───────────▼────────────┐
                   │ internal/app/          │  (Bootstrap / Composition Root, Router, Middlewares)
                   └───────────┬────────────┘
                               │
                   ┌───────────▼────────────┐
                   │ internal/domain/       │  (Entities, Usecases, Ports / Interfaces)
                   └───────────▲────────────┘
                               │ (Implements Ports)
                   ┌───────────┴────────────┐
                   │ internal/infrastructure│  (Postgres, Redis, Mongo, Kafka, OTel, slog)
                   └────────────────────────┘
```

### The Golden Rule: The Dependency Rule
- **Domain never imports infrastructure.**
- `internal/domain/sample/usecase/usecase.go` only relies on interfaces defined in `internal/domain/sample/interfaces.go` and `internal/domain/ports/telemetry.go`.
- `internal/infrastructure/` contains the actual third-party dependencies (e.g. `sqlx`, `go-redis`, `mongo-driver`, `segmentio/kafka-go`, `go.opentelemetry.io/otel`, `prometheus/client_golang`).
- `internal/app/bootstrap/` acts as the **Composition Root**: it initializes concrete infrastructure adapters and injects them into usecases and handlers.

---

## 2. SDK vs. Microservice Decision Framework

When scaling from a single service template to multiple microservices, you face two extremes:

1. **Copy-Paste Everything (No SDK)**:
   - *Problem*: High maintenance overhead, divergent log formats, broken distributed tracing across services, inconsistent error responses, and configuration drift.
2. **The "God SDK" Anti-Pattern (One massive shared repo)**:
   - *Problem*: Tight coupling. A minor update to a database driver forces services that don't even use databases to recompile. Dependency version conflicts (e.g., gRPC / protobuf versions) propagate across all teams.
3. **The Right Approach — Fine-Grained, Focused SDK Modules**:
   - Extract **generic, cross-cutting infrastructure and transport boilerplate** into small, focused modules.
   - Keep **all domain logic, schemas, queries, and business metrics** local to each service.

---

## 3. What to Extract into Shared SDKs (With File References)

Below is a detailed breakdown of codebase components, their roles, and how to extract them.

### A. Observability & Telemetry (Logger, APM, Metrics)

#### 1. Structured Logging
- **Source File**: `internal/infrastructure/logger/logger.go`
- **What it does**:
  - Sets up `log/slog` with JSON formatting and minimum log levels (`logger.go:14-32`).
  - Carries context-aware loggers across request boundaries via `FromContext` and `WithContext` (`logger.go:34-44`).
- **SDK Extraction**: Extract into `sdk/logger` (or `pkg/observability/logger`). Every microservice imports this to ensure identical JSON log structures for Loki/Promtail/Elasticsearch.

#### 2. APM / Distributed Tracing
- **Source File**: `internal/infrastructure/apm/apm.go`
- **What it does**:
  - Initializes OpenTelemetry `TracerProvider`, OTLP gRPC exporter to Grafana Tempo, trace sampler ratio, and W3C context propagators (`apm.go:26-62`).
- **SDK Extraction**: Extract into `sdk/telemetry/tracer`. Every service can initialize OTel with 5 lines of code, guaranteeing uniform `traceparent` header propagation across HTTP and Kafka boundaries.

#### 3. Metrics (Generic RED vs. Domain Metrics)
- **Source Files**:
  - `internal/infrastructure/metrics/metrics.go`
  - `internal/domain/ports/telemetry.go`
  - `internal/infrastructure/telemetry/telemetry.go`
- **What it does**:
  - `metrics.go:12-26` defines generic HTTP/gRPC metrics (`HTTPRequestsTotal`, `HTTPRequestDuration`, `GRPCRequestsTotal`).
  - `metrics.go:28-66` defines domain-specific metrics (`SampleOperationsTotal`, `MongoSyncFailuresTotal`, `OutboxEventsPending`).
- **SDK Extraction**:
  - **Extract to SDK**: Standard HTTP/gRPC RED metrics (`http_requests_total`, `http_request_duration_seconds`) and the `/metrics` promhttp handler (`metrics.go:68-70`).
  - **Keep in Service**: Domain-specific counters and gauge vectors (`SampleOperationsTotal`, `MongoSyncFailuresTotal`).
  - **Keep in Service**: Domain port interfaces (`internal/domain/ports/telemetry.go:17-33`) so domain usecases remain decoupled from Prometheus packages.

---

### B. Standard Errors & API Response Envelopes

#### 1. Application Error Model
- **Source File**: `internal/pkg/errors/errors.go`
- **What it does**:
  - Standard error codes: `NOT_FOUND`, `CONFLICT`, `VALIDATION_ERROR`, `UNAUTHORIZED`, `INTERNAL_ERROR` (`errors.go:23-33`).
  - `AppError` struct with HTTP status mapping, details map, and internal error wrapping (`errors.go:35-80`).
- **SDK Extraction**: Extract into `sdk/errors`. Standardizing error codes across all microservices ensures API Gateways, BFFs, and frontend clients handle errors uniformly.

#### 2. HTTP Response Formatter
- **Source File**: `internal/pkg/formatter/response.go`
- **What it does**:
  - Uniform JSON envelope: `{ "success": true, "message": "...", "data": {...}, "meta": {...}, "error": {...} }` (`response.go:10-29`).
  - Helper functions `formatter.Success`, `formatter.Created`, `formatter.Error` (`response.go:37-75`).
- **SDK Extraction**: Extract into `sdk/http/formatter` or `sdk/response`.

---

### C. Transport Middlewares & Context Helpers

#### 1. HTTP Middlewares
- **Source File**: `internal/app/middleware/middleware.go`
- **What it does**:
  - `RequestID`: Injects or generates `X-Request-Id` and binds it to context (`middleware.go:31-40`).
  - `Recoverer`: Recovers from panics, logs stack trace via `slog`, and returns uniform 500 error response (`middleware.go:42-65`).
  - `Logger`: Logs request start, completion status, and duration (`middleware.go:67-90`).
  - `Tracer`: Starts an OpenTelemetry HTTP span and records route attributes (`middleware.go:92-113`).
  - `Metrics`: Records Prometheus duration histogram and counter for each HTTP path pattern (`middleware.go:115-132`).
- **SDK Extraction**: Extract into `sdk/http/middleware`. These middlewares are completely generic and required by all HTTP services.

#### 2. Context Metadata Helpers
- **Source File**: `internal/pkg/helper/context.go`
- **What it does**:
  - Context helpers for request ID, user ID, and auth tokens (`context.go:1-30`).
- **SDK Extraction**: Extract into `sdk/contextutil`.

---

### D. Eventing, Transactional Outbox & Consumer Idempotency

#### 1. Event Envelope Schema
- **Source File**: `internal/domain/sample/event/event.go`
- **What it does**:
  - Standard event structure: `event_id`, `aggregate_type`, `aggregate_id`, `event_type`, `version`, `timestamp`, and `payload` (`event.go:1-35`).
- **SDK Extraction**: Extract the generic `Envelope[T]` or `CloudEvent` schema into `sdk/messaging/event`. Domain-specific payload types (like `SampleCreatedPayload`) stay in each service.

#### 2. Transactional Outbox Relay Engine
- **Source File**: `internal/infrastructure/outbox/relay.go`
- **What it does**:
  - Background ticker loop claiming pending rows with `FOR UPDATE SKIP LOCKED`, publishing to Kafka, marking status (`published`/`failed`), handling retries, and stale claim resets (`relay.go:40-200`).
- **SDK Extraction**: Extract the **algorithm and engine** into `sdk/outbox`. The service provides the SQL connection and the publisher implementation; the SDK handles concurrency, retry exponential backoff, and state transitions.

#### 3. Idempotent Consumer & DLQ Pattern
- **Source Files**:
  - `internal/infrastructure/messaging/sample/idempotency.go`
  - `internal/infrastructure/messaging/sample/consumer_handler.go`
  - `internal/infrastructure/messaging/sample/errors.go`
  - `internal/infrastructure/kafka/kafka.go`
- **What it does**:
  - `idempotency.go:10-38` checks `processed_events` table before executing side-effects.
  - `consumer_handler.go:40-80` unmarshals events, validates idempotency, dispatches handlers, and diverts unparseable/poison events to DLQ via `PublishDLQ`.
- **SDK Extraction**: Extract the **Idempotency Guard & DLQ Handler Wrapper** into `sdk/messaging/consumer`. The microservice only implements the business handler function `Handle(ctx, payload)`.

---

### E. Resilient Inter-Service Clients

#### 1. HTTP Client Factory
- **Source File**: `internal/infrastructure/external/http.go`
- **What it does**:
  - HTTP client wrapper with configurable base URL and timeout (`http.go:1-38`).
- **SDK Extraction**: Expand into `sdk/client/httpclient` with built-in:
  - Automatic OpenTelemetry trace propagation headers (`W3C TraceContext`).
  - Automatic `X-Request-Id` forwarding.
  - Retries with exponential backoff and jitter.
  - Circuit breaking (e.g. `sony/gobreaker`).

#### 2. gRPC Client Dialer
- **Source File**: `internal/infrastructure/grpc/client.go`
- **What it does**:
  - Upstream gRPC client dialer and standard health check (`client.go:1-45`).
- **SDK Extraction**: Extract into `sdk/client/grpcclient` with standard OpenTelemetry client interceptors, load balancing, and keep-alive configurations.

---

### F. API Contracts (Protobuf / gRPC)

- **Source Files**: `internal/proto/sample/v1/` and `internal/proto/healthcheck/v1/`
- **Pattern**:
  - Store `.proto` files in a dedicated central schema repository (e.g. `github.com/your-org/proto-contracts`) or manage them via Buf Schema Registry.
  - Automatically generate and publish Go packages (e.g. `github.com/your-org/genproto/go/sample/v1`).
  - Microservices import the compiled stubs rather than generating them locally inside `internal/proto/`.

---

## 4. What Must ALWAYS Stay Inside Each Microservice

To avoid tight coupling, the following components must **never** be moved into a shared SDK:

| Component | Code Location in Template | Reason to Keep Local |
|---|---|---|
| **Domain Entities** | `internal/domain/sample/entity/sample.go` | Business models are unique to the service's bounded context. |
| **Usecases / Business Logic** | `internal/domain/sample/usecase/usecase.go` | Domain workflows must evolve independently. |
| **Domain Ports (Interfaces)** | `internal/domain/sample/interfaces.go` | Clean Architecture requires the domain to own its interface definitions. |
| **Database Migrations** | `migrations/00001_*.sql` | Each service owns its private database schema. |
| **Database Repositories** | `internal/infrastructure/persistence/sample/*` | SQL queries, Mongo BSON mappings, and Redis keys are service-private. |
| **Domain Metrics** | `internal/infrastructure/metrics/metrics.go` (Domain vars) | Counters like `sample_operations_total` or `orders_created_total` are service-specific. |
| **Bootstrap & Routing** | `internal/app/bootstrap/*`, `internal/app/router/*` | Each service defines its own DI container and route tree. |

---

## 5. Recommended Modular SDK Structure

If building a shared toolkit (e.g., `github.com/your-org/core-go`), organize it into decoupled packages:

```
core-go/
├── logger/                  # slog JSON configuration & context helpers
│   ├── logger.go
│   └── context.go
├── telemetry/               # OpenTelemetry TracerProvider & MeterProvider setup
│   ├── tracer.go
│   └── propagator.go
├── errors/                  # Standard error codes & AppError type
│   └── errors.go
├── response/                # Standard JSON response envelope & helpers
│   └── formatter.go
├── middleware/              # Chi / net/http standard middlewares
│   ├── request_id.go
│   ├── recoverer.go
│   ├── logging.go
│   ├── tracing.go
│   └── metrics.go
├── messaging/               # Event envelope & outbox relay engine
│   ├── envelope.go
│   ├── outbox_relay.go
│   └── idempotency.go
└── client/                  # Pre-configured HTTP and gRPC client dialers
    ├── httpclient.go
    └── grpcclient.go
```

> **Important**: Avoid creating a single `util` or `common` package where everything is dumped. Keep packages modular so a service needing only `errors` does not pull in `kafka-go` or database drivers.

---

## 6. Summary Checklist

When building new microservices based on this template:

1. **Import Shared SDK for**:
   - `slog` logger setup and request ID propagation.
   - OpenTelemetry tracer initialization.
   - Standard HTTP/gRPC middlewares (logging, metrics, tracing, panic recovery).
   - Standard error codes (`errors.AppError`) and JSON response envelopes.
   - Outbox relay background worker loop.
   - Shared gRPC protobuf client stubs.

2. **Implement Inside the Service**:
   - Domain entities, value objects, and business usecases.
   - Domain ports (`interfaces.go`, `ports/telemetry.go`).
   - SQL queries, DB migrations, and repository implementations.
   - Custom business metrics and alerts.
   - Route definitions and dependency injection wiring in `bootstrap`.
