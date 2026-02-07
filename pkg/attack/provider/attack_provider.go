package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// ATTACKProvider implements DataSourceProvider for ATT&CK data
type ATTACKProvider struct {
	config     *provider.ProviderConfig
	rateLimiter *provider.RateLimiter
	progress   *provider.ProviderProgress
	cancelFunc  context.CancelFunc
	mu          sync.RWMutex
	ctx         context.Context
	localPath   string
	rpcClient   *rpc.Client
}

// NewATTACKProvider creates a new ATT&CK provider
// localPath is the path to ATT&CK Excel file
func NewATTACKProvider(localPath string) (*ATTACKProvider, error) {
	baseConfig := provider.DefaultProviderConfig()
	baseConfig.Name = "ATT&CK"
	baseConfig.DataType = "ATT&CK"
	baseConfig.LocalPath = localPath
	baseConfig.BatchSize = 100
	baseConfig.MaxRetries = 3
	baseConfig.RetryDelay = 5 * time.Second
	baseConfig.RateLimitPermits = 10

	if localPath == "" {
		localPath = "assets/enterprise-attack.xlsx"
	}

	return &ATTACKProvider{
		config:      baseConfig,
		rateLimiter: provider.NewRateLimiter(baseConfig.RateLimitPermits),
		progress:     &provider.ProviderProgress{},
		rpcClient:   &rpc.Client{},
		localPath:    localPath,
	}, nil
}

// Initialize sets up the ATT&CK provider context
func (p *ATTACKProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancelFunc = context.WithCancel(ctx)
	return nil
}

// GetID returns the provider ID
func (p *ATTACKProvider) GetID() string {
	return p.config.DataType
}

// GetType returns the provider type
func (p *ATTACKProvider) GetType() string {
	return p.config.DataType
}

// GetState returns current state as a string
func (p *ATTACKProvider) GetState() string {
	return "IDLE"
}

// Start begins provider execution
func (p *ATTACKProvider) Start() error {
	return nil
}

// Pause pauses provider execution
func (p *ATTACKProvider) Pause() error {
	return nil
}

// Stop terminates provider execution
func (p *ATTACKProvider) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}

// OnQuotaRevoked handles quota revocation
func (p *ATTACKProvider) OnQuotaRevoked(revokedCount int) error {
	return nil
}

// OnQuotaGranted handles quota grant
func (p *ATTACKProvider) OnQuotaGranted(grantedCount int) error {
	return nil
}

// OnRateLimited handles rate limiting
func (p *ATTACKProvider) OnRateLimited(retryAfter time.Duration) error {
	return nil
}

// Execute performs the fetch operation
func (p *ATTACKProvider) Execute() error {
	return p.Fetch(p.ctx)
}

// Fetch reads and parses ATT&CK data from local Excel file
func (p *ATTACKProvider) Fetch(ctx context.Context) error {
	atomic.StoreInt64(&p.progress.Fetched, 0)

	data, err := os.ReadFile(p.localPath)
	if err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to read ATT&CK file: %w", err)
	}

	var attackList []attack.ATTACKItem
	if err := json.Unmarshal(data, &attackList); err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to unmarshal ATT&CK data: %w", err)
	}

	count := len(attackList)
	atomic.StoreInt64(&p.progress.Fetched, int64(count))
	p.progress.LastFetchAt = time.Now()

	return nil
}

// Store stores fetched ATT&CK data using local service RPC
func (p *ATTACKProvider) Store(ctx context.Context) error {
	atomic.StoreInt64(&p.progress.Stored, 0)
	atomic.StoreInt64(&p.progress.Failed, 0)

	data, err := os.ReadFile(p.localPath)
	if err != nil {
		return fmt.Errorf("failed to read ATT&CK file: %w", err)
	}

	var attackList []attack.ATTACKItem
	if err := json.Unmarshal(data, &attackList); err != nil {
		return fmt.Errorf("failed to unmarshal ATT&CK data: %w", err)
	}

	stored := 0
	for i, attackItem := range attackList {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if i > 0 && i%p.config.BatchSize == 0 {
			time.Sleep(1 * time.Second)
		}

		atomic.StoreInt64(&p.progress.Stored, int64(stored))

		if err := p.rateLimiter.Wait(ctx); err != nil {
			return err
		}

		attackItemJSON, err := json.Marshal(attackItem)
		if err != nil {
			atomic.AddInt64(&p.progress.Failed, 1)
			continue
		}

		// TODO: Use RPCStoreAttack to store each item
		atomic.StoreInt64(&p.progress.Stored, int64(stored))

	p.progress.LastStoreAt = time.Now()
	p.progress.StoreRate = float64(stored) / time.Since(p.progress.LastFetchAt).Seconds()

	return nil
}

// GetProgress returns a copy of current progress metrics
func (p *ATTACKProvider) GetProgress() *provider.ProviderProgress {
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

// GetConfig returns provider configuration
func (p *ATTACKProvider) GetConfig() *provider.ProviderConfig {
	return p.config
}

// Cleanup releases resources held by the ATT&CK provider
func (p *ATTACKProvider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}
