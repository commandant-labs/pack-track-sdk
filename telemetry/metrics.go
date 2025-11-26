package telemetry

import (
	"context"
	"time"
)

// Type is the metric type.
type Type int

const (
	Gauge Type = iota
	Counter
	Histogram
)

// Labels carries metric labels/tags.
type Labels map[string]string

// Metric represents a single metric data point.
type Metric struct {
	Time   time.Time
	Name   string
	Type   Type
	Value  float64
	Unit   string
	Labels Labels
}

// Recorder records metrics.
type Recorder interface {
	Record(ctx context.Context, m Metric) error
}
