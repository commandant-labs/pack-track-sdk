package packtrack

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IngestResponse represents a minimal response for ingestion.
type IngestResponse struct {
	StatusCode int
	Body       []byte
}

// Client is the main SDK surface for PackTrack ingestion.
type Client interface {
	IngestEvent(ctx context.Context, e Event) (IngestResponse, error)
	IngestBatch(ctx context.Context, events []Event) (IngestResponse, error)
	HealthCheck(ctx context.Context) bool
	Flush(ctx context.Context) error
	Close(ctx context.Context) error
}

type client struct {
	cfg    Config
	http   *http.Client
	closed bool
}

// NewClient constructs a synchronous Client.
func NewClient(opts ...Option) (Client, error) {
	cfg := Defaults()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key required: use WithAPIKey")
	}
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("base URL required: use WithBaseURL")
	}
	h := cfg.HTTPClient
	if h == nil {
		h = &http.Client{Timeout: cfg.Timeout}
	}
	return &client{cfg: cfg, http: h}, nil
}

func (c *client) IngestEvent(ctx context.Context, e Event) (IngestResponse, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return IngestResponse{}, fmt.Errorf("marshal event: %w", err)
	}
	return c.send(ctx, payload, false)
}

func (c *client) IngestBatch(ctx context.Context, events []Event) (IngestResponse, error) {
	payload, err := json.Marshal(events)
	if err != nil {
		return IngestResponse{}, fmt.Errorf("marshal batch: %w", err)
	}
	return c.send(ctx, payload, true)
}

func (c *client) HealthCheck(ctx context.Context) bool {
	if !c.cfg.HealthEnable {
		return false
	}
	url := strings.TrimRight(c.cfg.BaseURL, "/") + c.cfg.HealthPath
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	c.addCommonHeaders(req)
	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (c *client) Flush(ctx context.Context) error { return nil }
func (c *client) Close(ctx context.Context) error { c.closed = true; return nil }

func (c *client) send(ctx context.Context, payload []byte, isBatch bool) (IngestResponse, error) {
	var body io.Reader = bytes.NewReader(payload)
	var contentEncoding string
	if c.cfg.Compression == CompressionGzip {
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write(payload); err != nil {
			return IngestResponse{}, fmt.Errorf("gzip write: %w", err)
		}
		if err := zw.Close(); err != nil {
			return IngestResponse{}, fmt.Errorf("gzip close: %w", err)
		}
		body = &buf
		contentEncoding = "gzip"
	}

	url := strings.TrimRight(c.cfg.BaseURL, "/") + "/api/ingest"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return IngestResponse{}, fmt.Errorf("build request: %w", err)
	}
	c.addCommonHeaders(req)
	req.Header.Set("Content-Type", "application/json")
	if contentEncoding != "" {
		req.Header.Set("Content-Encoding", contentEncoding)
	}
	if c.cfg.IdempotencyKey != "" {
		req.Header.Set("Idempotency-Key", c.cfg.IdempotencyKey)
	}

	attempts := c.cfg.Retry.MaxAttempts
	if attempts <= 0 {
		attempts = 1
	}
	var lastErr error
	for i := 0; i < attempts; i++ {
		resp, err := c.http.Do(req)
		if err != nil {
			lastErr = &IngestError{Retryable: true, Cause: err}
		} else {
			b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB cap
			resp.Body.Close()
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				if c.cfg.MetricsHooks != nil && c.cfg.MetricsHooks.OnIngestSuccess != nil {
					c.cfg.MetricsHooks.OnIngestSuccess(1)
				}
				return IngestResponse{StatusCode: resp.StatusCode, Body: b}, nil
			}
			retryable := resp.StatusCode >= 500 || resp.StatusCode == 429
			lastErr = &IngestError{StatusCode: resp.StatusCode, Body: string(b), Retryable: retryable}
		}
		// No retry on client errors except 429
		ie, _ := lastErr.(*IngestError)
		if ie != nil && !ie.Retryable {
			break
		}
		if i < attempts-1 {
			c.sleepBackoff(i)
		}
	}
	if c.cfg.MetricsHooks != nil && c.cfg.MetricsHooks.OnIngestFailure != nil {
		c.cfg.MetricsHooks.OnIngestFailure(1)
	}
	return IngestResponse{}, lastErr
}

func (c *client) sleepBackoff(attempt int) {
	// Exponential backoff with jitter
	backoff := c.cfg.Retry.InitialBackoff
	if backoff <= 0 {
		backoff = 100 * time.Millisecond
	}
	max := c.cfg.Retry.MaxBackoff
	if max <= 0 {
		max = 2 * time.Second
	}
	// 2^attempt * initial, capped at max
	d := backoff << attempt
	if d > max {
		d = max
	}
	jitter := c.cfg.Retry.Jitter
	if jitter < 0 {
		jitter = 0
	} else if jitter > 1 {
		jitter = 1
	}
	// apply jitter: [d*(1-j), d*(1+j)]
	lo := float64(d) * (1 - jitter)
	hi := float64(d) * (1 + jitter)
	n := time.Duration(lo + (hi-lo)/2) // simple mid-point to avoid math/rand dependency
	time.Sleep(n)
}

func (c *client) addCommonHeaders(req *http.Request) {
	req.Header.Set("X-PackTrack-Key", c.cfg.APIKey)
	if c.cfg.UserAgent != "" {
		req.Header.Set("User-Agent", c.cfg.UserAgent)
	}
}
