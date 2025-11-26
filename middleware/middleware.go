package middleware

import "context"

// LogHandler is the terminal function that emits a log entry.
type LogHandler func(ctx context.Context, level int, message string, fields map[string]any) error

// LogMiddleware composes behavior around a LogHandler.
type LogMiddleware func(next LogHandler) LogHandler

// MetricHandler is the terminal function that records a metric.
type MetricHandler func(ctx context.Context, name string, value float64, labels map[string]string) error

// MetricMiddleware composes behavior around a MetricHandler.
type MetricMiddleware func(next MetricHandler) MetricHandler
