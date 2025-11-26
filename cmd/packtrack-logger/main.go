package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func stderr() *os.File { return os.Stderr }

func main() {
	cfg := &Config{}
	// Load env defaults before defining flags so flags get env-backed defaults
	cfg.LoadEnvDefaults()

	// Core flags
	flag.StringVar(&cfg.APIKey, "api-key", cfg.APIKey, "PackTrack API key (or set PACKTRACK_API_KEY)")
	flag.StringVar(&cfg.BaseURL, "base-url", cfg.BaseURL, "PackTrack base URL")
	flag.DurationVar(&cfg.Timeout, "timeout", defaultDuration(cfg.Timeout, 15*time.Second), "request timeout")
	flag.IntVar(&cfg.Retries, "retries", defaultInt(cfg.Retries, 3), "retry attempts")
	flag.DurationVar(&cfg.BackoffInitial, "backoff-initial", 100*time.Millisecond, "initial backoff")
	flag.DurationVar(&cfg.BackoffMax, "backoff-max", 2*time.Second, "maximum backoff")
	flag.Float64Var(&cfg.Jitter, "jitter", 0.2, "backoff jitter fraction 0..1")
	flag.StringVar(&cfg.UserAgent, "user-agent", cfg.UserAgent, "override User-Agent")
	flag.StringVar(&cfg.IdempotencyKey, "idempotency-key", "", "optional idempotency key")
	flag.BoolVar(&cfg.Gzip, "gzip", false, "enable gzip for batches")
	flag.BoolVar(&cfg.Verbose, "verbose", false, "verbose output to stderr")
	flag.BoolVar(&cfg.DryRun, "dry-run", false, "validate inputs without sending")
	flag.BoolVar(&cfg.Health, "health", false, "perform health check and exit")
	flag.StringVar(&cfg.HealthPath, "health-path", defaultString(cfg.HealthPath, "/api/health"), "health check path")

	// Event flags
	flag.StringVar(&cfg.Timestamp, "timestamp", "", "RFC3339 timestamp (default now)")
	flag.StringVar(&cfg.SourceSystem, "source-system", "", "event source system (required unless using --file/--stdin)")
	flag.StringVar(&cfg.SourceEnv, "source-env", "", "event source environment")
	flag.StringVar(&cfg.WorkflowID, "workflow-id", "", "workflow id (required unless using --file/--stdin)")
	flag.StringVar(&cfg.WorkflowName, "workflow-name", "", "workflow name")
	flag.StringVar(&cfg.RunID, "run-id", "", "workflow run id")
	flag.StringVar(&cfg.StepID, "step-id", "", "workflow step id")
	flag.StringVar(&cfg.ActorType, "actor-type", "", "actor type (required unless using --file/--stdin)")
	flag.StringVar(&cfg.ActorID, "actor-id", "", "actor id (required unless using --file/--stdin)")
	flag.StringVar(&cfg.ActorDisplay, "actor-display", "", "actor display name")
	flag.StringVar(&cfg.Severity, "severity", "", "severity [debug|info|warn|error]")
	flag.StringVar(&cfg.Status, "status", "", "status [success|running|error]")
	flag.StringVar(&cfg.Message, "message", "", "message text")
	flag.StringVar(&cfg.MetadataJSON, "metadata", "", "metadata JSON object")
	flag.StringVar(&cfg.ExtraJSON, "extra", "", "extra JSON object")

	// Input
	flag.StringVar(&cfg.File, "file", "", "read event(s) from JSON file (object or array); use --ndjson for NDJSON")
	flag.BoolVar(&cfg.Stdin, "stdin", false, "read event(s) from STDIN")
	flag.BoolVar(&cfg.NDJSON, "ndjson", false, "treat input as NDJSON for batch")

	// Async
	flag.BoolVar(&cfg.Async, "async", false, "use async batching")
	flag.IntVar(&cfg.BatchSize, "batch-size", 100, "batch size for async or split in sync mode")
	flag.DurationVar(&cfg.FlushInterval, "flush-interval", time.Second, "async flush interval")
	flag.IntVar(&cfg.QueueCapacity, "queue-capacity", 10000, "async queue capacity")

	// Misc
	showVersion := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("packtrack-logger %s\n", Version)
		os.Exit(0)
	}

	// Ensure API key presence when not dry-run/health
	if !cfg.DryRun && !cfg.Health && cfg.APIKey == "" {
		fmt.Fprintln(stderr(), "error: missing --api-key or PACKTRACK_API_KEY")
		os.Exit(ExitInvalid)
	}

	code := runCLI(cfg)
	os.Exit(code)
}

func defaultDuration(got, def time.Duration) time.Duration {
	if got > 0 {
		return got
	}
	return def
}
func defaultInt(got, def int) int {
	if got > 0 {
		return got
	}
	return def
}
func defaultString(got, def string) string {
	if got != "" {
		return got
	}
	return def
}
