package notes

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CreateMemoryCardFull creates a new memory card with all fields
func (s *MemoryCardService) CreateMemoryCardFull(ctx context.Context, bookmarkID uint, front, back, majorClass, minorClass, status, content, cardType, author string, isPrivate bool, metadata map[string]any) (*MemoryCardModel, error) {
	card := &MemoryCardModel{
		BookmarkID: bookmarkID,
		Front:      front,
		Back:       back,
		MajorClass: majorClass,
		MinorClass: minorClass,
		Status:     status,
		Content:    content,
		EaseFactor: 2.5,
		Interval:   1,
		Repetition: 0,
		// TODO: CardType, Author, IsPrivate, Metadata (if you add to model)
	}
	if err := s.db.WithContext(ctx).Create(card).Error; err != nil {
		return nil, fmt.Errorf("failed to create memory card: %w", err)
	}
	return card, nil
}

// UpdateMemoryCardFields updates a memory card by ID and arbitrary fields
func (s *MemoryCardService) UpdateMemoryCardFields(ctx context.Context, fields map[string]any) (*MemoryCardModel, error) {
	idAny, ok := fields["id"]
	if !ok {
		return nil, fmt.Errorf("missing id field")
	}
	id, ok := idAny.(float64)
	if !ok {
		return nil, fmt.Errorf("invalid id type")
	}
	// Perform updates transactionally and validate status transitions
	var updated MemoryCardModel
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var card MemoryCardModel
		if err := tx.First(&card, uint(id)).Error; err != nil {
			return err
		}

		// Handle status transition specially
		if rawStatus, ok := fields["status"]; ok {
			if statusStr, ok := rawStatus.(string); ok {
				parsed, perr := ParseCardStatus(statusStr)
				if perr != nil {
					return perr
				}
				// Use centralized transition logic within this transaction
				if err := s.transitionCardStatusTx(tx, card.ID, parsed); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("status must be a string")
			}
			delete(fields, "status")
		}

		// Delete id from fields map and apply other updates
		delete(fields, "id")
		if len(fields) > 0 {
			if err := tx.Model(&card).Updates(fields).Error; err != nil {
				return err
			}
		}

		// Reload the updated card to return
		if err := tx.First(&updated, card.ID).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

// DeleteMemoryCard deletes a memory card by ID
func (s *MemoryCardService) DeleteMemoryCard(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&MemoryCardModel{}, id).Error
}

// ListMemoryCardsFull lists memory cards with new filters
func (s *MemoryCardService) ListMemoryCardsFull(ctx context.Context, bookmarkID *uint, majorClass, minorClass, status, author *string, isPrivate *bool, offset, limit int) ([]*MemoryCardModel, int64, error) {
	var cards []*MemoryCardModel
	query := s.db.WithContext(ctx).Model(&MemoryCardModel{})
	if bookmarkID != nil {
		query = query.Where("bookmark_id = ?", *bookmarkID)
	}
	if majorClass != nil && *majorClass != "" {
		query = query.Where("major_class = ?", *majorClass)
	}
	if minorClass != nil && *minorClass != "" {
		query = query.Where("minor_class = ?", *minorClass)
	}
	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}
	// TODO: author, is_private if added to model
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}
	if err := query.Find(&cards).Error; err != nil {
		return nil, 0, err
	}
	return cards, total, nil
}

// BookmarkService handles all bookmark-related operations
type BookmarkService struct {
	db *gorm.DB
}

// NewBookmarkService creates a new BookmarkService instance
func NewBookmarkService(db *gorm.DB) *BookmarkService {
	return &BookmarkService{db: db}
}

// CreateBookmark creates a new bookmark, enforcing single bookmark per item
func (s *BookmarkService) CreateBookmark(ctx context.Context, globalItemID, itemType, itemID, title, description string) (*BookmarkModel, *MemoryCardModel, error) {
	// Check if bookmark already exists for this item
	var existingBookmark BookmarkModel
	err := s.db.WithContext(ctx).Where("global_item_id = ? AND item_type = ? AND item_id = ?", globalItemID, itemType, itemID).First(&existingBookmark).Error
	if err == nil {
		// Bookmark already exists, return the existing one
		return &existingBookmark, nil, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		// Some other database error occurred
		return nil, nil, fmt.Errorf("failed to check for existing bookmark: %w", err)
	}

	// No existing bookmark found, create new one
	now := time.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	// Generate URN for the bookmark
	urnStr := GenerateURN(itemType, itemID, "")

	bookmark := &BookmarkModel{
		GlobalItemID:  globalItemID,
		ItemType:      itemType,
		ItemID:        itemID,
		URN:           urnStr,
		Title:         title,
		Description:   description,
		LearningState: string(LearningStateToReview),
		MasteryLevel:  0.0,
		// Initialize stats in metadata
		Metadata: map[string]interface{}{
			"view_count":       0,
			"study_sessions":   0,
			"last_viewed":      nowStr,
			"first_bookmarked": nowStr,
		},
	}

	var createdCard *MemoryCardModel

	// Use a transaction so bookmark, history and auto-created memory card are atomic
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(bookmark).Error; err != nil {
			return fmt.Errorf("failed to create bookmark: %w", err)
		}

		// Create history entry for creation
		history := &BookmarkHistoryModel{
			BookmarkID: bookmark.ID,
			Action:     string(BookmarkActionCreated),
			NewValue:   string(LearningStateToReview),
			Timestamp:  time.Now(),
		}
		if err := tx.Create(history).Error; err != nil {
			return fmt.Errorf("failed to create bookmark history: %w", err)
		}

		// Auto-create a memory card for this bookmark using title/description
		card := &MemoryCardModel{
			BookmarkID: bookmark.ID,
			URN:        urnStr,
			Front:      title,
			Back:       description,
			EaseFactor: 2.5,
			Interval:   1,
			Repetition: 0,
		}
		if err := tx.Create(card).Error; err != nil {
			return fmt.Errorf("failed to create memory card: %w", err)
		}
		createdCard = card

		return nil
	})

	if err != nil {
		return nil, nil, err
	}

	return bookmark, createdCard, nil
}

// UpdateBookmarkStats updates the statistics for a bookmark
func (s *BookmarkService) UpdateBookmarkStats(ctx context.Context, bookmarkID uint, viewIncrement int, studyIncrement int) error {
	var bookmark BookmarkModel
	if err := s.db.WithContext(ctx).First(&bookmark, bookmarkID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bookmark with ID %d not found", bookmarkID)
		}
		return fmt.Errorf("failed to find bookmark: %w", err)
	}

	// Initialize metadata if nil
	if bookmark.Metadata == nil {
		bookmark.Metadata = make(map[string]interface{})
	}

	// Update view count
	viewCount, _ := bookmark.Metadata["view_count"].(float64)
	bookmark.Metadata["view_count"] = int(viewCount) + viewIncrement

	// Update study sessions
	studySessions, _ := bookmark.Metadata["study_sessions"].(float64)
	bookmark.Metadata["study_sessions"] = int(studySessions) + studyIncrement

	// Update last viewed timestamp
	bookmark.Metadata["last_viewed"] = time.Now().UTC().Format(time.RFC3339)

	// Update the bookmark
	if err := s.db.WithContext(ctx).Save(&bookmark).Error; err != nil {
		return fmt.Errorf("failed to update bookmark stats: %w", err)
	}

	return nil
}

// GetBookmarkStats retrieves the statistics for a bookmark
func (s *BookmarkService) GetBookmarkStats(ctx context.Context, bookmarkID uint) (map[string]interface{}, error) {
	var bookmark BookmarkModel
	if err := s.db.WithContext(ctx).Select("metadata").First(&bookmark, bookmarkID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bookmark with ID %d not found", bookmarkID)
		}
		return nil, fmt.Errorf("failed to get bookmark stats: %w", err)
	}

	if bookmark.Metadata == nil {
		return make(map[string]interface{}), nil
	}

	return bookmark.Metadata, nil
}

// GetBookmarkByID retrieves a bookmark by its ID
func (s *BookmarkService) GetBookmarkByID(ctx context.Context, id uint) (*BookmarkModel, error) {
	var bookmark BookmarkModel
	err := s.db.WithContext(ctx).Preload("Notes").Preload("History").First(&bookmark, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("bookmark with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}
	return &bookmark, nil
}

// GetBookmarksByGlobalItemID retrieves all bookmarks for a specific global item
func (s *BookmarkService) GetBookmarksByGlobalItemID(ctx context.Context, globalItemID string) ([]*BookmarkModel, error) {
	var bookmarks []*BookmarkModel
	err := s.db.WithContext(ctx).Preload("Notes").Preload("History").Where("global_item_id = ?", globalItemID).Find(&bookmarks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks by global item ID: %w", err)
	}
	return bookmarks, nil
}

// UpdateBookmark updates an existing bookmark
func (s *BookmarkService) UpdateBookmark(ctx context.Context, bookmark *BookmarkModel) error {
	// Get the original bookmark to track changes
	var original BookmarkModel
	if err := s.db.WithContext(ctx).First(&original, bookmark.ID).Error; err != nil {
		return fmt.Errorf("bookmark with ID %d not found", bookmark.ID)
	}

	// Update the bookmark
	if err := s.db.WithContext(ctx).Save(bookmark).Error; err != nil {
		return fmt.Errorf("failed to update bookmark: %w", err)
	}

	// Create history entry if learning state changed
	if original.LearningState != bookmark.LearningState {
		history := &BookmarkHistoryModel{
			BookmarkID: bookmark.ID,
			Action:     string(BookmarkActionLearningStateChanged),
			OldValue:   original.LearningState,
			NewValue:   bookmark.LearningState,
			Timestamp:  time.Now(),
		}
		if err := s.db.WithContext(ctx).Create(history).Error; err != nil {
			return fmt.Errorf("failed to create bookmark history: %w", err)
		}
	}

	return nil
}

// DeleteBookmark deletes a bookmark
func (s *BookmarkService) DeleteBookmark(ctx context.Context, id uint) error {
	bookmark := &BookmarkModel{ID: id}
	if err := s.db.WithContext(ctx).First(bookmark).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("bookmark with ID %d not found", id)
		}
		return fmt.Errorf("failed to find bookmark: %w", err)
	}

	// Create history entry before deletion
	history := &BookmarkHistoryModel{
		BookmarkID: bookmark.ID,
		Action:     string(BookmarkActionDeleted),
		OldValue:   bookmark.LearningState,
		Timestamp:  time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("failed to create bookmark history: %w", err)
	}

	// Delete the bookmark (soft delete)
	if err := s.db.WithContext(ctx).Delete(bookmark).Error; err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	return nil
}

// UpdateLearningState updates the learning state of a bookmark
func (s *BookmarkService) UpdateLearningState(ctx context.Context, bookmarkID uint, newState LearningState) error {
	// Get the original bookmark to track changes
	var original BookmarkModel
	if err := s.db.WithContext(ctx).First(&original, bookmarkID).Error; err != nil {
		return fmt.Errorf("bookmark with ID %d not found", bookmarkID)
	}

	// Update the learning state
	if err := s.db.WithContext(ctx).Model(&BookmarkModel{}).Where("id = ?", bookmarkID).Update("learning_state", string(newState)).Error; err != nil {
		return fmt.Errorf("failed to update learning state: %w", err)
	}

	// Create history entry
	history := &BookmarkHistoryModel{
		BookmarkID: bookmarkID,
		Action:     string(BookmarkActionLearningStateChanged),
		OldValue:   original.LearningState,
		NewValue:   string(newState),
		Timestamp:  time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(history).Error; err != nil {
		return fmt.Errorf("failed to create bookmark history: %w", err)
	}

	return nil
}

// GetBookmarksByLearningState retrieves bookmarks by learning state
func (s *BookmarkService) GetBookmarksByLearningState(ctx context.Context, state LearningState) ([]*BookmarkModel, error) {
	var bookmarks []*BookmarkModel
	err := s.db.WithContext(ctx).Preload("Notes").Preload("History").Where("learning_state = ?", string(state)).Find(&bookmarks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get bookmarks by learning state: %w", err)
	}
	return bookmarks, nil
}

// ListBookmarks lists bookmarks with optional state filtering and pagination
func (s *BookmarkService) ListBookmarks(ctx context.Context, state string, offset, limit int) ([]*BookmarkModel, int64, error) {
	var bookmarks []*BookmarkModel

	query := s.db.WithContext(ctx).Preload("Notes").Preload("History")

	if state != "" {
		query = query.Where("learning_state = ?", state)
	}

	// Get total count
	var total int64
	if err := query.Model(&BookmarkModel{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count bookmarks: %w", err)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&bookmarks).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list bookmarks: %w", err)
	}

	return bookmarks, total, nil
}

// NoteService handles all note-related operations
type NoteService struct {
	db *gorm.DB
}

// NewNoteService creates a new NoteService instance
func NewNoteService(db *gorm.DB) *NoteService {
	return &NoteService{db: db}
}

// AddNote adds a note to a bookmark
func (s *NoteService) AddNote(ctx context.Context, bookmarkID uint, content string, author *string, isPrivate bool) (*NoteModel, error) {
	note := &NoteModel{
		BookmarkID: bookmarkID,
		Content:    content,
		Author:     author,
		IsPrivate:  isPrivate,
	}

	if err := s.db.WithContext(ctx).Create(note).Error; err != nil {
		return nil, fmt.Errorf("failed to add note: %w", err)
	}

	// Create history entry
	bookmarkHistory := &BookmarkHistoryModel{
		BookmarkID: bookmarkID,
		Action:     string(BookmarkActionNoteAdded),
		NewValue:   fmt.Sprintf("Note ID: %d", note.ID),
		Timestamp:  time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(bookmarkHistory).Error; err != nil {
		return nil, fmt.Errorf("failed to create bookmark history: %w", err)
	}

	return note, nil
}

// GetNotesByBookmarkID retrieves all notes for a specific bookmark
func (s *NoteService) GetNotesByBookmarkID(ctx context.Context, bookmarkID uint) ([]*NoteModel, error) {
	var notes []*NoteModel
	err := s.db.WithContext(ctx).Where("bookmark_id = ?", bookmarkID).Find(&notes).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get notes by bookmark ID: %w", err)
	}
	return notes, nil
}

// UpdateNote updates an existing note
func (s *NoteService) UpdateNote(ctx context.Context, note *NoteModel) error {
	if err := s.db.WithContext(ctx).Save(note).Error; err != nil {
		return fmt.Errorf("failed to update note: %w", err)
	}
	return nil
}

// GetNoteByID retrieves a note by its ID
func (s *NoteService) GetNoteByID(ctx context.Context, id uint) (*NoteModel, error) {
	var note NoteModel
	err := s.db.WithContext(ctx).First(&note, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("note with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get note: %w", err)
	}
	return &note, nil
}

// GetByURN retrieves a note by its URN
func (s *NoteService) GetByURN(ctx context.Context, urn string) (*NoteModel, error) {
	var note NoteModel
	err := s.db.WithContext(ctx).Where("urn = ?", urn).First(&note).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("note with URN %s not found", urn)
		}
		return nil, fmt.Errorf("failed to get note by URN: %w", err)
	}
	return &note, nil
}

// DeleteNote deletes a note
func (s *NoteService) DeleteNote(ctx context.Context, id uint) error {
	note := &NoteModel{ID: id}
	if err := s.db.WithContext(ctx).First(note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("note with ID %d not found", id)
		}
		return fmt.Errorf("failed to find note: %w", err)
	}

	if err := s.db.WithContext(ctx).Delete(note).Error; err != nil {
		return fmt.Errorf("failed to delete note: %w", err)
	}

	return nil
}

// HistoryService handles all history-related operations
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new HistoryService instance
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// GetHistoryByBookmarkID retrieves all history entries for a specific bookmark
func (s *HistoryService) GetHistoryByBookmarkID(ctx context.Context, bookmarkID uint) ([]*BookmarkHistoryModel, error) {
	var history []*BookmarkHistoryModel
	err := s.db.WithContext(ctx).Where("bookmark_id = ?", bookmarkID).Order("timestamp DESC").Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get history by bookmark ID: %w", err)
	}
	return history, nil
}

// RevertBookmarkState reverts a bookmark to a previous state
func (s *HistoryService) RevertBookmarkState(ctx context.Context, bookmarkID uint, timestamp interface{}) error {
	var bookmark BookmarkModel
	if err := s.db.WithContext(ctx).First(&bookmark, bookmarkID).Error; err != nil {
		return fmt.Errorf("bookmark with ID %d not found: %w", bookmarkID, err)
	}

	var timestampTime time.Time
	switch t := timestamp.(type) {
	case time.Time:
		timestampTime = t
	case *time.Time:
		if t != nil {
			timestampTime = *t
		} else {
			// Get the most recent history entry before current state
			var history BookmarkHistoryModel
			err := s.db.WithContext(ctx).Where("bookmark_id = ?", bookmarkID).Order("timestamp DESC").First(&history).Error
			if err != nil {
				return fmt.Errorf("failed to get most recent history: %w", err)
			}
			timestampTime = history.Timestamp
		}
	case string:
		var err error
		timestampTime, err = time.Parse(time.RFC3339, t)
		if err != nil {
			return fmt.Errorf("invalid timestamp format: %w", err)
		}
	default:
		return fmt.Errorf("invalid timestamp type")
	}

	// Find the closest history entry before the given timestamp
	var historyEntry BookmarkHistoryModel
	err := s.db.WithContext(ctx).Where("bookmark_id = ? AND timestamp <= ?", bookmarkID, timestampTime).
		Order("timestamp DESC").First(&historyEntry).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no history entry found before the given timestamp")
		}
		return fmt.Errorf("failed to find history entry: %w", err)
	}

	// Update the bookmark to the state from the history entry
	bookmark.LearningState = historyEntry.OldValue
	if err := s.db.WithContext(ctx).Save(&bookmark).Error; err != nil {
		return fmt.Errorf("failed to revert bookmark state: %w", err)
	}

	// Create a history entry for the revert action
	revertHistory := &BookmarkHistoryModel{
		BookmarkID: bookmarkID,
		Action:     string(BookmarkActionStateReverted),
		OldValue:   historyEntry.NewValue,
		NewValue:   bookmark.LearningState,
		Timestamp:  time.Now(),
	}
	if err := s.db.WithContext(ctx).Create(revertHistory).Error; err != nil {
		return fmt.Errorf("failed to create revert history entry: %w", err)
	}

	return nil
}

// CrossReferenceService handles cross-reference operations between items
type CrossReferenceService struct {
	db *gorm.DB
}

// NewCrossReferenceService creates a new CrossReferenceService instance
func NewCrossReferenceService(db *gorm.DB) *CrossReferenceService {
	return &CrossReferenceService{db: db}
}

// CreateCrossReference creates a new cross-reference between two items
func (s *CrossReferenceService) CreateCrossReference(ctx context.Context, sourceItemID, targetItemID, sourceType, targetType, relationshipType string, strength float32, description *string) (*CrossReferenceModel, error) {
	crossRef := &CrossReferenceModel{
		SourceItemID:     sourceItemID,
		TargetItemID:     targetItemID,
		SourceType:       sourceType,
		TargetType:       targetType,
		RelationshipType: relationshipType,
		Strength:         strength,
		Description:      description,
		CreatedAt:        time.Now(),
	}

	if err := s.db.WithContext(ctx).Create(crossRef).Error; err != nil {
		return nil, fmt.Errorf("failed to create cross-reference: %w", err)
	}

	return crossRef, nil
}

// GetCrossReferencesBySource retrieves all cross-references from a specific source item
func (s *CrossReferenceService) GetCrossReferencesBySource(ctx context.Context, sourceItemID string) ([]*CrossReferenceModel, error) {
	var crossRefs []*CrossReferenceModel
	err := s.db.WithContext(ctx).Where("source_item_id = ?", sourceItemID).Find(&crossRefs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-references by source: %w", err)
	}
	return crossRefs, nil
}

// GetCrossReferencesByTarget retrieves all cross-references pointing to a specific target item
func (s *CrossReferenceService) GetCrossReferencesByTarget(ctx context.Context, targetItemID string) ([]*CrossReferenceModel, error) {
	var crossRefs []*CrossReferenceModel
	err := s.db.WithContext(ctx).Where("target_item_id = ?", targetItemID).Find(&crossRefs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-references by target: %w", err)
	}
	return crossRefs, nil
}

// GetCrossReferencesByType retrieves cross-references by relationship type
func (s *CrossReferenceService) GetCrossReferencesByType(ctx context.Context, relationshipType RelationshipType) ([]*CrossReferenceModel, error) {
	var crossRefs []*CrossReferenceModel
	err := s.db.WithContext(ctx).Where("relationship_type = ?", string(relationshipType)).Find(&crossRefs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cross-references by type: %w", err)
	}
	return crossRefs, nil
}

// GetBidirectionalCrossReferences gets all cross-references between two items in either direction
func (s *CrossReferenceService) GetBidirectionalCrossReferences(ctx context.Context, itemID1, itemID2 string) ([]*CrossReferenceModel, error) {
	var crossRefs []*CrossReferenceModel
	err := s.db.WithContext(ctx).
		Where("(source_item_id = ? AND target_item_id = ?) OR (source_item_id = ? AND target_item_id = ?)",
			itemID1, itemID2, itemID2, itemID1).
		Find(&crossRefs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get bidirectional cross-references: %w", err)
	}
	return crossRefs, nil
}

// MemoryCardService handles memory card operations for spaced repetition
type MemoryCardService struct {
	db *gorm.DB
}

// NewMemoryCardService creates a new MemoryCardService instance
func NewMemoryCardService(db *gorm.DB) *MemoryCardService {
	return &MemoryCardService{db: db}
}

// CreateMemoryCard creates a new memory card for a bookmark
func (s *MemoryCardService) CreateMemoryCard(ctx context.Context, bookmarkID uint, front, back string) (*MemoryCardModel, error) {
	card := &MemoryCardModel{
		BookmarkID: bookmarkID,
		Front:      front,
		Back:       back,
		Status:     string(StatusNew),
		Content:    "{}",
		EaseFactor: 2.5,
		Interval:   1,
		Repetition: 0,
	}

	if err := s.db.WithContext(ctx).Create(card).Error; err != nil {
		return nil, fmt.Errorf("failed to create memory card: %w", err)
	}

	return card, nil
}

// GetMemoryCardByID retrieves a memory card by its ID
func (s *MemoryCardService) GetMemoryCardByID(ctx context.Context, id uint) (*MemoryCardModel, error) {
	var card MemoryCardModel
	err := s.db.WithContext(ctx).First(&card, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("memory card with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get memory card: %w", err)
	}
	return &card, nil
}

// GetMemoryCardByURN retrieves a memory card by its URN
func (s *MemoryCardService) GetMemoryCardByURN(ctx context.Context, urn string) (*MemoryCardModel, error) {
	var card MemoryCardModel
	err := s.db.WithContext(ctx).Where("urn = ?", urn).First(&card).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("memory card with URN %s not found", urn)
		}
		return nil, fmt.Errorf("failed to get memory card by URN: %w", err)
	}
	return &card, nil
}

// GetMemoryCardsByBookmarkID retrieves all memory cards for a specific bookmark
func (s *MemoryCardService) GetMemoryCardsByBookmarkID(ctx context.Context, bookmarkID uint) ([]*MemoryCardModel, error) {
	var cards []*MemoryCardModel
	err := s.db.WithContext(ctx).Where("bookmark_id = ?", bookmarkID).Find(&cards).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get memory cards by bookmark ID: %w", err)
	}
	return cards, nil
}

// GetCardsForReview retrieves memory cards that are due for review
func (s *MemoryCardService) GetCardsForReview(ctx context.Context) ([]*MemoryCardModel, error) {
	now := time.Now()
	var cards []*MemoryCardModel
	err := s.db.WithContext(ctx).
		Joins("JOIN bookmark_models ON memory_card_models.bookmark_id = bookmark_models.id").
		Where("bookmark_models.learning_state = ? AND (memory_card_models.next_review <= ? OR memory_card_models.next_review IS NULL)",
			string(LearningStateLearning), now.Format("2006-01-02 15:04:05")).
		Find(&cards).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cards for review: %w", err)
	}
	return cards, nil
}

// UpdateCardAfterReview updates a memory card based on the user's rating after review
func (s *MemoryCardService) UpdateCardAfterReview(ctx context.Context, cardID uint, rating CardRating) error {
	card, err := s.GetMemoryCardByID(ctx, cardID)
	if err != nil {
		return err
	}

	// Get the bookmark to update mastery level
	var bookmark BookmarkModel
	if err := s.db.WithContext(ctx).First(&bookmark, card.BookmarkID).Error; err != nil {
		return fmt.Errorf("failed to get bookmark: %w", err)
	}

	// Implement spaced repetition algorithm (simplified SM-2 algorithm)
	switch rating {
	case CardRatingAgain:
		// Reset card to beginning
		card.Repetition = 0
		card.Interval = 1
		card.EaseFactor = max(1.3, card.EaseFactor-0.2)
	case CardRatingHard:
		// Slow down the interval increase
		if card.Repetition == 0 {
			card.Interval = 1
		} else if card.Repetition == 1 {
			card.Interval = 2 // Changed from 3 to 2 for "hard" rating
		} else {
			card.Interval = int(max(1, float32(card.Interval)*card.EaseFactor*0.8)) // Minimum interval of 1 day
		}
		card.Repetition += 1
	case CardRatingGood:
		// Normal interval increase
		if card.Repetition == 0 {
			card.Interval = 1
		} else if card.Repetition == 1 {
			card.Interval = 3
		} else {
			card.Interval = int(float32(card.Interval) * card.EaseFactor)
		}
		card.Repetition += 1
	case CardRatingEasy:
		// Increase ease factor and interval
		card.EaseFactor += 0.15
		if card.Repetition == 0 {
			card.Interval = 4
		} else if card.Repetition == 1 {
			card.Interval = 6
		} else {
			card.Interval = int(float32(card.Interval) * card.EaseFactor * 1.3)
		}
		card.Repetition += 1
	}

	// Calculate next review date
	nextReview := time.Now().AddDate(0, 0, card.Interval)
	card.NextReview = &nextReview

	// Determine logical next status from current state and repetition
	var nextStatus CardStatus
	currentStatus := CardStatus(card.Status)

	// For new cards, move to learning state
	if currentStatus == StatusNew {
		nextStatus = StatusLearning
	} else if currentStatus == StatusLearning {
		// Cards in learning transition to reviewed or mastered if enough repetitions
		if card.Repetition >= 5 {
			nextStatus = StatusMastered
		} else {
			nextStatus = StatusReviewed
		}
	} else if currentStatus == StatusReviewed {
		// Reviewed cards can go back to learning (more practice) or mastered (if enough repetitions)
		if card.Repetition >= 5 {
			nextStatus = StatusMastered
		} else {
			nextStatus = StatusLearning
		}
	} else {
		// Default fallback for other states
		if card.Repetition >= 5 {
			nextStatus = StatusMastered
		} else {
			nextStatus = StatusReviewed
		}
	}

	// Persist changes transactionally using optimistic version bump
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Refresh within transaction
		var cur MemoryCardModel
		if err := tx.First(&cur, card.ID).Error; err != nil {
			return err
		}

		// Validate transition
		if !CanTransition(CardStatus(cur.Status), nextStatus) {
			return ErrInvalidTransition
		}

		// Attempt update with version check
		res := tx.Model(&MemoryCardModel{}).
			Where("id = ? AND version = ?", cur.ID, cur.Version).
			Updates(map[string]any{
				"ease_factor": card.EaseFactor,
				"interval":    card.Interval,
				"repetition":  card.Repetition,
				"next_review": card.NextReview,
				"status":      string(nextStatus),
				"version":     cur.Version + 1,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return fmt.Errorf("concurrent update detected")
		}

		// Update bookmark mastery (no error return value)
		updateBookmarkMastery(ctx, tx, &bookmark, rating)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// transitionCardStatusTx performs a versioned status transition within an existing transaction.
func (s *MemoryCardService) transitionCardStatusTx(tx *gorm.DB, cardID uint, next CardStatus) error {
	var cur MemoryCardModel
	if err := tx.First(&cur, cardID).Error; err != nil {
		return err
	}
	if !CanTransition(CardStatus(cur.Status), next) {
		return ErrInvalidTransition
	}
	res := tx.Model(&MemoryCardModel{}).
		Where("id = ? AND version = ?", cur.ID, cur.Version).
		Updates(map[string]any{"status": string(next), "version": cur.Version + 1})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrConcurrentUpdate
	}
	return nil
}

// TransitionCardStatus transitions a card to the given status using optimistic concurrency.
// If expectedVersion is non-nil, the transition will only succeed if the current
// version matches expectedVersion; otherwise it uses the current version observed
// in the database for the versioned update.
func (s *MemoryCardService) TransitionCardStatus(ctx context.Context, cardID uint, expectedVersion *int, next CardStatus) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cur MemoryCardModel
		if err := tx.First(&cur, cardID).Error; err != nil {
			return err
		}
		if expectedVersion != nil && cur.Version != *expectedVersion {
			return ErrConcurrentUpdate
		}
		if !CanTransition(CardStatus(cur.Status), next) {
			return ErrInvalidTransition
		}
		// versioned update
		res := tx.Model(&MemoryCardModel{}).
			Where("id = ? AND version = ?", cur.ID, cur.Version).
			Updates(map[string]any{"status": string(next), "version": cur.Version + 1})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return ErrConcurrentUpdate
		}
		return nil
	})
}

// GetCardsByLearningState retrieves memory cards for bookmarks in a specific learning state
func (s *MemoryCardService) GetCardsByLearningState(ctx context.Context, state LearningState) ([]*MemoryCardModel, error) {
	var cards []*MemoryCardModel
	err := s.db.WithContext(ctx).
		Joins("JOIN bookmark_models ON memory_card_models.bookmark_id = bookmark_models.id").
		Where("bookmark_models.learning_state = ?", string(state)).
		Find(&cards).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get cards by learning state: %w", err)
	}
	return cards, nil
}

// ListMemoryCards retrieves memory cards with optional filters and pagination
func (s *MemoryCardService) ListMemoryCards(ctx context.Context, bookmarkID *uint, learningState *string, author *string, isPrivate *bool, offset, limit int) ([]*MemoryCardModel, int64, error) {
	var cards []*MemoryCardModel

	query := s.db.WithContext(ctx).Table("memory_card_models")

	// Apply filters if provided
	if bookmarkID != nil {
		query = query.Where("bookmark_id = ?", *bookmarkID)
	}
	if learningState != nil && *learningState != "" {
		query = query.Joins("JOIN bookmark_models ON memory_card_models.bookmark_id = bookmark_models.id").
			Where("bookmark_models.learning_state = ?", *learningState)
	}
	if author != nil && *author != "" {
		// Assuming author is stored in a field, adjust as needed
		query = query.Where("author = ?", *author)
	}
	if isPrivate != nil {
		query = query.Where("is_private = ?", *isPrivate)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count memory cards: %w", err)
	}

	// Apply pagination
	if limit > 0 {
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&cards).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list memory cards: %w", err)
	}

	return cards, total, nil
}

// Helper function to update bookmark mastery level based on card review
func updateBookmarkMastery(ctx context.Context, db *gorm.DB, bookmark *BookmarkModel, rating CardRating) {
	// Get all cards for this bookmark to calculate average mastery
	var allCards []*MemoryCardModel
	db.WithContext(ctx).Where("bookmark_id = ?", bookmark.ID).Find(&allCards)

	if len(allCards) == 0 {
		return
	}

	// Calculate average ease factor
	totalEase := float32(0)
	for _, card := range allCards {
		totalEase += card.EaseFactor
	}
	avgEase := totalEase / float32(len(allCards))

	// Map ease factor to mastery level (0.0 to 1.0)
	// Ease factor of 1.3 (hardest) = 0.0 mastery, 3.0+ (easiest) = 1.0 mastery
	mastery := min(1.0, max(0.0, (avgEase-1.3)/(3.0-1.3)))
	bookmark.MasteryLevel = mastery

	// Update learning state based on mastery
	if mastery >= 0.9 {
		bookmark.LearningState = string(LearningStateMastered)
	} else if mastery >= 0.7 {
		bookmark.LearningState = string(LearningStateLearning)
	}

	// Save updated bookmark
	db.WithContext(ctx).Save(bookmark)
}

// Helper functions
func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}
