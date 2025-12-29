package internal

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type rateLimiter struct {
	mu    sync.Mutex
	store map[string]*rateEntry
	rps   float64
	burst float64
	ttl   time.Duration
}

type rateEntry struct {
	tokens float64
	last   time.Time
}

func NewRateLimitHandler(next http.Handler, rps int64, burst int64, ttl time.Duration) http.Handler {
	if rps <= 0 {
		return next
	}
	limiter := &rateLimiter{
		store: make(map[string]*rateEntry),
		rps:   float64(rps),
		burst: float64(burst),
		ttl:   ttl,
	}
	if limiter.burst <= 0 {
		limiter.burst = limiter.rps
		if limiter.burst < 1 {
			limiter.burst = 1
		}
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.allow(clientIP(r)) {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (l *rateLimiter) allow(key string) bool {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.store[key]
	if !ok {
		l.store[key] = &rateEntry{tokens: l.burst - 1, last: now}
		return true
	}

	elapsed := now.Sub(entry.last).Seconds()
	entry.tokens += elapsed * l.rps
	if entry.tokens > l.burst {
		entry.tokens = l.burst
	}
	entry.last = now

	if entry.tokens < 1 {
		return false
	}
	entry.tokens -= 1
	return true
}

func clientIP(r *http.Request) string {
	if r == nil {
		return ""
	}
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		parts := strings.Split(fwd, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}
	if ip := r.Header.Get("X-Real-Ip"); ip != "" {
		return ip
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
