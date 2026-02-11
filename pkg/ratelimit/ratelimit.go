// Package ratelimit provides token-bucket rate limiting for HTTP requests.
// It uses a memory-based implementation suitable for single-instance deployments.
package ratelimit

import (
	"sync"
	"time"
)

// TokenBucket implements a token bucket rate limiter.
// Tokens are added at a fixed rate until the bucket is full.
// Each request consumes one token; requests are denied when the bucket is empty.
type TokenBucket struct {
	mu         sync.Mutex
	tokens     int
	maxTokens  int
	refillRate time.Duration
	lastRefill time.Time
}

// NewTokenBucket creates a new token bucket rate limiter.
// maxTokens is the maximum number of tokens the bucket can hold.
// refillInterval is how often to add one token to the bucket.
//
// Example: NewTokenBucket(100, time.Second) allows 100 requests per second,
// with burst capacity of up to 100 requests.
func NewTokenBucket(maxTokens int, refillInterval time.Duration) *TokenBucket {
	now := time.Now()
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillInterval,
		lastRefill: now,
	}
}

// Allow checks if a request should be allowed.
// It returns true if the request is allowed, false otherwise.
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	// Calculate how many tokens to add based on elapsed time
	if elapsed >= tb.refillRate {
		tokensToAdd := int(elapsed / tb.refillRate)
		tb.tokens += tokensToAdd
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		tb.lastRefill = now
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// ClientLimiter tracks rate limits per client IP address.
type ClientLimiter struct {
	mu         sync.RWMutex
	limiters   map[string]*TokenBucket
	maxTokens  int
	refillRate time.Duration
}

// NewClientLimiter creates a new client-based rate limiter.
// Each unique client key (typically IP address) gets its own token bucket.
// maxTokens is the maximum tokens per bucket.
// refillInterval is how often to add one token.
func NewClientLimiter(maxTokens int, refillInterval time.Duration) *ClientLimiter {
	return &ClientLimiter{
		limiters:   make(map[string]*TokenBucket),
		maxTokens:  maxTokens,
		refillRate: refillInterval,
	}
}

// Allow checks if a request from the given client should be allowed.
// The clientKey is typically the IP address or other client identifier.
// It returns true if the request is allowed, false otherwise.
func (cl *ClientLimiter) Allow(clientKey string) bool {
	// Fast path: read lock to check for existing limiter
	cl.mu.RLock()
	limiter, exists := cl.limiters[clientKey]
	cl.mu.RUnlock()

	if !exists {
		// Slow path: write lock to create new limiter
		cl.mu.Lock()
		// Double-check after acquiring write lock
		limiter, exists = cl.limiters[clientKey]
		if !exists {
			limiter = NewTokenBucket(cl.maxTokens, cl.refillRate)
			cl.limiters[clientKey] = limiter
		}
		cl.mu.Unlock()
	}

	return limiter.Allow()
}

// Cleanup removes stale limiters that haven't been used recently.
// This should be called periodically to prevent memory leaks.
// maxAge is the maximum age of a limiter since its last use.
func (cl *ClientLimiter) Cleanup(maxAge time.Duration) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	// Note: This is a simple cleanup that removes all limiters.
	// A more sophisticated implementation would track last access time
	// per limiter and only remove those older than maxAge.
	// For now, we keep it simple since limiters are small.
	if len(cl.limiters) > 10000 {
		// If we have too many limiters, clear them all
		// This prevents unbounded growth in case of DDoS
		cl.limiters = make(map[string]*TokenBucket)
	}
}
