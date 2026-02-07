package provider

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/provider"
	"github.com/cyw0ng95/v2e/pkg/ssg/remote"
)

// SSGProvider implements DataSourceProvider for SSG data
type SSGProvider struct {
	config     *provider.ProviderConfig
	rateLimiter *provider.RateLimiter
	progress   *provider.ProviderProgress
	cancelFunc  context.CancelFunc
	mu          sync.RWMutex
	ctx         context.Context
	gitClient *remote.GitClient
}

// NewSSGProvider creates a new SSG provider
// repoURL is the Git repository URL
func NewSSGProvider(repoURL string) (*SSGProvider, error) {
	baseConfig := provider.DefaultProviderConfig()
	baseConfig.Name = "SSG"
	baseConfig.DataType = "SSG"
	baseConfig.BatchSize = 10
	baseConfig.MaxRetries = 3
	baseConfig.RetryDelay = 10 * time.Second
	baseConfig.RateLimitPermits = 5

	if repoURL == "" {
		repoURL = "https://github.com/OWASP/wg-ssg"
	}

	return &SSGProvider{
		config:      baseConfig,
		rateLimiter: provider.NewRateLimiter(baseConfig.RateLimitPermits),
		progress:     &provider.ProviderProgress{},
		gitClient:    remote.NewGitClient(repoURL, ""),
	}, nil
}

// Initialize sets up the SSG provider context
func (p *SSGProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancelFunc = context.WithCancel(ctx)
	return nil
}

// GetID returns the provider ID
func (p *SSGProvider) GetID() string {
	return p.config.DataType
}

// GetType returns the provider type
func (p *SSGProvider) GetType() string {
	return p.config.DataType
}

// GetState returns the current state as a string
func (p *SSGProvider) GetState() string {
	return "IDLE"
}

// Start begins provider execution
func (p *SSGProvider) Start() error {
	return nil
}

// Pause pauses provider execution
func (p *SSGProvider) Pause() error {
	return nil
}

// Stop terminates provider execution
func (p *SSGProvider) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}

// OnQuotaRevoked handles quota revocation
func (p *SSGProvider) OnQuotaRevoked(revokedCount int) error {
	return nil
}

// OnQuotaGranted handles quota grant
func (p *SSGProvider) OnQuotaGranted(grantedCount int) error {
	return nil
}

// OnRateLimited handles rate limiting
func (p *SSGProvider) OnRateLimited(retryAfter time.Duration) error {
	return nil
}

// Execute performs the fetch operation
func (p *SSGProvider) Execute() error {
	return p.Fetch(p.ctx)
}

// Fetch performs a Git pull to get latest SSG data
func (p *SSGProvider) Fetch(ctx context.Context) error {
	// Reset counter
	atomic.StoreInt64(&p.progress.Fetched, 0)

	err := p.gitClient.Pull()
	if err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to pull SSG repository: %w", err)
	}

	guideFiles, err := p.gitClient.ListGuideFiles()
	if err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to list guide files: %w", err)
	}

	count := len(guideFiles)
	atomic.StoreInt64(&p.progress.Fetched, int64(count))
	p.progress.LastFetchAt = time.Now()

	return nil
}

// Store stores fetched SSG data using local service RPC
// TODO: Implement store logic using RPC calls to ssg service
func (p *SSGProvider) Store(ctx context.Context) error {
	return fmt.Errorf("SSG store not yet implemented")
}

// GetProgress returns a copy of current progress metrics
func (p *SSGProvider) GetProgress() *provider.ProviderProgress {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return &provider.ProviderProgress{
		Fetched:      atomic.LoadInt64(&p.progress.Fetched),
			Stored:       atomic.LoadInt64(&p.progress.Stored),
		Failed:       atomic.LoadInt64(&p.progress.Failed),
		LastFetchAt:  p.progress.LastFetchAt,
		LastStoreAt:  p.progress.LastStoreAt,
		FetchRate:    p.progress.FetchRate,
		StoreRate:    p.progress.StoreRate,
	}
}

// GetConfig returns the provider configuration
func (p *SSGProvider) GetConfig() *provider.ProviderConfig {
	return p.config
}

// Cleanup releases resources held by the SSG provider
func (p *SSGProvider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}
