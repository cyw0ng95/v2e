package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cce"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CCEProvider implements ProviderFSM for CCE data
type CCEProvider struct {
	*fsm.BaseProviderFSM
	localPath  string
	batchSize  int
	maxRetries int
	retryDelay time.Duration
	rpcClient  *rpc.Client
	mu         sync.RWMutex
}

// NewCCEProvider creates a new CCE provider with FSM support
// localPath is the path to the CCE Excel file
func NewCCEProvider(localPath string, store *storage.Store) (*CCEProvider, error) {
	if localPath == "" {
		localPath = "assets/cce-raw.xlsx"
	}

	provider := &CCEProvider{
		localPath:  localPath,
		rpcClient:  &rpc.Client{},
		batchSize:  100,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
	}

	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "cce",
		ProviderType: "cce",
		Storage:      store,
		Executor:     provider.execute,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// Initialize sets up the CCE provider context
func (p *CCEProvider) Initialize(ctx context.Context) error {
	return nil
}

// execute performs CCE fetch and store operations
func (p *CCEProvider) execute() error {
	currentState := p.GetState()

	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	p.mu.RLock()
	localPath := p.localPath
	batchSize := p.batchSize
	rpcClient := p.rpcClient
	p.mu.RUnlock()

	parser := cce.NewParser(localPath)

	totalCount, err := parser.ParseRowCount()
	if err != nil {
		return fmt.Errorf("failed to get CCE row count: %w", err)
	}

	offset := 0

	for offset < totalCount {
		if p.GetState() != fsm.ProviderRunning {
			break
		}

		entries, _, err := parser.ParseBatch(offset, batchSize)
		if err != nil {
			return fmt.Errorf("failed to parse CCE batch at offset %d: %w", offset, err)
		}

		for _, entry := range entries {
			itemURN := urn.MustParse(fmt.Sprintf("v2e::cce::%s", entry.ID))

			entryData, err := json.Marshal(entry)
			if err != nil {
				return fmt.Errorf("failed to marshal CCE entry: %w", err)
			}

			_, err = rpcClient.InvokeRPC(context.Background(), "local", "RPCImportCCE", map[string]interface{}{
				"cceData": entryData,
			})

			success := true
			errorMsg := ""
			if err != nil {
				success = false
				errorMsg = fmt.Sprintf("failed to store CCE: %v", err)
			}

			if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
				return fmt.Errorf("failed to save checkpoint for %s: %w", entry.ID, err)
			}
		}

		if offset > 0 && offset%batchSize == 0 {
			time.Sleep(1 * time.Second)
		}

		offset += len(entries)
	}

	return nil
}

// Fetch performs the fetch operation
func (p *CCEProvider) Fetch(ctx context.Context) error {
	return p.Execute()
}

// Store performs the store operation
func (p *CCEProvider) Store(ctx context.Context) error {
	return p.Execute()
}

// GetStats returns provider statistics
func (p *CCEProvider) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]interface{}{
		"batch_size": p.batchSize,
	}
}

// Cleanup releases any resources held by the provider
func (p *CCEProvider) Cleanup(ctx context.Context) error {
	return nil
}

// SetBatchSize sets the batch size
func (p *CCEProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// GetLocalPath returns the local file path
func (p *CCEProvider) GetLocalPath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.localPath
}

// SetLocalPath sets the local file path
func (p *CCEProvider) SetLocalPath(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.localPath = path
}

// GetBatchSize returns the batch size
func (p *CCEProvider) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}
