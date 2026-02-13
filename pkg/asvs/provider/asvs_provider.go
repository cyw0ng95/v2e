package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// ASVSProvider implements ProviderFSM for ASVS data
type ASVSProvider struct {
	*fsm.BaseProviderFSM
	csvURL     string
	batchSize  int
	maxRetries int
	retryDelay time.Duration
	mu         sync.RWMutex
}

// NewASVSProvider creates a new ASVS provider with FSM support
// csvURL is URL to ASVS CSV file
func NewASVSProvider(csvURL string, store *storage.Store) (*ASVSProvider, error) {
	if csvURL == "" {
		csvURL = "https://raw.githubusercontent.com/OWASP/ASVS/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv"
	}

	provider := &ASVSProvider{
		csvURL:     csvURL,
		batchSize:  100,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}

	// Create base FSM with custom executor
	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "asvs",
		ProviderType: "asvs",
		Storage:      store,
		Executor:     provider.execute,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// Initialize sets up the ASVS provider context
func (p *ASVSProvider) Initialize(ctx context.Context) error {
	return nil
}

// execute performs ASVS fetch and store operations
func (p *ASVSProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	p.mu.RLock()
	csvURL := p.csvURL
	p.mu.RUnlock()

	// Fetch ASVS CSV file
	resp, err := http.Get(csvURL)
	if err != nil {
		return fmt.Errorf("failed to download ASVS CSV: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read ASVS CSV: %w", err)
	}

	// Parse CSV and process rows
	// For now, just validate we got data
	if len(data) == 0 {
		return fmt.Errorf("ASVS CSV file is empty")
	}

	// TODO: Parse CSV rows and import via RPC to v2local service
	// The v2local service provides RPCImportASVS for ASVS storage

	return nil
}

// Cleanup releases any resources held by the provider
func (p *ASVSProvider) Cleanup(ctx context.Context) error {
	return nil
}

// Fetch performs the fetch operation
func (p *ASVSProvider) Fetch(ctx context.Context) error {
	return p.Execute()
}

// Store performs the store operation
func (p *ASVSProvider) Store(ctx context.Context) error {
	return p.Execute()
}

// GetStats returns provider statistics
func (p *ASVSProvider) GetStats() map[string]interface{} {
	stats := p.BaseProviderFSM.GetStats()
	stats["batch_size"] = p.batchSize
	stats["csv_url"] = p.csvURL
	return stats
}

// SetBatchSize sets the batch size
func (p *ASVSProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// GetBatchSize returns the batch size
func (p *ASVSProvider) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}

// GetCSVURL returns the CSV URL
func (p *ASVSProvider) GetCSVURL() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.csvURL
}

// SetCSVURL sets the CSV URL
func (p *ASVSProvider) SetCSVURL(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.csvURL = url
}
