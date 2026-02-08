package provider

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
)

// CAPECProvider implements DataSourceProvider for CAPEC data
type CAPECProvider struct {
	config       *provider.ProviderConfig
	rateLimiter  *provider.RateLimiter
	progress     *provider.ProviderProgress
	cancelFunc   context.CancelFunc
	mu           sync.RWMutex
	ctx          context.Context
	localPath    string
	eventHandler func(*fsm.Event) error
}

// NewCAPECProvider creates a new CAPEC provider
// localPath is the path to CAPEC XML file
func NewCAPECProvider(localPath string) (*CAPECProvider, error) {
	baseConfig := provider.DefaultProviderConfig()
	baseConfig.Name = "CAPEC"
	baseConfig.DataType = "CAPEC"
	baseConfig.LocalPath = localPath
	baseConfig.BatchSize = 50
	baseConfig.MaxRetries = 3
	baseConfig.RetryDelay = 5 * time.Second
	baseConfig.RateLimitPermits = 10

	if localPath == "" {
		localPath = "assets/capec_contents_latest.xml"
	}

	return &CAPECProvider{
		config:      baseConfig,
		rateLimiter: provider.NewRateLimiter(baseConfig.RateLimitPermits),
		progress:    &provider.ProviderProgress{},
		localPath:   localPath,
	}, nil
}

// Initialize sets up the CAPEC provider context
func (p *CAPECProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancelFunc = context.WithCancel(ctx)
	return nil
}

// GetID returns provider ID
func (p *CAPECProvider) GetID() string {
	return p.config.DataType
}

// GetType returns provider type
func (p *CAPECProvider) GetType() string {
	return p.config.DataType
}

// GetState returns the current state as a string
func (p *CAPECProvider) GetState() fsm.ProviderState {
	return fsm.ProviderIdle
}

// Start begins provider execution
func (p *CAPECProvider) Start() error {
	return nil
}

// Pause pauses provider execution
func (p *CAPECProvider) Pause() error {
	return nil
}

// Resume resumes provider execution
func (p *CAPECProvider) Resume() error {
	return nil
}

// SetEventHandler sets the callback for event bubbling to MacroFSM
func (p *CAPECProvider) SetEventHandler(handler func(*fsm.Event) error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.eventHandler = handler
}

// Transition attempts to transition to a new state
func (p *CAPECProvider) Transition(newState fsm.ProviderState) error {
	// TODO: Implement state transition logic with validation
	return nil
}

// Stop terminates provider execution
func (p *CAPECProvider) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}

// OnQuotaRevoked handles quota revocation
func (p *CAPECProvider) OnQuotaRevoked(revokedCount int) error {
	return nil
}

// OnQuotaGranted handles quota grant
func (p *CAPECProvider) OnQuotaGranted(grantedCount int) error {
	return nil
}

// OnRateLimited handles rate limiting
func (p *CAPECProvider) OnRateLimited(retryAfter time.Duration) error {
	return nil
}

// Execute performs the fetch operation
func (p *CAPECProvider) Execute() error {
	return p.Fetch(p.ctx)
}

// Fetch reads and parses CAPEC data from local XML file
func (p *CAPECProvider) Fetch(ctx context.Context) error {
	atomic.StoreInt64(&p.progress.Fetched, 0)

	data, err := os.ReadFile(p.localPath)
	if err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to read CAPEC file: %w", err)
	}

	capecData := capec.Root{}
	if err := xml.Unmarshal(data, &capecData); err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to unmarshal CAPEC data: %w", err)
	}

	// Count attack patterns
	count := len(capecData.AttackPatterns.AttackPattern)
	atomic.StoreInt64(&p.progress.Fetched, int64(count))
	p.progress.LastFetchAt = time.Now()

	return nil
}

// Store stores fetched CAPEC data using local service RPC
func (p *CAPECProvider) Store(ctx context.Context) error {
	atomic.StoreInt64(&p.progress.Stored, 0)
	atomic.StoreInt64(&p.progress.Failed, 0)

	data, err := os.ReadFile(p.localPath)
	if err != nil {
		return fmt.Errorf("failed to read CAPEC file: %w", err)
	}

	capecData := capec.Root{}
	if err := xml.Unmarshal(data, &capecData); err != nil {
		return fmt.Errorf("failed to unmarshal CAPEC data: %w", err)
	}

	stored := 0
	for i, attackPattern := range capecData.AttackPatterns.AttackPattern {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if i > 0 && i%p.config.BatchSize == 0 {
			time.Sleep(1 * time.Second)
		}

		// TODO: Use RPCStoreCAPEC to store each item
		_, err := json.Marshal(attackPattern)
		if err != nil {
			atomic.AddInt64(&p.progress.Failed, 1)
			continue
		}

		if err := p.rateLimiter.Wait(ctx); err != nil {
			return err
		}

		stored++
		atomic.StoreInt64(&p.progress.Stored, int64(stored))
	}

	p.progress.LastStoreAt = time.Now()
	p.progress.StoreRate = float64(stored) / time.Since(p.progress.LastFetchAt).Seconds()

	return nil
}

// GetProgress returns a copy of current progress metrics
func (p *CAPECProvider) GetProgress() *provider.ProviderProgress {
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
func (p *CAPECProvider) GetConfig() *provider.ProviderConfig {
	return p.config
}

// Cleanup releases resources held by the CAPEC provider
func (p *CAPECProvider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}
