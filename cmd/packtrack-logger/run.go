package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	packtrack "github.com/commandant-labs/pack-track-sdk"
)

const (
	ExitOK           = 0
	ExitInvalid      = 1
	ExitSDKError     = 2
	ExitRetryExhaust = 3
	ExitInputError   = 4
)

func buildClient(cfg *Config) (packtrack.Client, error) {
	opts := []packtrack.Option{
		packtrack.WithBaseURL(nonEmpty(cfg.BaseURL, packtrack.Defaults().BaseURL)),
		packtrack.WithAPIKey(cfg.APIKey),
		packtrack.WithTimeout(nonZeroDuration(cfg.Timeout, packtrack.Defaults().Timeout)),
		packtrack.WithRetry(nonZeroInt(cfg.Retries, packtrack.Defaults().Retry.MaxAttempts),
			nonZeroDuration(cfg.BackoffInitial, packtrack.Defaults().Retry.InitialBackoff),
			nonZeroDuration(cfg.BackoffMax, packtrack.Defaults().Retry.MaxBackoff),
			clampFloat(cfg.Jitter, 0.0, 1.0)),
	}
	if cfg.UserAgent != "" {
		opts = append(opts, packtrack.WithUserAgent(cfg.UserAgent+" packtrack-logger/"+Version))
	}
	if cfg.IdempotencyKey != "" {
		opts = append(opts, packtrack.WithIdempotencyKey(cfg.IdempotencyKey))
	}
	if cfg.Gzip {
		opts = append(opts, packtrack.WithCompression(packtrack.CompressionGzip))
	}
	if cfg.HealthEnable {
		opts = append(opts, packtrack.WithHealthEnabled(true))
	}
	if cfg.HealthPath != "" {
		opts = append(opts, packtrack.WithHealthPath(cfg.HealthPath))
	}
	return packtrack.NewClient(opts...)
}

func nonEmpty(s, def string) string {
	if s != "" {
		return s
	}
	return def
}
func nonZeroDuration(d, def time.Duration) time.Duration {
	if d > 0 {
		return d
	}
	return def
}
func nonZeroInt(n, def int) int {
	if n > 0 {
		return n
	}
	return def
}
func clampFloat(f, lo, hi float64) float64 {
	if f < lo {
		return lo
	}
	if f > hi {
		return hi
	}
	return f
}

func runCLI(cfg *Config) int {
	// Health check mode
	if cfg.Health {
		cfg.HealthEnable = true
		cl, err := buildClient(cfg)
		if err != nil {
			fmt.Fprintf(stderr(), "error: %v\n", err)
			return ExitInvalid
		}
		ok := cl.HealthCheck(context.Background())
		if cfg.Verbose {
			fmt.Fprintf(stderr(), "health: %v\n", ok)
		}
		if ok {
			return ExitOK
		}
		return ExitInvalid
	}

	// Input source
	var events []packtrack.Event
	if cfg.File != "" {
		ev, err := readEventsFromFile(cfg.File, cfg.NDJSON)
		if err != nil {
			fmt.Fprintf(stderr(), "input error: %v\n", err)
			return ExitInputError
		}
		events = ev
	} else if cfg.Stdin {
		ev, err := readEventsFromStdin(cfg.NDJSON)
		if err != nil {
			fmt.Fprintf(stderr(), "input error: %v\n", err)
			return ExitInputError
		}
		events = ev
	} else {
		// Build single event from flags
		ev, err := buildEventFromConfig(cfg)
		if err != nil {
			fmt.Fprintf(stderr(), "invalid: %v\n", err)
			return ExitInvalid
		}
		events = []packtrack.Event{ev}
	}

	if cfg.DryRun {
		if cfg.Verbose {
			fmt.Fprintln(stderr(), "dry-run: validation passed")
		}
		return ExitOK
	}

	cl, err := buildClient(cfg)
	if err != nil {
		fmt.Fprintf(stderr(), "error: %v\n", err)
		return ExitInvalid
	}

	ctx := context.Background()
	if cfg.Async {
		ac, err := packtrack.NewAsyncClient(cl,
			packtrack.WithBatchSize(nonZeroInt(cfg.BatchSize, 100)),
			packtrack.WithFlushInterval(nonZeroDuration(cfg.FlushInterval, time.Second)),
			packtrack.WithQueueCapacity(nonZeroInt(cfg.QueueCapacity, 10000)))
		if err != nil {
			fmt.Fprintf(stderr(), "error: %v\n", err)
			return ExitInvalid
		}
		for _, e := range events {
			if err := ac.Enqueue(e); err != nil {
				fmt.Fprintf(stderr(), "enqueue error: %v\n", err)
				return ExitSDKError
			}
		}
		if err := ac.Flush(ctx); err != nil {
			fmt.Fprintf(stderr(), "flush error: %v\n", err)
			return classifyErr(err)
		}
		if err := ac.Close(ctx); err != nil {
			fmt.Fprintf(stderr(), "close error: %v\n", err)
			return classifyErr(err)
		}
		return ExitOK
	}

	// Sync
	if len(events) == 1 {
		_, err := cl.IngestEvent(ctx, events[0])
		if err != nil {
			fmt.Fprintf(stderr(), "ingest error: %v\n", err)
			return classifyErr(err)
		}
		return ExitOK
	}

	// If batchSize specified, chunk
	if cfg.BatchSize > 0 && cfg.BatchSize < len(events) {
		for i := 0; i < len(events); i += cfg.BatchSize {
			end := i + cfg.BatchSize
			if end > len(events) {
				end = len(events)
			}
			if _, err := cl.IngestBatch(ctx, events[i:end]); err != nil {
				fmt.Fprintf(stderr(), "batch ingest error: %v\n", err)
				return classifyErr(err)
			}
		}
		return ExitOK
	}
	_, err = cl.IngestBatch(ctx, events)
	if err != nil {
		fmt.Fprintf(stderr(), "batch ingest error: %v\n", err)
		return classifyErr(err)
	}
	return ExitOK
}

func classifyErr(err error) int {
	var ie *packtrack.IngestError
	if errors.As(err, &ie) {
		if ie.Retryable {
			return ExitRetryExhaust
		}
		return ExitSDKError
	}
	// Fallback: network/unknown
	s := err.Error()
	if strings.Contains(s, "timeout") || strings.Contains(s, "context deadline") {
		return ExitSDKError
	}
	return ExitSDKError
}
