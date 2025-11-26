package backoff

import "time"

// Strategy computes a backoff duration from an attempt number.
type Strategy interface {
	Next(attempt int) time.Duration
}
