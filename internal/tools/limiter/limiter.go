package limiter

import (
	"sync"
	"time"
)

type Limiter interface {
	Allow() bool
}
type RateLimiter struct {
	mu        sync.Mutex
	times     []time.Time
	threshold int           // max notifications allowed
	interval  time.Duration // sliding window duration
}

func NewRateLimiter(threshold int, interval time.Duration) *RateLimiter {
	return &RateLimiter{
		times:     make([]time.Time, 0, threshold),
		threshold: threshold,
		interval:  interval,
	}
}

func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.interval)

	// Remove outdated timestamps
	i := 0
	for _, t := range r.times {
		if t.After(cutoff) {
			break
		}
		i++
	}
	r.times = r.times[i:]

	if len(r.times) >= r.threshold {
		return false
	}

	// Add the new timestamp
	r.times = append(r.times, now)
	return true
}
