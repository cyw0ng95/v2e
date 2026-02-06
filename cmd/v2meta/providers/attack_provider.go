package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve/taskflow"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// ATTACKProvider implements Provider for ATT&CK technique data import
type ATTACKProvider struct {
	*fsm.BaseProviderFSM
	rpcClient          taskflow.RPCInvoker
	logger             *common.Logger
	batchSize          int
	checkpointInterval int
	filePath           string // Path to ATT&CK technique data
	lastModStartDate   string // For incremental updates
	errorCount         int64
	totalProcessed     int64
	failureThreshold   float64 // Auto-pause if error rate > threshold (default 0.1 = 10%)
	currentOffset      int     // Current position in file processing
}

// ATTACKProviderConfig holds configuration for ATT&CK provider
type ATTACKProviderConfig struct {
	ID                 string
	Storage            *storage.Store
	RPCClient          taskflow.RPCInvoker
	Logger             *common.Logger
	BatchSize          int
	CheckpointInterval int
	FilePath           string
	LastModStartDate   string
	FailureThreshold   float64
}

// NewATTACKProvider creates a new ATT&CK provider
func NewATTACKProvider(config ATTACKProviderConfig) (*ATTACKProvider, error) {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.CheckpointInterval <= 0 {
		config.CheckpointInterval = 100
	}
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 0.1 // 10% default
	}
	if config.FilePath == "" {
		config.FilePath = "assets/attack-techniques.json"
	}

	// Create executor function that will be called by BaseProviderFSM
	var provider *ATTACKProvider
	executor := func() error {
		if provider == nil {
			return fmt.Errorf("provider not initialized")
		}
		return provider.executeBatch()
	}

	baseFSM, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           config.ID,
		ProviderType: "attack",
		Storage:      config.Storage,
		Executor:     executor,
	})
	if err != nil {
		return nil, err
	}

	provider = &ATTACKProvider{
		BaseProviderFSM:    baseFSM,
		rpcClient:          config.RPCClient,
		logger:             config.Logger,
		batchSize:          config.BatchSize,
		checkpointInterval: config.CheckpointInterval,
		filePath:           config.FilePath,
		lastModStartDate:   config.LastModStartDate,
		failureThreshold:   config.FailureThreshold,
	}

	// Load last checkpoint to resume from
	if err := provider.loadLastCheckpoint(); err != nil {
		config.Logger.Warn("Failed to load checkpoint, starting fresh: %v", err)
	}

	return provider, nil
}

// executeBatch performs one batch of ATT&CK technique import
func (p *ATTACKProvider) executeBatch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check error rate and auto-pause if threshold exceeded
	if err := p.checkErrorThreshold(); err != nil {
		return err
	}

	// Import ATT&CK techniques with batch processing
	params := map[string]interface{}{
		"file_path": p.filePath,
		"offset":    p.currentOffset,
		"limit":     p.batchSize,
	}

	// Use incremental import if lastModStartDate is set
	if p.lastModStartDate != "" {
		params["lastModStartDate"] = p.lastModStartDate
	}

	p.logger.Info("Importing ATT&CK batch: file=%s, offset=%d, size=%d", p.filePath, p.currentOffset, p.batchSize)

	resp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCImportATTACKBatch", params)
	if err != nil {
		p.errorCount++
		p.logger.Error("Failed to import ATT&CK batch: %v", err)
		return fmt.Errorf("failed to import ATT&CK batch: %w", err)
	}

	// Check for error response
	if isErr, errMsg := subprocess.IsErrorResponse(resp.(*subprocess.Message)); isErr {
		p.errorCount++
		p.logger.Error("ATT&CK import returned error: %s", errMsg)
		return fmt.Errorf("ATT&CK import failed: %s", errMsg)
	}

	// Extract techniques from response
	var batchResp struct {
		Techniques []map[string]interface{} `json:"techniques"`
	}
	if err := subprocess.UnmarshalPayload(resp.(*subprocess.Message), &batchResp); err != nil {
		p.errorCount++
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	techniques := batchResp.Techniques
	if len(techniques) == 0 {
		p.logger.Info("No more ATT&CK techniques to import, provider completed")
		return nil // No more data, provider will transition to TERMINATED
	}

	// Process and save each technique
	for i, techniqueMap := range techniques {
		techniqueID, ok := techniqueMap["technique_id"].(string)
		if !ok {
			p.errorCount++
			p.logger.Warn("Missing technique ID at index %d", i)
			continue
		}

		// Save to local storage
		if err := p.saveATTACKTechnique(ctx, techniqueMap); err != nil {
			p.errorCount++
			p.logger.Error("Failed to save ATT&CK technique %s: %v", techniqueID, err)
			continue
		}

		p.totalProcessed++

		// Save checkpoint every N items
		if p.totalProcessed%int64(p.checkpointInterval) == 0 {
			itemURN, err := urn.Parse(fmt.Sprintf("v2e::mitre::attack::%s", techniqueID))
			if err != nil {
				p.logger.Error("Failed to parse URN for ATT&CK technique %s: %v", techniqueID, err)
			} else {
				if err := p.SaveCheckpoint(itemURN, true, ""); err != nil {
					p.logger.Error("Failed to save checkpoint: %v", err)
				} else {
					p.logger.Info("Checkpoint saved at %s (processed: %d)", itemURN.Key(), p.totalProcessed)
				}
			}
		}

		// Update lastModStartDate for incremental updates
		if lastMod, ok := techniqueMap["last_modified"].(string); ok {
			p.lastModStartDate = lastMod
		}
	}

	// Update offset for next batch
	p.currentOffset += len(techniques)

	p.logger.Info("Processed ATT&CK batch: %d items, total: %d, errors: %d", len(techniques), p.totalProcessed, p.errorCount)
	return nil
}

// saveATTACKTechnique saves an ATT&CK technique to local storage via RPC
func (p *ATTACKProvider) saveATTACKTechnique(ctx context.Context, techniqueData map[string]interface{}) error {
	// Check if we should use field-level diffing
	techniqueID, _ := techniqueData["technique_id"].(string)

	// First, try to get existing technique
	existingResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCGetATTACKTechnique", map[string]interface{}{
		"technique_id": techniqueID,
	})

	// If exists and not error, perform diff
	if err == nil && existingResp != nil {
		if isErr, _ := subprocess.IsErrorResponse(existingResp.(*subprocess.Message)); !isErr {
			// Technique exists, unmarshal and perform field-level diffing
			var existingTechnique map[string]interface{}
			if err := subprocess.UnmarshalPayload(existingResp.(*subprocess.Message), &existingTechnique); err == nil {
				updateParams := p.diffFields(existingTechnique, techniqueData)
				if len(updateParams) == 0 {
					p.logger.Debug("ATT&CK technique %s unchanged, skipping update", techniqueID)
					return nil
				}
				updateParams["technique_id"] = techniqueID

				updateResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCUpdateATTACKTechnique", updateParams)
				if err != nil {
					return fmt.Errorf("failed to update ATT&CK technique: %w", err)
				}

				// Check for error response
				if isErr, errMsg := subprocess.IsErrorResponse(updateResp.(*subprocess.Message)); isErr {
					return fmt.Errorf("update ATT&CK technique failed: %s", errMsg)
				}

				return nil
			}
		}
	}

	// Technique doesn't exist or error occurred, create new
	saveResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCSaveATTACKTechnique", techniqueData)
	if err != nil {
		return fmt.Errorf("failed to save ATT&CK technique: %w", err)
	}

	// Check for error response
	if isErr, errMsg := subprocess.IsErrorResponse(saveResp.(*subprocess.Message)); isErr {
		return fmt.Errorf("save ATT&CK technique failed: %s", errMsg)
	}

	return nil
}

// diffFields compares two ATT&CK technique objects and returns only changed fields
// Implements Field-Level Data Diffing
func (p *ATTACKProvider) diffFields(existing, incoming map[string]interface{}) map[string]interface{} {
	changed := make(map[string]interface{})

	for key, newVal := range incoming {
		if key == "technique_id" {
			continue // Don't compare ID
		}

		oldVal, exists := existing[key]
		if !exists || !deepEqual(oldVal, newVal) {
			changed[key] = newVal
		}
	}

	return changed
}

// loadLastCheckpoint loads the last checkpoint from storage
func (p *ATTACKProvider) loadLastCheckpoint() error {
	stats := p.GetStats()
	checkpoint, _ := stats["last_checkpoint"].(string)

	if checkpoint != "" {
		_, err := urn.Parse(checkpoint)
		if err == nil {
			p.logger.Info("Resuming from checkpoint: %s", checkpoint)
			// URN format: v2e::mitre::attack::T####
		}
	}

	return nil
}

// checkErrorThreshold checks if error rate exceeds threshold
// Implements Pause-on-Error Threshold
func (p *ATTACKProvider) checkErrorThreshold() error {
	if p.totalProcessed == 0 {
		return nil // No data yet
	}

	errorRate := float64(p.errorCount) / float64(p.totalProcessed)
	if errorRate > p.failureThreshold {
		p.logger.Error("Error rate %.2f%% exceeds threshold %.2f%%, auto-pausing provider",
			errorRate*100, p.failureThreshold*100)

		// Transition to PAUSED state
		if err := p.Transition(fsm.ProviderPaused); err != nil {
			return fmt.Errorf("failed to pause provider: %w", err)
		}

		return fmt.Errorf("provider auto-paused due to high error rate: %.2f%%", errorRate*100)
	}

	return nil
}

// GetProgress returns current progress metrics
func (p *ATTACKProvider) GetProgress() map[string]interface{} {
	errorRate := 0.0
	if p.totalProcessed > 0 {
		errorRate = float64(p.errorCount) / float64(p.totalProcessed)
	}

	stats := p.GetStats()
	checkpoint, _ := stats["last_checkpoint"].(string)

	return map[string]interface{}{
		"total_processed": p.totalProcessed,
		"error_count":     p.errorCount,
		"error_rate":      errorRate,
		"last_checkpoint": checkpoint,
		"batch_size":      p.batchSize,
		"current_offset":  p.currentOffset,
	}
}
