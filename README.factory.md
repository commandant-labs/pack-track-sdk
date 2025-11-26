# PackTrack Integration Guide (Factory Team)

This standalone guide explains what PackTrack is and how the Factory team can log AI agent activity into PackTrack using the packtrack-logger CLI. It includes a brief platform overview, installation, full environment setup, and ready-to-run examples.

## What Is PackTrack?

PackTrack is an observability platform for AI agent workflows. It ingests structured events from your agents/services so you can:
- Trace workflows and steps across systems
- Track status: running, success, error
- Track severity: debug, info, warn, error
- Attach structured metadata (order IDs, robot IDs, durations, etc.) for analysis

Events are sent via HTTPS to PackTrack's ingest API. This repository provides a CLI (packtrack-logger) that wraps the official Go SDK to submit events reliably with retries and backoff.

## What You Will Do

- Install the packtrack-logger CLI
- Configure environment variables once (environment-first)
- Call the CLI from your agents/scripts to emit events (as simple as: `packtrack-logger --message "Hello"`)

## Install the CLI

Option A: Go install (requires Go 1.21+)
```
go install github.com/commandant-labs/pack-track-sdk/cmd/packtrack-logger@latest
```
Ensure your `$GOPATH/bin` or `$(go env GOPATH)/bin` is on your PATH.

Option B: Build from source
```
make build-cli
./bin/packtrack-logger -version
```

## Event Model (Summary)

Each event has the following key fields:
- timestamp: RFC3339 string (default: now in UTC)
- source: { system (required), env (optional) }
- workflow: { id (required), name (optional), run_id (optional), step_id (optional) }
- actor: { type (required), id (required), display_name (optional) }
- severity: debug | info | warn | error (required)
- status: success | running | error (required)
- message: short human-readable string (required)
- metadata: JSON object (optional)
- extra: JSON object (optional)

## Environment Setup (Environment-First)

Set these environment variables so you can run with only a `--message` flag.

Minimal recommended set:
```
export PACKTRACK_API_KEY=pt_live_XXXXXXXXXXXXXXXX
export PACKTRACK_SOURCE_SYSTEM=factory-agents
export PACKTRACK_WORKFLOW_ID=factory-workflow
export PACKTRACK_ACTOR_TYPE=agent
export PACKTRACK_ACTOR_ID=agent-123
export PACKTRACK_SEVERITY=info
export PACKTRACK_STATUS=running
```
Common optional:
- `PACKTRACK_BASE_URL` (default `https://pack.shimcounty.com`)
- `PACKTRACK_SOURCE_ENV` (e.g., prod, staging)
- `PACKTRACK_WORKFLOW_NAME`, `PACKTRACK_RUN_ID`, `PACKTRACK_STEP_ID`
- `PACKTRACK_ACTOR_DISPLAY` (friendly name)
- `PACKTRACK_METADATA`, `PACKTRACK_EXTRA` (JSON strings)
- `PACKTRACK_IDEMPOTENCY_KEY` (optional dedupe ID), `PACKTRACK_GZIP` ("true" to compress batches)

Full environment variable list (flags override env):
- Core:
  - `PACKTRACK_API_KEY`, `PACKTRACK_BASE_URL`, `PACKTRACK_TIMEOUT`, `PACKTRACK_RETRIES`
  - `PACKTRACK_BACKOFF_INITIAL`, `PACKTRACK_BACKOFF_MAX`, `PACKTRACK_JITTER`
  - `PACKTRACK_USER_AGENT`, `PACKTRACK_IDEMPOTENCY_KEY`, `PACKTRACK_GZIP`
  - `PACKTRACK_VERBOSE`, `PACKTRACK_DRY_RUN`, `PACKTRACK_HEALTH`, `PACKTRACK_HEALTH_ENABLE`, `PACKTRACK_HEALTH_PATH`
- Event:
  - `PACKTRACK_TIMESTAMP`, `PACKTRACK_SOURCE_SYSTEM`, `PACKTRACK_SOURCE_ENV`
  - `PACKTRACK_WORKFLOW_ID`, `PACKTRACK_WORKFLOW_NAME`, `PACKTRACK_RUN_ID`, `PACKTRACK_STEP_ID`
  - `PACKTRACK_ACTOR_TYPE`, `PACKTRACK_ACTOR_ID`, `PACKTRACK_ACTOR_DISPLAY`
  - `PACKTRACK_SEVERITY`, `PACKTRACK_STATUS`, `PACKTRACK_MESSAGE`, `PACKTRACK_METADATA`, `PACKTRACK_EXTRA`
- Input:
  - `PACKTRACK_FILE`, `PACKTRACK_STDIN`, `PACKTRACK_NDJSON`
- Async (optional):
  - `PACKTRACK_ASYNC`, `PACKTRACK_BATCH_SIZE`, `PACKTRACK_FLUSH_INTERVAL`, `PACKTRACK_QUEUE_CAPACITY`

Windows PowerShell equivalents:
```
$env:PACKTRACK_API_KEY = "pt_live_XXXXXXXXXXXXXXXX"
$env:PACKTRACK_SOURCE_SYSTEM = "factory-agents"
$env:PACKTRACK_WORKFLOW_ID = "factory-workflow"
$env:PACKTRACK_ACTOR_TYPE = "agent"
$env:PACKTRACK_ACTOR_ID = "agent-123"
$env:PACKTRACK_SEVERITY = "info"
$env:PACKTRACK_STATUS = "running"
```

## Quick Start Commands

Send a basic event using env vars + message:
```
packtrack-logger --message "agent started"
```

Override status/severity at call time:
```
packtrack-logger --status success --severity info --message "step completed"
packtrack-logger --status error --severity error --message "failed to pick item" \
  --metadata '{"error":"picker jam","order_id":"ORD-42"}'
```

## Using Files and NDJSON

Single event JSON file (event.json):
```
{
  "timestamp": "2025-11-25T12:34:56Z",
  "source": {"system": "factory-agents", "env": "prod"},
  "workflow": {"id": "factory-workflow", "run_id": "RUN-001", "step_id": "pick"},
  "actor": {"type": "agent", "id": "agent-123", "display_name": "PickerBot"},
  "severity": "info",
  "status": "running",
  "message": "picking started",
  "metadata": {"order_id": "ORD-42"}
}
```
Send it:
```
packtrack-logger --file ./event.json
```

JSON array of events (events.json):
```
[
  { "timestamp": "2025-11-25T12:34:56Z", "source": {"system": "factory-agents"},
    "workflow": {"id": "factory-workflow"}, "actor": {"type": "agent", "id": "agent-123"},
    "severity": "info", "status": "running", "message": "picking started" },
  { "timestamp": "2025-11-25T12:36:00Z", "source": {"system": "factory-agents"},
    "workflow": {"id": "factory-workflow"}, "actor": {"type": "agent", "id": "agent-123"},
    "severity": "info", "status": "success", "message": "picking complete" }
]
```
Send it:
```
packtrack-logger --file ./events.json
```

NDJSON from stdin (events.ndjson):
```
{"timestamp":"2025-11-25T12:34:56Z","source":{"system":"factory-agents"},"workflow":{"id":"factory-workflow"},"actor":{"type":"agent","id":"agent-123"},"severity":"info","status":"running","message":"picking started"}
{"timestamp":"2025-11-25T12:36:00Z","source":{"system":"factory-agents"},"workflow":{"id":"factory-workflow"},"actor":{"type":"agent","id":"agent-123"},"severity":"info","status":"success","message":"picking complete"}
```
Send it:
```
cat events.ndjson | packtrack-logger --stdin --ndjson
```

## Optional Features

- Health check (connectivity/auth):
```
packtrack-logger --health --health-path /api/health
```
- Async batching:
```
packtrack-logger --async --batch-size 100 --flush-interval 1s --queue-capacity 10000 \
  --message "processed batch"
```
- Gzip compression (mainly for batches):
```
packtrack-logger --gzip --message "compressed batch"
```
- Idempotency (if re-sending the same event):
```
packtrack-logger --idempotency-key abc123 --message "dedupe-safe"
```

## Exit Codes

- 0: success
- 1: invalid arguments or validation error
- 2: non-retryable SDK/HTTP error (e.g., 400/401/403)
- 3: retries exhausted (retryable error, e.g., 429/5xx after backoff)
- 4: input/read/parse error (files or stdin)

## Troubleshooting

- 401/403: Invalid or missing `PACKTRACK_API_KEY`
- 404: Verify `PACKTRACK_BASE_URL` and API path
- 429: Rate limited; CLI automatically retries with backoff
- 5xx: Transient server issue; CLI retries with backoff
- Timeout: Increase `PACKTRACK_TIMEOUT` (e.g., `30s`)
- Validation errors: Ensure required env/flags are set (`source-system`, `workflow-id`, `actor-type`, `actor-id`, `severity`, `status`, `message`)

## Best Practices

- Use consistent IDs:
  - `workflow.id` for the workflow (e.g., `factory-workflow`)
  - `run_id` per execution/job
  - `step_id` for stages (e.g., `pick`, `pack`, `ship`)
- Severity vs Status:
  - `severity` = importance of the log (debug/info/warn/error)
  - `status` = lifecycle state (running/success/error)
- Redact secrets: never put credentials/tokens in message or metadata
- Prefer UTC; CLI defaults to UTC now if `timestamp` is omitted
- Use metadata for structured fields (order_id, robot_id, SKU, durations)

## Verification

- Check health:
```
packtrack-logger --health --health-path /api/health
```
- Send a test event:
```
packtrack-logger --message "factory integration test"
```
Exit code `0` indicates success.

## Reference

- To see all flags: `packtrack-logger -h`
- Flags always override environment variables.

If you need access or have questions, contact the PackTrack team.
