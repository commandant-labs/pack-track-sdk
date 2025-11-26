package retry

import "time"

// Policy describes a retry backoff policy.
type Policy interface {
	// NextBackoff returns the delay for the given attempt (0-based).
	NextBackoff(attempt int) time.Duration
	// MaxAttempts returns the maximum attempts before giving up.
	MaxAttempts() int
}
