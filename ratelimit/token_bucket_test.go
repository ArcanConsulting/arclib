package ratelimit

import (
	"sync"
	"testing"
	"time"
)

func TestNewTokenBucket(t *testing.T) {
	tb := NewTokenBucket(10.0, 5)
	if tb.rate != 10.0 {
		t.Errorf("expected rate 10.0, got %f", tb.rate)
	}
	if tb.capacity != 5.0 {
		t.Errorf("expected capacity 5.0, got %f", tb.capacity)
	}
	if tb.Available() != 5 {
		t.Errorf("expected 5 available tokens, got %d", tb.Available())
	}
}

func TestTokenBucketAllow(t *testing.T) {
	tb := NewTokenBucket(10.0, 3)

	// Should allow first 3 requests (bucket starts full).
	for i := range 3 {
		if !tb.Allow() {
			t.Errorf("expected Allow() to succeed on call %d", i+1)
		}
	}

	// Bucket is now empty; next call should fail.
	if tb.Allow() {
		t.Error("expected Allow() to fail when bucket is empty")
	}
}

func TestTokenBucketAllowN(t *testing.T) {
	tb := NewTokenBucket(10.0, 10)

	if !tb.AllowN(5) {
		t.Error("expected AllowN(5) to succeed")
	}
	if tb.Available() != 5 {
		t.Errorf("expected 5 tokens remaining, got %d", tb.Available())
	}

	// Try to consume more than available.
	if tb.AllowN(6) {
		t.Error("expected AllowN(6) to fail with only 5 tokens")
	}

	// Exact remaining should work.
	if !tb.AllowN(5) {
		t.Error("expected AllowN(5) to succeed with exactly 5 tokens")
	}
}

func TestTokenBucketRefill(t *testing.T) {
	tb := NewTokenBucket(100.0, 10)

	// Drain the bucket.
	tb.AllowN(10)
	if tb.Available() != 0 {
		t.Fatalf("expected 0 tokens, got %d", tb.Available())
	}

	// Simulate time passing by moving lastRefill back.
	tb.mu.Lock()
	tb.lastRefill = tb.lastRefill.Add(-100 * time.Millisecond)
	tb.mu.Unlock()

	// At rate=100 tokens/sec, 100ms = 10 tokens, capped at capacity 10.
	avail := tb.Available()
	if avail < 9 || avail > 10 {
		t.Errorf("expected ~10 tokens after refill, got %d", avail)
	}
}

func TestTokenBucketCapacity(t *testing.T) {
	tb := NewTokenBucket(1000.0, 5)

	// Even after a long time, tokens should not exceed capacity.
	tb.mu.Lock()
	tb.lastRefill = tb.lastRefill.Add(-10 * time.Second)
	tb.mu.Unlock()

	avail := tb.Available()
	if avail != 5 {
		t.Errorf("expected tokens capped at 5, got %d", avail)
	}
}

func TestTokenBucketConcurrency(t *testing.T) {
	tb := NewTokenBucket(1000.0, 100)

	var wg sync.WaitGroup
	allowed := make(chan bool, 200)

	for range 200 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- tb.Allow()
		}()
	}

	wg.Wait()
	close(allowed)

	trueCount := 0
	for v := range allowed {
		if v {
			trueCount++
		}
	}

	// With capacity 100 and high rate, we should get at least 100 allowed.
	// The exact number depends on timing but must not exceed what's possible.
	if trueCount < 100 {
		t.Errorf("expected at least 100 allowed, got %d", trueCount)
	}
}
