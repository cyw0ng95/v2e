package notes

import (
	"context"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Run migrations
	if err := MigrateNotesTables(db); err != nil {
		return nil, err
	}

	return db, nil
}

func TestBookmarkService(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	bookmarkService := NewBookmarkService(db)

	ctx := context.Background()

	t.Run("CreateBookmark", func(t *testing.T) {
		bookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-item-123", "CVE", "CVE-2021-1234", "Test CVE", "A test CVE for bookmarking")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		if bookmark.GlobalItemID != "global-item-123" {
			t.Errorf("Expected GlobalItemID to be 'global-item-123', got '%s'", bookmark.GlobalItemID)
		}

		if bookmark.ItemType != "CVE" {
			t.Errorf("Expected ItemType to be 'CVE', got '%s'", bookmark.ItemType)
		}

		if bookmark.LearningState != string(LearningStateToReview) {
			t.Errorf("Expected LearningState to be '%s', got '%s'", string(LearningStateToReview), bookmark.LearningState)
		}
	})

	t.Run("GetBookmarkByID", func(t *testing.T) {
		// First create a bookmark
		createdBookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-item-456", "CWE", "CWE-123", "Test CWE", "A test CWE for bookmarking")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Then get it by ID
		retrievedBookmark, err := bookmarkService.GetBookmarkByID(ctx, createdBookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get bookmark: %v", err)
		}

		if retrievedBookmark.ID != createdBookmark.ID {
			t.Errorf("Expected ID to be %d, got %d", createdBookmark.ID, retrievedBookmark.ID)
		}

		if retrievedBookmark.Title != "Test CWE" {
			t.Errorf("Expected Title to be 'Test CWE', got '%s'", retrievedBookmark.Title)
		}
	})

	t.Run("GetBookmarksByGlobalItemID", func(t *testing.T) {
		globalItemID := "global-item-789"

		// Create multiple bookmarks with the same global item ID
		_, _, err := bookmarkService.CreateBookmark(ctx, globalItemID, "CAPEC", "CAPEC-123", "Test CAPEC 1", "A test CAPEC for bookmarking")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		_, _, err = bookmarkService.CreateBookmark(ctx, globalItemID, "CAPEC", "CAPEC-124", "Test CAPEC 2", "Another test CAPEC for bookmarking")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Get all bookmarks for the global item ID
		bookmarks, err := bookmarkService.GetBookmarksByGlobalItemID(ctx, globalItemID)
		if err != nil {
			t.Fatalf("Failed to get bookmarks by global item ID: %v", err)
		}

		if len(bookmarks) != 2 {
			t.Errorf("Expected 2 bookmarks, got %d", len(bookmarks))
		}
	})

	t.Run("UpdateBookmark", func(t *testing.T) {
		// Create a bookmark first
		bookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-item-update", "ATT&CK", "T1001", "Test ATT&CK Technique", "A test ATT&CK technique for bookmarking")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Update the bookmark
		bookmark.Title = "Updated ATT&CK Technique"
		bookmark.LearningState = string(LearningStateLearning)
		err = bookmarkService.UpdateBookmark(ctx, bookmark)
		if err != nil {
			t.Fatalf("Failed to update bookmark: %v", err)
		}

		// Verify the update
		updatedBookmark, err := bookmarkService.GetBookmarkByID(ctx, bookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get updated bookmark: %v", err)
		}

		if updatedBookmark.Title != "Updated ATT&CK Technique" {
			t.Errorf("Expected updated title to be 'Updated ATT&CK Technique', got '%s'", updatedBookmark.Title)
		}

		if updatedBookmark.LearningState != string(LearningStateLearning) {
			t.Errorf("Expected updated learning state to be '%s', got '%s'", string(LearningStateLearning), updatedBookmark.LearningState)
		}
	})

	t.Run("UpdateLearningState", func(t *testing.T) {
		// Create a bookmark first
		bookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-item-state", "CVE", "CVE-2022-1234", "Test CVE State", "A test CVE for state updates")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Update the learning state
		err = bookmarkService.UpdateLearningState(ctx, bookmark.ID, LearningStateMastered)
		if err != nil {
			t.Fatalf("Failed to update learning state: %v", err)
		}

		// Verify the update
		updatedBookmark, err := bookmarkService.GetBookmarkByID(ctx, bookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get updated bookmark: %v", err)
		}

		if updatedBookmark.LearningState != string(LearningStateMastered) {
			t.Errorf("Expected learning state to be '%s', got '%s'", string(LearningStateMastered), updatedBookmark.LearningState)
		}
	})

	t.Run("GetBookmarksByLearningState", func(t *testing.T) {
		// Create a bookmark with a specific learning state
		_, _, err := bookmarkService.CreateBookmark(ctx, "global-item-filter", "CWE", "CWE-456", "Test Filter", "A test item for filtering")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Create another bookmark and update its state
		bookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-item-filter2", "CAPEC", "CAPEC-456", "Test Filter 2", "Another test item for filtering")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}
		err = bookmarkService.UpdateLearningState(ctx, bookmark.ID, LearningStateLearning)
		if err != nil {
			t.Fatalf("Failed to update learning state: %v", err)
		}

		// Get bookmarks by learning state
		learningBookmarks, err := bookmarkService.GetBookmarksByLearningState(ctx, LearningStateLearning)
		if err != nil {
			t.Fatalf("Failed to get bookmarks by learning state: %v", err)
		}

		// Check if at least one bookmark has the learning state (could be more from other tests)
		if len(learningBookmarks) < 1 {
			t.Errorf("Expected at least 1 bookmark with learning state '%s', got %d", string(LearningStateLearning), len(learningBookmarks))
		}
	})
}

func TestNoteService(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	bookmarkService := NewBookmarkService(db)
	noteService := NewNoteService(db)

	ctx := context.Background()

	// Create a bookmark first to attach notes to
	bookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-note-test", "CVE", "CVE-2023-1234", "Note Test CVE", "A test CVE for note testing")
	if err != nil {
		t.Fatalf("Failed to create bookmark: %v", err)
	}

	t.Run("AddNote", func(t *testing.T) {
		note, err := noteService.AddNote(ctx, bookmark.ID, "This is a test note", nil, false)
		if err != nil {
			t.Fatalf("Failed to add note: %v", err)
		}

		if note.Content != "This is a test note" {
			t.Errorf("Expected note content to be 'This is a test note', got '%s'", note.Content)
		}

		if note.BookmarkID != bookmark.ID {
			t.Errorf("Expected BookmarkID to be %d, got %d", bookmark.ID, note.BookmarkID)
		}
	})

	t.Run("GetNotesByBookmarkID", func(t *testing.T) {
		// Add a couple of notes
		_, err := noteService.AddNote(ctx, bookmark.ID, "First note", nil, false)
		if err != nil {
			t.Fatalf("Failed to add note: %v", err)
		}

		_, err = noteService.AddNote(ctx, bookmark.ID, "Second note", nil, true)
		if err != nil {
			t.Fatalf("Failed to add note: %v", err)
		}

		// Get all notes for the bookmark
		notes, err := noteService.GetNotesByBookmarkID(ctx, bookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get notes by bookmark ID: %v", err)
		}

		if len(notes) != 3 { // Including the note from the previous test
			t.Errorf("Expected 3 notes, got %d", len(notes))
		}
	})

	t.Run("UpdateNote", func(t *testing.T) {
		// Add a note first
		note, err := noteService.AddNote(ctx, bookmark.ID, "Original note content", nil, false)
		if err != nil {
			t.Fatalf("Failed to add note: %v", err)
		}

		// Update the note
		note.Content = "Updated note content"
		err = noteService.UpdateNote(ctx, note)
		if err != nil {
			t.Fatalf("Failed to update note: %v", err)
		}

		// Since we don't have a GetNoteByID method, we'll verify by getting all notes for the bookmark
		notes, err := noteService.GetNotesByBookmarkID(ctx, bookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get notes by bookmark ID: %v", err)
		}

		// Find our updated note
		found := false
		for _, n := range notes {
			if n.ID == note.ID && n.Content == "Updated note content" {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("Could not find updated note with content 'Updated note content'")
		}
	})
}

func TestCrossReferenceService(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	crossRefService := NewCrossReferenceService(db)

	ctx := context.Background()

	t.Run("CreateCrossReference", func(t *testing.T) {
		description := "Test relationship"
		crossRef, err := crossRefService.CreateCrossReference(ctx, "global-source-123", "global-target-456", "CVE", "CWE", string(RelationshipTypeExploits), 0.8, &description)
		if err != nil {
			t.Fatalf("Failed to create cross-reference: %v", err)
		}

		if crossRef.SourceItemID != "global-source-123" {
			t.Errorf("Expected SourceItemID to be 'global-source-123', got '%s'", crossRef.SourceItemID)
		}

		if crossRef.TargetItemID != "global-target-456" {
			t.Errorf("Expected TargetItemID to be 'global-target-456', got '%s'", crossRef.TargetItemID)
		}

		if crossRef.RelationshipType != string(RelationshipTypeExploits) {
			t.Errorf("Expected RelationshipType to be '%s', got '%s'", string(RelationshipTypeExploits), crossRef.RelationshipType)
		}

		if crossRef.Strength != 0.8 {
			t.Errorf("Expected Strength to be 0.8, got %f", crossRef.Strength)
		}

		if *crossRef.Description != "Test relationship" {
			t.Errorf("Expected Description to be 'Test relationship', got '%s'", *crossRef.Description)
		}
	})

	t.Run("GetCrossReferencesBySource", func(t *testing.T) {
		sourceID := "global-source-get-test"

		// Create multiple cross-references from the same source
		_, err := crossRefService.CreateCrossReference(ctx, sourceID, "global-target-1", "CVE", "CWE", string(RelationshipTypeExploits), 0.7, nil)
		if err != nil {
			t.Fatalf("Failed to create cross-reference: %v", err)
		}

		_, err = crossRefService.CreateCrossReference(ctx, sourceID, "global-target-2", "CVE", "CAPEC", string(RelationshipTypeRelatedTo), 0.9, nil)
		if err != nil {
			t.Fatalf("Failed to create cross-reference: %v", err)
		}

		// Get all cross-references from the source
		crossRefs, err := crossRefService.GetCrossReferencesBySource(ctx, sourceID)
		if err != nil {
			t.Fatalf("Failed to get cross-references by source: %v", err)
		}

		if len(crossRefs) != 2 {
			t.Errorf("Expected 2 cross-references, got %d", len(crossRefs))
		}
	})

	t.Run("GetCrossReferencesByType", func(t *testing.T) {
		// Create a cross-reference with a specific type
		_, err := crossRefService.CreateCrossReference(ctx, "global-src-type", "global-target-type", "CWE", "CVE", string(RelationshipTypeMitigates), 0.6, nil)
		if err != nil {
			t.Fatalf("Failed to create cross-reference: %v", err)
		}

		// Get cross-references by type
		crossRefs, err := crossRefService.GetCrossReferencesByType(ctx, RelationshipTypeMitigates)
		if err != nil {
			t.Fatalf("Failed to get cross-references by type: %v", err)
		}

		if len(crossRefs) != 1 {
			t.Errorf("Expected 1 cross-reference with type '%s', got %d", string(RelationshipTypeMitigates), len(crossRefs))
		}
	})
}

func TestMemoryCardService(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	bookmarkService := NewBookmarkService(db)
	memoryCardService := NewMemoryCardService(db)

	ctx := context.Background()

	// Create a bookmark first to attach memory cards to
	bookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-card-test", "CVE", "CVE-2024-1234", "Card Test CVE", "A test CVE for card testing")
	if err != nil {
		t.Fatalf("Failed to create bookmark: %v", err)
	}

	t.Run("CreateMemoryCard", func(t *testing.T) {
		card, err := memoryCardService.CreateMemoryCard(ctx, bookmark.ID, "What is CVE-2024-1234?", "A test vulnerability in a software component")
		if err != nil {
			t.Fatalf("Failed to create memory card: %v", err)
		}

		if card.BookmarkID != bookmark.ID {
			t.Errorf("Expected BookmarkID to be %d, got %d", bookmark.ID, card.BookmarkID)
		}

		if card.Front != "What is CVE-2024-1234?" {
			t.Errorf("Expected Front to be 'What is CVE-2024-1234?', got '%s'", card.Front)
		}

		if card.Back != "A test vulnerability in a software component" {
			t.Errorf("Expected Back to be 'A test vulnerability in a software component', got '%s'", card.Back)
		}

		if card.EaseFactor != 2.5 {
			t.Errorf("Expected EaseFactor to be 2.5, got %f", card.EaseFactor)
		}
	})

	t.Run("GetMemoryCardsByBookmarkID", func(t *testing.T) {
		// Create multiple memory cards for the same bookmark
		_, err := memoryCardService.CreateMemoryCard(ctx, bookmark.ID, "Front 1", "Back 1")
		if err != nil {
			t.Fatalf("Failed to create memory card: %v", err)
		}

		_, err = memoryCardService.CreateMemoryCard(ctx, bookmark.ID, "Front 2", "Back 2")
		if err != nil {
			t.Fatalf("Failed to create memory card: %v", err)
		}

		// Get all memory cards for the bookmark
		cards, err := memoryCardService.GetMemoryCardsByBookmarkID(ctx, bookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get memory cards by bookmark ID: %v", err)
		}

		// We expect 3 cards total (1 from the previous test + 2 from this test)
		if len(cards) != 3 {
			t.Errorf("Expected 3 memory cards, got %d", len(cards))
		}
	})

	t.Run("UpdateCardAfterReview", func(t *testing.T) {
		// Create a memory card
		card, err := memoryCardService.CreateMemoryCard(ctx, bookmark.ID, "Review Test Front", "Review Test Back")
		if err != nil {
			t.Fatalf("Failed to create memory card: %v", err)
		}

		// Update the card after a review with "Good" rating
		err = memoryCardService.UpdateCardAfterReview(ctx, card.ID, CardRatingGood)
		if err != nil {
			t.Fatalf("Failed to update card after review: %v", err)
		}

		// Get the updated card
		updatedCard, err := memoryCardService.GetMemoryCardByID(ctx, card.ID)
		if err != nil {
			t.Fatalf("Failed to get updated card: %v", err)
		}

		// Verify the card was updated appropriately
		if updatedCard.Repetition != 1 {
			t.Errorf("Expected Repetition to be 1 after 'Good' rating, got %d", updatedCard.Repetition)
		}

		// For the first "Good" rating, the interval should be 1 based on our algorithm
		// since Repetition was 0 initially and this is the first good rating
		if updatedCard.Interval != 1 { // Based on the algorithm for first good review
			t.Errorf("Expected Interval to be 1 after first 'Good' rating, got %d", updatedCard.Interval)
		}
	})
}

func TestHistoryService(t *testing.T) {
	db, err := setupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
	}

	bookmarkService := NewBookmarkService(db)
	historyService := NewHistoryService(db)

	ctx := context.Background()

	// Create a bookmark first
	bookmark, _, err := bookmarkService.CreateBookmark(ctx, "global-history-test", "CWE", "CWE-789", "History Test CWE", "A test CWE for history testing")
	if err != nil {
		t.Fatalf("Failed to create bookmark: %v", err)
	}

	t.Run("GetHistoryByBookmarkID", func(t *testing.T) {
		// The bookmark creation should have created a history entry
		history, err := historyService.GetHistoryByBookmarkID(ctx, bookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get history by bookmark ID: %v", err)
		}

		if len(history) == 0 {
			t.Fatalf("Expected at least 1 history entry for the newly created bookmark, got 0")
		}

		// Check that the first entry is for creation
		if history[0].Action != string(BookmarkActionCreated) {
			t.Errorf("Expected first history action to be '%s', got '%s'", string(BookmarkActionCreated), history[0].Action)
		}
	})

	t.Run("UpdateBookmarkCreatesHistory", func(t *testing.T) {
		// Update the learning state which should create a history entry
		err := bookmarkService.UpdateLearningState(ctx, bookmark.ID, LearningStateLearning)
		if err != nil {
			t.Fatalf("Failed to update learning state: %v", err)
		}

		// Get history again
		history, err := historyService.GetHistoryByBookmarkID(ctx, bookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get history by bookmark ID: %v", err)
		}

		// We should now have at least 2 history entries
		if len(history) < 2 {
			t.Errorf("Expected at least 2 history entries, got %d", len(history))
		}

		// The latest entry should be for learning state change
		latestEntry := history[0] // First in the list because we ordered by timestamp DESC
		if latestEntry.Action != string(BookmarkActionLearningStateChanged) {
			t.Errorf("Expected latest history action to be '%s', got '%s'", string(BookmarkActionLearningStateChanged), latestEntry.Action)
		}
	})
}
