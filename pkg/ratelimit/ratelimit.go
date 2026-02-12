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
	if maxTokens <= 0 || refillInterval <= 0 {
		return &TokenBucket{
			tokens:     1,
			maxTokens:  1,
			refillRate: time.Second,
			lastRefill: time.Now(),
		}
	}
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

// AllowWithRetryAfter checks if a request should be allowed and returns retry-after duration.
// If allowed, retryAfter is 0. If denied, retryAfter indicates time until next token is available.
func (tb *TokenBucket) AllowWithRetryAfter() (allowed bool, retryAfter time.Duration) {
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
		return true, 0
	}

	// Calculate time until next token is available
	timeSinceLastRefill := now.Sub(tb.lastRefill)
	retryAfter = tb.refillRate - timeSinceLastRefill
	if retryAfter < 0 {
		retryAfter = 0
	}

	return false, retryAfter
}

// ClientLimiter tracks rate limits per client IP address.
type ClientLimiter struct {
	mu         sync.RWMutex
	limiters   map[string]*TokenBucket
	lastAccess map[string]time.Time
	maxTokens  int
	refillRate time.Duration
}

// NewClientLimiter creates a new client-based rate limiter.
// Each unique client key (typically IP address) gets its own token bucket.
// maxTokens is the maximum tokens per bucket.
// refillInterval is how often to add one token.
func NewClientLimiter(maxTokens int, refillInterval time.Duration) *ClientLimiter {
	if maxTokens <= 0 || refillInterval <= 0 {
		return &ClientLimiter{
			limiters:   make(map[string]*TokenBucket),
			lastAccess: make(map[string]time.Time),
			maxTokens:  1,
			refillRate: time.Second,
		}
	}
	return &ClientLimiter{
		limiters:   make(map[string]*TokenBucket),
		lastAccess: make(map[string]time.Time),
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
		cl.lastAccess[clientKey] = time.Now()
		cl.mu.Unlock()
	} else {
		cl.mu.Lock()
		cl.lastAccess[clientKey] = time.Now()
		cl.mu.Unlock()
	}

	return limiter.Allow()
}

// AllowWithRetryAfter checks if a request from the given client should be allowed.
// Returns whether allowed and time until next token is available if denied.
func (cl *ClientLimiter) AllowWithRetryAfter(clientKey string) (bool, time.Duration) {
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
		cl.lastAccess[clientKey] = time.Now()
		cl.mu.Unlock()
	} else {
		cl.mu.Lock()
		cl.lastAccess[clientKey] = time.Now()
		cl.mu.Unlock()
	}

	return limiter.AllowWithRetryAfter()
}

// Cleanup removes stale limiters that haven't been used recently.
// This should be called periodically to prevent memory leaks.
// maxAge is the maximum age of a limiter since its last use.
func (cl *ClientLimiter) Cleanup(maxAge time.Duration) {
	cl.mu.Lock()
	defer cl.mu.Unlock()

	now := time.Now()
	for key, lastAccess := range cl.lastAccess {
		if now.Sub(lastAccess) > maxAge {
			delete(cl.limiters, key)
			delete(cl.lastAccess, key)
		}
	}
}
