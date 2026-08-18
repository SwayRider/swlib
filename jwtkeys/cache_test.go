package jwtkeys

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	log "github.com/swayrider/swlib/logger"
)

func TestMain(m *testing.M) {
	log.SetOutput(io.Discard)
	os.Exit(m.Run())
}

const (
	modeOK = iota
	modeErr
	modeBlock
	modePanic
)

// fakeFetcher simulates authclient behavior: success, error, a call that
// blocks forever, or a panic.
type fakeFetcher struct {
	mu    sync.Mutex
	mode  int
	keys  []string
	err   error
	block chan struct{}
	calls int
}

func (f *fakeFetcher) PublicKeys() ([]string, error) {
	f.mu.Lock()
	f.calls++
	mode := f.mode
	block := f.block
	keys, err := f.keys, f.err
	f.mu.Unlock()

	switch mode {
	case modeErr:
		return nil, err
	case modeBlock:
		<-block
		return keys, nil
	case modePanic:
		panic("boom")
	default:
		return keys, nil
	}
}

func (f *fakeFetcher) set(mode int, keys []string, err error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.mode = mode
	f.keys = keys
	f.err = err
}

// newTestCache returns a Cache configured with a short fetch timeout so
// tests exercising the blocking/timeout path stay fast. The refresh
// interval is also kept short but is irrelevant to most tests below, which
// drive refreshes directly via refresh(fetcher) rather than waiting on the
// ticker.
func newTestCache() *Cache {
	c := New(log.New())
	c.Configure(50*time.Millisecond, 50*time.Millisecond)
	return c
}

func TestCache_RefreshSuccess(t *testing.T) {
	f := &fakeFetcher{}
	f.set(modeOK, []string{"key-1"}, nil)
	c := newTestCache()
	c.Start(context.Background(), f)

	keys := c.Keys()
	if len(keys) != 1 || keys[0] != "key-1" {
		t.Errorf("Keys() = %v, want [key-1]", keys)
	}
}

func TestCache_RefreshErrorKeepsKeysAndRecovers(t *testing.T) {
	f := &fakeFetcher{}
	f.set(modeOK, []string{"key-1"}, nil)
	c := newTestCache()
	c.Start(context.Background(), f)

	// A failed refresh must not clear the cached keys.
	f.set(modeErr, nil, errors.New("authservice down"))
	c.refresh(f)
	if keys := c.Keys(); len(keys) != 1 || keys[0] != "key-1" {
		t.Errorf("Keys() = %v, want cached [key-1] preserved", keys)
	}

	// A later successful refresh replaces them.
	f.set(modeOK, []string{"key-2"}, nil)
	c.refresh(f)
	if keys := c.Keys(); len(keys) != 1 || keys[0] != "key-2" {
		t.Errorf("Keys() = %v, want [key-2]", keys)
	}
}

func TestCache_StuckCallDoesNotBlockRefresh(t *testing.T) {
	f := &fakeFetcher{}
	blocked := make(chan struct{})
	f.mu.Lock()
	f.mode = modeBlock
	f.keys = []string{"key-1"}
	f.block = blocked
	f.mu.Unlock()

	c := newTestCache()
	start := time.Now()
	c.Start(context.Background(), f) // initial refresh must time out, not hang
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("Start() blocked for %v; refresh timeout not honored", elapsed)
	}
	if keys := c.Keys(); len(keys) != 0 {
		t.Fatalf("Keys() = %v, want empty (fetch never completed)", keys)
	}

	// Unblock the stuck call; a later refresh must succeed.
	close(blocked)
	f.set(modeOK, []string{"key-2"}, nil)
	c.refresh(f)
	if keys := c.Keys(); len(keys) != 1 || keys[0] != "key-2" {
		t.Errorf("Keys() = %v, want [key-2] after recovery", keys)
	}
}

func TestCache_PanicInFetchDoesNotCrash(t *testing.T) {
	f := &fakeFetcher{}
	f.set(modePanic, nil, nil)
	c := newTestCache()
	c.Start(context.Background(), f) // must return, not crash the process

	if keys := c.Keys(); len(keys) != 0 {
		t.Fatalf("Keys() = %v, want empty after panicking fetch", keys)
	}

	// A later successful refresh must work.
	f.set(modeOK, []string{"key-3"}, nil)
	c.refresh(f)
	if keys := c.Keys(); len(keys) != 1 || keys[0] != "key-3" {
		t.Errorf("Keys() = %v, want [key-3] after recovery", keys)
	}
}

func TestCache_StaleKeysAreReported(t *testing.T) {
	f := &fakeFetcher{}
	f.set(modeOK, []string{"key-1"}, nil)
	c := newTestCache()
	c.Start(context.Background(), f)

	// Simulate a long stall: last successful refresh was hours ago, and the
	// next refresh fails.
	c.mu.Lock()
	c.lastSuccess = time.Now().Add(-2 * time.Hour)
	c.mu.Unlock()
	f.set(modeErr, nil, errors.New("down"))
	c.refresh(f)

	c.mu.RLock()
	stale := c.staleLogged
	c.mu.RUnlock()
	if !stale {
		t.Errorf("stale keys were never reported")
	}

	// Recovery clears the flag.
	f.set(modeOK, []string{"key-2"}, nil)
	c.refresh(f)
	c.mu.RLock()
	stale = c.staleLogged
	c.mu.RUnlock()
	if stale {
		t.Errorf("stale flag not cleared after recovery")
	}
}

func TestCache_GetPublicKeysErrorsWhenEmpty(t *testing.T) {
	c := New(log.New())
	if keys, err := c.GetPublicKeys(); err == nil {
		t.Fatalf("GetPublicKeys() = %v, nil; want an error on an empty cache", keys)
	}

	f := &fakeFetcher{}
	f.set(modeOK, []string{"key-1"}, nil)
	c.Configure(50*time.Millisecond, 50*time.Millisecond)
	c.Start(context.Background(), f)

	keys, err := c.GetPublicKeys()
	if err != nil || len(keys) != 1 || keys[0] != "key-1" {
		t.Errorf("GetPublicKeys() = %v, %v; want [key-1], nil", keys, err)
	}
}

// TestCache_ConfigureControlsTickerCadence asserts that Configure, called
// before Run/Start, actually changes the interval the refresh loop ticks
// on — Cache has no such knob in its predecessor, so this is new coverage
// for the New/Configure/Run split.
func TestCache_ConfigureControlsTickerCadence(t *testing.T) {
	f := &fakeFetcher{}
	f.set(modeOK, []string{"key-1"}, nil)
	c := New(log.New())
	c.Configure(20*time.Millisecond, 50*time.Millisecond)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Start(ctx, f)

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		calls := f.calls
		f.mu.Unlock()
		if calls >= 5 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}

	f.mu.Lock()
	calls := f.calls
	f.mu.Unlock()
	t.Fatalf("expected several ticks at the configured 20ms interval within 500ms, got %d calls", calls)
}
