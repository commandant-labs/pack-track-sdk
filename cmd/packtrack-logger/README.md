# packtrack-logger

A small CLI to submit PackTrack events using the Go SDK.

## Install

```
go install github.com/commandant-labs/pack-track-sdk/cmd/packtrack-logger@latest
```

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
