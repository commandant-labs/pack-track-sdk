package packtrack

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Async options

type AsyncOption func(*AsyncConfig)

type AsyncConfig struct {
	BatchSize     int
	FlushInterval time.Duration
	QueueCapacity int
}

func defaultAsyncConfig() AsyncConfig {
	return AsyncConfig{BatchSize: 100, FlushInterval: time.Second, QueueCapacity: 10000}
}

func WithBatchSize(n int) AsyncOption { return func(a *AsyncConfig) { a.BatchSize = n } }
func WithFlushInterval(d time.Duration) AsyncOption {
	return func(a *AsyncConfig) { a.FlushInterval = d }
}
func WithQueueCapacity(n int) AsyncOption { return func(a *AsyncConfig) { a.QueueCapacity = n } }

// AsyncClient wraps a sync Client with background batching.
type AsyncClient interface {
	Enqueue(e Event) error
	Flush(ctx context.Context) error
	Close(ctx context.Context) error
}

type asyncClient struct {
	base   Client
	cfg    AsyncConfig
	q      chan Event
	wg     sync.WaitGroup
	mu     sync.Mutex
	closed bool
}

// NewAsyncClient creates an AsyncClient on top of an existing Client.
func NewAsyncClient(base Client, opts ...AsyncOption) (AsyncClient, error) {
	if base == nil {
		return nil, errors.New("base client is nil")
	}
	cfg := defaultAsyncConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	ac := &asyncClient{
		base: base,
		cfg:  cfg,
		q:    make(chan Event, cfg.QueueCapacity),
	}
	ac.wg.Add(1)
	go ac.worker()
	return ac, nil
}

func (a *asyncClient) Enqueue(e Event) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return errors.New("async client closed")
	}
	select {
	case a.q <- e:
		return nil
	default:
		return errors.New("queue full")
	}
}

func (a *asyncClient) Flush(ctx context.Context) error {
	// Signal flush via special timer path by draining current queue into a batch
	// Implemented by spinning a temporary batch on demand.
	var batch []Event
	for {
		select {
		case e := <-a.q:
			batch = append(batch, e)
			if len(batch) >= a.cfg.BatchSize {
				_, err := a.base.IngestBatch(ctx, batch)
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		default:
			if len(batch) > 0 {
				_, err := a.base.IngestBatch(ctx, batch)
				return err
			}
			return nil
		}
	}
}

func (a *asyncClient) Close(ctx context.Context) error {
	a.mu.Lock()
	a.closed = true
	close(a.q)
	a.mu.Unlock()
	a.wg.Wait()
	return a.base.Close(ctx)
}

func (a *asyncClient) worker() {
	defer a.wg.Done()
	var batch []Event
	flush := func() {
		if len(batch) == 0 {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, _ = a.base.IngestBatch(ctx, batch)
		cancel()
		batch = batch[:0]
	}
	var t *time.Timer
	resetTimer := func() {
		if a.cfg.FlushInterval <= 0 {
			return
		}
		if t == nil {
			t = time.NewTimer(a.cfg.FlushInterval)
		} else {
			if !t.Stop() {
				select {
				case <-t.C:
				default:
				}
			}
			t.Reset(a.cfg.FlushInterval)
		}
	}
	for {
		if a.cfg.FlushInterval > 0 && t == nil {
			t = time.NewTimer(a.cfg.FlushInterval)
		}
		select {
		case e, ok := <-a.q:
			if !ok {
				flush()
				return
			}
			batch = append(batch, e)
			if len(batch) >= a.cfg.BatchSize {
				flush()
				resetTimer()
			}
		case <-func() <-chan time.Time {
			if t != nil {
				return t.C
			}
			return make(chan time.Time)
		}():
			flush()
			resetTimer()
		}
	}
}
