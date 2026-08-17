package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_AllowsWithinBurstThenBlocks(t *testing.T) {
	l := New(1, 3, time.Minute)

	for i := range 3 {
		if !l.Allow("key-a") {
			t.Fatalf("request %d: expected allowed within burst", i)
		}
	}
	if l.Allow("key-a") {
		t.Error("expected request beyond burst to be blocked")
	}
}

func TestLimiter_RefillsOverTime(t *testing.T) {
	l := New(1000, 1, time.Minute) // ~1ms per token refill

	if !l.Allow("key-a") {
		t.Fatal("expected first request to be allowed")
	}
	if l.Allow("key-a") {
		t.Fatal("expected immediate second request to be blocked (burst exhausted)")
	}

	time.Sleep(10 * time.Millisecond)
	if !l.Allow("key-a") {
		t.Error("expected request to be allowed after refill window")
	}
}

func TestLimiter_KeysAreIsolated(t *testing.T) {
	l := New(1, 1, time.Minute)

	if !l.Allow("key-a") {
		t.Fatal("expected key-a's first request to be allowed")
	}
	if l.Allow("key-a") {
		t.Fatal("expected key-a's second request to be blocked")
	}
	if !l.Allow("key-b") {
		t.Error("expected key-b to have its own independent bucket")
	}
}

func TestLimiter_Evict_RemovesOnlyIdleEntries(t *testing.T) {
	l := New(1, 1, 50*time.Millisecond)

	l.Allow("idle-key")
	time.Sleep(60 * time.Millisecond)
	l.Allow("fresh-key")

	l.Evict()

	l.mu.Lock()
	_, idleStillPresent := l.buckets["idle-key"]
	_, freshStillPresent := l.buckets["fresh-key"]
	l.mu.Unlock()

	if idleStillPresent {
		t.Error("expected idle-key's bucket to be evicted")
	}
	if !freshStillPresent {
		t.Error("expected fresh-key's bucket to survive eviction")
	}
}
