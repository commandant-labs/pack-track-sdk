package packtrack

import (
	"time"
)

// Severity is the event severity level.
type Severity string

const (
	SeverityDebug Severity = "debug"
	SeverityInfo  Severity = "info"
	SeverityWarn  Severity = "warn"
	SeverityError Severity = "error"
)

// Status represents high-level execution status for an event.
type Status string

const (
	StatusOK       Status = "ok"
	StatusFailed   Status = "failed"
	StatusSkipped  Status = "skipped"
	StatusTimeout  Status = "timeout"
	StatusRetrying Status = "retrying"
	StatusUnknown  Status = "unknown"
)

// Source identifies the system emitting the event.
type Source struct {
	System string `json:"system"`
	Env    string `json:"env,omitempty"`
}

// Workflow identifies a workflow context for the event.
type Workflow struct {
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	RunID  string `json:"run_id,omitempty"`
	StepID string `json:"step_id,omitempty"`
}

// Actor identifies who/what produced the event.
type Actor struct {
	Type        string `json:"type"`
	ID          string `json:"id"`
	DisplayName string `json:"display_name,omitempty"`
}

// Event is the ingestion payload for PackTrack.
type Event struct {
	Timestamp time.Time      `json:"timestamp"`
	Source    Source         `json:"source"`
	Workflow  Workflow       `json:"workflow"`
	Actor     Actor          `json:"actor"`
	Severity  Severity       `json:"severity"`
	Status    Status         `json:"status"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Extra     map[string]any `json:"extra,omitempty"` // future-proof extensions
}
