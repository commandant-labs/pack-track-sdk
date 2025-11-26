package logging

import (
	"context"
	"time"
)

// Level represents log severity.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)

// Fields carries structured log fields.
type Fields map[string]any

// Event describes a structured log event.
type Event struct {
	Time    time.Time
	Level   Level
	Message string
	Fields  Fields
	TraceID string
	SpanID  string
}

// Logger is an optional logging-only surface that SDK users may use directly.
type Logger interface {
	Log(ctx context.Context, ev Event) error
	// With returns a logger that always includes the given fields.
	With(fields Fields) Logger
}
