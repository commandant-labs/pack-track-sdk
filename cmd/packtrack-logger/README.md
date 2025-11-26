# packtrack-logger

A small CLI to submit PackTrack events using the Go SDK.

## Install

```
go install github.com/commandant-labs/pack-track-sdk/cmd/packtrack-logger@latest
```

## Environment-First Defaults
If your environment is fully configured, you can send an event with only a message:
```
packtrack-logger --message "Hello"
```
The CLI checks environment variables first (flags override). Recommended minimal set:
- PACKTRACK_API_KEY
- PACKTRACK_SOURCE_SYSTEM
- PACKTRACK_WORKFLOW_ID
- PACKTRACK_RUN_ID
- PACKTRACK_ACTOR_TYPE

- PACKTRACK_ACTOR_ID
- PACKTRACK_SEVERITY (debug|info|warn|error)
- PACKTRACK_STATUS (ok|failed|skipped|timeout|retrying|unknown)

Optional (commonly used):
- PACKTRACK_BASE_URL (default https://pack.shimcounty.com)
- PACKTRACK_METADATA, PACKTRACK_EXTRA (JSON)
- PACKTRACK_USER_AGENT, PACKTRACK_IDEMPOTENCY_KEY

Example .env snippet:
```
export PACKTRACK_API_KEY=pt_live_XXXXXXXXXXXXXXXX
export PACKTRACK_SOURCE_SYSTEM=my-service
export PACKTRACK_WORKFLOW_ID=wf-123
export PACKTRACK_RUN_ID=run-1
export PACKTRACK_ACTOR_TYPE=agent
export PACKTRACK_ACTOR_ID=a-42
export PACKTRACK_SEVERITY=info
export PACKTRACK_STATUS=ok
```

Full list of supported env vars (flags override):
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

## Examples

Single event (flags):
```
packtrack-logger \
  --api-key "$PACKTRACK_API_KEY" \
  --base-url https://pack.shimcounty.com \
  --source-system my-service --workflow-id wf-1 \
  --actor-type agent --actor-id a-1 \
  --severity info --status ok \
  --message "hello from CLI"
```

From JSON file:
```
# Single event object
packtrack-logger --api-key "$PACKTRACK_API_KEY" --file ./event.json

# JSON array of events
packtrack-logger --api-key "$PACKTRACK_API_KEY" --file ./events.json
```

NDJSON from stdin:
```
cat events.ndjson | packtrack-logger --api-key "$PACKTRACK_API_KEY" --stdin --ndjson
```

Health check:
```
packtrack-logger --api-key "$PACKTRACK_API_KEY" --health --health-path /api/health
```

Async batching:
```
packtrack-logger --async --batch-size 100 --flush-interval 1s --queue-capacity 10000 \
  --message "processed batch" \
  # assumes environment provides required event fields
```

Gzip + Idempotency:
```
packtrack-logger --gzip --idempotency-key abc123 --message "compressed batch"
```

Dry-run and verbose:
```
packtrack-logger --dry-run --verbose --message "test only"
```

## Flags
- Flags mirror the environment variables above; run `packtrack-logger -h` to see all flags.
- Flags override environment values.

## Exit Codes
- 0: success
- 1: invalid arguments
- 2: SDK error (non-retryable)
- 3: retries exhausted (retryable)
- 4: input/read/parse error

## Build From Source
```
make build-cli
./bin/packtrack-logger -version
```

## Releases
- Download prebuilt binaries from the GitHub Releases page for your OS/arch: darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64, windows-arm64.
- Each archive contains the `packtrack-logger` binary and READMEs. Verify with the provided `checksums.txt`.

## Cutting a Release (maintainers)
1) Update `cmd/packtrack-logger/version.go` if needed.
2) Tag and push: `git tag vX.Y.Z && git push origin vX.Y.Z`.
3) GitHub Actions will run GoReleaser and publish binaries to the tag's release.
4) Optional: test locally without publishing: `goreleaser release --clean --skip=publish`.
5) If your org restricts the default `GITHUB_TOKEN`, create a repo secret `GORELEASER_TOKEN` with a PAT (classic) that has `repo` scope; the workflow will use it automatically.
