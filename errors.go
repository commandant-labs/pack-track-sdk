package packtrack

import (
	"errors"
	"fmt"
)

// IngestError is a typed error for ingestion failures.
type IngestError struct {
	StatusCode int    // HTTP status code if available
	Body       string // Response body (capped by caller)
	Retryable  bool   // Whether retrying might succeed
	Cause      error  // Underlying error cause
}

func (e *IngestError) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.StatusCode > 0 {
		return fmt.Sprintf("ingest error: status=%d retryable=%t: %s", e.StatusCode, e.Retryable, e.Body)
	}
	return fmt.Sprintf("ingest transport error: retryable=%t: %v", e.Retryable, e.Cause)
}

// Unwrap returns the underlying cause for errors.Is/As.
func (e *IngestError) Unwrap() error { return e.Cause }

// JoinIngestError joins an error into an IngestError cause chain.
func JoinIngestError(e *IngestError, err error) *IngestError {
	if e == nil {
		return &IngestError{Cause: err}
	}
	e.Cause = errors.Join(e.Cause, err)
	return e
}
