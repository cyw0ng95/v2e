package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CWEProvider implements ProviderFSM for CWE data
type CWEProvider struct {
	*fsm.BaseProviderFSM
	localPath  string
	batchSize  int
	maxRetries int
	retryDelay time.Duration
	rpcClient  *rpc.Client
	mu         sync.RWMutex
}

// NewCWEProvider creates a new CWE provider with FSM support
// localPath is the path to the CWE JSON file
func NewCWEProvider(localPath string, store *storage.Store) (*CWEProvider, error) {
	if localPath == "" {
		localPath = "assets/cwe-raw.json"
	}

	provider := &CWEProvider{
		localPath:  localPath,
		rpcClient:  &rpc.Client{},
		batchSize:  100,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}

	// Create base FSM with custom executor
	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "cwe",
		ProviderType: "cwe",
		Storage:      store,
		Executor:     provider.execute,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// Initialize sets up the CWE provider context
func (p *CWEProvider) Initialize(ctx context.Context) error {
	return nil
}

// execute performs CWE fetch and store operations
func (p *CWEProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	p.mu.RLock()
	localPath := p.localPath
	batchSize := p.batchSize
	rpcClient := p.rpcClient
	p.mu.RUnlock()

	// Read and parse CWE data
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read CWE file: %w", err)
	}

	var cweList []cwe.CWEItem
	if err := json.Unmarshal(data, &cweList); err != nil {
		return fmt.Errorf("failed to unmarshal CWE data: %w", err)
	}

	totalCount := len(cweList)

	// Process CWEs in batches
	for i := 0; i < totalCount; i++ {
		// Check for cancellation or state change
		if p.GetState() != fsm.ProviderRunning {
			break
		}

		cweItem := cweList[i]

		// Create URN for checkpointing
		itemURN := urn.MustParse(fmt.Sprintf("v2e::mitre::cwe::%s", cweItem.ID))

		// Store CWE via RPC
		cweItemData, err := json.Marshal(cweItem)
		if err != nil {
			return fmt.Errorf("failed to marshal CWE item: %w", err)
		}

		// Simulate batch delay
		if i > 0 && i%batchSize == 0 {
			time.Sleep(1 * time.Second)
		}

		// Call RPC to store CWE
		_, err = rpcClient.InvokeRPC(context.Background(), "local", "RPCImportCWE", map[string]interface{}{
			"cweData": cweItemData,
		})

		success := true
		errorMsg := ""
		if err != nil {
			success = false
			errorMsg = fmt.Sprintf("failed to store CWE: %v", err)
		}

		// Save checkpoint
		if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
			return fmt.Errorf("failed to save checkpoint for %s: %w", cweItem.ID, err)
		}
	}

	return nil
}

// SetBatchSize sets the batch size
func (p *CWEProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// GetLocalPath returns the local file path
func (p *CWEProvider) GetLocalPath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.localPath
}

// SetLocalPath sets the local file path
func (p *CWEProvider) SetLocalPath(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.localPath = path
}

// GetBatchSize returns the batch size
func (p *CWEProvider) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}

// SetBatchSize sets the batch size
func (p *CWEProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}
