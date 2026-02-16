// Package ratelimit provides weight-based rate limiting for exchange APIs.
// This package is specifically designed for Binance's weight-based rate limiting system.
package ratelimit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// Binance default rate limits (weight-based)
// Documentation: https://binance-docs.github.io/apidocs/spot/en/#limits
// Last verified: 2026-02-16
const (
	// DefaultMaxWeight is the default maximum weight per minute for Binance
	// (1200 weight per minute for most endpoints)
	DefaultMaxWeight = 1200
	// RefillInterval is how often the weight bucket refills
	RefillInterval = time.Minute
)

// LimiterStats contains statistics about the rate limiter state.
type LimiterStats struct {
	CurrentWeight int           `json:"current_weight"`
	MaxWeight     int           `json:"max_weight"`
	Available     int           `json:"available"`
	WaitTime      time.Duration `json:"wait_time"`
}

// WeightedLimiter implements weight-based rate limiting.
// This is designed for Binance's API which uses weight-based rate limiting
// instead of simple request counting.
//
// Key features:
//   - Tracks current weight usage via X-MBX-USED-WEIGHT-* headers
//   - Uses token bucket for smooth rate limiting
//   - Supports blocking (Wait) and non-blocking (Allow) operations
//   - Thread-safe for concurrent use
type WeightedLimiter struct {
	maxWeight     int64
	currentWeight atomic.Int64
	limiter       *rate.Limiter
	mu            sync.RWMutex
}

// NewWeightedLimiter creates a new weight-based rate limiter.
// maxWeight is the maximum weight allowed per interval (default 1200 for Binance).
// The limiter uses a token bucket that refills at a rate of maxWeight per minute.
func NewWeightedLimiter(maxWeight int) *WeightedLimiter {
	if maxWeight <= 0 {
		maxWeight = DefaultMaxWeight
	}

	wl := &WeightedLimiter{
		maxWeight: int64(maxWeight),
	}
	wl.currentWeight.Store(0)

	// Create a limiter that allows maxWeight tokens per minute
	// This provides smooth rate limiting rather than bursty minute boundaries
	// rate = maxWeight / 60 (tokens per second)
	tokensPerSecond := float64(maxWeight) / 60.0
	wl.limiter = rate.NewLimiter(rate.Limit(tokensPerSecond), maxWeight)

	return wl
}

// Wait blocks until weight is available or context is cancelled.
// This is the recommended method for rate limiting as it handles backpressure
// automatically and respects context cancellation.
//
// Returns nil if the request can proceed, or ctx.Err() if the context
// was cancelled before weight became available.
func (wl *WeightedLimiter) Wait(ctx context.Context, weight int) error {
	if weight <= 0 {
		return nil
	}

	// Reserve the weight from the token bucket
	return wl.limiter.WaitN(ctx, weight)
}

// Allow checks if weight is available without blocking.
// Returns true if the request can proceed immediately, false otherwise.
// Use this for non-critical requests that can be skipped when rate limited.
func (wl *WeightedLimiter) Allow(weight int) bool {
	if weight <= 0 {
		return true
	}

	return wl.limiter.AllowN(time.Now(), weight)
}

// UpdateWeight updates the current weight usage from server response.
// This should be called with the X-MBX-USED-WEIGHT-* header value
// from each Binance API response to keep the limiter in sync with
// the server's view of weight usage.
//
// The server-provided weight is used for monitoring and metrics,
// while the token bucket handles actual rate limiting.
func (wl *WeightedLimiter) UpdateWeight(weight int) {
	if weight >= 0 {
		wl.currentWeight.Store(int64(weight))
	}
}

// CurrentWeight returns the current weight usage as reported by the server.
// This reflects the X-MBX-USED-WEIGHT-* header value from the last response.
func (wl *WeightedLimiter) CurrentWeight() int {
	return int(wl.currentWeight.Load())
}

// MaxWeight returns the maximum allowed weight per interval.
func (wl *WeightedLimiter) MaxWeight() int {
	return int(wl.maxWeight)
}

// WaitTime returns the estimated time to wait for the given weight.
// This is useful for logging and monitoring rate limit pressure.
func (wl *WeightedLimiter) WaitTime(weight int) time.Duration {
	if weight <= 0 {
		return 0
	}

	// Calculate how long until weight is available
	now := time.Now()
	r := wl.limiter.ReserveN(now, weight)
	if !r.OK() {
		// Request is too large, would never succeed
		return -1
	}

	delay := r.DelayFrom(now)
	r.CancelAt(now) // Don't actually consume the reservation
	return delay
}

// Stats returns current statistics about the rate limiter.
func (wl *WeightedLimiter) Stats() LimiterStats {
	current := wl.CurrentWeight()
	waitTime := wl.WaitTime(1)

	return LimiterStats{
		CurrentWeight: current,
		MaxWeight:     wl.MaxWeight(),
		Available:     wl.MaxWeight() - current,
		WaitTime:      waitTime,
	}
}

// Reset resets the limiter to its initial state.
// This is useful after extended periods of inactivity or when
// reconnecting to the exchange.
func (wl *WeightedLimiter) Reset() {
	wl.currentWeight.Store(0)
	// Recreate the limiter to reset the token bucket
	tokensPerSecond := float64(wl.maxWeight) / 60.0
	wl.mu.Lock()
	wl.limiter = rate.NewLimiter(rate.Limit(tokensPerSecond), int(wl.maxWeight))
	wl.mu.Unlock()
}
