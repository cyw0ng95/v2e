//go:build CONFIG_USE_LIBXML2

package capec

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// StreamingBatchConfig configures batch processing for CAPEC import
type StreamingBatchConfig struct {
	BatchSize int // Number of Attack_Patterns to batch before committing
}

// DefaultStreamingBatchConfig returns default batch configuration
func DefaultStreamingBatchConfig() StreamingBatchConfig {
	return StreamingBatchConfig{
		BatchSize: 100, // Commit every 100 Attack_Patterns
	}
}

// StreamingCAPECParser implements streaming XML parsing with batch processing
// to minimize memory usage and transaction overhead for large CAPEC files
type StreamingCAPECParser struct {
	db        *gorm.DB
	config    StreamingBatchConfig
	tx        *gorm.DB
	batch     batchBuffer
	patterns  []CAPECAttackPattern
	itemCount int
}

type batchBuffer struct {
	items       []CAPECItemModel
	weaknesses  []CAPECRelatedWeaknessModel
	examples    []CAPECExampleModel
	mitigations []CAPECMitigationModel
	references  []CAPECReferenceModel
}

// NewStreamingCAPECParser creates a new streaming parser
func NewStreamingCAPECParser(db *gorm.DB, config StreamingBatchConfig) *StreamingCAPECParser {
	if config.BatchSize <= 0 {
		config = DefaultStreamingBatchConfig()
	}
	return &StreamingCAPECParser{
		db:     db,
		config: config,
		tx:     nil,
		batch: batchBuffer{
			items:       make([]CAPECItemModel, 0, config.BatchSize),
			weaknesses:  make([]CAPECRelatedWeaknessModel, 0, config.BatchSize*5), // Estimate 5 weaknesses per pattern
			examples:    make([]CAPECExampleModel, 0, config.BatchSize*3),
			mitigations: make([]CAPECMitigationModel, 0, config.BatchSize*2),
			references:  make([]CAPECReferenceModel, 0, config.BatchSize*3),
		},
		patterns: make([]CAPECAttackPattern, 0, config.BatchSize),
	}
}

// Parse begins streaming XML parsing from the provided reader
func (p *StreamingCAPECParser) Parse(r io.Reader) error {
	// Create limited reader for security
	limitedReader := io.LimitReader(r, MaxXMLFileSize)

	decoder := xml.NewDecoder(limitedReader)
	decoder.Strict = true
	decoder.AutoClose = xml.HTMLAutoClose

	// Begin transaction
	p.tx = p.db.Begin()
	defer func() {
		if p.tx != nil {
			p.tx.Rollback()
		}
	}()

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("xml token error: %w", err)
		}

		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Attack_Pattern" {
				var ap CAPECAttackPattern
				if err := decoder.DecodeElement(&ap, &se); err != nil {
					return fmt.Errorf("decode attack pattern: %w", err)
				}
				if err := p.addPattern(ap); err != nil {
					return fmt.Errorf("add pattern: %w", err)
				}
			}
		}
	}

	// Flush remaining batch
	if err := p.flushBatch(); err != nil {
		return err
	}

	// Commit transaction
	if err := p.tx.Commit().Error; err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	p.tx = nil
	return nil
}

// addPattern adds a pattern to the current batch
func (p *StreamingCAPECParser) addPattern(ap CAPECAttackPattern) error {
	// Prepare summary
	summary := ap.Summary
	if strings.TrimSpace(summary) == "" {
		summary = truncateString(strings.TrimSpace(ap.Description.XML), 200)
	}

	// Add to items batch
	item := CAPECItemModel{
		CAPECID:         ap.ID,
		Name:            ap.Name,
		Summary:         summary,
		Description:     ap.Description.XML,
		Status:          ap.Status,
		Abstraction:     ap.Abstraction,
		Likelihood:      ap.Likelihood,
		TypicalSeverity: ap.TypicalSeverity,
	}
	p.batch.items = append(p.batch.items, item)

	// Collect related weaknesses with deduplication
	seenCWEs := make(map[string]bool)
	for _, cwe := range ap.RelatedWeaknesses {
		if cwe.CWEID != "" && !seenCWEs[cwe.CWEID] {
			seenCWEs[cwe.CWEID] = true
			p.batch.weaknesses = append(p.batch.weaknesses, CAPECRelatedWeaknessModel{
				CAPECID: ap.ID,
				CWEID:   cwe.CWEID,
			})
		}
	}

	// Collect examples
	for _, ex := range ap.Examples {
		e := strings.TrimSpace(ex.XML)
		if e != "" {
			p.batch.examples = append(p.batch.examples, CAPECExampleModel{
				CAPECID:     ap.ID,
				ExampleText: e,
			})
		}
	}

	// Collect mitigations
	for _, m := range ap.Mitigations {
		mm := strings.TrimSpace(m.XML)
		if mm != "" {
			p.batch.mitigations = append(p.batch.mitigations, CAPECMitigationModel{
				CAPECID:        ap.ID,
				MitigationText: mm,
			})
		}
	}

	// Collect references with deduplication
	seenRefs := make(map[string]bool)
	for _, rref := range ap.References {
		ref := rref.ExternalRef
		if ref != "" && !seenRefs[ref] {
			seenRefs[ref] = true
			p.batch.references = append(p.batch.references, CAPECReferenceModel{
				CAPECID:         ap.ID,
				ExternalReference: ref,
				URL:              "",
			})
		}
	}

	p.patterns = append(p.patterns, ap)
	p.itemCount++

	// Flush batch if size limit reached
	if p.itemCount >= p.config.BatchSize {
		return p.flushBatch()
	}
	return nil
}

// flushBatch commits the current batch to database
func (p *StreamingCAPECParser) flushBatch() error {
	if len(p.batch.items) == 0 {
		return nil
	}

	// Collect CAPEC IDs for deletion
	capecIDs := make([]int, len(p.batch.items))
	for i, item := range p.batch.items {
		capecIDs[i] = item.CAPECID
	}

	// Delete existing related data for these CAPEC IDs
	if err := p.tx.Where("capec_id IN ?", capecIDs).Delete(&CAPECRelatedWeaknessModel{}).Error; err != nil {
		p.tx.Rollback()
		return fmt.Errorf("delete old weaknesses: %w", err)
	}
	if err := p.tx.Where("capec_id IN ?", capecIDs).Delete(&CAPECExampleModel{}).Error; err != nil {
		p.tx.Rollback()
		return fmt.Errorf("delete old examples: %w", err)
	}
	if err := p.tx.Where("capec_id IN ?", capecIDs).Delete(&CAPECMitigationModel{}).Error; err != nil {
		p.tx.Rollback()
		return fmt.Errorf("delete old mitigations: %w", err)
	}
	if err := p.tx.Where("capec_id IN ?", capecIDs).Delete(&CAPECReferenceModel{}).Error; err != nil {
		p.tx.Rollback()
		return fmt.Errorf("delete old references: %w", err)
	}

	// Batch insert main items
	if err := p.tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&p.batch.items).Error; err != nil {
		p.tx.Rollback()
		return fmt.Errorf("insert items: %w", err)
	}

	// Batch insert related data
	if len(p.batch.weaknesses) > 0 {
		if err := p.tx.Create(&p.batch.weaknesses).Error; err != nil {
			p.tx.Rollback()
			return fmt.Errorf("insert weaknesses: %w", err)
		}
	}
	if len(p.batch.examples) > 0 {
		if err := p.tx.Create(&p.batch.examples).Error; err != nil {
			p.tx.Rollback()
			return fmt.Errorf("insert examples: %w", err)
		}
	}
	if len(p.batch.mitigations) > 0 {
		if err := p.tx.Create(&p.batch.mitigations).Error; err != nil {
			p.tx.Rollback()
			return fmt.Errorf("insert mitigations: %w", err)
		}
	}
	if len(p.batch.references) > 0 {
		if err := p.tx.Create(&p.batch.references).Error; err != nil {
			p.tx.Rollback()
			return fmt.Errorf("insert references: %w", err)
		}
	}

	// Clear batch buffers
	p.batch.items = p.batch.items[:0]
	p.batch.weaknesses = p.batch.weaknesses[:0]
	p.batch.examples = p.batch.examples[:0]
	p.batch.mitigations = p.batch.mitigations[:0]
	p.batch.references = p.batch.references[:0]
	p.patterns = p.patterns[:0]
	p.itemCount = 0

	return nil
}

// SetCatalogMeta persists catalog metadata after successful import
func (p *StreamingCAPECParser) SetCatalogMeta(version, source string, importedAtUTC int64) error {
	if p.db == nil {
		return fmt.Errorf("parser not initialized")
	}
	meta := CAPECCatalogMeta{
		ID:            1,
		Version:       version,
		Source:        source,
		ImportedAtUTC: importedAtUTC,
	}
	return p.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(&meta).Error
}
