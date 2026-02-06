package providers

import (
	"context"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CVEProvider implements Provider for CVE data fetching
type CVEProvider struct {
	*fsm.BaseProviderFSM
	rpcClient          RPCInterface
	logger             *common.Logger
	batchSize          int
	checkpointInterval int
	lastModStartDate   string // For incremental fetching
	errorCount         int64
	totalProcessed     int64
	failureThreshold   float64 // Auto-pause if error rate > threshold (default 0.1 = 10%)
}

// CVEProviderConfig holds configuration for CVE provider
type CVEProviderConfig struct {
	ID                 string
	Storage            *storage.Store
	RPCClient          RPCInterface
	Logger             *common.Logger
	BatchSize          int
	CheckpointInterval int
	LastModStartDate   string
	FailureThreshold   float64
}

// NewCVEProvider creates a new CVE provider
func NewCVEProvider(config CVEProviderConfig) (*CVEProvider, error) {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}
	if config.CheckpointInterval <= 0 {
		config.CheckpointInterval = 100
	}
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 0.1 // 10% default
	}

	// Create executor function that will be called by BaseProviderFSM
	var provider *CVEProvider
	executor := func() error {
		if provider == nil {
			return fmt.Errorf("provider not initialized")
		}
		return provider.executeBatch()
	}

	baseFSM, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           config.ID,
		ProviderType: "cve",
		Storage:      config.Storage,
		Executor:     executor,
	})
	if err != nil {
		return nil, err
	}

	provider = &CVEProvider{
		BaseProviderFSM:    baseFSM,
		rpcClient:          config.RPCClient,
		logger:             config.Logger,
		batchSize:          config.BatchSize,
		checkpointInterval: config.CheckpointInterval,
		lastModStartDate:   config.LastModStartDate,
		failureThreshold:   config.FailureThreshold,
	}

	// Load last checkpoint to resume from
	if err := provider.loadLastCheckpoint(); err != nil {
		config.Logger.Warn("Failed to load checkpoint, starting fresh: %v", err)
	}

	return provider, nil
}

// executeBatch performs one batch of CVE fetching
func (p *CVEProvider) executeBatch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check error rate and auto-pause if threshold exceeded
	if err := p.checkErrorThreshold(); err != nil {
		return err
	}

	// Fetch CVEs with incremental update support
	params := map[string]interface{}{
		"limit": p.batchSize,
	}
	
	// Use incremental fetching if lastModStartDate is set
	if p.lastModStartDate != "" {
		params["lastModStartDate"] = p.lastModStartDate
	}

	p.logger.Info("Fetching CVE batch: size=%d, lastModStartDate=%s", p.batchSize, p.lastModStartDate)

	resp, err := p.rpcClient.InvokeRPC(ctx, "remote", "RPCFetchCVEBatch", params)
	if err != nil {
		p.errorCount++
		p.logger.Error("Failed to fetch CVE batch: %v", err)
		return fmt.Errorf("failed to fetch CVE batch: %w", err)
	}

	// Check for error response
	if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
		p.errorCount++
		p.logger.Error("CVE fetch returned error: %s", errMsg)
		return fmt.Errorf("CVE fetch failed: %s", errMsg)
	}

	// Extract CVEs from response
	var batchResp struct {
		CVEs []map[string]interface{} `json:"cves"`
	}
	if err := subprocess.UnmarshalPayload(resp, &batchResp); err != nil {
		p.errorCount++
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	cves := batchResp.CVEs
	if len(cves) == 0 {
		p.logger.Info("No more CVEs to fetch, provider completed")
		return nil // No more data, provider will transition to TERMINATED
	}

	// Process and save each CVE
	for i, cveMap := range cves {
		cveID, ok := cveMap["cve_id"].(string)
		if !ok {
			p.errorCount++
			p.logger.Warn("Missing CVE ID at index %d", i)
			continue
		}

		// Save to local storage
		if err := p.saveCVE(ctx, cveMap); err != nil {
			p.errorCount++
			p.logger.Error("Failed to save CVE %s: %v", cveID, err)
			continue
		}

		p.totalProcessed++

		// Save checkpoint every N items
		if p.totalProcessed%int64(p.checkpointInterval) == 0 {
			itemURN, err := urn.Parse(fmt.Sprintf("v2e::nvd::cve::%s", cveID))
			if err != nil {
				p.logger.Error("Failed to parse URN for CVE %s: %v", cveID, err)
			} else {
				if err := p.SaveCheckpoint(itemURN, true, ""); err != nil {
					p.logger.Error("Failed to save checkpoint: %v", err)
				} else {
					p.logger.Info("Checkpoint saved at %s (processed: %d)", itemURN.Key(), p.totalProcessed)
				}
			}
		}

		// Update lastModStartDate for incremental updates
		if lastMod, ok := cveMap["last_modified"].(string); ok {
			p.lastModStartDate = lastMod
		}
	}

	p.logger.Info("Processed CVE batch: %d items, total: %d, errors: %d", len(cves), p.totalProcessed, p.errorCount)
	return nil
}

// saveCVE saves a CVE to local storage via RPC
func (p *CVEProvider) saveCVE(ctx context.Context, cveData map[string]interface{}) error {
	// Check if we should use field-level diffing (Requirement 15)
	cveID, _ := cveData["cve_id"].(string)

	// First, try to get existing CVE
	existingResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCGetCVE", map[string]interface{}{
		"cve_id": cveID,
	})

	// If exists and not error, perform diff
	if err == nil && existingResp != nil {
		if isErr, _ := subprocess.IsErrorResponse(existingResp); !isErr {
			// CVE exists, unmarshal and perform field-level diffing
			var existingCVE map[string]interface{}
			if err := subprocess.UnmarshalPayload(existingResp, &existingCVE); err == nil {
				updateParams := p.diffFields(existingCVE, cveData)
				if len(updateParams) == 0 {
					p.logger.Debug("CVE %s unchanged, skipping update", cveID)
					return nil
				}
				updateParams["cve_id"] = cveID

				updateResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCUpdateCVE", updateParams)
				if err != nil {
					return fmt.Errorf("failed to update CVE: %w", err)
				}
				
				// Check for error response
				if isErr, errMsg := subprocess.IsErrorResponse(updateResp); isErr {
					return fmt.Errorf("update CVE failed: %s", errMsg)
				}
				
				return nil
			}
		}
	}

	// CVE doesn't exist or error occurred, create new
	saveResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCSaveCVE", cveData)
	if err != nil {
		return fmt.Errorf("failed to save CVE: %w", err)
	}

	// Check for error response
	if isErr, errMsg := subprocess.IsErrorResponse(saveResp); isErr {
		return fmt.Errorf("save CVE failed: %s", errMsg)
	}

	return nil
}

// diffFields compares two CVE objects and returns only changed fields
// Implements Requirement 15: Field-Level Data Diffing
func (p *CVEProvider) diffFields(existing, incoming map[string]interface{}) map[string]interface{} {
	changed := make(map[string]interface{})
	
	for key, newVal := range incoming {
		if key == "cve_id" {
			continue // Don't compare ID
		}
		
		oldVal, exists := existing[key]
		if !exists || !deepEqual(oldVal, newVal) {
			changed[key] = newVal
		}
	}
	
	return changed
}

// deepEqual performs deep comparison of interface{} values
func deepEqual(a, b interface{}) bool {
	// Simple comparison - can be enhanced with reflect.DeepEqual if needed
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// loadLastCheckpoint loads the last checkpoint from storage
func (p *CVEProvider) loadLastCheckpoint() error {
	stats := p.GetStats()
	checkpoint, _ := stats["last_checkpoint"].(string)

	if checkpoint != "" {
		_, err := urn.Parse(checkpoint)
		if err == nil {
			p.logger.Info("Resuming from checkpoint: %s", checkpoint)
			// Extract CVE ID from URN for incremental fetching
			// URN format: v2e::nvd::cve::CVE-XXXX-XXXX
		}
	}

	return nil
}

// checkErrorThreshold checks if error rate exceeds threshold
// Implements Requirement 14: Pause-on-Error Threshold
func (p *CVEProvider) checkErrorThreshold() error {
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
func (p *CVEProvider) GetProgress() map[string]interface{} {
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
	}
}
