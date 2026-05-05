package ratelimit

import (
	"sync"
	"time"
)

// SlidingWindow implements a per-key sliding window rate limiter.
// It tracks request timestamps and enforces a maximum number of requests
// within a rolling time window.
type SlidingWindow struct {
	window      time.Duration
	maxRequests int
	timestamps  map[string][]time.Time
	mu          sync.Mutex
}

// NewSlidingWindow creates a new SlidingWindow that allows at most maxRequests
// within the given window duration, tracked independently per key.
func NewSlidingWindow(window time.Duration, maxRequests int) *SlidingWindow {
	return &SlidingWindow{
		window:      window,
		maxRequests: maxRequests,
		timestamps:  make(map[string][]time.Time),
	}
}

// Allow checks whether a request for the given key is allowed.
// If allowed, it records the request and returns true.
func (sw *SlidingWindow) Allow(key string) bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	sw.timestamps[key] = pruneExpired(sw.timestamps[key], cutoff)

	if len(sw.timestamps[key]) >= sw.maxRequests {
		return false
	}

	sw.timestamps[key] = append(sw.timestamps[key], now)
	return true
}

// Count returns the number of requests recorded for the given key
// within the current window.
func (sw *SlidingWindow) Count(key string) int {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	cutoff := time.Now().Add(-sw.window)
	sw.timestamps[key] = pruneExpired(sw.timestamps[key], cutoff)
	return len(sw.timestamps[key])
}

// Cleanup removes all expired entries from all keys. Keys with no remaining
// timestamps are deleted entirely.
func (sw *SlidingWindow) Cleanup() {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	cutoff := time.Now().Add(-sw.window)
	for key, stamps := range sw.timestamps {
		pruned := pruneExpired(stamps, cutoff)
		if len(pruned) == 0 {
			delete(sw.timestamps, key)
		} else {
			sw.timestamps[key] = pruned
		}
	}
}

// pruneExpired removes timestamps that are at or before the cutoff.
func pruneExpired(stamps []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for _, t := range stamps {
		if t.After(cutoff) {
			stamps[i] = t
			i++
		}
	}
	return stamps[:i]
}
