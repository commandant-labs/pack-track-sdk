package packtrack

import (
	"context"
	"time"
)

// LogLevel represents severity for log events.
type LogLevel int

const (
	LevelDebug LogLevel = iota
	LevelInfo
	LevelWarn
	LevelError
)

// Fields carries structured log fields.
type Fields map[string]any

// Labels carries metric labels/tags.
type Labels map[string]string

// Client is the main SDK surface for logging and telemetry.
type Client interface {
	// Log emits a structured log entry.
	Log(ctx context.Context, level LogLevel, message string, fields Fields) error
	// Metric records a single metric data point.
	Metric(ctx context.Context, name string, value float64, labels Labels) error
	// Flush ensures all buffered data is persisted/transported.
	Flush(ctx context.Context) error
	// Close flushes and releases resources.
	Close(ctx context.Context) error
}

// Options configures a Client instance.
type Options struct {
	// Endpoint is the base URL for the Pack-Track ingest service.
	Endpoint string
	// APIKey is used for authenticating requests to the service.
	APIKey string
	// Timeout controls per-request timeout to the ingest service.
	Timeout time.Duration
	// MaxBatch controls how many items are batched per send.
	MaxBatch int
	// BackgroundFlushInterval flushes buffers periodically when > 0.
	BackgroundFlushInterval time.Duration
	// DisableLogging disables the logging pipeline when true.
	DisableLogging bool
	// DisableTelemetry disables the telemetry pipeline when true.
	DisableTelemetry bool
}

// Option is the functional option type used to build Options.
type Option func(*Options)

// New constructs a Client. Implementation added in subsequent iterations.
func New(opts ...Option) (Client, error) {
	_ = applyOptions(opts...)
	// TODO: actual client construction wired with transport, buffers, etc.
	return nil, nil
}

func applyOptions(opts ...Option) *Options {
	o := &Options{
		Timeout:                 10 * time.Second,
		MaxBatch:                100,
		BackgroundFlushInterval: 0,
		DisableLogging:          false,
		DisableTelemetry:        false,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}
	return o
}

// WithEndpoint sets the ingest endpoint.
func WithEndpoint(u string) Option {
	return func(o *Options) { o.Endpoint = u }
}

// WithAPIKey sets the API key.
func WithAPIKey(k string) Option {
	return func(o *Options) { o.APIKey = k }
}

// WithTimeout sets the per-request timeout.
func WithTimeout(d time.Duration) Option {
	return func(o *Options) { o.Timeout = d }
}

// WithMaxBatch sets the max batch size.
func WithMaxBatch(n int) Option {
	return func(o *Options) { o.MaxBatch = n }
}

// WithBackgroundFlushInterval sets periodic flush interval.
func WithBackgroundFlushInterval(d time.Duration) Option {
	return func(o *Options) { o.BackgroundFlushInterval = d }
}

// DisableLogging turns off logging pipeline.
func DisableLogging() Option {
	return func(o *Options) { o.DisableLogging = true }
}

// DisableTelemetry turns off telemetry pipeline.
func DisableTelemetry() Option {
	return func(o *Options) { o.DisableTelemetry = true }
}
