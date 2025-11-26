# Product Requirements Document (PRD)

## Project Overview

This PRD defines the requirements for `pack-track-sdk` — an official Go (Golang) client library for interacting with PackTrack. The SDK enables Go applications and services to ingest events into PackTrack and, over time, access read endpoints as they are formalized.

- Project Name: pack-track-sdk (Go)
- Version: 0.1.0 (PRD)
- Date: 2025-11-25
- Author: PackTrack Team

### Background

PackTrack provides real-time observability for AI agent workflows. Today, event ingestion is performed via HTTP requests to the `/api/ingest` endpoint using an API key (header `X-PackTrack-Key`). This SDK standardizes that interaction for Go programs, offering a strongly-typed, ergonomic, and resilient client.

## Objectives

1. Provide an idiomatic Go client for PackTrack with first-class support for event ingestion.
2. Offer production-grade reliability features: retries, backoff with jitter, batching, optional async queue, and graceful shutdown/flush.
3. Expose a simple, well-typed API aligned to the PackTrack event schema.
4. Make the SDK easy to adopt: clear documentation, examples, and semver releases.
5. Prepare a stable foundation to add read APIs (workflows, runs, events) as public endpoints are finalized.

## Scope (v1)

In-scope for v1:

- Authenticated event ingestion to `/api/ingest` (single and batch).
- Client configuration (base URL, API key, timeouts, retries, transport, user-agent).
- Optional asynchronous ingestion with batching and periodic flush.
- Graceful shutdown with `Flush()` and `Close()` semantics.
- Minimal health probe (optional, configurable path; disabled by default).
- Typed error model with retryability hints.
- Context-aware APIs and cancellation.

Out of scope for v1 (may be v1.x/v2):

- Read APIs for events/workflows/runs (awaiting public endpoints).
- Admin operations (API key management, alerts CRUD, etc.).
- Non-HTTP transports (gRPC, NATS, Kafka, etc.).

## Target Audience

- Go service developers producing PackTrack events (microservices, workers, CLIs).
- SRE/Platform teams integrating observability into Go-based systems.

## Technical Requirements

### Language and Compatibility

- Go version: ≥ 1.21 (prefer 1.22 for toolchain parity).
- Module path: `github.com/<org>/pack-track-sdk-go` (exact path to be finalized).
- OS: Linux, macOS, Windows; architectures supported by stdlib net/http.

### Authentication

- API key provided by caller and sent as header `X-PackTrack-Key: <key>` on every request.
- No token refresh or OAuth flows in v1.

### Endpoints

- Ingest (required): `POST {BaseURL}/api/ingest`
  - Request: JSON event or array of events (batch).
  - Headers: `Content-Type: application/json`, `X-PackTrack-Key: <key>`.
  - Response: 2xx on success; body may include IDs or summary where applicable.
- Health (optional): `GET {BaseURL}/api/health` (if available) or configurable path. Disabled by default.

### Event Schema (summary)

Match the PackTrack event schema documented in README/docs (key fields below):

- `timestamp`: RFC3339 string
- `source`: `{ system: string, env?: string }`
- `workflow`: `{ id: string, name?: string, run_id?: string, step_id?: string }`
- `actor`: `{ type: "agent"|"system"|string, id: string, display_name?: string }`
- `severity`: `debug|info|warn|error`
- `status`: `ok|failed|skipped|timeout|retrying|unknown`
- `message`: string
- `metadata`: object (optional)

The SDK should define Go structs reflecting this schema, with JSON tags matching server expectations. It must allow additional fields via `map[string]any` for forward compatibility.

### Public API (proposed)

```go
// Client construction
c, err := packtrack.NewClient(
  packtrack.WithBaseURL("https://pack.shimcounty.com"),
  packtrack.WithAPIKey(os.Getenv("PACKTRACK_API_KEY")),
  packtrack.WithTimeout(15*time.Second),
  packtrack.WithRetry(3, packtrack.WithExponentialBackoff(100*time.Millisecond, 2*time.Second)),
  packtrack.WithUserAgent("my-service/1.2.3"),
)

// Synchronous ingestion (single)
resp, err := c.IngestEvent(ctx, packtrack.Event{ /* fields */ })

// Synchronous ingestion (batch)
resp, err := c.IngestBatch(ctx, []packtrack.Event{e1, e2, e3})

// Asynchronous ingestion with batching
ac, _ := packtrack.NewAsyncClient(c,
  packtrack.WithBatchSize(100),
  packtrack.WithFlushInterval(1*time.Second),
  packtrack.WithQueueCapacity(10_000),
  packtrack.WithCompression(packtrack.CompressionGzip),
)
_ = ac.Enqueue(e1)
_ = ac.Enqueue(e2)
_ = ac.Flush(ctx)   // optional on-demand flush
_ = ac.Close(ctx)   // flush and stop workers

// Optional health check (disabled by default)
ok := c.HealthCheck(ctx)
```

### Configuration Options (minimum)

- Base URL (string), default `https://pack.shimcounty.com` (override for self-hosted/local).
- API Key (string), required.
- Timeout (default 15s), per-request via `context.Context` override.
- Retries: default 3; exponential backoff with jitter; retry on network errors and HTTP 5xx; do not retry 4xx (except 429 with backoff).
- HTTP Transport: allow custom `*http.Client` injection.
- User-Agent: default `pack-track-sdk-go/<version>` with override.
- Compression: optional gzip for batch payloads.

### Reliability & Performance

- Thread-safe client usable from multiple goroutines.
- Async mode: background worker(s) aggregate events into batches (`batchSize`, `flushInterval`); flush on `Close()`.
- Backpressure: bounded queue with configurable capacity. Drop policy configurable (reject enqueue vs. drop-oldest), default reject enqueue with explicit error.
- Idempotency: allow optional caller-provided idempotency key header for future server-side dedupe (pass-through; no client store required in v1).
- Target throughput (guideline): 10k events/min per process with batch size 100 and gzip enabled.

### Observability & Logging

- Pluggable logger interface (leveled); disabled by default.
- Minimal metrics hooks (callbacks) for success/failure counts and queue depth (optional, non-breaking).

### Error Model

- Typed errors with fields: `StatusCode` (int), `Body` (string up to limit), `Retryable` (bool), `Cause` (error).
- Wrap using `errors.Join` / `%w`; never panic in library code.

### Security

- Never log API keys or full payloads by default.
- TLS verification on by default; allow custom CA roots via injected transport only.
- Respect context cancellation and deadlines.

## Documentation & Examples

- README with quick start, configuration, sync/async examples.
- API reference via GoDoc (godoc.org/pkg).
- Examples directory with runnable samples.

## Testing Strategy

- Unit tests: payload marshaling, retry logic, backoff, async queue behavior, gzip.
- HTTP tests: use `httptest.Server` for success/failure (2xx/4xx/5xx/429) and network error scenarios.
- Concurrency tests: race detector in CI (`-race`).
- Integration (optional): gated tests against a local PackTrack dev environment.
- Load test scenario (optional): validate batching throughput targets.

## CI/CD

- GitHub Actions: lint (golangci-lint), unit tests (Linux, macOS, Windows), Go matrix (1.21, 1.22), race detector, coverage badge.
- Release process: tag with semver (v0.x initially), GitHub Release notes, changelog.

## Versioning & Compatibility

- Semantic Versioning; v0.x for initial iterations, v1.0.0 when API is stable.
- Backward-compatible additions only in minor versions; breaking changes require major bump.

## Milestones

1. v0.1.0 (MVP): sync ingestion, config, retries/backoff, basic docs, unit tests.
2. v0.2.0: async queue + batching + gzip; flush/close; examples.
3. v0.3.0: typed errors, metrics hooks, improved docs.
4. v1.0.0: stabilization; finalize public API; performance validation.
5. Future: read APIs once server endpoints are available.

## Acceptance Criteria (v0.1.0)

- NewClient, IngestEvent, IngestBatch implemented and documented.
- Retries with exponential backoff and jitter; configurable attempt count.
- Context cancellation respected for all requests.
- Unit test coverage ≥ 80% for core ingestion paths.
- README includes quick start and code examples.

## Risks & Mitigations

- Server read endpoints not finalized → Limit v1 to ingestion; design client surface to extend without breaking.
- Misuse of async queue leading to drops → Provide clear defaults, explicit errors on backpressure, metrics hooks.
- API key leakage → Never log secrets; redaction helpers; document best practices.

## Open Questions

- Health endpoint path and semantics (is `/api/health` available?) — default to disabled; make path configurable.
- Preferred default batch size/flush interval for async mode — propose 100 events / 1s; validate with load tests.
- Idempotency header support on server — if/when added, client can pass through.

---

This PRD defines the initial scope and quality bar for the Go SDK enabling PackTrack ingestion from Go applications. It prioritizes reliability, idiomatic design, and ease of adoption while leaving room for future read APIs.
