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

// MemoryCardProvider implements Provider for automatic memory card generation from bookmarks
// This provider creates spaced-repetition cards for existing bookmarks with rich content
type MemoryCardProvider struct {
	*fsm.BaseProviderFSM
	rpcClient          taskflow.RPCInvoker
	logger             *common.Logger
	batchSize          int
	checkpointInterval int
	errorCount         int64
	totalProcessed     int64
	currentOffset      int
	failureThreshold   float64
}

// MemoryCardProviderConfig holds configuration for memory card provider
type MemoryCardProviderConfig struct {
	ID                 string
	Storage            *storage.Store
	RPCClient          taskflow.RPCInvoker
	Logger             *common.Logger
	BatchSize          int
	CheckpointInterval int
	FailureThreshold   float64
}

// NewMemoryCardProvider creates a new memory card provider
func NewMemoryCardProvider(config MemoryCardProviderConfig) (*MemoryCardProvider, error) {
	if config.BatchSize <= 0 {
		config.BatchSize = 50
	}
	if config.CheckpointInterval <= 0 {
		config.CheckpointInterval = 50
	}
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 0.1
	}

	var provider *MemoryCardProvider
	executor := func() error {
		if provider == nil {
			return fmt.Errorf("provider not initialized")
		}
		return provider.executeBatch()
	}

	baseFSM, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           config.ID,
		ProviderType: "memorycard",
		Storage:      config.Storage,
		Executor:     executor,
	})
	if err != nil {
		return nil, err
	}

	provider = &MemoryCardProvider{
		BaseProviderFSM:    baseFSM,
		rpcClient:          config.RPCClient,
		logger:             config.Logger,
		batchSize:          config.BatchSize,
		checkpointInterval: config.CheckpointInterval,
		failureThreshold:   config.FailureThreshold,
	}

	if err := provider.loadLastCheckpoint(); err != nil {
		config.Logger.Warn("Failed to load checkpoint, starting fresh: %v", err)
	}

	return provider, nil
}

// executeBatch performs one batch of memory card generation
func (p *MemoryCardProvider) executeBatch() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := p.checkErrorThreshold(); err != nil {
		return err
	}

	params := map[string]interface{}{
		"offset": p.currentOffset,
		"limit":  p.batchSize,
	}

	p.logger.Info("Generating memory cards: offset=%d, size=%d", p.currentOffset, p.batchSize)

	resp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCListBookmarks", params)
	if err != nil {
		p.errorCount++
		p.logger.Error("Failed to list bookmarks: %v", err)
		return fmt.Errorf("failed to list bookmarks: %w", err)
	}

	if isErr, errMsg := subprocess.IsErrorResponse(resp.(*subprocess.Message)); isErr {
		p.errorCount++
		p.logger.Error("List bookmarks returned error: %s", errMsg)
		return fmt.Errorf("list bookmarks failed: %s", errMsg)
	}

	var batchResp struct {
		Bookmarks []map[string]interface{} `json:"bookmarks"`
		Total     int64                    `json:"total"`
	}
	if err := subprocess.UnmarshalPayload(resp.(*subprocess.Message), &batchResp); err != nil {
		p.errorCount++
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	bookmarks := batchResp.Bookmarks
	if len(bookmarks) == 0 {
		p.logger.Info("No more bookmarks to process, provider completed")
		return nil
	}

	for i, bookmark := range bookmarks {
		bookmarkIDFloat, ok := bookmark["id"].(float64)
		if !ok {
			p.errorCount++
			p.logger.Warn("Missing bookmark ID at index %d", i)
			continue
		}
		bookmarkID := uint(bookmarkIDFloat)

		itemType, _ := bookmark["item_type"].(string)
		itemID, _ := bookmark["item_id"].(string)
		title, _ := bookmark["title"].(string)

		if err := p.generateMemoryCards(ctx, bookmarkID, itemType, itemID, title, bookmark); err != nil {
			p.errorCount++
			p.logger.Error("Failed to generate memory cards for bookmark %d: %v", bookmarkID, err)
			continue
		}

		p.totalProcessed++

		if p.totalProcessed%int64(p.checkpointInterval) == 0 {
			itemURN, err := urn.Parse(fmt.Sprintf("v2e::notes::bookmark::%d", bookmarkID))
			if err != nil {
				p.logger.Error("Failed to parse URN for bookmark %d: %v", bookmarkID, err)
			} else {
				if err := p.SaveCheckpoint(itemURN, true, ""); err != nil {
					p.logger.Error("Failed to save checkpoint: %v", err)
				} else {
					p.logger.Info("Checkpoint saved at %s (processed: %d)", itemURN.Key(), p.totalProcessed)
				}
			}
		}
	}

	p.currentOffset += len(bookmarks)
	p.logger.Info("Processed memory card generation batch: %d bookmarks, total: %d, errors: %d", len(bookmarks), p.totalProcessed, p.errorCount)
	return nil
}

// generateMemoryCards creates multiple memory cards for a single bookmark
func (p *MemoryCardProvider) generateMemoryCards(ctx context.Context, bookmarkID uint, itemType, itemID, title string, bookmark map[string]interface{}) error {
	description, _ := bookmark["description"].(string)

	cards := p.generateCardTemplates(itemType, itemID, title, description)

	for _, card := range cards {
		createResp, err := p.rpcClient.InvokeRPC(ctx, "local", "RPCCreateMemoryCard", map[string]interface{}{
			"bookmark_id":   bookmarkID,
			"front_content": card.Front,
			"back_content":  card.Back,
			"major_class":   card.MajorClass,
			"minor_class":   card.MinorClass,
			"status":        card.Status,
			"content":       card.Content,
			"card_type":     card.CardType,
			"author":        card.Author,
			"is_private":    card.IsPrivate,
			"metadata":      card.Metadata,
		})
		if err != nil {
			return fmt.Errorf("failed to create memory card: %w", err)
		}

		if isErr, errMsg := subprocess.IsErrorResponse(createResp.(*subprocess.Message)); isErr {
			if errMsg == "memory card already exists" {
				p.logger.Debug("Memory card already exists for bookmark %d", bookmarkID)
				continue
			}
			return fmt.Errorf("create memory card failed: %s", errMsg)
		}

		p.logger.Debug("Created memory card for bookmark %d: %s", bookmarkID, card.MajorClass)
	}

	return nil
}

// CardTemplate holds template data for a memory card
type CardTemplate struct {
	Front      string
	Back       string
	MajorClass string
	MinorClass string
	Status     string
	Content    string
	CardType   string
	Author     string
	IsPrivate  bool
	Metadata   map[string]interface{}
}

// generateCardTemplates generates card templates based on item type
func (p *MemoryCardProvider) generateCardTemplates(itemType, itemID, title, description string) []CardTemplate {
	var cards []CardTemplate

	switch itemType {
	case "CVE":
		cards = p.generateCVETemplates(itemID, title, description)
	case "CWE":
		cards = p.generateCWETemplates(itemID, title, description)
	case "CAPEC":
		cards = p.generateCAPECTemplates(itemID, title, description)
	case "ATT&CK":
		cards = p.generateATTACKTemplates(itemID, title, description)
	default:
		cards = []CardTemplate{p.generateGenericTemplate(itemType, itemID, title, description)}
	}

	return cards
}

// generateCVETemplates generates cards for CVE items
func (p *MemoryCardProvider) generateCVETemplates(cveID, title, description string) []CardTemplate {
	return []CardTemplate{
		{
			Front:      fmt.Sprintf("What is CVE-%s?", cveID),
			Back:       fmt.Sprintf("%s\n\nDescription: %s", title, description),
			MajorClass: "CVE",
			MinorClass: "Identification",
			Status:     "new",
			Content:    "{}",
			CardType:   "basic",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"cve_id": cveID,
				"type":   "identification",
			},
		},
		{
			Front:      fmt.Sprintf("What is the description of CVE-%s?", cveID),
			Back:       description,
			MajorClass: "CVE",
			MinorClass: "Description",
			Status:     "new",
			Content:    "{}",
			CardType:   "basic",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"cve_id": cveID,
				"type":   "description",
			},
		},
	}
}

// generateCWETemplates generates cards for CWE items
func (p *MemoryCardProvider) generateCWETemplates(cweID, title, description string) []CardTemplate {
	return []CardTemplate{
		{
			Front:      fmt.Sprintf("What is CWE-%s?", cweID),
			Back:       fmt.Sprintf("%s\n\nDescription: %s", title, description),
			MajorClass: "CWE",
			MinorClass: "Identification",
			Status:     "new",
			Content:    "{}",
			CardType:   "basic",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"cwe_id": cweID,
				"type":   "identification",
			},
		},
		{
			Front:      fmt.Sprintf("What weakness category does CWE-%s belong to?", cweID),
			Back:       description,
			MajorClass: "CWE",
			MinorClass: "Classification",
			Status:     "new",
			Content:    "{}",
			CardType:   "classification",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"cwe_id": cweID,
				"type":   "classification",
			},
		},
	}
}

// generateCAPECTemplates generates cards for CAPEC items
func (p *MemoryCardProvider) generateCAPECTemplates(capecID, title, description string) []CardTemplate {
	return []CardTemplate{
		{
			Front:      fmt.Sprintf("What is CAPEC-%s?", capecID),
			Back:       fmt.Sprintf("%s\n\nDescription: %s", title, description),
			MajorClass: "CAPEC",
			MinorClass: "Identification",
			Status:     "new",
			Content:    "{}",
			CardType:   "basic",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"capec_id": capecID,
				"type":     "identification",
			},
		},
		{
			Front:      fmt.Sprintf("What attack pattern does CAPEC-%s describe?", capecID),
			Back:       description,
			MajorClass: "CAPEC",
			MinorClass: "Attack Pattern",
			Status:     "new",
			Content:    "{}",
			CardType:   "pattern",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"capec_id": capecID,
				"type":     "pattern",
			},
		},
	}
}

// generateATTACKTemplates generates cards for ATT&CK items
func (p *MemoryCardProvider) generateATTACKTemplates(attackID, title, description string) []CardTemplate {
	return []CardTemplate{
		{
			Front:      fmt.Sprintf("What is ATT&CK technique %s?", attackID),
			Back:       fmt.Sprintf("%s\n\nDescription: %s", title, description),
			MajorClass: "ATT&CK",
			MinorClass: "Technique",
			Status:     "new",
			Content:    "{}",
			CardType:   "basic",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"attack_id": attackID,
				"type":      "technique",
			},
		},
		{
			Front:      fmt.Sprintf("What tactics are associated with %s?", title),
			Back:       description,
			MajorClass: "ATT&CK",
			MinorClass: "Tactics",
			Status:     "new",
			Content:    "{}",
			CardType:   "tactic",
			Author:     "system",
			IsPrivate:  false,
			Metadata: map[string]interface{}{
				"attack_id": attackID,
				"type":      "tactics",
			},
		},
	}
}

// generateGenericTemplate generates a generic card for unknown item types
func (p *MemoryCardProvider) generateGenericTemplate(itemType, itemID, title, description string) CardTemplate {
	return CardTemplate{
		Front:      fmt.Sprintf("What is %s %s?", itemType, itemID),
		Back:       fmt.Sprintf("%s\n\nDescription: %s", title, description),
		MajorClass: itemType,
		MinorClass: "General",
		Status:     "new",
		Content:    "{}",
		CardType:   "basic",
		Author:     "system",
		IsPrivate:  false,
		Metadata: map[string]interface{}{
			"item_type": itemType,
			"item_id":   itemID,
			"type":      "general",
		},
	}
}

// loadLastCheckpoint loads the last checkpoint from storage
func (p *MemoryCardProvider) loadLastCheckpoint() error {
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
func (p *MemoryCardProvider) checkErrorThreshold() error {
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
func (p *MemoryCardProvider) GetProgress() map[string]interface{} {
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
		"provider_type":   "memorycard",
	}
}
