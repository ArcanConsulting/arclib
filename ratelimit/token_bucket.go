package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implements the token bucket rate limiting algorithm.
// Tokens are added at a fixed rate up to a maximum capacity.
// Each allowed request consumes one or more tokens.
type TokenBucket struct {
	rate       float64
	capacity   float64
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket creates a new TokenBucket that adds tokens at the given rate
// (tokens per second) up to the specified capacity.
func NewTokenBucket(rate float64, capacity int) *TokenBucket {
	return &TokenBucket{
		rate:       rate,
		capacity:   float64(capacity),
		tokens:     float64(capacity),
		lastRefill: time.Now(),
	}
}

// Allow consumes one token and returns true if a token was available.
func (tb *TokenBucket) Allow() bool {
	return tb.AllowN(1)
}

// AllowN consumes n tokens and returns true if enough tokens were available.
func (tb *TokenBucket) AllowN(n int) bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()

	requested := float64(n)
	if tb.tokens < requested {
		return false
	}
	tb.tokens -= requested
	return true
}

// Available returns the current number of available tokens (truncated to int).
func (tb *TokenBucket) Available() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.refill()
	return int(tb.tokens)
}

// refill adds tokens based on elapsed time since the last refill.
// Must be called with tb.mu held.
func (tb *TokenBucket) refill() {
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	if elapsed <= 0 {
		return
	}

	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastRefill = now
}
