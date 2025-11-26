package packtrack

import (
	"net/http"
	"time"
)

// CompressionType represents payload compression.
type CompressionType int

const (
	CompressionNone CompressionType = iota
	CompressionGzip
)

// RetryConfig controls retry behavior.
type RetryConfig struct {
	MaxAttempts int
	// Base/backoff settings
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
	Jitter         float64 // 0..1 fraction
}

// Config holds client configuration.
type Config struct {
	BaseURL    string
	APIKey     string
	Timeout    time.Duration
	HTTPClient *http.Client
	UserAgent  string

	Retry       RetryConfig
	Compression CompressionType

	// Optional idempotency header value for future use
	IdempotencyKey string

	// Optional hooks
	Logger       Logger
	MetricsHooks *MetricsHooks

	// Health check options
	HealthPath   string
	HealthEnable bool
}

// Option configures the Client via functional options.
type Option func(*Config)

// Defaults returns default config values.
func Defaults() Config {
	return Config{
		BaseURL:   "https://pack.shimcounty.com",
		Timeout:   15 * time.Second,
		UserAgent: "pack-track-sdk-go/0.1.0",
		Retry: RetryConfig{
			MaxAttempts:    3,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     2 * time.Second,
			Jitter:         0.2,
		},
		Compression: CompressionNone,
		HealthPath:  "/api/health",
	}
}

func WithBaseURL(u string) Option        { return func(c *Config) { c.BaseURL = u } }
func WithAPIKey(k string) Option         { return func(c *Config) { c.APIKey = k } }
func WithTimeout(d time.Duration) Option { return func(c *Config) { c.Timeout = d } }
func WithHTTPClient(h *http.Client) Option {
	return func(c *Config) { c.HTTPClient = h }
}
func WithUserAgent(ua string) Option { return func(c *Config) { c.UserAgent = ua } }

func WithRetry(max int, initial, maxBackoff time.Duration, jitter float64) Option {
	return func(c *Config) {
		c.Retry = RetryConfig{MaxAttempts: max, InitialBackoff: initial, MaxBackoff: maxBackoff, Jitter: jitter}
	}
}

func WithExponentialBackoff(initial, maxBackoff time.Duration) Option {
	return func(c *Config) {
		c.Retry.InitialBackoff = initial
		c.Retry.MaxBackoff = maxBackoff
	}
}

func WithCompression(ct CompressionType) Option { return func(c *Config) { c.Compression = ct } }

func WithIdempotencyKey(k string) Option { return func(c *Config) { c.IdempotencyKey = k } }

func WithLogger(l Logger) Option { return func(c *Config) { c.Logger = l } }

func WithMetricsHooks(h *MetricsHooks) Option { return func(c *Config) { c.MetricsHooks = h } }

func WithHealthEnabled(enabled bool) Option { return func(c *Config) { c.HealthEnable = enabled } }

func WithHealthPath(p string) Option { return func(c *Config) { c.HealthPath = p } }
