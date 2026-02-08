package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve/remote"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CVEProvider implements ProviderFSM for CVE data
type CVEProvider struct {
	*fsm.BaseProviderFSM
	fetcher    *remote.Fetcher
	batchSize  int
	maxRetries int
	retryDelay time.Duration
	apiKey     string
	mu         sync.RWMutex
}

// NewCVEProvider creates a new CVE provider with FSM support
// apiKey is optional NVD API key for higher rate limits
func NewCVEProvider(apiKey string, store *storage.Store) (*CVEProvider, error) {
	fetcher, err := remote.NewFetcher(apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create fetcher: %w", err)
	}

	provider := &CVEProvider{
		fetcher:    fetcher,
		apiKey:     apiKey,
		batchSize:  2000,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}

	// Create base FSM with custom executor
	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "cve",
		ProviderType: "cve",
		Storage:      store,
		Executor:     provider.execute,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// Initialize sets up the CVE provider context
func (p *CVEProvider) Initialize(ctx context.Context) error {
	// BaseProviderFSM doesn't need explicit initialization
	return nil
}

// execute performs the CVE fetch and store operations
func (p *CVEProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	p.mu.RLock()
	batchSize := p.batchSize
	fetcher := p.fetcher
	p.mu.RUnlock()

	startIndex := 0

	for {
		// Check for cancellation or state change
		if p.GetState() != fsm.ProviderRunning {
			break
		}

		// Fetch batch from NVD API
		resp, err := fetcher.FetchCVEs(startIndex, batchSize)
		if err != nil {
			// Handle rate limiting
			if err == remote.ErrRateLimited {
				if stateErr := p.OnRateLimited(30 * time.Second); stateErr != nil {
					return fmt.Errorf("rate limit handling failed: %w", stateErr)
				}
				// Backoff and retry
				time.Sleep(30 * time.Second)
				continue
			}

			// Handle other errors
			return fmt.Errorf("failed to fetch CVEs at index %d: %w", startIndex, err)
		}

		// Process each CVE in batch
		count := len(resp.Vulnerabilities)
		for _, vuln := range resp.Vulnerabilities {
			cveID := vuln.CVE.ID

			// Create URN for checkpointing
			itemURN := urn.MustParse(fmt.Sprintf("v2e::nvd::cve::%s", cveID))

			// Store CVE (TODO: implement actual storage via RPC)
			// For now, just simulate success
			success := true
			errorMsg := ""

			// Save checkpoint
			if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
				return fmt.Errorf("failed to save checkpoint for %s: %w", cveID, err)
			}
		}

		// Check if we've fetched all CVEs
		if startIndex+count >= resp.TotalResults {
			break
		}

		startIndex += count
	}

	return nil
}

// GetBatchSize returns the batch size for fetching
func (p *CVEProvider) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}

// SetBatchSize sets the batch size for fetching
func (p *CVEProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// GetAPIKey returns the NVD API key
func (p *CVEProvider) GetAPIKey() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.apiKey
}

// GetFetcher returns the NVD fetcher
func (p *CVEProvider) GetFetcher() *remote.Fetcher {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.fetcher
}
