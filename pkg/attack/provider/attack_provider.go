package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// ATTACKProvider implements ProviderFSM for ATT&CK data
type ATTACKProvider struct {
	*fsm.BaseProviderFSM
	localPath string
}

// NewATTACKProvider creates a new ATT&CK provider with FSM support
// localPath is the path to ATT&CK Excel file
func NewATTACKProvider(localPath string, store *storage.Store) (*ATTACKProvider, error) {
	if localPath == "" {
		localPath = "assets/enterprise-attack.xlsx"
	}

	provider := &ATTACKProvider{
		localPath: localPath,
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

// execute performs ATT&CK fetch and store operations
func (p *ATTACKProvider) execute() error {
	currentState := p.GetState()

	// Check if we should be running
	if currentState != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", currentState)
	}

	localPath := p.localPath
	batchSize := p.GetBatchSize()

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

		// Store ATT&CK technique via RPC call to v2local service
		// Note: The v2local service provides RPCImportATTACKs for bulk import from XLSX
		// For now, marshal the item until RPC integration is complete
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

		// Save checkpoint
		if err := p.SaveCheckpoint(itemURN, success, errorMsg); err != nil {
			return fmt.Errorf("failed to save checkpoint for %s: %w", techniqueID, err)
		}
	}

	return nil
}

// GetLocalPath returns the local file path
func (p *ATTACKProvider) GetLocalPath() string {
	return p.localPath
}

// SetLocalPath sets the local file path
func (p *ATTACKProvider) SetLocalPath(path string) {
	p.localPath = path
}
