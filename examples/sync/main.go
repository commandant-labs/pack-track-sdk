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
	)
	if err != nil {
		log.Fatal(err)
	}
	ev := packtrack.Event{
		Timestamp: time.Now().UTC(),
		Source:    packtrack.Source{System: "example"},
		Workflow:  packtrack.Workflow{ID: "wf-1"},
		Actor:     packtrack.Actor{Type: "agent", ID: "example"},
		Severity:  packtrack.SeverityInfo,
		Status:    packtrack.StatusSuccess,
		Message:   "hello from sync example",
	}
	ctx := context.Background()
	if _, err := c.IngestEvent(ctx, ev); err != nil {
		log.Fatal(err)
	}
}
