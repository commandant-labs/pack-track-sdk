package packtrack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func newTestEvent() Event {
	return Event{
		Timestamp: time.Now().UTC(),
		Source:    Source{System: "test"},
		Workflow:  Workflow{ID: "wf", RunID: "run-1"},
		Actor:     Actor{Type: "agent", ID: "a1"},
		Severity:  SeverityInfo,
		Status:    StatusOK,
		Message:   "ok",
	}
}

func TestIngestEvent_Success(t *testing.T) {
	var called int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		if r.Header.Get("X-PackTrack-Key") == "" {
			t.Fatalf("missing api key")
		}
		w.WriteHeader(200)
		w.Write([]byte("{\"ok\":true}"))
	}))
	defer ts.Close()

	c, err := NewClient(
		WithBaseURL(ts.URL),
		WithAPIKey("k"),
		WithRetry(1, 10*time.Millisecond, 10*time.Millisecond, 0),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	resp, err := c.IngestEvent(context.Background(), newTestEvent())
	if err != nil {
		t.Fatalf("ingest: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("status=%d", resp.StatusCode)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected 1 call")
	}
}

func TestIngestEvent_RetryOn500(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(204)
	}))
	defer ts.Close()
	c, _ := NewClient(
		WithBaseURL(ts.URL),
		WithAPIKey("k"),
		WithRetry(2, 1*time.Millisecond, 2*time.Millisecond, 0),
	)
	_, err := c.IngestEvent(context.Background(), newTestEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 attempts")
	}
}

func TestIngestEvent_NoRetryOn400(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(400)
	}))
	defer ts.Close()
	c, _ := NewClient(
		WithBaseURL(ts.URL),
		WithAPIKey("k"),
		WithRetry(3, 1*time.Millisecond, 2*time.Millisecond, 0),
	)
	_, err := c.IngestEvent(context.Background(), newTestEvent())
	if err == nil {
		t.Fatalf("expected error")
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 attempt")
	}
}

func TestIngestEvent_RetryOn429(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 2 {
			w.WriteHeader(429)
			return
		}
		w.WriteHeader(200)
	}))
	defer ts.Close()
	c, _ := NewClient(
		WithBaseURL(ts.URL),
		WithAPIKey("k"),
		WithRetry(2, 1*time.Millisecond, 2*time.Millisecond, 0),
	)
	_, err := c.IngestEvent(context.Background(), newTestEvent())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Fatalf("expected 2 attempts")
	}
}

func TestIngestBatch_Success(t *testing.T) {
	var called int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&called, 1)
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("content-type")
		}
		w.WriteHeader(201)
	}))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"))
	_, err := c.IngestBatch(context.Background(), []Event{newTestEvent(), newTestEvent()})
	if err != nil {
		t.Fatalf("batch ingest: %v", err)
	}
	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected 1 call")
	}
}

func TestHealthCheck_DisabledByDefault(t *testing.T) {
	c, _ := NewClient(WithBaseURL("http://example"), WithAPIKey("k"))
	if c.HealthCheck(context.Background()) {
		t.Fatalf("expected disabled health")
	}
}

func TestHealthCheck_EnabledOK(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"), WithHealthEnabled(true), WithHealthPath("/"))
	if !c.HealthCheck(context.Background()) {
		t.Fatalf("expected health ok")
	}
}

func TestRetryExhaustion_ReturnsIngestError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"), WithRetry(2, 1*time.Millisecond, 2*time.Millisecond, 0))
	_, err := c.IngestEvent(context.Background(), newTestEvent())
	if err == nil {
		t.Fatalf("expected error")
	}
	ie, ok := err.(*IngestError)
	if !ok {
		t.Fatalf("expected IngestError")
	}
	if !ie.Retryable {
		t.Fatalf("expected retryable error")
	}
}

func TestMetricsHooks(t *testing.T) {
	var success, failure int32
	tsFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer tsFail.Close()
	hooks := &MetricsHooks{
		OnIngestSuccess: func(count int) { atomic.AddInt32(&success, int32(count)) },
		OnIngestFailure: func(count int) { atomic.AddInt32(&failure, int32(count)) },
	}
	c, _ := NewClient(WithBaseURL(tsFail.URL), WithAPIKey("k"), WithRetry(1, 1*time.Millisecond, 1*time.Millisecond, 0), WithMetricsHooks(hooks))
	_, _ = c.IngestEvent(context.Background(), newTestEvent())
	if atomic.LoadInt32(&success) != 0 {
		t.Fatalf("unexpected success metric")
	}
	if atomic.LoadInt32(&failure) != 1 {
		t.Fatalf("expected failure metric = 1")
	}
	// success path
	success = 0
	failure = 0
	tsOK := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) }))
	defer tsOK.Close()
	c2, _ := NewClient(WithBaseURL(tsOK.URL), WithAPIKey("k"), WithMetricsHooks(hooks))
	_, _ = c2.IngestEvent(context.Background(), newTestEvent())
	if atomic.LoadInt32(&success) != 1 {
		t.Fatalf("expected success metric = 1")
	}
	if atomic.LoadInt32(&failure) != 0 {
		t.Fatalf("unexpected failure metric")
	}
}

func TestCompression_GzipHeader(t *testing.T) {
	var gotENC string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotENC = r.Header.Get("Content-Encoding")
		w.WriteHeader(200)
	}))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"), WithCompression(CompressionGzip))
	_, err := c.IngestBatch(context.Background(), []Event{newTestEvent()})
	if err != nil {
		t.Fatal(err)
	}
	if gotENC != "gzip" {
		t.Fatalf("expected gzip, got %q", gotENC)
	}
}

func TestIdempotencyKeyHeader(t *testing.T) {
	var got string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Get("Idempotency-Key")
		w.WriteHeader(200)
	}))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"), WithIdempotencyKey("abc123"))
	_, err := c.IngestEvent(context.Background(), newTestEvent())
	if err != nil {
		t.Fatal(err)
	}
	if got != "abc123" {
		t.Fatalf("expected idempotency header")
	}
}

func TestUserAgent_Default(t *testing.T) {
	var ua string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		w.WriteHeader(200)
	}))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"))
	_, err := c.IngestEvent(context.Background(), newTestEvent())
	if err != nil {
		t.Fatal(err)
	}
	if ua == "" {
		t.Fatalf("expected default user agent")
	}
}
