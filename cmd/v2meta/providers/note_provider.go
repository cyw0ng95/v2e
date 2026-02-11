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

// NoteProvider implements Provider for automatic note generation from populated security data
// This provider creates structured learning notes from CVE/CWE/CAPEC/ATT&CK items
type NoteProvider struct {
	*fsm.BaseProviderFSM
	rpcClient          taskflow.RPCInvoker
	logger             *common.Logger
	batchSize          int
	checkpointInterval int
	itemType           string // "CVE", "CWE", "CAPEC", "ATT&CK"
	errorCount         int64
	totalProcessed     int64
	currentOffset      int
	failureThreshold   float64
}

// NoteProviderConfig holds configuration for note provider
type NoteProviderConfig struct {
	ID                 string
	Storage            *storage.Store
	RPCClient          taskflow.RPCInvoker
	Logger             *common.Logger
	BatchSize          int
	CheckpointInterval int
	ItemType           string
	FailureThreshold   float64
}

// NewNoteProvider creates a new note provider
func NewNoteProvider(config NoteProviderConfig) (*NoteProvider, error) {
	if config.BatchSize <= 0 {
		config.BatchSize = 50
	}
	if config.CheckpointInterval <= 0 {
		config.CheckpointInterval = 50
	}
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 0.1
	}

	var provider *NoteProvider
	executor := func() error {
		if provider == nil {
			return fmt.Errorf("provider not initialized")
		}
		return provider.executeBatch()
	}

	baseFSM, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           config.ID,
		ProviderType: "note-" + config.ItemType,
		Storage:      config.Storage,
		Executor:     executor,
	})
	if err != nil {
		return nil, err
	}

	provider = &NoteProvider{
		BaseProviderFSM:    baseFSM,
		rpcClient:          config.RPCClient,
		logger:             config.Logger,
		batchSize:          config.BatchSize,
		checkpointInterval: config.CheckpointInterval,
		itemType:           config.ItemType,
		failureThreshold:   config.FailureThreshold,
	}

	if err := provider.loadLastCheckpoint(); err != nil {
		config.Logger.Warn("Failed to load checkpoint, starting fresh: %v", err)
	}

	return provider, nil
}

// executeBatch performs one batch of note generation
func (p *NoteProvider) executeBatch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := p.checkErrorThreshold(); err != nil {
		return err
	}

	params := map[string]interface{}{
		"offset": p.currentOffset,
		"limit":  p.batchSize,
	}

	p.logger.Info("Generating notes for %s: offset=%d, size=%d", p.itemType, p.currentOffset, p.batchSize)

	var listMethod string
	switch p.itemType {
	case "CVE":
		listMethod = "RPCListCVEs"
	case "CWE":
		listMethod = "RPCListCWEs"
	case "CAPEC":
		listMethod = "RPCListCAPECs"
	case "ATT&CK":
		listMethod = "RPCListATTACKs"
	default:
		return fmt.Errorf("unsupported item type: %s", p.itemType)
	}

	resp, err := p.rpcClient.InvokeRPC(ctx, "local", listMethod, params)
	if err != nil {
		p.errorCount++
		p.logger.Error("Failed to list %s items: %v", p.itemType, err)
		return fmt.Errorf("failed to list %s items: %w", p.itemType, err)
	}

	if isErr, errMsg := subprocess.IsErrorResponse(resp.(*subprocess.Message)); isErr {
		p.errorCount++
		p.logger.Error("List %s returned error: %s", p.itemType, errMsg)
		return fmt.Errorf("list %s failed: %s", p.itemType, errMsg)
	}

	var batchResp struct {
		Items []map[string]interface{} `json:"items"`
		Total int64                    `json:"total"`
	}
	if err := subprocess.UnmarshalPayload(resp.(*subprocess.Message), &batchResp); err != nil {
		p.errorCount++
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	items := batchResp.Items
	if len(items) == 0 {
		p.logger.Info("No more %s items to process, provider completed", p.itemType)
		return nil
	}

	for i, item := range items {
		itemID, ok := item["id"].(string)
		if !ok {
			itemID, ok = item["cve_id"].(string)
		}
		if !ok {
			itemID, ok = item["cwe_id"].(string)
		}
		if !ok {
			p.errorCount++
			p.logger.Warn("Missing ID at index %d", i)
			continue
		}

		if err := p.generateNote(ctx, itemID, item); err != nil {
			p.errorCount++
			p.logger.Error("Failed to generate note for %s %s: %v", p.itemType, itemID, err)
			continue
		}

		p.totalProcessed++

		if p.totalProcessed%int64(p.checkpointInterval) == 0 {
			itemURN, err := p.createURN(itemID)
			if err != nil {
				p.logger.Error("Failed to parse URN for %s: %v", itemID, err)
			} else {
				if err := p.SaveCheckpoint(itemURN, true, ""); err != nil {
					p.logger.Error("Failed to save checkpoint: %v", err)
				} else {
					p.logger.Info("Checkpoint saved at %s (processed: %d)", itemURN.Key(), p.totalProcessed)
				}
			}
		}
	}

	p.currentOffset += len(items)
	p.logger.Info("Processed note generation batch: %d items, total: %d, errors: %d", len(items), p.totalProcessed, p.errorCount)
	return nil
}

// generateNote creates a bookmark and note for an item
func (p *NoteProvider) generateNote(ctx context.Context, itemID string, itemData map[string]interface{}) error {
	title := p.extractTitle(itemData)
	description := p.extractDescription(itemData)
	globalItemID := fmt.Sprintf("%s::%s", p.itemType, itemID)

	createResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCCreateBookmark", map[string]interface{}{
		"global_item_id": globalItemID,
		"item_type":      p.itemType,
		"item_id":        itemID,
		"title":          title,
		"description":    description,
	})
	if err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	if isErr, errMsg := subprocess.IsErrorResponse(createResp.(*subprocess.Message)); isErr {
		if errMsg == "bookmark already exists" {
			p.logger.Debug("Bookmark already exists for %s %s", p.itemType, itemID)
			return nil
		}
		return fmt.Errorf("create bookmark failed: %s", errMsg)
	}

	noteContent := p.generateNoteContent(itemData)
	if noteContent == "" {
		p.logger.Debug("No note content generated for %s %s", p.itemType, itemID)
		return nil
	}

	var bookmarkID uint
	var result map[string]interface{}
	if err := subprocess.UnmarshalPayload(createResp.(*subprocess.Message), &result); err == nil {
		if bookmark, ok := result["bookmark"].(map[string]interface{}); ok {
			if id, ok := bookmark["id"].(float64); ok {
				bookmarkID = uint(id)
			}
		}
	}

	if bookmarkID == 0 {
		return fmt.Errorf("failed to get bookmark ID")
	}

	noteResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCAddNote", map[string]interface{}{
		"bookmark_id": bookmarkID,
		"content":     noteContent,
		"author":      nil,
		"is_private":  false,
	})
	if err != nil {
		return fmt.Errorf("failed to add note: %w", err)
	}

	if isErr, errMsg := subprocess.IsErrorResponse(noteResp.(*subprocess.Message)); isErr {
		return fmt.Errorf("add note failed: %s", errMsg)
	}

	p.logger.Debug("Created note for %s %s", p.itemType, itemID)
	return nil
}

// extractTitle extracts title from item data
func (p *NoteProvider) extractTitle(item map[string]interface{}) string {
	if title, ok := item["title"].(string); ok && title != "" {
		return title
	}
	if title, ok := item["name"].(string); ok && title != "" {
		return title
	}
	if id, ok := item["cve_id"].(string); ok && id != "" {
		return fmt.Sprintf("%s: %s", p.itemType, id)
	}
	if id, ok := item["cwe_id"].(string); ok && id != "" {
		return fmt.Sprintf("%s: %s", p.itemType, id)
	}
	return fmt.Sprintf("%s Item", p.itemType)
}

// extractDescription extracts description from item data
func (p *NoteProvider) extractDescription(item map[string]interface{}) string {
	if desc, ok := item["description"].(string); ok && desc != "" {
		return desc
	}
	return ""
}

// generateNoteContent generates rich text note content from item data
func (p *NoteProvider) generateNoteContent(item map[string]interface{}) string {
	var content string

	switch p.itemType {
	case "CVE":
		content = p.generateCVENoteContent(item)
	case "CWE":
		content = p.generateCWENoteContent(item)
	case "CAPEC":
		content = p.generateCAPECNoteContent(item)
	case "ATT&CK":
		content = p.generateATTACKNoteContent(item)
	}

	return content
}

// generateCVENoteContent generates note content for CVE
func (p *NoteProvider) generateCVENoteContent(item map[string]interface{}) string {
	cveID, _ := item["cve_id"].(string)
	content := fmt.Sprintf("# %s\n\n", cveID)

	if desc, ok := item["description"].(string); ok {
		content += fmt.Sprintf("## Description\n%s\n\n", desc)
	}

	if metrics, ok := item["metrics"].(map[string]interface{}); ok {
		if cvss, ok := metrics["cvss_metric_v30"].(map[string]interface{}); ok {
			if baseScore, ok := cvss["baseScore"].(float64); ok {
				content += fmt.Sprintf("## CVSS Score\n**%.1f**\n\n", baseScore)
			}
			if severity, ok := cvss["baseSeverity"].(string); ok {
				content += fmt.Sprintf("Severity: **%s**\n\n", severity)
			}
		}
	}

	if refs, ok := item["references"].([]interface{}); ok && len(refs) > 0 {
		content += "## References\n"
		for _, ref := range refs {
			if refMap, ok := ref.(map[string]interface{}); ok {
				if url, ok := refMap["url"].(string); ok {
					content += fmt.Sprintf("- [%s](%s)\n", url, url)
				}
			}
		}
		content += "\n"
	}

	return content
}

// generateCWENoteContent generates note content for CWE
func (p *NoteProvider) generateCWENoteContent(item map[string]interface{}) string {
	cweID, _ := item["cwe_id"].(string)
	content := fmt.Sprintf("# %s\n\n", cweID)

	if desc, ok := item["description"].(string); ok {
		content += fmt.Sprintf("## Description\n%s\n\n", desc)
	}

	if name, ok := item["name"].(string); ok {
		content += fmt.Sprintf("## Name\n%s\n\n", name)
	}

	if status, ok := item["status"].(string); ok {
		content += fmt.Sprintf("## Status\n%s\n\n", status)
	}

	return content
}

// generateCAPECNoteContent generates note content for CAPEC
func (p *NoteProvider) generateCAPECNoteContent(item map[string]interface{}) string {
	capecID, _ := item["capec_id"].(string)
	content := fmt.Sprintf("# %s\n\n", capecID)

	if desc, ok := item["description"].(string); ok {
		content += fmt.Sprintf("## Description\n%s\n\n", desc)
	}

	if name, ok := item["name"].(string); ok {
		content += fmt.Sprintf("## Name\n%s\n\n", name)
	}

	return content
}

// generateATTACKNoteContent generates note content for ATT&CK
func (p *NoteProvider) generateATTACKNoteContent(item map[string]interface{}) string {
	attackID, _ := item["attack_id"].(string)
	content := fmt.Sprintf("# %s\n\n", attackID)

	if desc, ok := item["description"].(string); ok {
		content += fmt.Sprintf("## Description\n%s\n\n", desc)
	}

	if name, ok := item["name"].(string); ok {
		content += fmt.Sprintf("## Name\n%s\n\n", name)
	}

	if tactics, ok := item["tactics"].([]interface{}); ok && len(tactics) > 0 {
		content += "## Tactics\n"
		for _, tactic := range tactics {
			if tacticStr, ok := tactic.(string); ok {
				content += fmt.Sprintf("- %s\n", tacticStr)
			}
		}
		content += "\n"
	}

	return content
}

// createURN creates a URN for an item
func (p *NoteProvider) createURN(itemID string) (*urn.URN, error) {
	var provider, resourceType string

	switch p.itemType {
	case "CVE":
		provider = "nvd"
		resourceType = "cve"
	case "CWE", "CAPEC", "ATT&CK":
		provider = "mitre"
		switch p.itemType {
		case "CWE":
			resourceType = "cwe"
		case "CAPEC":
			resourceType = "capec"
		case "ATT&CK":
			resourceType = "attack"
		}
	}

	return urn.Parse(fmt.Sprintf("v2e::%s::%s::%s", provider, resourceType, itemID))
}

// loadLastCheckpoint loads the last checkpoint from storage
func (p *NoteProvider) loadLastCheckpoint() error {
	stats := p.GetStats()
	checkpoint, _ := stats["last_checkpoint"].(string)

	if checkpoint != "" {
		_, err := urn.Parse(checkpoint)
		if err == nil {
			p.logger.Info("Resuming from checkpoint: %s", checkpoint)
		}
	}

	return nil
}

// checkErrorThreshold checks if error rate exceeds threshold
func (p *NoteProvider) checkErrorThreshold() error {
	if p.totalProcessed == 0 {
		return nil
	}

	errorRate := float64(p.errorCount) / float64(p.totalProcessed)
	if errorRate > p.failureThreshold {
		p.logger.Error("Error rate %.2f%% exceeds threshold %.2f%%, auto-pausing provider",
			errorRate*100, p.failureThreshold*100)

		if err := p.Transition(fsm.ProviderPaused); err != nil {
			return fmt.Errorf("failed to pause provider: %w", err)
		}

		return fmt.Errorf("provider auto-paused due to high error rate: %.2f%%", errorRate*100)
	}

	return nil
}

// GetProgress returns current progress metrics
func (p *NoteProvider) GetProgress() map[string]interface{} {
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
		"item_type":       p.itemType,
	}
}
