package packtrack

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestAsync_BatchBySize(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"))
	ac, _ := NewAsyncClient(c, WithBatchSize(2), WithFlushInterval(time.Hour), WithQueueCapacity(10))
	defer ac.Close(context.Background())
	if err := ac.Enqueue(newTestEvent()); err != nil {
		t.Fatal(err)
	}
	if err := ac.Enqueue(newTestEvent()); err != nil {
		t.Fatal(err)
	}
	// allow worker to process
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 batch call")
	}
}

func TestAsync_QueueFull(t *testing.T) {
	c, _ := NewClient(WithBaseURL("http://example"), WithAPIKey("k"))
	ac, _ := NewAsyncClient(c, WithQueueCapacity(1))
	defer ac.Close(context.Background())
	_ = ac.Enqueue(newTestEvent())
	if err := ac.Enqueue(newTestEvent()); err == nil {
		t.Fatalf("expected queue full error")
	}
}

func TestAsync_Flush(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(200)
	}))
	defer ts.Close()
	c, _ := NewClient(WithBaseURL(ts.URL), WithAPIKey("k"))
	ac, _ := NewAsyncClient(c, WithBatchSize(100), WithFlushInterval(time.Hour), WithQueueCapacity(10))
	_ = ac.Enqueue(newTestEvent())
	_ = ac.Enqueue(newTestEvent())
	if err := ac.Flush(context.Background()); err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Fatalf("expected 1 call after flush")
	}
}
