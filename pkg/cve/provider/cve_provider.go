package provider

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
)

// CVEProvider implements DataSourceProvider for CVE data
type CVEProvider struct {
	config      *provider.ProviderConfig
	rateLimiter *provider.RateLimiter
	progress    *provider.ProviderProgress
	cancelFunc  context.CancelFunc
	mu          sync.RWMutex
	ctx         context.Context
	fetcher     *remote.Fetcher
}

// NewCVEProvider creates a new CVE provider
// apiKey is optional NVD API key for higher rate limits
func NewCVEProvider(apiKey string) (*CVEProvider, error) {
	baseConfig := provider.DefaultProviderConfig()
	baseConfig.Name = "CVE"
	baseConfig.DataType = "CVE"
	baseConfig.BaseURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"
	baseConfig.APIKey = apiKey
	baseConfig.BatchSize = 2000
	baseConfig.MaxRetries = 3
	baseConfig.RetryDelay = 5 * time.Second
	baseConfig.RateLimitPermits = 10

	fetcher, err := remote.NewFetcher(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetcher: %w", err)
	}

	return &CVEProvider{
		config:      baseConfig,
		rateLimiter: provider.NewRateLimiter(baseConfig.RateLimitPermits),
		progress:    &provider.ProviderProgress{},
		fetcher:     fetcher,
	}, nil
}

// Initialize sets up the CVE provider context
func (p *CVEProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancelFunc = context.WithCancel(ctx)
	return nil
}

// GetID returns the provider ID
func (p *CVEProvider) GetID() string {
	return p.config.DataType
}

// GetType returns the provider type
func (p *CVEProvider) GetType() string {
	return p.config.DataType
}

// GetState returns the current state as a string
func (p *CVEProvider) GetState() string {
	return "IDLE"
}

// Start begins provider execution
func (p *CVEProvider) Start() error {
	return nil
}

// Pause pauses provider execution
func (p *CVEProvider) Pause() error {
	return nil
}

// Stop terminates provider execution
func (p *CVEProvider) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}

// OnQuotaRevoked handles quota revocation
func (p *CVEProvider) OnQuotaRevoked(revokedCount int) error {
	return nil
}

// OnQuotaGranted handles quota grant
func (p *CVEProvider) OnQuotaGranted(grantedCount int) error {
	return nil
}

// OnRateLimited handles rate limiting
func (p *CVEProvider) OnRateLimited(retryAfter time.Duration) error {
	return nil
}

// Execute performs the fetch operation
func (p *CVEProvider) Execute() error {
	return p.Fetch(p.ctx)
}

// Fetch retrieves CVEs from NVD API
// This method implements the fetch loop with rate limiting and retry logic
func (p *CVEProvider) Fetch(ctx context.Context) error {
	config := p.config
	startIndex := 0
	pageSize := config.BatchSize

	atomic.StoreInt64(&p.progress.Fetched, 0)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := p.rateLimiter.Wait(ctx); err != nil {
			return err
		}

		resp, err := p.fetcher.FetchCVEs(startIndex, pageSize)
		if err != nil {
			if err == remote.ErrRateLimited {
				time.Sleep(30 * time.Second)
				continue
			}

			atomic.AddInt64(&p.progress.Failed, 1)
			return fmt.Errorf("failed to fetch CVEs: %w", err)
		}

		count := len(resp.Vulnerabilities)
		atomic.AddInt64(&p.progress.Fetched, int64(count))
		p.progress.LastFetchAt = time.Now()

		if startIndex+count >= resp.TotalResults {
			break
		}

		startIndex += count
	}

	return nil
}

// Store stores fetched CVEs using local service
// TODO: Implement store logic using RPC calls to local service
func (p *CVEProvider) Store(ctx context.Context) error {
	return fmt.Errorf("CVE store not yet implemented")
}

// GetProgress returns a copy of current progress metrics
func (p *CVEProvider) GetProgress() *provider.ProviderProgress {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &provider.ProviderProgress{
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
func (p *CVEProvider) GetConfig() *provider.ProviderConfig {
	return p.config
}

// Cleanup releases resources held by the CVE provider
func (p *CVEProvider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}
