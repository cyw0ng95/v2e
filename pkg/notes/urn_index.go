package notes

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// URNLink represents a bidirectional relationship between a note/card and a security item
type URNLink struct {
	ID         uint   `gorm:"primaryKey"`
	SourceURN  string `gorm:"index;index:idx_source_urn_type;not null"`
	SourceType string `gorm:"index:idx_source_urn_type;not null"` // "note", "card"
	TargetURN  string `gorm:"index;idx_target_urn;not null"`
	TargetType string `gorm:"index;not null"` // "cve", "cwe", "capec", "attack"
	CreatedAt  int64  `gorm:"autoCreateTime:millisecond"`
}

// URNIndex maintains bidirectional relationships between notes/cards and security items
type URNIndex struct {
	db *gorm.DB
	mu sync.RWMutex
}

// NewURNIndex creates a new URN index
func NewURNIndex(db *gorm.DB) *URNIndex {
	return &URNIndex{db: db}
}

// AddLink creates a bidirectional link between a note/card and a security item
func (idx *URNIndex) AddLink(sourceURN, sourceType, targetURN, targetType string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	link := &URNLink{
		SourceURN:  sourceURN,
		SourceType: sourceType,
		TargetURN:  targetURN,
		TargetType: targetType,
	}

	if err := idx.db.Create(link).Error; err != nil {
		return fmt.Errorf("failed to create URN link: %w", err)
	}

	return nil
}

// RemoveLink deletes a link
func (idx *URNIndex) RemoveLink(sourceURN, targetURN string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if err := idx.db.Where("source_urn = ? AND target_urn = ?", sourceURN, targetURN).Delete(&URNLink{}).Error; err != nil {
		return fmt.Errorf("failed to remove URN link: %w", err)
	}

	return nil
}

// GetNotesForURN returns all notes linked to a security item
func (idx *URNIndex) GetNotesForURN(targetURN string) ([]NoteModel, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var links []URNLink
	if err := idx.db.Where("target_urn = ? AND source_type = ?", targetURN, "note").Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to query note links: %w", err)
	}

	var noteIDs []uint
	for _, link := range links {
		// Extract note ID from URN format: v2e::note::<id>
		var id uint
		if _, err := fmt.Sscanf(link.SourceURN, "v2e::note::%d", &id); err == nil {
			noteIDs = append(noteIDs, id)
		}
	}

	if len(noteIDs) == 0 {
		return []NoteModel{}, nil
	}

	var notes []NoteModel
	if err := idx.db.Where("id IN ?", noteIDs).Find(&notes).Error; err != nil {
		return nil, fmt.Errorf("failed to load notes: %w", err)
	}

	return notes, nil
}

// GetCardsForURN returns all memory cards linked to a security item
func (idx *URNIndex) GetCardsForURN(targetURN string) ([]MemoryCardModel, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var links []URNLink
	if err := idx.db.Where("target_urn = ? AND source_type = ?", targetURN, "card").Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to query card links: %w", err)
	}

	var cardIDs []uint
	for _, link := range links {
		// Extract card ID from URN format: v2e::card::<id>
		var id uint
		if _, err := fmt.Sscanf(link.SourceURN, "v2e::card::%d", &id); err == nil {
			cardIDs = append(cardIDs, id)
		}
	}

	if len(cardIDs) == 0 {
		return []MemoryCardModel{}, nil
	}

	var cards []MemoryCardModel
	if err := idx.db.Where("id IN ?", cardIDs).Find(&cards).Error; err != nil {
		return nil, fmt.Errorf("failed to load cards: %w", err)
	}

	return cards, nil
}

// GetLinkedURNs returns all URNs linked to a note or card
func (idx *URNIndex) GetLinkedURNs(sourceURN string) ([]string, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var links []URNLink
	if err := idx.db.Where("source_urn = ?", sourceURN).Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}

	urns := make([]string, len(links))
	for i, link := range links {
		urns[i] = link.TargetURN
	}

	return urns, nil
}

// GetLinksBySource returns all links for a source URN
func (idx *URNIndex) GetLinksBySource(sourceURN string) ([]URNLink, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var links []URNLink
	if err := idx.db.Where("source_urn = ?", sourceURN).Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}

	return links, nil
}

// GetLinksByTarget returns all links pointing to a target URN
func (idx *URNIndex) GetLinksByTarget(targetURN string) ([]URNLink, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var links []URNLink
	if err := idx.db.Where("target_urn = ?", targetURN).Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to query links: %w", err)
	}

	return links, nil
}

// GetAllLinks returns all links (for debugging/export)
func (idx *URNIndex) GetAllLinks() ([]URNLink, error) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var links []URNLink
	if err := idx.db.Find(&links).Error; err != nil {
		return nil, fmt.Errorf("failed to query all links: %w", err)
	}

	return links, nil
}

// DeleteAllLinksForSource removes all links for a source URN
func (idx *URNIndex) DeleteAllLinksForSource(sourceURN string) error {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if err := idx.db.Where("source_urn = ?", sourceURN).Delete(&URNLink{}).Error; err != nil {
		return fmt.Errorf("failed to delete links: %w", err)
	}

	return nil
}

// MigrateURNLinks runs auto-migration for URN links table
func MigrateURNLinks(db *gorm.DB) error {
	return db.AutoMigrate(&URNLink{})
}
