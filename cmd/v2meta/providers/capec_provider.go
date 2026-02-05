package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CAPECProvider implements Provider for CAPEC data import from XML file
type CAPECProvider struct {
	*fsm.BaseProviderFSM
	rpcClient          *rpc.Client
	logger             *common.Logger
	batchSize          int
	checkpointInterval int
	filePath           string // Path to CAPEC XML file
	lastModStartDate   string // For incremental updates
	errorCount         int64
	totalProcessed     int64
	failureThreshold   float64 // Auto-pause if error rate > threshold (default 0.1 = 10%)
	currentOffset      int     // Current position in file processing
}

// CAPECProviderConfig holds configuration for CAPEC provider
type CAPECProviderConfig struct {
	ID                 string
	Storage            *storage.Store
	RPCClient          *rpc.Client
	Logger             *common.Logger
	BatchSize          int
	CheckpointInterval int
	FilePath           string
	LastModStartDate   string
	FailureThreshold   float64
}

// NewCAPECProvider creates a new CAPEC provider
func NewCAPECProvider(config CAPECProviderConfig) (*CAPECProvider, error) {
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
		config.FilePath = "assets/capec.xml"
	}

	// Create executor function that will be called by BaseProviderFSM
	var provider *CAPECProvider
	executor := func() error {
		if provider == nil {
			return fmt.Errorf("provider not initialized")
		}
		return provider.executeBatch()
	}

	baseFSM, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           config.ID,
		ProviderType: "capec",
		Storage:      config.Storage,
		Executor:     executor,
	})
	if err != nil {
		return nil, err
	}

	provider = &CAPECProvider{
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

// executeBatch performs one batch of CAPEC import
func (p *CAPECProvider) executeBatch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check error rate and auto-pause if threshold exceeded
	if err := p.checkErrorThreshold(); err != nil {
		return err
	}

	// Import CAPECs from XML file with batch processing
	params := map[string]interface{}{
		"file_path": p.filePath,
		"offset":    p.currentOffset,
		"limit":     p.batchSize,
	}

	// Use incremental import if lastModStartDate is set
	if p.lastModStartDate != "" {
		params["lastModStartDate"] = p.lastModStartDate
	}

	p.logger.Info("Importing CAPEC batch: file=%s, offset=%d, size=%d", p.filePath, p.currentOffset, p.batchSize)

	resp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCImportCAPECBatch", params)
	if err != nil {
		p.errorCount++
		p.logger.Error("Failed to import CAPEC batch: %v", err)
		return fmt.Errorf("failed to import CAPEC batch: %w", err)
	}

	// Check for error response
	if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
		p.errorCount++
		p.logger.Error("CAPEC import returned error: %s", errMsg)
		return fmt.Errorf("CAPEC import failed: %s", errMsg)
	}

	// Extract CAPECs from response
	var batchResp struct {
		CAPECs []map[string]interface{} `json:"capecs"`
	}
	if err := subprocess.UnmarshalPayload(resp, &batchResp); err != nil {
		p.errorCount++
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	capecs := batchResp.CAPECs
	if len(capecs) == 0 {
		p.logger.Info("No more CAPECs to import, provider completed")
		return nil // No more data, provider will transition to TERMINATED
	}

	// Process and save each CAPEC
	for i, capecMap := range capecs {
		capecID, ok := capecMap["capec_id"].(string)
		if !ok {
			p.errorCount++
			p.logger.Warn("Missing CAPEC ID at index %d", i)
			continue
		}

		// Save to local storage
		if err := p.saveCAPEC(ctx, capecMap); err != nil {
			p.errorCount++
			p.logger.Error("Failed to save CAPEC %s: %v", capecID, err)
			continue
		}

		p.totalProcessed++

		// Save checkpoint every N items
		if p.totalProcessed%int64(p.checkpointInterval) == 0 {
			itemURN, err := urn.Parse(fmt.Sprintf("v2e::mitre::capec::%s", capecID))
			if err != nil {
				p.logger.Error("Failed to parse URN for CAPEC %s: %v", capecID, err)
			} else {
				if err := p.SaveCheckpoint(itemURN, true, ""); err != nil {
					p.logger.Error("Failed to save checkpoint: %v", err)
				} else {
					p.logger.Info("Checkpoint saved at %s (processed: %d)", itemURN.Key(), p.totalProcessed)
				}
			}
		}

		// Update lastModStartDate for incremental updates
		if lastMod, ok := capecMap["last_modified"].(string); ok {
			p.lastModStartDate = lastMod
		}
	}

	// Update offset for next batch
	p.currentOffset += len(capecs)

	p.logger.Info("Processed CAPEC batch: %d items, total: %d, errors: %d", len(capecs), p.totalProcessed, p.errorCount)
	return nil
}

// saveCAPEC saves a CAPEC to local storage via RPC
func (p *CAPECProvider) saveCAPEC(ctx context.Context, capecData map[string]interface{}) error {
	// Check if we should use field-level diffing
	capecID, _ := capecData["capec_id"].(string)

	// First, try to get existing CAPEC
	existingResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCGetCAPEC", map[string]interface{}{
		"capec_id": capecID,
	})

	// If exists and not error, perform diff
	if err == nil && existingResp != nil {
		if isErr, _ := subprocess.IsErrorResponse(existingResp); !isErr {
			// CAPEC exists, unmarshal and perform field-level diffing
			var existingCAPEC map[string]interface{}
			if err := subprocess.UnmarshalPayload(existingResp, &existingCAPEC); err == nil {
				updateParams := p.diffFields(existingCAPEC, capecData)
				if len(updateParams) == 0 {
					p.logger.Debug("CAPEC %s unchanged, skipping update", capecID)
					return nil
				}
				updateParams["capec_id"] = capecID

				updateResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCUpdateCAPEC", updateParams)
				if err != nil {
					return fmt.Errorf("failed to update CAPEC: %w", err)
				}
				
				// Check for error response
				if isErr, errMsg := subprocess.IsErrorResponse(updateResp); isErr {
					return fmt.Errorf("update CAPEC failed: %s", errMsg)
				}
				
				return nil
			}
		}
	}

	// CAPEC doesn't exist or error occurred, create new
	saveResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCSaveCAPEC", capecData)
	if err != nil {
		return fmt.Errorf("failed to save CAPEC: %w", err)
	}

	// Check for error response
	if isErr, errMsg := subprocess.IsErrorResponse(saveResp); isErr {
		return fmt.Errorf("save CAPEC failed: %s", errMsg)
	}

	return nil
}

// diffFields compares two CAPEC objects and returns only changed fields
// Implements Field-Level Data Diffing
func (p *CAPECProvider) diffFields(existing, incoming map[string]interface{}) map[string]interface{} {
	changed := make(map[string]interface{})

	for key, newVal := range incoming {
		if key == "capec_id" {
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
func (p *CAPECProvider) loadLastCheckpoint() error {
	stats := p.GetStats()
	checkpoint, _ := stats["last_checkpoint"].(string)

	if checkpoint != "" {
		_, err := urn.Parse(checkpoint)
		if err == nil {
			p.logger.Info("Resuming from checkpoint: %s", checkpoint)
			// URN format: v2e::mitre::capec::CAPEC-##
		}
	}

	return nil
}

// checkErrorThreshold checks if error rate exceeds threshold
// Implements Pause-on-Error Threshold
func (p *CAPECProvider) checkErrorThreshold() error {
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
func (p *CAPECProvider) GetProgress() map[string]interface{} {
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
