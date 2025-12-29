package internal

import (
	"testing"
	"time"
)

func TestRateLimiterAllow(t *testing.T) {
	limiter := &rateLimiter{
		store: make(map[string]*rateEntry),
		rps:   1,
		burst: 1,
	}

	if !limiter.allow("client") {
		t.Fatalf("expected first request to be allowed")
	}
	if limiter.allow("client") {
		t.Fatalf("expected second request to be rate limited")
	}

	time.Sleep(1100 * time.Millisecond)

	if !limiter.allow("client") {
		t.Fatalf("expected request after refill to be allowed")
	}
}
