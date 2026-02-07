package provider

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// BaseProvider implements DataSourceProvider with common functionality
// All data source providers should embed this struct
type BaseProvider struct {
	*fsm.BaseProviderFSM

	config      *ProviderConfig
	progress    *ProviderProgress
	rateLimiter *RateLimiter
	mu          sync.RWMutex
	cancelFunc  context.CancelFunc
	ctx         context.Context
}

// NewBaseProvider creates a new BaseProvider with given configuration
func NewBaseProvider(config *ProviderConfig, storage *storage.Store) (*BaseProvider, error) {
	if config == nil {
		config = DefaultProviderConfig()
	}

	baseFSM, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           config.DataType,
		ProviderType: config.DataType,
		Storage:      storage,
		Executor:     nil, // Will be set by concrete providers
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create base provider FSM: %w", err)
	}

	return &BaseProvider{
		BaseProviderFSM: baseFSM,
		config:          config,
		progress:        &ProviderProgress{},
		rateLimiter:     NewRateLimiter(config.RateLimitPermits),
	}, nil
}

// Initialize sets up the provider context
func (p *BaseProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancelFunc = context.WithCancel(ctx)
	return nil
}

// GetProgress returns a copy of current progress metrics
func (p *BaseProvider) GetProgress() *ProviderProgress {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &ProviderProgress{
		Fetched:     atomic.LoadInt64(&p.progress.Fetched),
		Stored:      atomic.LoadInt64(&p.progress.Stored),
		Failed:      atomic.LoadInt64(&p.progress.Failed),
		LastFetchAt: p.progress.LastFetchAt,
		LastStoreAt: p.progress.LastStoreAt,
		FetchRate:   p.progress.FetchRate,
		StoreRate:   p.progress.StoreRate,
	}
}

// GetConfig returns the provider configuration
func (p *BaseProvider) GetConfig() *ProviderConfig {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.config
}

// incrementFetched atomically increments the fetched counter
func (p *BaseProvider) incrementFetched(count int64) {
	atomic.AddInt64(&p.progress.Fetched, count)
	p.progress.LastFetchAt = time.Now()
	p.updateFetchRate()
}

// incrementStored atomically increments the stored counter
func (p *BaseProvider) incrementStored(count int64) {
	atomic.AddInt64(&p.progress.Stored, count)
	p.progress.LastStoreAt = time.Now()
	p.updateStoreRate()
}

// incrementFailed atomically increments the failed counter
func (p *BaseProvider) incrementFailed() {
	atomic.AddInt64(&p.progress.Failed, 1)
}

// updateFetchRate calculates the current fetch rate
func (p *BaseProvider) updateFetchRate() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.progress.LastFetchAt.IsZero() {
		return
	}

	duration := time.Since(p.progress.LastFetchAt)
	if duration > 0 {
		p.progress.FetchRate = float64(p.progress.Fetched) / duration.Seconds()
	}
}

// updateStoreRate calculates the current store rate
func (p *BaseProvider) updateStoreRate() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.progress.LastStoreAt.IsZero() {
		return
	}

	duration := time.Since(p.progress.LastStoreAt)
	if duration > 0 {
		p.progress.StoreRate = float64(p.progress.Stored) / duration.Seconds()
	}
}

// waitForPermit waits for a permit from the rate limiter
func (p *BaseProvider) waitForPermit(ctx context.Context) error {
	return p.rateLimiter.Wait(ctx)
}

// executeWithErrorHandling executes a function with error handling and retry logic
func (p *BaseProvider) executeWithErrorHandling(operation string, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < p.config.MaxRetries; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Log error
		p.incrementFailed()

		// Check if we should retry
		providerErr, ok := err.(*ProviderError)
		if !ok || !providerErr.Retryable {
			return err
		}

		// Wait before retry
		if attempt < p.config.MaxRetries-1 {
			select {
			case <-p.ctx.Done():
				return fmt.Errorf("provider canceled during retry: %w", p.ctx.Err())
			case <-time.After(p.config.RetryDelay):
			}
		}
	}

	return fmt.Errorf("operation '%s' failed after %d attempts: %w", operation, p.config.MaxRetries, lastErr)
}

// Cleanup releases resources held by the provider
func (p *BaseProvider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}

// checkContext checks if the provider context is canceled
func (p *BaseProvider) checkContext() error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.ctx != nil {
		select {
		case <-p.ctx.Done():
			return fmt.Errorf("provider context canceled: %w", p.ctx.Err())
		default:
		}
	}
	return nil
}
