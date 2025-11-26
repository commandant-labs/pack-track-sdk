package main

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds CLI configuration and event fields.
type Config struct {
	// Core
	APIKey         string
	BaseURL        string
	Timeout        time.Duration
	Retries        int
	BackoffInitial time.Duration
	BackoffMax     time.Duration
	Jitter         float64
	UserAgent      string
	IdempotencyKey string
	Gzip           bool
	Verbose        bool
	DryRun         bool
	Health         bool
	HealthEnable   bool
	HealthPath     string

	// Event fields (single event mode)
	Timestamp    string
	SourceSystem string
	SourceEnv    string
	WorkflowID   string
	WorkflowName string
	RunID        string
	StepID       string
	ActorType    string
	ActorID      string
	ActorDisplay string
	Severity     string
	Status       string
	Message      string
	MetadataJSON string
	ExtraJSON    string

	// Input/batch
	File   string
	Stdin  bool
	NDJSON bool

	// Async
	Async         bool
	BatchSize     int
	FlushInterval time.Duration
	QueueCapacity int
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func envBool(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	switch strings.ToLower(v) {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return def
	}
}

// LoadEnvDefaults populates defaults from environment variables.
func (c *Config) LoadEnvDefaults() {
	// Core
	c.APIKey = envOrDefault("PACKTRACK_API_KEY", c.APIKey)
	c.BaseURL = envOrDefault("PACKTRACK_BASE_URL", c.BaseURL)
	c.Timeout = envDuration("PACKTRACK_TIMEOUT", c.Timeout)
	c.Retries = envInt("PACKTRACK_RETRIES", c.Retries)
	c.BackoffInitial = envDuration("PACKTRACK_BACKOFF_INITIAL", c.BackoffInitial)
	c.BackoffMax = envDuration("PACKTRACK_BACKOFF_MAX", c.BackoffMax)
	c.Jitter = envFloat("PACKTRACK_JITTER", c.Jitter)
	c.UserAgent = envOrDefault("PACKTRACK_USER_AGENT", c.UserAgent)
	c.IdempotencyKey = envOrDefault("PACKTRACK_IDEMPOTENCY_KEY", c.IdempotencyKey)
	c.Gzip = envBool("PACKTRACK_GZIP", c.Gzip)
	c.Verbose = envBool("PACKTRACK_VERBOSE", c.Verbose)
	c.DryRun = envBool("PACKTRACK_DRY_RUN", c.DryRun)
	c.Health = envBool("PACKTRACK_HEALTH", c.Health)
	c.HealthEnable = envBool("PACKTRACK_HEALTH_ENABLE", c.HealthEnable)
	c.HealthPath = envOrDefault("PACKTRACK_HEALTH_PATH", c.HealthPath)

	// Event fields
	c.Timestamp = envOrDefault("PACKTRACK_TIMESTAMP", c.Timestamp)
	c.SourceSystem = envOrDefault("PACKTRACK_SOURCE_SYSTEM", c.SourceSystem)
	c.SourceEnv = envOrDefault("PACKTRACK_SOURCE_ENV", c.SourceEnv)
	c.WorkflowID = envOrDefault("PACKTRACK_WORKFLOW_ID", c.WorkflowID)
	c.WorkflowName = envOrDefault("PACKTRACK_WORKFLOW_NAME", c.WorkflowName)
	c.RunID = envOrDefault("PACKTRACK_RUN_ID", c.RunID)
	c.StepID = envOrDefault("PACKTRACK_STEP_ID", c.StepID)
	c.ActorType = envOrDefault("PACKTRACK_ACTOR_TYPE", c.ActorType)
	c.ActorID = envOrDefault("PACKTRACK_ACTOR_ID", c.ActorID)
	c.ActorDisplay = envOrDefault("PACKTRACK_ACTOR_DISPLAY", c.ActorDisplay)
	c.Severity = envOrDefault("PACKTRACK_SEVERITY", c.Severity)
	c.Status = envOrDefault("PACKTRACK_STATUS", c.Status)
	c.Message = envOrDefault("PACKTRACK_MESSAGE", c.Message)
	c.MetadataJSON = envOrDefault("PACKTRACK_METADATA", c.MetadataJSON)
	c.ExtraJSON = envOrDefault("PACKTRACK_EXTRA", c.ExtraJSON)

	// Input
	c.File = envOrDefault("PACKTRACK_FILE", c.File)
	c.Stdin = envBool("PACKTRACK_STDIN", c.Stdin)
	c.NDJSON = envBool("PACKTRACK_NDJSON", c.NDJSON)

	// Async
	c.Async = envBool("PACKTRACK_ASYNC", c.Async)
	c.BatchSize = envInt("PACKTRACK_BATCH_SIZE", c.BatchSize)
	c.FlushInterval = envDuration("PACKTRACK_FLUSH_INTERVAL", c.FlushInterval)
	c.QueueCapacity = envInt("PACKTRACK_QUEUE_CAPACITY", c.QueueCapacity)
}
