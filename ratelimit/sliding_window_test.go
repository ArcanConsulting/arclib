package ratelimit

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestNewSlidingWindow(t *testing.T) {
	sw := NewSlidingWindow(time.Minute, 10)
	if sw.window != time.Minute {
		t.Errorf("expected window of 1m, got %v", sw.window)
	}
	if sw.maxRequests != 10 {
		t.Errorf("expected maxRequests 10, got %d", sw.maxRequests)
	}
}

func TestSlidingWindowAllow(t *testing.T) {
	sw := NewSlidingWindow(time.Minute, 3)

	for i := range 3 {
		if !sw.Allow("user1") {
			t.Errorf("expected Allow to succeed on call %d", i+1)
		}
	}

	// Fourth request should be denied.
	if sw.Allow("user1") {
		t.Error("expected Allow to fail after maxRequests reached")
	}
}

func TestSlidingWindowPerKey(t *testing.T) {
	sw := NewSlidingWindow(time.Minute, 2)

	// Fill up user1.
	sw.Allow("user1")
	sw.Allow("user1")

	// user1 is exhausted.
	if sw.Allow("user1") {
		t.Error("expected user1 to be rate limited")
	}

	// user2 should be independent.
	if !sw.Allow("user2") {
		t.Error("expected user2 to be allowed independently")
	}
}

func TestSlidingWindowCount(t *testing.T) {
	sw := NewSlidingWindow(time.Minute, 10)

	sw.Allow("key1")
	sw.Allow("key1")
	sw.Allow("key1")

	if c := sw.Count("key1"); c != 3 {
		t.Errorf("expected count 3, got %d", c)
	}
	if c := sw.Count("key2"); c != 0 {
		t.Errorf("expected count 0 for unknown key, got %d", c)
	}
}

func TestSlidingWindowExpiry(t *testing.T) {
	sw := NewSlidingWindow(time.Second, 2)

	// Inject timestamps in the past (already expired).
	sw.mu.Lock()
	past := time.Now().Add(-2 * time.Second)
	sw.timestamps["user1"] = []time.Time{past, past}
	sw.mu.Unlock()

	// Despite having 2 entries, they are expired so Allow should succeed.
	if !sw.Allow("user1") {
		t.Error("expected Allow to succeed after entries expired")
	}
}

func TestSlidingWindowCleanup(t *testing.T) {
	sw := NewSlidingWindow(time.Second, 100)

	// Add expired entries for multiple keys.
	sw.mu.Lock()
	past := time.Now().Add(-2 * time.Second)
	sw.timestamps["expired1"] = []time.Time{past, past}
	sw.timestamps["expired2"] = []time.Time{past}
	sw.timestamps["active"] = []time.Time{time.Now()}
	sw.mu.Unlock()

	sw.Cleanup()

	sw.mu.Lock()
	defer sw.mu.Unlock()

	if _, exists := sw.timestamps["expired1"]; exists {
		t.Error("expected expired1 to be cleaned up")
	}
	if _, exists := sw.timestamps["expired2"]; exists {
		t.Error("expected expired2 to be cleaned up")
	}
	if _, exists := sw.timestamps["active"]; !exists {
		t.Error("expected active to remain after cleanup")
	}
}

func TestSlidingWindowConcurrency(t *testing.T) {
	sw := NewSlidingWindow(time.Minute, 50)

	var wg sync.WaitGroup
	allowed := make(chan bool, 100)

	for i := range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- sw.Allow(fmt.Sprintf("user%d", i%5))
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

	// 5 users, 50 max each, 100 total requests (20 per user) - all should pass.
	if trueCount != 100 {
		t.Errorf("expected all 100 requests allowed, got %d", trueCount)
	}
}
