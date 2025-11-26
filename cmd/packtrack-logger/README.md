# packtrack-logger

A small CLI to submit PackTrack events using the Go SDK.

## Install

```
go install github.com/commandant-labs/pack-track-sdk/cmd/packtrack-logger@latest
```

## Environment-First Defaults
If your environment is fully configured, you can run:
```
packtrack-logger --message "Hello"
```
Environment variables checked (flags override):
- PACKTRACK_API_KEY, PACKTRACK_BASE_URL, PACKTRACK_TIMEOUT, PACKTRACK_RETRIES
- PACKTRACK_BACKOFF_INITIAL, PACKTRACK_BACKOFF_MAX, PACKTRACK_JITTER
- PACKTRACK_USER_AGENT, PACKTRACK_IDEMPOTENCY_KEY, PACKTRACK_GZIP
- PACKTRACK_VERBOSE, PACKTRACK_DRY_RUN, PACKTRACK_HEALTH, PACKTRACK_HEALTH_ENABLE, PACKTRACK_HEALTH_PATH
- PACKTRACK_TIMESTAMP, PACKTRACK_SOURCE_SYSTEM, PACKTRACK_SOURCE_ENV
- PACKTRACK_WORKFLOW_ID, PACKTRACK_WORKFLOW_NAME, PACKTRACK_RUN_ID, PACKTRACK_STEP_ID
- PACKTRACK_ACTOR_TYPE, PACKTRACK_ACTOR_ID, PACKTRACK_ACTOR_DISPLAY
- PACKTRACK_SEVERITY, PACKTRACK_STATUS, PACKTRACK_MESSAGE, PACKTRACK_METADATA, PACKTRACK_EXTRA
- PACKTRACK_FILE, PACKTRACK_STDIN, PACKTRACK_NDJSON
- PACKTRACK_ASYNC, PACKTRACK_BATCH_SIZE, PACKTRACK_FLUSH_INTERVAL, PACKTRACK_QUEUE_CAPACITY

## Usage (single event)

```
packtrack-logger \
  --api-key "$PACKTRACK_API_KEY" \
  --base-url https://pack.shimcounty.com \
  --source-system my-service --workflow-id wf-1 \
  --actor-type agent --actor-id a-1 \
  --severity info --status success \
  --message "hello from CLI"
```

## From file / STDIN

- Single event file: `packtrack-logger --api-key ... --file ./event.json`
- JSON array file: `packtrack-logger --api-key ... --file ./events.json`
- NDJSON stdin: `cat events.ndjson | packtrack-logger --api-key ... --stdin --ndjson`

## Exit Codes
- 0: success
- 1: invalid arguments
- 2: SDK error (non-retryable)
- 3: retries exhausted (retryable)
- 4: input/read/parse error
