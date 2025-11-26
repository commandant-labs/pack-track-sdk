# PackTrack Integration Guide (Factory Team)

This guide explains what PackTrack is and how the Factory team can log AI agent activity into PackTrack using the packtrack-logger CLI. It includes initial setup, environment variables, and common usage patterns.

## What Is PackTrack?

PackTrack is an observability platform for AI agent workflows. It captures structured events from your agents and services so you can:
- Trace workflows and steps across systems
- Track status (running, success, error) and severity (debug, info, warn, error)
- Attach structured metadata for debugging and analytics

You integrate by sending JSON events to PackTrack's ingest API. This repo provides a Go SDK and a CLI, packtrack-logger, that wraps the SDK for quick adoption from scripts and services.

## What Is packtrack-logger?

packtrack-logger is a small command-line tool included in this repo (cmd/packtrack-logger) that submits events to PackTrack via the Go SDK. With environment variables set, you can send an event with a single flag:

```
packtrack-logger --message "Hello"
```

It supports single events, JSON files (single or array), and NDJSON (newline-delimited JSON) for batches. Retries, backoff with jitter, gzip, and idempotency are supported.

## Install the CLI

Option 1: Go install (requires Go 1.21+)
```
go install github.com/commandant-labs/pack-track-sdk/cmd/packtrack-logger@latest
```
This installs packtrack-logger into your GOPATH/bin (ensure it is on PATH).

Option 2: Build from source
```
make build-cli
./bin/packtrack-logger -version
```

## Event Model (Summary)

An event has the following key fields (mapped by the CLI):
- timestamp (RFC3339) — default now (UTC)
- source: system (required), env (optional)
- workflow: id (required), name (optional), run_id (optional), step_id (optional)
- actor: type (required), id (required), display_name (optional)
- severity: debug | info | warn | error (required)
- status: success | running | error (required)
- message: short human-readable message (required)
- metadata: JSON object (optional)
- extra: JSON object (optional)

See cmd/packtrack-logger/README.md for full details.

## Initial Setup (Environment-First)

Set environment variables first so engineers can run the CLI with only a message. Minimal recommended set:

```
export PACKTRACK_API_KEY=pt_live_XXXXXXXXXXXXXXXX
export PACKTRACK_SOURCE_SYSTEM=factory-agents
export PACKTRACK_WORKFLOW_ID=factory-workflow
export PACKTRACK_ACTOR_TYPE=agent
export PACKTRACK_ACTOR_ID=agent-123
export PACKTRACK_SEVERITY=info
export PACKTRACK_STATUS=running
```

Common optional variables:
- PACKTRACK_BASE_URL (default https://pack.shimcounty.com)
- PACKTRACK_SOURCE_ENV (e.g., prod, staging)
- PACKTRACK_WORKFLOW_NAME, PACKTRACK_RUN_ID, PACKTRACK_STEP_ID
- PACKTRACK_ACTOR_DISPLAY (friendly name)
- PACKTRACK_METADATA (JSON), PACKTRACK_EXTRA (JSON)
- PACKTRACK_IDEMPOTENCY_KEY, PACKTRACK_GZIP

Full env variable list is documented in cmd/packtrack-logger/README.md. Flags always override env values.

## Quick Start

1) Export env vars (as above)
2) Send a message:
```
packtrack-logger --message "agent started"
```
3) Send success and error events by changing status/severity or providing flags to override envs:
```
# Mark step success
packtrack-logger --status success --severity info --message "step completed"

# Record an error
packtrack-logger --status error --severity error --message "failed to pick item" \
  --metadata '{"error":"picker jam","order_id":"ORD-42"}'
```

## File and NDJSON Inputs

- From a JSON file (single event):
```
packtrack-logger --file ./event.json
```
- From a JSON file (array of events):
```
packtrack-logger --file ./events.json
```
- NDJSON (newline-delimited JSON) from stdin:
```
cat events.ndjson | packtrack-logger --stdin --ndjson
```

## Health Check

Verify connectivity and auth:
```
packtrack-logger --health --health-path /api/health
```
Exit code 0 indicates OK; non-zero indicates a problem.

## Exit Codes

- 0: success
- 1: invalid arguments or validation error
- 2: non-retryable SDK/HTTP error (e.g., 400/401/403)
- 3: retries exhausted (retryable error, e.g., 429/5xx after backoff)
- 4: input/read/parse error (files or stdin)

## Best Practices for Factory Agents

- Use consistent IDs:
  - workflow.id for the workflow definition (e.g., "factory-workflow")
  - run_id for a specific execution (e.g., UUID per job)
  - step_id for activity stage (e.g., "pick", "pack", "ship")
- Severity vs Status:
  - severity reflects log importance (debug/info/warn/error)
  - status reflects state (running/success/error)
- Redact secrets: Do not put credentials/tokens in message or metadata
- Use metadata for structured fields (order_id, robot_id, SKU, durations)
- Use idempotency when repeating sends for the same event
- Prefer UTC timestamps; the CLI defaults to now in UTC if none provided

## Troubleshooting

- 401/403: Invalid or missing PACKTRACK_API_KEY
- 404: Check PACKTRACK_BASE_URL and API path
- 429: Backoff and retry; CLI does this automatically
- 5xx: Transient server issue; CLI retries with exponential backoff
- Timeout: Adjust PACKTRACK_TIMEOUT (e.g., "30s")
- Validation errors: Ensure required env or flags are set (source-system, workflow-id, actor-type, actor-id, severity, status, message)

## CI/CD Usage

Set env variables in your CI secret store and invoke the CLI in steps:
```
- name: Log build start
  run: packtrack-logger --message "factory build started"

- name: Log build end
  run: packtrack-logger --status success --message "factory build finished"
```

## Where To Learn More

- SDK quick start: README.md
- CLI docs with full flags and env list: cmd/packtrack-logger/README.md

If you have questions or need access to PackTrack, contact the PackTrack team.
