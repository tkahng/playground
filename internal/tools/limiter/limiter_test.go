package limiter

import (
	"testing"
	"time"

	"github.com/tkahng/playground/internal/test"
)

func TestNewRateLimiter(t *testing.T) {
	test.Parallel(t)
	got := NewRateLimiter(10, time.Minute)
	if got == nil {
		t.Fatal("expected not to be nil")
	}
	if got.threshold != 10 {
		t.Fatalf("expected threshold to be %d, got %d", 10, got.threshold)
	}
	if got.interval != time.Minute {
		t.Fatalf("expected interval to be %v, got %v", time.Minute, got.interval)
	}
}

func TestRateLimiter_Allow(t *testing.T) {
	test.Parallel(t)
	t.Run("6 per minute", func(t *testing.T) {
		rateLimiter := NewRateLimiter(5, 10*time.Second)
		for range 5 {
			if !rateLimiter.Allow() {
				t.Fatal("expected to be allowed")
			}
		}
		if rateLimiter.Allow() {
			t.Fatal("expected to be not allowed")
		}
	})
}
