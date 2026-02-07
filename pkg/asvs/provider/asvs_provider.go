package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/asvs"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
)

// ASVSProvider implements DataSourceProvider for ASVS data
type ASVSProvider struct {
	config      *provider.ProviderConfig
	rateLimiter *provider.RateLimiter
	progress    *provider.ProviderProgress
	cancelFunc  context.CancelFunc
	mu          sync.RWMutex
	ctx         context.Context
	csvURL      string
}

// NewASVSProvider creates a new ASVS provider
// csvURL is URL to ASVS CSV file
func NewASVSProvider(csvURL string) (*ASVSProvider, error) {
	baseConfig := provider.DefaultProviderConfig()
	baseConfig.Name = "ASVS"
	baseConfig.DataType = "ASVS"
	baseConfig.CSVURL = csvURL
	baseConfig.BatchSize = 100
	baseConfig.MaxRetries = 3
	baseConfig.RetryDelay = 5 * time.Second
	baseConfig.RateLimitPermits = 10

	if csvURL == "" {
		csvURL = "https://raw.githubusercontent.com/OWASP/ASVS/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv"
	}

	return &ASVSProvider{
		config:      baseConfig,
		rateLimiter: provider.NewRateLimiter(baseConfig.RateLimitPermits),
		progress:    &provider.ProviderProgress{},
		csvURL:      csvURL,
	}, nil
}

// Initialize sets up the ASVS provider context
func (p *ASVSProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancelFunc = context.WithCancel(ctx)
	return nil
}

// GetID returns the provider ID
func (p *ASVSProvider) GetID() string {
	return p.config.DataType
}

// GetType returns the provider type
func (p *ASVSProvider) GetType() string {
	return p.config.DataType
}

// GetState returns the current state as a string
func (p *ASVSProvider) GetState() string {
	return "IDLE"
}

// Start begins provider execution
func (p *ASVSProvider) Start() error {
	return nil
}

// Pause pauses provider execution
func (p *ASVSProvider) Pause() error {
	return nil
}

// Stop terminates provider execution
func (p *ASVSProvider) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}

// OnQuotaRevoked handles quota revocation
func (p *ASVSProvider) OnQuotaRevoked(revokedCount int) error {
	return nil
}

// OnQuotaGranted handles quota grant
func (p *ASVSProvider) OnQuotaGranted(grantedCount int) error {
	return nil
}

// OnRateLimited handles rate limiting
func (p *ASVSProvider) OnRateLimited(retryAfter time.Duration) error {
	return nil
}

// Execute performs fetch and store operations
func (p *ASVSProvider) Execute() error {
	return p.Fetch(p.ctx)
}

// Fetch downloads ASVS CSV file from URL
func (p *ASVSProvider) Fetch(ctx context.Context) error {
	atomic.StoreInt64(&p.progress.Fetched, 0)

	resp, err := http.Get(p.csvURL)
	if err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to download ASVS CSV: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to read ASVS CSV: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		atomic.AddInt64(&p.progress.Fetched, 0)
		return fmt.Errorf("ASVS CSV file is empty")
	}

	atomic.StoreInt64(&p.progress.Fetched, int64(len(lines)))
	p.progress.LastFetchAt = time.Now()

	return nil
}

// Store stores fetched ASVS data using local service RPC
func (p *ASVSProvider) Store(ctx context.Context) error {
	return fmt.Errorf("ASVS store not yet implemented")
}

// GetProgress returns a copy of current progress metrics
func (p *ASVSProvider) GetProgress() *provider.ProviderProgress {
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

// GetConfig returns provider configuration
func (p *ASVSProvider) GetConfig() *provider.ProviderConfig {
	return p.config
}

// Cleanup releases resources held by ASVS provider
func (p *ASVSProvider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}
