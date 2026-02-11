package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CWEProvider implements ProviderFSM for CWE data
type CWEProvider struct {
	*fsm.BaseProviderFSM
	localPath string
	sp        *subprocess.Subprocess
}

// NewCWEProvider creates a new CWE provider with FSM support
// localPath is the path to the CWE JSON file
// sp is the subprocess for RPC communication (can be nil for testing)
func NewCWEProvider(localPath string, store *storage.Store, sp *subprocess.Subprocess) (*CWEProvider, error) {
	if localPath == "" {
		localPath = "assets/cwe-raw.json"
	}

	provider := &CWEProvider{
		localPath: localPath,
		sp:        sp,
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

// execute performs CWE fetch and store operations
func (p *CWEProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	// Require subprocess for RPC calls
	if p.sp == nil {
		return fmt.Errorf("subprocess not configured, cannot make RPC calls")
	}

	localPath := p.localPath
	batchSize := p.GetBatchSize()

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
	errorCount := 0

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

		// Build RPC request to store individual CWE
		// Note: We use RPCImportCWE which expects cweData parameter
		params := map[string]interface{}{
			"cweData": cweItemData,
		}

		// Send RPC request to broker to route to local service
		payload, merr := subprocess.MarshalFast(params)
		if merr != nil {
			success := false
			errorMsg := fmt.Sprintf("failed to marshal RPC request: %v", merr)
			if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
				return fmt.Errorf("failed to save checkpoint for %s: %w", cweItem.ID, err)
			}
			return fmt.Errorf("marshal error: %w", merr)
		}

		rpcMsg := &subprocess.Message{
			Type:    subprocess.MessageTypeRequest,
			ID:      "RPCImportCWE",
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
			errorCount++
			// Return error if RPC fails - this is the bug fix
			if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
				return fmt.Errorf("failed to save checkpoint for %s: %w", cweItem.ID, err)
			}
			return fmt.Errorf("RPC call failed for CWE %s: %w", cweItem.ID, err)
		}

		// Save checkpoint
		if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
			return fmt.Errorf("failed to save checkpoint for %s: %w", cweItem.ID, err)
		}
	}

	if errorCount > 0 {
		return fmt.Errorf("completed with %d errors out of %d CWEs", errorCount, totalCount)
	}

	return nil
}

// GetLocalPath returns the local file path
func (p *CWEProvider) GetLocalPath() string {
	return p.localPath
}

// SetLocalPath sets the local file path
func (p *CWEProvider) SetLocalPath(path string) {
	p.localPath = path
}
