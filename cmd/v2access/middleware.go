package main

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/ratelimit"
	"github.com/gin-gonic/gin"
)

// Rate limit configuration constants
const (
	defaultMaxTokens    = 100               // Maximum requests per window
	defaultRefillRate   = time.Second       // Add one token per second
	defaultCleanupIntvl = time.Minute * 5   // Cleanup interval
	maxTokensPerClient  = 50                // Per-client burst limit
)

// RateLimiterConfig holds configuration for the rate limiter middleware
type RateLimiterConfig struct {
	// MaxTokens is the maximum number of tokens per client
	MaxTokens int
	// RefillInterval is how often to add one token
	RefillInterval time.Duration
	// CleanupInterval is how often to clean up stale limiters
	CleanupInterval time.Duration
	// TrustedProxies are CIDR ranges that are exempt from rate limiting
	TrustedProxies []string
	// ExcludedPaths are path prefixes that are excluded from rate limiting
	ExcludedPaths []string
}

// DefaultRateLimiterConfig returns the default rate limiter configuration
func DefaultRateLimiterConfig() *RateLimiterConfig {
	return &RateLimiterConfig{
		MaxTokens:       maxTokensPerClient,
		RefillInterval:  defaultRefillRate,
		CleanupInterval: defaultCleanupIntvl,
		TrustedProxies:  []string{"127.0.0.1", "::1"},
		ExcludedPaths:   []string{"/restful/health"},
	}
}

// RateLimiterMiddleware creates a Gin middleware for rate limiting
func RateLimiterMiddleware(config *RateLimiterConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimiterConfig()
	}

	limiter := ratelimit.NewClientLimiter(config.MaxTokens, config.RefillInterval)

	// Start cleanup goroutine
	go func() {
		ticker := time.NewTicker(config.CleanupInterval)
		defer ticker.Stop()
		for range ticker.C {
			limiter.Cleanup(config.CleanupInterval)
		}
	}()

	common.Info(LogMsgRateLimiterStarted, config.MaxTokens, config.RefillInterval)

	return func(c *gin.Context) {
		// Check if path is excluded
		for _, prefix := range config.ExcludedPaths {
			if strings.HasPrefix(c.Request.URL.Path, prefix) {
				c.Next()
				return
			}
		}

		// Get client IP
		clientIP := getClientIP(c)

		// Check if client is trusted proxy
		if isTrustedProxy(clientIP, config.TrustedProxies) {
			c.Next()
			return
		}

		// Check rate limit
		if !limiter.Allow(clientIP) {
			common.Warn(LogMsgRateLimitExceeded, clientIP)
			c.Header("X-RateLimit-Limit", strconv.Itoa(config.MaxTokens))
			c.Header("X-RateLimit-Refill", strconv.Itoa(int(config.RefillInterval.Seconds())))
			c.Header("Retry-After", strconv.Itoa(int(config.RefillInterval.Seconds())))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"retcode": 429,
				"message": "Rate limit exceeded. Please retry later.",
				"payload": nil,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first (for proxied requests)
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if idx := strings.Index(xff, ","); idx != -1 {
			return strings.TrimSpace(xff[:idx])
		}
		return strings.TrimSpace(xff)
	}

	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	return c.ClientIP()
}

// isTrustedProxy checks if the given IP is in the trusted proxy list
func isTrustedProxy(ip string, trustedProxies []string) bool {
	for _, trusted := range trustedProxies {
		if ip == trusted {
			return true
		}
	}
	return false
}
