package provider

import (
	"context"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
)

// DataSourceProvider defines the interface for all data source providers
// Each data source (CVE, CWE, CAPEC, ATT&CK, SSG, ASVS) must implement this interface
type DataSourceProvider interface {
	fsm.ProviderFSM

	// Initialize provider-specific resources (clients, connections, etc.)
	Initialize(ctx context.Context) error

	// Fetch data from remote source or local file
	// This method should handle rate limiting, retries, and error recovery
	Fetch(ctx context.Context) error

	// Process and store fetched data
	// This method should handle batch operations and error handling
	Store(ctx context.Context) error

	// Get current progress metrics for monitoring
	GetProgress() *ProviderProgress

	// Get configuration for this provider
	GetConfig() *ProviderConfig

	// Cleanup resources when provider is stopped
	Cleanup(ctx context.Context) error
}

// ProviderProgress represents progress metrics for a data source provider
type ProviderProgress struct {
	// Fetched is the number of items fetched from the source
	Fetched int64

	// Stored is the number of items successfully stored
	Stored int64

	// Failed is the number of items that failed to process
	Failed int64

	// LastFetchAt is the timestamp of the last successful fetch
	LastFetchAt time.Time

	// LastStoreAt is the timestamp of the last successful store operation
	LastStoreAt time.Time

	// FetchRate is the current fetch rate (items/second)
	FetchRate float64

	// StoreRate is the current store rate (items/second)
	StoreRate float64
}

// ProviderConfig holds configuration for a data source provider
type ProviderConfig struct {
	// Name is the human-readable name of the provider (e.g., "CVE", "CWE")
	Name string

	// DataType is the type of data this provider handles
	DataType string

	// BaseURL is the base URL for remote data sources (empty for local sources)
	BaseURL string

	// APIKey is the optional API key for authenticated requests
	APIKey string

	// LocalPath is the path to local data files (for local sources)
	LocalPath string

	// BatchSize is the number of items to process in each batch
	BatchSize int

	// MaxRetries is the maximum number of retry attempts for failed operations
	MaxRetries int

	// RetryDelay is the delay between retry attempts
	RetryDelay time.Duration

	// RateLimitPermits is the number of permits for rate limiting
	RateLimitPermits int

	// FetchInterval is the interval between fetch operations (for polling-based sources)
	FetchInterval time.Duration

	// Timeout is the timeout for individual operations
	Timeout time.Duration
}

// DefaultProviderConfig returns a sensible default configuration
func DefaultProviderConfig() *ProviderConfig {
	return &ProviderConfig{
		BatchSize:        100,
		MaxRetries:       3,
		RetryDelay:       5 * time.Second,
		RateLimitPermits: 10,
		FetchInterval:    time.Hour,
		Timeout:          30 * time.Second,
	}
}

// ProviderError represents an error that occurred during provider operations
type ProviderError struct {
	// ProviderID is the ID of the provider that encountered the error
	ProviderID string

	// Operation is the operation being performed (e.g., "fetch", "store")
	Operation string

	// Err is the underlying error
	Err error

	// Retryable indicates whether the operation can be retried
	Retryable bool

	// Timestamp is when the error occurred
	Timestamp time.Time
}

// Error wraps an error with provider context
func (e *ProviderError) Error() string {
	return e.Err.Error()
}

// Unwrap returns the underlying error
func (e *ProviderError) Unwrap() error {
	return e.Err
}

// NewProviderError creates a new ProviderError
func NewProviderError(providerID, operation string, err error, retryable bool) *ProviderError {
	return &ProviderError{
		ProviderID: providerID,
		Operation:  operation,
		Err:        err,
		Retryable:  retryable,
		Timestamp:  time.Now(),
	}
}
