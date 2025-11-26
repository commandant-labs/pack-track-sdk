package main

import (
	"os"
	"strconv"
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

// LoadEnvDefaults populates defaults from environment variables.
func (c *Config) LoadEnvDefaults() {
	c.APIKey = envOrDefault("PACKTRACK_API_KEY", c.APIKey)
	c.BaseURL = envOrDefault("PACKTRACK_BASE_URL", c.BaseURL)
	c.Timeout = envDuration("PACKTRACK_TIMEOUT", c.Timeout)
	c.Retries = envInt("PACKTRACK_RETRIES", c.Retries)
	c.UserAgent = envOrDefault("PACKTRACK_USER_AGENT", c.UserAgent)
	if os.Getenv("PACKTRACK_HEALTH_ENABLE") != "" {
		c.HealthEnable = os.Getenv("PACKTRACK_HEALTH_ENABLE") == "1" || os.Getenv("PACKTRACK_HEALTH_ENABLE") == "true"
	}
	c.HealthPath = envOrDefault("PACKTRACK_HEALTH_PATH", c.HealthPath)
}
