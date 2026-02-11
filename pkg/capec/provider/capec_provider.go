package provider

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CAPECProvider implements ProviderFSM for CAPEC data
type CAPECProvider struct {
	*fsm.BaseProviderFSM
	localPath  string
	batchSize  int
	maxRetries int
	retryDelay time.Duration
	sp         *subprocess.Subprocess
	mu         sync.RWMutex
}

// NewCAPECProvider creates a new CAPEC provider with FSM support
// localPath is the path to CAPEC XML file
// sp is the subprocess for RPC communication (can be nil for testing)
// dependencies is a list of provider IDs that must complete before this provider
func NewCAPECProvider(localPath string, store *storage.Store, sp *subprocess.Subprocess, dependencies []string) (*CAPECProvider, error) {
	if localPath == "" {
		localPath = "assets/capec_contents_latest.xml"
	}

	provider := &CAPECProvider{
		localPath:  localPath,
		batchSize:  50,
		maxRetries: 3,
		retryDelay: 5 * time.Second,
		sp:         sp,
	}

	// Create base FSM with custom executor and dependencies
	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "capec",
		ProviderType: "capec",
		Storage:      store,
		Executor:     provider.execute,
		Dependencies: dependencies,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// Initialize sets up to CAPEC provider context
func (p *CAPECProvider) Initialize(ctx context.Context) error {
	return nil
}

// execute performs CAPEC fetch and store operations
func (p *CAPECProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	// Require subprocess for RPC calls
	if p.sp == nil {
		return fmt.Errorf("subprocess not configured, cannot make RPC calls")
	}

	p.mu.RLock()
	localPath := p.localPath
	batchSize := p.batchSize
	p.mu.RUnlock()

	// Read and parse CAPEC data
	data, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read CAPEC file: %w", err)
	}

	capecData := capec.Root{}
	if err := xml.Unmarshal(data, &capecData); err != nil {
		return fmt.Errorf("failed to unmarshal CAPEC data: %w", err)
	}

	attackPatterns := capecData.AttackPatterns.AttackPattern
	totalCount := len(attackPatterns)

	// Process attack patterns in batches
	for i := 0; i < totalCount; i++ {
		// Check for cancellation or state change
		if p.GetState() != fsm.ProviderRunning {
			break
		}

		attackPattern := attackPatterns[i]

		// Create URN for checkpointing
		capecID := attackPattern.ID
		itemURN := urn.MustParse(fmt.Sprintf("v2e::mitre::capec::%d", capecID))

		// Marshal CAPEC data for RPC call
		capecJSON, err := json.Marshal(attackPattern)
		if err != nil {
			return fmt.Errorf("failed to marshal CAPEC item: %w", err)
		}

		// Simulate batch delay
		if i > 0 && i%batchSize == 0 {
			time.Sleep(1 * time.Second)
		}

		// Store CAPEC via RPC call to v2local service
		// Use RPCImportCAPEC for individual CAPEC import

		params := map[string]interface{}{
			"capecData": capecJSON,
		}

		// Send RPC request to broker to route to local service
		payload, merr := subprocess.MarshalFast(params)
		if merr != nil {
			success := false
			errorMsg := fmt.Sprintf("failed to marshal RPC request: %v", merr)
			if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
				return fmt.Errorf("failed to save checkpoint for %d: %w", capecID, err)
			}
			return fmt.Errorf("marshal error: %w", merr)
		}

		rpcMsg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCImportCAPEC",
			Payload: payload,
			Target:  "local",
			Source:  p.sp.ID,
		}

		err = p.sp.SendMessage(rpcMsg)

		success := true
		errorMsg := ""
		if err != nil {
			success = false
			errorMsg = fmt.Sprintf("failed to send RPC request: %v", err)
		}

		// Save checkpoint
		if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
			return fmt.Errorf("failed to save checkpoint for %d: %w", capecID, err)
		}

		if !success {
			return fmt.Errorf("failed to store CAPEC %d: %s", capecID, errorMsg)
		}
	}

	return nil
}

// GetLocalPath returns to local file path
func (p *CAPECProvider) GetLocalPath() string {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.localPath
}

// SetLocalPath sets to local file path
func (p *CAPECProvider) SetLocalPath(path string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.localPath = path
}

// GetBatchSize returns to batch size
func (p *CAPECProvider) GetBatchSize() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.batchSize
}

// SetBatchSize sets to batch size
func (p *CAPECProvider) SetBatchSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.batchSize = size
}

// Cleanup releases any resources held by the provider
func (p *CAPECProvider) Cleanup(ctx context.Context) error {
	return nil
}

// Fetch performs the fetch operation
func (p *CAPECProvider) Fetch(ctx context.Context) error {
	return p.Execute()
}

// Store performs the store operation
func (p *CAPECProvider) Store(ctx context.Context) error {
	return p.Execute()
}

// GetStats returns provider statistics
func (p *CAPECProvider) GetStats() map[string]interface{} {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return map[string]interface{}{
		"batch_size": p.batchSize,
	}
}
