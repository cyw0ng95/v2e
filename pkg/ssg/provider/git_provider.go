package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/ssg/remote"
)

// SSGProvider implements ProviderFSM for SSG data
type SSGProvider struct {
	*fsm.BaseProviderFSM
	repoURL    string
	batchSize  int
	maxRetries int
	retryDelay time.Duration
	mu         sync.RWMutex
	gitClient  *remote.GitClient
}

// NewSSGProvider creates a new SSG provider with FSM support
// repoURL is the Git repository URL
func NewSSGProvider(repoURL string, store *storage.Store) (*SSGProvider, error) {
	if repoURL == "" {
		repoURL = "https://github.com/OWASP/wg-ssg"
	}

	provider := &SSGProvider{
		repoURL:    repoURL,
		batchSize:  10,
		maxRetries: 3,
		retryDelay: 10 * time.Second,
		gitClient:  remote.NewGitClient(repoURL, ""),
	}

	// Create base FSM with custom executor
	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "ssg",
		ProviderType: "ssg",
		Storage:      store,
		Executor:     provider.execute,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// Initialize sets up the SSG provider context
func (p *SSGProvider) Initialize(ctx context.Context) error {
	return nil
}

// execute performs SSG fetch and store operations
func (p *SSGProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	p.mu.RLock()
	gitClient := p.gitClient
	p.mu.RUnlock()

	// Fetch latest SSG data from Git
	err := gitClient.Pull()
	if err != nil {
		return fmt.Errorf("failed to pull SSG repository: %w", err)
	}

	guideFiles, err := gitClient.ListGuideFiles()
	if err != nil {
		return fmt.Errorf("failed to list guide files: %w", err)
	}

	if len(guideFiles) == 0 {
		return fmt.Errorf("no guide files found in SSG repository")
	}

	// TODO: Import each guide via RPC to v2local service
	// The v2local service provides RPCSSGImportGuide for SSG storage
	// For each guideFile in guideFiles:
	//   - Call RPCSSGImportGuide with the guide path

	return nil
}

// Cleanup releases any resources held by the provider
func (p *SSGProvider) Cleanup(ctx context.Context) error {
	return nil
}

// Fetch performs the fetch operation
func (p *SSGProvider) Fetch(ctx context.Context) error {
	return p.Execute()
}

// Store performs the store operation
func (p *SSGProvider) Store(ctx context.Context) error {
	return p.Execute()
}

// GetStats returns provider statistics
func (p *SSGProvider) GetStats() map[string]interface{} {
	stats := p.BaseProviderFSM.GetStats()
	stats["batch_size"] = p.batchSize
	stats["repo_url"] = p.repoURL
	return stats
}

// SetBatchSize sets the batch size
func (p *SSGProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// GetBatchSize returns the batch size
func (p *SSGProvider) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}

// GetRepoURL returns the repository URL
func (p *SSGProvider) GetRepoURL() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.repoURL
}

// SetRepoURL sets the repository URL
func (p *SSGProvider) SetRepoURL(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.repoURL = url
	// Also update the git client
	p.gitClient = remote.NewGitClient(url, "")
}
