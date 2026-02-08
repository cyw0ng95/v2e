package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/provider"
	proc "github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

// CWEProvider implements DataSourceProvider for CWE data
type CWEProvider struct {
	config       *provider.ProviderConfig
	rateLimiter  *provider.RateLimiter
	progress     *provider.ProviderProgress
	cancelFunc   context.CancelFunc
	mu           sync.RWMutex
	ctx          context.Context
	localPath    string
	rpcClient    *rpc.Client
	eventHandler func(*fsm.Event) error
}

// NewCWEProvider creates a new CWE provider
// localPath is the path to the CWE JSON file
func NewCWEProvider(localPath string) (*CWEProvider, error) {
	baseConfig := provider.DefaultProviderConfig()
	baseConfig.Name = "CWE"
	baseConfig.DataType = "CWE"
	baseConfig.LocalPath = localPath
	baseConfig.BatchSize = 100
	baseConfig.MaxRetries = 3
	baseConfig.RetryDelay = 5 * time.Second
	baseConfig.RateLimitPermits = 10

	if localPath == "" {
		localPath = "assets/cwe-raw.json"
	}

	return &CWEProvider{
		config:      baseConfig,
		rateLimiter: provider.NewRateLimiter(baseConfig.RateLimitPermits),
		progress:    &provider.ProviderProgress{},
		rpcClient:   &rpc.Client{},
		localPath:   localPath,
	}, nil
}

// Initialize sets up the CWE provider context
func (p *CWEProvider) Initialize(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.ctx, p.cancelFunc = context.WithCancel(ctx)
	return nil
}

// GetID returns the provider ID
func (p *CWEProvider) GetID() string {
	return p.config.DataType
}

// GetType returns the provider type
func (p *CWEProvider) GetType() string {
	return p.config.DataType
}

// GetState returns the current state as a string
func (p *CWEProvider) GetState() fsm.ProviderState {
	return fsm.ProviderIdle
}

// Start begins provider execution
func (p *CWEProvider) Start() error {
	return nil
}

// Pause pauses provider execution
func (p *CWEProvider) Pause() error {
	return nil
}

// Resume resumes provider execution
func (p *CWEProvider) Resume() error {
	return nil
}

// SetEventHandler sets the callback for event bubbling to MacroFSM
func (p *CWEProvider) SetEventHandler(handler func(*fsm.Event) error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.eventHandler = handler
}

// Transition attempts to transition to a new state
func (p *CWEProvider) Transition(newState fsm.ProviderState) error {
	// TODO: Implement state transition logic with validation
	return nil
}

// Stop terminates provider execution
func (p *CWEProvider) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}

// OnQuotaRevoked handles quota revocation
func (p *CWEProvider) OnQuotaRevoked(revokedCount int) error {
	return nil
}

// OnQuotaGranted handles quota grant
func (p *CWEProvider) OnQuotaGranted(grantedCount int) error {
	return nil
}

// OnRateLimited handles rate limiting
func (p *CWEProvider) OnRateLimited(retryAfter time.Duration) error {
	return nil
}

// Execute performs the fetch and store operations
func (p *CWEProvider) Execute() error {
	return p.Fetch(p.ctx)
}

// Fetch reads and parses CWE data from local JSON file
func (p *CWEProvider) Fetch(ctx context.Context) error {
	atomic.StoreInt64(&p.progress.Fetched, 0)

	data, err := os.ReadFile(p.localPath)
	if err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to read CWE file: %w", err)
	}

	var cweList []cwe.CWEItem
	if err := json.Unmarshal(data, &cweList); err != nil {
		atomic.AddInt64(&p.progress.Failed, 1)
		return fmt.Errorf("failed to unmarshal CWE data: %w", err)
	}

	count := len(cweList)
	atomic.StoreInt64(&p.progress.Fetched, int64(count))
	p.progress.LastFetchAt = time.Now()

	return nil
}

// Store stores fetched CWE data using local service RPC
func (p *CWEProvider) Store(ctx context.Context) error {
	atomic.StoreInt64(&p.progress.Stored, 0)
	atomic.StoreInt64(&p.progress.Failed, 0)

	data, err := os.ReadFile(p.localPath)
	if err != nil {
		return fmt.Errorf("failed to read CWE file: %w", err)
	}

	var cweList []cwe.CWEItem
	if err := json.Unmarshal(data, &cweList); err != nil {
		return fmt.Errorf("failed to unmarshal CWE data: %w", err)
	}

	stored := 0
	for i, cweItem := range cweList {
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

		cweItemData, err := json.Marshal(cweItem)
		if err != nil {
			atomic.AddInt64(&p.progress.Failed, 1)
			continue
		}

		msg, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCImportCWE", map[string]interface{}{
			"cweData": cweItemData,
		})

		if err != nil {
			atomic.AddInt64(&p.progress.Failed, 1)
			continue
		} else if msg.Type == proc.MessageTypeError {
			stored++
		} else {
			stored++
		}
	}

	p.progress.LastStoreAt = time.Now()
	p.progress.StoreRate = float64(stored) / time.Since(p.progress.LastFetchAt).Seconds()

	return nil
}

// GetProgress returns a copy of current progress metrics
func (p *CWEProvider) GetProgress() *provider.ProviderProgress {
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
func (p *CWEProvider) GetConfig() *provider.ProviderConfig {
	return p.config
}

// Cleanup releases resources held by the CWE provider
func (p *CWEProvider) Cleanup(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancelFunc != nil {
		p.cancelFunc()
	}

	return nil
}
