// Package ratelimit provides a coarse, per-key token-bucket rate limiter.
//
// It is deliberately independent of swlib/app so it can be used and tested
// on its own; callers drive eviction on their own schedule (e.g. via an
// app.BackgroundRoutine ticker) rather than the package owning a goroutine.
package ratelimit

import (
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// entry pairs a per-key token bucket with the last time it was used, so
// Evict can reclaim buckets for keys that have gone idle.
type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// Limiter is a per-key token-bucket rate limiter. The zero value is not
// usable; construct with New.
type Limiter struct {
	mu             sync.Mutex
	buckets        map[string]*entry
	requestsPerSec float64
	burst          int
	idleTTL        time.Duration
}

// New creates a Limiter allowing requestsPerSecond sustained requests per
// key, with a burst allowance of burst. Buckets idle for longer than idleTTL
// are eligible for removal by Evict.
func New(requestsPerSecond float64, burst int, idleTTL time.Duration) *Limiter {
	return &Limiter{
		buckets:        make(map[string]*entry),
		requestsPerSec: requestsPerSecond,
		burst:          burst,
		idleTTL:        idleTTL,
	}
}

// Allow reports whether a request for key may proceed, consuming one token
// from its bucket if so. A key's bucket is created lazily on first use.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	e, ok := l.buckets[key]
	if !ok {
		e = &entry{limiter: rate.NewLimiter(rate.Limit(l.requestsPerSec), l.burst)}
		l.buckets[key] = e
	}
	e.lastSeen = time.Now()
	return e.limiter.Allow()
}

// Evict removes buckets that have been idle for longer than idleTTL. It
// performs a single synchronous sweep; callers are expected to invoke it
// periodically (e.g. from a ticker) to bound memory under sustained,
// high-cardinality traffic.
func (l *Limiter) Evict() {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := time.Now().Add(-l.idleTTL)
	for key, e := range l.buckets {
		if e.lastSeen.Before(cutoff) {
			delete(l.buckets, key)
		}
	}
}
