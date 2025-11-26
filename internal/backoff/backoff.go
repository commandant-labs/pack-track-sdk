package backoff

import (
	"math"
	"math/rand"
	"time"
)

// Strategy computes a backoff duration from an attempt number.
type Strategy interface {
	Next(attempt int) time.Duration
}

// Exponential implements exponential backoff with jitter.
type Exponential struct {
	Initial time.Duration
	Max     time.Duration
	Jitter  float64 // 0..1
}

func (e Exponential) Next(attempt int) time.Duration {
	if e.Initial <= 0 {
		e.Initial = 100 * time.Millisecond
	}
	if e.Max <= 0 {
		e.Max = 2 * time.Second
	}
	base := float64(e.Initial) * math.Pow(2, float64(attempt))
	if base > float64(e.Max) {
		base = float64(e.Max)
	}
	j := e.Jitter
	if j < 0 {
		j = 0
	} else if j > 1 {
		j = 1
	}
	if j == 0 {
		return time.Duration(base)
	}
	min := base * (1 - j)
	max := base * (1 + j)
	return time.Duration(min + rand.Float64()*(max-min))
}
