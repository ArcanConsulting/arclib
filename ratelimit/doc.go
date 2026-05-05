// Package ratelimit provides rate limiting implementations.
//
// Two algorithms are provided:
//
//   - TokenBucket: a classic token bucket that refills at a steady rate,
//     allowing bursts up to the bucket capacity.
//   - SlidingWindow: a per-key sliding window counter that tracks request
//     timestamps and enforces a maximum request count within a time window.
//
// Both implementations are safe for concurrent use.
package ratelimit
