# Product Requirements Document (PRD)

## Project: packtrack-logger (Go CLI)
- Version: 0.1.0 (PRD)
- Date: 2025-11-25
- Author: PackTrack Team
- Depends on: github.com/commandant-labs/pack-track-sdk

### Overview
A simple, reliable command-line program that submits PackTrack events using the official Go SDK. Engineers and scripts can invoke the executable to emit a single event (and optionally batches) to PackTrack. This tool targets quick integrations, shell pipelines, and operational workflows.

### Objectives
1. Provide an ergonomic CLI to send PackTrack events via the SDK.
2. Support explicit field flags and JSON input (stdin/file) to form events.
3. Offer production-grade reliability (retries, timeouts), secure defaults, and clear exit codes.
4. Ship prebuilt binaries for major platforms and basic documentation.

### Non-Objectives (v1)
- Full-fledged log shipping/agent with tailing, rotation, filters.
- Streaming ingestion or high-throughput daemons (async mode optional, off by default).
- Read APIs (workflows/runs/events). This CLI is write-only in v1.

## Scope (v1)
- Single-event submission from flags or a JSON file/stdin.
- Optional batch submission from a JSON array file or NDJSON via stdin.
- Authentication using API key header `X-PackTrack-Key`.
- Config via flags and environment variables with secure precedence.
- Retries with exponential backoff and jitter; respect context timeouts.
- Clear process exit codes and minimal output by default; verbose logs optional.

Out-of-scope (v1):
- File tailing; log discovery; local buffering to disk.
- Non-HTTP transports; proxy auto-discovery (use OS/http defaults only).

## Target Audience
- Engineers, SREs, and CI/CD jobs needing quick event submission.
- Shell scripts and operational tooling in heterogeneous environments.

## Technical Requirements
### Language and Compatibility
- Go ≥ 1.21 (build with 1.22 in CI).
- OS: Linux, macOS, Windows (amd64/arm64 where supported).
- Single static binary per OS/arch.

### Authentication
- API Key required on all requests via `X-PackTrack-Key: <key>`.
- Provided via `--api-key` or env `PACKTRACK_API_KEY` (flag wins).

### Endpoints
- Uses the SDK default BaseURL `https://pack.shimcounty.com` (override supported).
- Ingest: POST `{BaseURL}/api/ingest` (via SDK `IngestEvent`/`IngestBatch`).
- Optional health probe for diagnostics only (no-op by default).

### Event Schema
- Conform to SDK `packtrack.Event` structure:
  - timestamp (RFC3339), source.system, source.env (optional),
  - workflow.id (required), optional name/run_id/step_id,
  - actor.type, actor.id, optional display_name,
  - severity: debug|info|warn|error,
  - status: ok|failed|skipped|timeout|retrying|unknown,
  - message: string,
  - metadata: object (optional), extra: object (optional).
- CLI provides flags for all common fields; metadata/extra accepted as JSON.

## CLI Interface
Executable name: `packtrack-logger`

### Usage (single event)
```
packtrack-logger \
  --api-key "$PACKTRACK_API_KEY" \
  --base-url https://pack.shimcounty.com \
  --timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
  --source-system my-service --source-env prod \
  --workflow-id wf-123 --workflow-name "Checkout" --run-id run-1 --step-id step-1 \
  --actor-type agent --actor-id a-42 --actor-display "cart-stepper" \
  --severity info --status ok \
  --message "checkout completed" \
  --metadata '{"cart_total":123.45,"currency":"USD"}'
```

### Usage (from file/STDIN)
- JSON file (single event): `packtrack-logger --api-key ... --file ./event.json`
- JSON file (batch array): `packtrack-logger --api-key ... --file ./events.json`
- NDJSON from STDIN (batch): `cat events.ndjson | packtrack-logger --api-key ... --stdin --ndjson`

### Flags
- Core:
  - `--api-key` string (required if PACKTRACK_API_KEY not set)
  - `--base-url` string (default `https://pack.shimcounty.com`)
  - `--timeout` duration (default 15s)
  - `--retries` int (default 3)
  - `--backoff-initial` duration (default 100ms)
  - `--backoff-max` duration (default 2s)
  - `--jitter` float (0..1, default 0.2)
  - `--user-agent` string (default `pack-track-sdk-go/<version> packtrack-logger/<version>`)
  - `--idempotency-key` string (optional)
  - `--gzip` (bool; compress batch payloads)

- Event construction (single record):
  - `--timestamp` RFC3339 (default now UTC)
  - `--source-system` string (required)
  - `--source-env` string
  - `--workflow-id` string (required)
  - `--workflow-name` string
  - `--run-id` string
  - `--step-id` string
  - `--actor-type` string (required)
  - `--actor-id` string (required)
  - `--actor-display` string
  - `--severity` enum [debug|info|warn|error] (required)
  - `--status` enum [ok|failed|skipped|timeout|retrying|unknown] (required)
  - `--message` string (required)
  - `--metadata` JSON object (string)
  - `--extra` JSON object (string)

- Input sources:
  - `--file` path (single JSON event or JSON array)
  - `--stdin` (read from STDIN)
  - `--ndjson` (treat STDIN/--file as newline-delimited JSON; batch)

- Diagnostics and UX:
  - `--verbose` (print debug info to stderr; never log secrets)
  - `--dry-run` (parse/validate only; do not send)
  - `--health` (run a health check then exit 0/1)

- Async (optional, off by default):
  - `--async` (enable async batching)
  - `--batch-size` int (default 100)
  - `--flush-interval` duration (default 1s)
  - `--queue-capacity` int (default 10000)

### Environment Variables
- `PACKTRACK_API_KEY` (required if `--api-key` omitted)
- `PACKTRACK_BASE_URL` (default override)
- `PACKTRACK_TIMEOUT`, `PACKTRACK_RETRIES`, `PACKTRACK_USER_AGENT`
- `PACKTRACK_HEALTH_ENABLE`, `PACKTRACK_HEALTH_PATH`

Precedence: flags > environment variables > defaults.

### Exit Codes
- 0: success (all events submitted)
- 1: invalid arguments/validation error
- 2: network/SDK error (non-retryable)
- 3: retries exhausted (retryable error)
- 4: input/read/parse error (file/stdin)

## Behavior Details
- Single event mode: construct an `Event` from flags and call `IngestEvent`.
- Batch mode:
  - File contains array -> unmarshal to []Event -> `IngestBatch`.
  - NDJSON -> read lines, each JSON -> batch submit (either single `IngestBatch` or multiple batches if `--batch-size` provided, even in sync mode).
- Async mode:
  - Wrap base client with `NewAsyncClient` and enqueue events; call `Flush` at the end unless `--no-flush` is set (default is to flush).
- Health mode: if `--health` set, perform `HealthCheck` using SDK (enabled temporarily) and exit 0/1 without sending events.
- Sensitive data: never echo API key or full payload when `--verbose`.

## Reliability & Security
- Retries per SDK config; respect context deadlines (timeout flag maps to http.Client timeout).
- TLS verification on; trust roots per OS; custom CA only via custom HTTP client injection is out-of-scope for v1.
- Idempotency: pass `--idempotency-key` when provided.
- Gzip: enable if `--gzip` and batching is used; harmless if single event but allowed.

## Logging
- Default: minimal output (quiet). On success, print nothing unless `--verbose`.
- Verbose: print attempt info, retry decisions, and summarized responses (status code only). Do not print API keys or full payloads.

## Documentation & Examples
- README:
  - Installation via `go install` and prebuilt binaries links.
  - Examples for single event, JSON file, and NDJSON/STDIN.
  - Description of exit codes and environment variables.

## Testing Strategy
- Unit:
  - Flag parsing and validation (required fields; JSON parsing for metadata/extra).
  - Event construction correctness.
- SDK/HTTP (with httptest.Server):
  - Success (2xx), client errors (4xx), server errors (5xx), 429 with retry.
  - Gzip header when `--gzip` with batch.
- STDIN/Files:
  - Single JSON event file; JSON array file; NDJSON lines.
  - Error cases: malformed JSON; empty input; mixed types.
- Exit codes: assert correct codes for each failure/success scenario.
- Race detector in CI.

## CI/CD
- GitHub Actions matrix (linux/macos/windows; Go 1.21, 1.22): build, vet, test -race.
- Release workflow to produce signed archives (optional) for darwin/linux/windows amd64/arm64.
- Optional GoReleaser config for cross builds.

## Versioning
- CLI version v0.x until the interface stabilizes; embed version in `--version` output.

## Milestones
1. v0.1.0: single event + batch via file/stdin; retries; docs; tests.
2. v0.2.0: async mode; NDJSON batching; gzip flag; more examples.
3. v1.0.0: stabilization, UX polish, robust error messages and exit codes.

## Acceptance Criteria (v0.1.0)
- `packtrack-logger` binary sends a single event via flags and returns exit code 0.
- Supports `--file` with single event and JSON array inputs.
- Respects `PACKTRACK_API_KEY`; rejects when auth missing; never logs secrets.
- Retries as configured; does not retry on 4xx (except 429); uses backoff.
- Unit tests for argument parsing, basic HTTP paths, and exit codes.
- README documents installations and working examples.

## Open Questions
- Should CLI default to NDJSON mode when `--stdin` without `--ndjson`? (Proposed: require explicit `--ndjson`.)
- Should we support a `--drop-oldest` backpressure policy in async mode? (Proposed: not in v1; default reject.)
- Should we add structured log output (JSON) for machine consumption? (Proposed: add `--json` in v0.2.0.)
