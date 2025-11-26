package main

import (
	"context"
	"log"
	"time"

	packtrack "github.com/commandant-labs/pack-track-sdk"
)

func main() {
	c, err := packtrack.NewClient(
		packtrack.WithBaseURL("https://pack.shimcounty.com"),
		packtrack.WithAPIKey("YOUR_API_KEY"),
		packtrack.WithCompression(packtrack.CompressionGzip),
	)
	if err != nil {
		log.Fatal(err)
	}
	ac, err := packtrack.NewAsyncClient(c,
		packtrack.WithBatchSize(100),
		packtrack.WithFlushInterval(1*time.Second),
		packtrack.WithQueueCapacity(10000),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	defer ac.Close(ctx)
	for i := 0; i < 10; i++ {
		ev := packtrack.Event{
			Timestamp: time.Now().UTC(),
			Source:    packtrack.Source{System: "example"},
			Workflow:  packtrack.Workflow{ID: "wf-1", RunID: "run-1"},
			Actor:     packtrack.Actor{Type: "agent", ID: "example"},
			Severity:  packtrack.SeverityInfo,
			Status:    packtrack.StatusOK,
			Message:   "hello from async example",
		}

		if err := ac.Enqueue(ev); err != nil {
			log.Fatal(err)
		}
	}
	_ = ac.Flush(ctx)
}
