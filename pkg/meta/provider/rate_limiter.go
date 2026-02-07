package provider

import (
	"context"
	"sync"
	"time"
)

// RateLimiter controls the rate at which operations can be performed
// It implements a token bucket algorithm for rate limiting
type RateLimiter struct {
	permits    int32
	maxPermits int32
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
// maxPermits is the maximum number of operations that can be performed before waiting
// refillRate is the time interval at which permits are refilled (e.g., 1 permit per second)
func NewRateLimiter(maxPermits int) *RateLimiter {
	return &RateLimiter{
		maxPermits: int32(maxPermits),
		permits:    int32(maxPermits),
		refillRate: time.Second / time.Duration(maxPermits),
		lastRefill: time.Now(),
	}
}

// Wait blocks until a permit is available or context is canceled
// If permits are immediately available, this returns without blocking
// Otherwise, it blocks until a permit is refilled or context is canceled
func (rl *RateLimiter) Wait(ctx context.Context) error {
	for {
		if rl.tryAcquire() {
			return nil
		}

		// Calculate when the next permit will be available
		nextRefill := rl.calculateNextRefill()
		waitDuration := time.Until(nextRefill)

		// Wait for refill or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
			// Refill permits
			rl.refill()
		}
	}
}

// tryAcquire attempts to acquire a permit without blocking
// Returns true if a permit was acquired, false otherwise
func (rl *RateLimiter) tryAcquire() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Refill permits if needed
	rl.refill()

	if rl.permits > 0 {
		rl.permits--
		return true
	}

	return false
}

// Acquire attempts to acquire a permit without blocking
// Returns immediately with true if a permit was acquired, false otherwise
// Unlike Wait, this method never blocks
func (rl *RateLimiter) Acquire() bool {
	return rl.tryAcquire()
}

// Release returns a permit back to the pool
// This is useful if an operation is aborted and the permit should be returned
func (rl *RateLimiter) Release() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.permits < rl.maxPermits {
		rl.permits++
	}
}

// refill adds permits based on the elapsed time since last refill
// This implements the token bucket algorithm
func (rl *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	if elapsed >= rl.refillRate {
		// Calculate how many permits to add
		permitsToAdd := int32(elapsed / rl.refillRate)

		// Add permits but don't exceed max
		rl.permits = minInt32(rl.permits+permitsToAdd, rl.maxPermits)
		rl.lastRefill = now
	}
}

// calculateNextRefill calculates when the next permit will be available
func (rl *RateLimiter) calculateNextRefill() time.Time {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if rl.permits > 0 {
		return time.Now()
	}

	// Calculate time until next permit
	timeUntilRefill := rl.refillRate
	return rl.lastRefill.Add(timeUntilRefill)
}

// AvailablePermits returns the current number of available permits
func (rl *RateLimiter) AvailablePermits() int {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return int(rl.permits)
}

// SetMaxPermits changes the maximum number of permits
func (rl *RateLimiter) SetMaxPermits(maxPermits int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.maxPermits = int32(maxPermits)

	// If current permits exceed new max, cap them
	if rl.permits > rl.maxPermits {
		rl.permits = rl.maxPermits
	}
}

// minInt32 returns the minimum of two int32 values
func minInt32(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}
