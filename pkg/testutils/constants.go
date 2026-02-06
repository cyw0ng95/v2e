package testutils

import "time"

const (
	// Test Levels
	TestLevel1 = 1 // Basic logic, no external dependencies
	TestLevel2 = 2 // Database operations
	TestLevel3 = 3 // External APIs, E2E tests

	// Timeout Constants
	DefaultTimeout  = 5 * time.Second
	DatabaseTimeout = 10 * time.Second
	APITimeout      = 30 * time.Second

	// Retry Configuration
	MaxRetries = 3
	RetryDelay = 100 * time.Millisecond

	// Database
	DefaultDBPath = ":memory:"
	TestDBPrefix  = "/tmp/test_"
)
