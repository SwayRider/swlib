// Package jwtkeys provides a hardened, configurable cache of JWT
// verification public keys fetched from authservice.
//
// It refreshes on a configurable interval, bounds every fetch with a
// timeout so a stuck downstream call can never wedge the refresh loop,
// recovers from panics in both the fetch and the loop itself, retains the
// last known-good keys across a transient failure, and logs when the cache
// has been stale for longer than its refresh interval.
package jwtkeys

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/swayrider/swlib/jwt"
	log "github.com/swayrider/swlib/logger"
)

// PublicKeyFetcher is satisfied structurally by *grpcclients/authclient.Client
// (its PublicKeys method has a value receiver, so both Client and *Client
// qualify) — no import of grpcclients is needed here.
type PublicKeyFetcher interface {
	PublicKeys() ([]string, error)
}

// Defaults applied by New and used until Configure is called.
const (
	DefaultRefreshInterval = 5 * time.Minute
	DefaultFetchTimeout    = 15 * time.Second
)

// Cache holds the current set of JWT verification public keys and keeps
// them refreshed in the background.
type Cache struct {
	mu           sync.RWMutex
	keys         []string
	lastSuccess  time.Time
	staleLogged  bool
	lastStaleLog time.Time

	refreshInterval time.Duration
	refreshTimeout  time.Duration

	l *log.Logger
}

// New constructs a Cache with default tuning and no fetcher. It is
// intentionally constructible synchronously with nothing but a logger, so
// its GetPublicKeys method can be handed to a consumer (e.g.
// app.NewGrpcConfig) immediately, before config is parsed and before any
// fetcher (e.g. an authservice client resolved via a service framework) is
// available. Call Configure to tune cadence before the refresh loop starts,
// and pass the fetcher to Run or Start once it exists.
func New(l *log.Logger) *Cache {
	return &Cache{
		l:               l.Derive(log.WithComponent("jwtkeys")),
		refreshInterval: DefaultRefreshInterval,
		refreshTimeout:  DefaultFetchTimeout,
	}
}

// Configure tunes the refresh cadence and per-fetch timeout. Safe to call
// at any time; zero or negative values are ignored (the current value is
// kept). Must be called before Run/Start for the new refreshInterval to
// apply to the loop's ticker.
func (c *Cache) Configure(refreshInterval, fetchTimeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if refreshInterval > 0 {
		c.refreshInterval = refreshInterval
	}
	if fetchTimeout > 0 {
		c.refreshTimeout = fetchTimeout
	}
}

// Run performs a synchronous first refresh, then blocks running the refresh
// loop until ctx is cancelled. Intended for callers that already occupy a
// dedicated, tracked goroutine (e.g. a background-routine framework) and
// want Run to own that goroutine until shutdown.
func (c *Cache) Run(ctx context.Context, fetcher PublicKeyFetcher) {
	c.refresh(fetcher)
	c.runLoop(ctx, fetcher)
}

// Start performs a synchronous first refresh, then runs the refresh loop in
// its own detached goroutine and returns immediately. Intended for bespoke
// bootstraps that must continue past this call synchronously.
func (c *Cache) Start(ctx context.Context, fetcher PublicKeyFetcher) {
	c.refresh(fetcher)
	go c.runLoop(ctx, fetcher)
}

// runLoop is the outer supervisor: it runs the ticker loop and, if that
// loop panics, logs, backs off, and resumes — all within this same
// goroutine, so a caller tracking this goroutine (e.g. via a WaitGroup)
// sees it exit exactly once, when ctx is cancelled.
func (c *Cache) runLoop(ctx context.Context, fetcher PublicKeyFetcher) {
	for {
		if c.tick(ctx, fetcher) {
			return
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Minute):
		}
	}
}

// tick runs the ticker select-loop until ctx is cancelled (returns true) or
// it recovers from a panic (returns false; runLoop backs off and resumes).
func (c *Cache) tick(ctx context.Context, fetcher PublicKeyFetcher) (stopped bool) {
	defer func() {
		if r := recover(); r != nil {
			c.l.Errorf("public key refresh loop panicked, resuming after a delay: %v", r)
			stopped = false
		}
	}()
	ticker := time.NewTicker(c.getRefreshInterval())
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.refresh(fetcher)
		case <-ctx.Done():
			return true
		}
	}
}

type keysResult struct {
	keys []string
	err  error
}

// refresh performs one fetch, bounded by the configured fetch timeout via
// select/time.After so a stuck downstream call can't wedge the loop; the
// fetch itself runs in a child goroutine with its own panic recovery. On
// success it replaces the cached keys and resets staleness tracking. On
// failure or an empty result it calls noteStale and leaves the cached keys
// UNTOUCHED — a transient failure never clears previously-cached good keys.
func (c *Cache) refresh(fetcher PublicKeyFetcher) {
	lg := c.l.Derive(log.WithFunction("refresh"))

	ch := make(chan keysResult, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				ch <- keysResult{err: fmt.Errorf("panic in public key fetch: %v", r)}
			}
		}()
		keys, err := fetcher.PublicKeys()
		ch <- keysResult{keys: keys, err: err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			lg.Warnf("failed to refresh public keys: %v", res.err)
			c.noteStale()
			return
		}
		if len(res.keys) == 0 {
			lg.Warnln("authservice returned no public keys")
			c.noteStale()
			return
		}
		c.mu.Lock()
		c.keys = res.keys
		c.lastSuccess = time.Now()
		c.staleLogged = false
		c.mu.Unlock()
		lg.Infof("refreshed %d public key(s)", len(res.keys))
	case <-time.After(c.getFetchTimeout()):
		lg.Warnf("public key refresh timed out after %s; will retry", c.getFetchTimeout())
		c.noteStale()
	}
}

// noteStale logs once when time-since-last-success exceeds the configured
// refresh interval, then re-logs at most hourly while still stale.
func (c *Cache) noteStale() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.lastSuccess.IsZero() && time.Since(c.lastSuccess) < c.refreshInterval {
		return
	}
	now := time.Now()
	if !c.staleLogged || now.Sub(c.lastStaleLog) >= time.Hour {
		c.staleLogged = true
		c.lastStaleLog = now
		c.l.Errorf(
			"public keys have not refreshed successfully for over %s; JWT verification may fail until it recovers",
			c.refreshInterval)
	}
}

func (c *Cache) getRefreshInterval() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.refreshInterval
}

func (c *Cache) getFetchTimeout() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.refreshTimeout
}

// Keys returns a snapshot of the current public keys (PEM-encoded).
func (c *Cache) Keys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, len(c.keys))
	copy(out, c.keys)
	return out
}

// GetPublicKeys adapts Keys to the func() ([]string, error) shape expected
// by swlib/security.PublicKeysFn. It returns an error on an empty cache
// (e.g. before the first successful refresh completes).
func (c *Cache) GetPublicKeys() ([]string, error) {
	keys := c.Keys()
	if len(keys) == 0 {
		return nil, errors.New("no public keys available")
	}
	return keys, nil
}

// Verify tries each cached key in turn, supporting rotation overlap.
func (c *Cache) Verify(token string) (*jwt.Claims, error) {
	keys := c.Keys()
	if len(keys) == 0 {
		return nil, errors.New("no public keys available")
	}
	var lastErr error
	for _, key := range keys {
		claims, err := jwt.VerifyToken(token, key, jwt.VerifyDefault)
		if err == nil {
			return claims, nil
		}
		lastErr = err
	}
	return nil, lastErr
}
