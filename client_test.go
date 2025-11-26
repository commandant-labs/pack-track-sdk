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
		Workflow:  Workflow{ID: "wf"},
		Actor:     Actor{Type: "agent", ID: "a1"},
		Severity:  SeverityInfo,
		Status:    StatusSuccess,
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
