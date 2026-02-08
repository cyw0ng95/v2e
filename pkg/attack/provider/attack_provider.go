package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// ATTACKProvider implements ProviderFSM for ATT&CK data
type ATTACKProvider struct {
	*fsm.BaseProviderFSM
	localPath  string
	batchSize  int
	maxRetries int
	retryDelay time.Duration
	rpcClient  *rpc.Client
	mu         sync.RWMutex
}

// NewATTACKProvider creates a new ATT&CK provider with FSM support
// localPath is the path to ATT&CK Excel file
func NewATTACKProvider(localPath string, store *storage.Store) (*ATTACKProvider, error) {
	if localPath == "" {
		localPath = "assets/enterprise-attack.xlsx"
	}

	provider := &ATTACKProvider{
		localPath:  localPath,
		rpcClient:  &rpc.Client{},
		batchSize:  100,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}

	// Create base FSM with custom executor
	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "attack",
		ProviderType: "attack",
		Storage:      store,
		Executor:     provider.execute,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// Initialize sets up the ATT&CK provider context
func (p *ATTACKProvider) Initialize(ctx context.Context) error {
	return nil
}

// execute performs ATT&CK fetch and store operations
func (p *ATTACKProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	p.mu.RLock()
	localPath := p.localPath
	batchSize := p.batchSize
	p.mu.RUnlock()

	// Read and parse ATT&CK data
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read ATT&CK file: %w", err)
	}

	var attackList []attack.AttackTechnique
	if err := json.Unmarshal(data, &attackList); err != nil {
		return fmt.Errorf("failed to unmarshal ATT&CK data: %w", err)
	}

	totalCount := len(attackList)

	// Process attack techniques in batches
	for i := 0; i < totalCount; i++ {
		// Check for cancellation or state change
		if p.GetState() != fsm.ProviderRunning {
			break
		}

		attackItem := attackList[i]

		// Create URN for checkpointing
		techniqueID := attackItem.ID
		itemURN := urn.MustParse(fmt.Sprintf("v2e::mitre::attack::%s", techniqueID))

		// Store ATT&CK (TODO: implement actual storage via RPC)
		_, err := json.Marshal(attackItem)
		if err != nil {
			return fmt.Errorf("failed to marshal attack item: %w", err)
		}

		// Simulate batch delay
		if i > 0 && i%batchSize == 0 {
			time.Sleep(1 * time.Second)
		}

		success := true
		errorMsg := ""
		// TODO: Use RPCStoreAttack to store each item

		// Save checkpoint
		if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
			return fmt.Errorf("failed to save checkpoint for %s: %w", techniqueID, err)
		}
	}

	return nil
}

// GetLocalPath returns to local file path
func (p *ATTACKProvider) GetLocalPath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.localPath
}

// SetLocalPath sets to local file path
func (p *ATTACKProvider) SetLocalPath(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.localPath = path
}

// GetBatchSize returns to batch size
func (p *ATTACKProvider) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}

// SetBatchSize sets to batch size
func (p *ATTACKProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// Cleanup releases any resources held by the provider
func (p *ATTACKProvider) Cleanup(ctx context.Context) error {
	return nil
}

// Fetch performs the fetch operation
func (p *ATTACKProvider) Fetch(ctx context.Context) error {
	return p.Execute()
}

// Store performs the store operation
func (p *ATTACKProvider) Store(ctx context.Context) error {
	return p.Execute()
}

// GetStats returns provider statistics
func (p *ATTACKProvider) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]interface{}{
		"batch_size": p.batchSize,
	}
}

// GetConfig returns provider configuration
func (p *ATTACKProvider) GetConfig() *provider.ProviderConfig {
	return p.BaseProviderFSM.GetConfig()
}

// Fetch performs the fetch operation (delegates to FSM Execute)
func (p *ATTACKProvider) Fetch(ctx context.Context) error {
	return p.Execute()
}

// Store performs the store operation (delegates to FSM Execute)
func (p *ATTACKProvider) Store(ctx context.Context) error {
	return p.Execute()
}

// GetStats returns provider statistics
func (p *ATTACKProvider) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]interface{}{
		"batch_size": p.batchSize,
	}
}
