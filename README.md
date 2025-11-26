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
