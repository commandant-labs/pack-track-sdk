# PackTrack Go SDK

Official Go SDK for PackTrack ingestion for AI agent observability.

## Install

```
go get github.com/commandant-labs/pack-track-sdk
```

## Quick Start (Sync)

```go
c, err := packtrack.NewClient(
    packtrack.WithBaseURL("https://pack.shimcounty.com"),
    packtrack.WithAPIKey(os.Getenv("PACKTRACK_API_KEY")),
)
if err != nil { log.Fatal(err) }

ctx := context.Background()
ev := packtrack.Event{ /* fill fields */ }
_, err = c.IngestEvent(ctx, ev)
```

## Async Batching

```go
ac, _ := packtrack.NewAsyncClient(c,
    packtrack.WithBatchSize(100),
    packtrack.WithFlushInterval(time.Second),
    packtrack.WithQueueCapacity(10000),
)
_ = ac.Enqueue(ev)
_ = ac.Flush(ctx)
_ = ac.Close(ctx)
```

## CLI: packtrack-logger
A companion CLI is included at `cmd/packtrack-logger` to submit events from the shell.

Install:
```
go install github.com/commandant-labs/pack-track-sdk/cmd/packtrack-logger@latest
```

Environment-first defaults let you run with only a message when env is set:
```
export PACKTRACK_API_KEY=...
export PACKTRACK_SOURCE_SYSTEM=my-service
export PACKTRACK_WORKFLOW_ID=wf-1
export PACKTRACK_ACTOR_TYPE=agent
export PACKTRACK_ACTOR_ID=a-1
export PACKTRACK_SEVERITY=info
export PACKTRACK_STATUS=ok
packtrack-logger --message "Hello"
```
See full CLI docs: cmd/packtrack-logger/README.md

## Configuration
- Base URL (default https://pack.shimcounty.com)
- API Key (required) via `X-PackTrack-Key`
- Timeout (default 15s)
- Retries (default 3) with exponential backoff and jitter
- HTTP client injection
- User-Agent override
- Optional gzip compression for batch payloads
- Optional health check (disabled by default)

## License
Apache-2.0
