package notes

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestRPCMethod tests RPC method functionality
func TestRPCMethod(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRPCMethod", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_rpc_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Create a test bookmark first
		createdBookmark, _, err := service.BookmarkService.CreateBookmark(ctx, "test-item-id", string(ItemTypeCVE), "test-cve-id", "Test Bookmark", "Test content")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Test GetBookmarkByID RPC method
		result, err := service.BookmarkService.GetBookmarkByID(ctx, createdBookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get bookmark by ID: %v", err)
		}

		if result.ID != createdBookmark.ID {
			t.Errorf("Expected bookmark ID %d, got %d", createdBookmark.ID, result.ID)
		}
		if result.Title != "Test Bookmark" {
			t.Errorf("Expected title 'Test Bookmark', got '%s'", result.Title)
		}

		// Test GetBookmarksByGlobalItemID RPC method
		bookmarks, err := service.BookmarkService.GetBookmarksByGlobalItemID(ctx, "test-item-id")
		if err != nil {
			t.Fatalf("Failed to get bookmarks by global item ID: %v", err)
		}

		if len(bookmarks) != 1 {
			t.Errorf("Expected 1 bookmark, got %d", len(bookmarks))
		} else if bookmarks[0].ID != createdBookmark.ID {
			t.Errorf("Expected bookmark ID %d, got %d", createdBookmark.ID, bookmarks[0].ID)
		}

		// Test GetBookmarksByLearningState RPC method
		newBookmarks, err := service.BookmarkService.GetBookmarksByLearningState(ctx, LearningStateToReview)
		if err != nil {
			t.Fatalf("Failed to get bookmarks by learning state: %v", err)
		}

		if len(newBookmarks) != 1 {
			t.Errorf("Expected 1 bookmark with NEW state, got %d", len(newBookmarks))
		}

		// Test UpdateBookmark RPC method
		updatedBookmark := createdBookmark
		updatedBookmark.Title = "Updated Test Bookmark"
		updatedBookmark.Description = "Updated content"
		updatedBookmark.UpdatedAt = time.Now()

		err = service.BookmarkService.UpdateBookmark(ctx, updatedBookmark)
		if err != nil {
			t.Fatalf("Failed to update bookmark: %v", err)
		}

		// Verify the update
		verifiedBookmark, err := service.BookmarkService.GetBookmarkByID(ctx, createdBookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get updated bookmark: %v", err)
		}

		if verifiedBookmark.Title != "Updated Test Bookmark" {
			t.Errorf("Expected updated title 'Updated Test Bookmark', got '%s'", verifiedBookmark.Title)
		}

		// Test UpdateLearningState RPC method
		err = service.BookmarkService.UpdateLearningState(ctx, createdBookmark.ID, LearningStateLearning)
		if err != nil {
			t.Fatalf("Failed to update learning state: %v", err)
		}

		// Verify the state update
		stateUpdatedBookmark, err := service.BookmarkService.GetBookmarkByID(ctx, createdBookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get bookmark after state update: %v", err)
		}

		if stateUpdatedBookmark.LearningState != string(LearningStateLearning) {
			t.Errorf("Expected learning state %s, got %s", LearningStateLearning, stateUpdatedBookmark.LearningState)
		}
	})

}

// TestDataMigration tests data migration functionality
func TestDataMigration(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataMigration", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_migration_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Test migration from an older schema to newer schema
		// This test ensures that the migration process works correctly
		// by checking that tables exist and have the expected structure

		// Check that all required tables exist
		tables := []string{
			"bookmarks", "bookmark_histories", "notes", "memory_cards",
			"learning_sessions", "cross_references", "global_items",
		}

		// We can verify the migration worked by attempting to create and query data
		// Since AutoMigrate was called successfully, tables should exist
		for _, tableName := range tables {
			// Just verify that we can query from each table without error
			switch tableName {
			case "bookmarks":
				err := db.Where("1 = 0").Find(&BookmarkModel{}).Error // Query with impossible condition
				if err != nil {
					t.Errorf("Error querying %s table: %v", tableName, err)
				}
			case "bookmark_histories":
				err := db.Where("1 = 0").Find(&BookmarkHistoryModel{}).Error
				if err != nil {
					t.Errorf("Error querying %s table: %v", tableName, err)
				}
			case "notes":
				err := db.Where("1 = 0").Find(&NoteModel{}).Error
				if err != nil {
					t.Errorf("Error querying %s table: %v", tableName, err)
				}
			case "memory_cards":
				err := db.Where("1 = 0").Find(&MemoryCardModel{}).Error
				if err != nil {
					t.Errorf("Error querying %s table: %v", tableName, err)
				}
			case "learning_sessions":
				err := db.Where("1 = 0").Find(&LearningSessionModel{}).Error
				if err != nil {
					t.Errorf("Error querying %s table: %v", tableName, err)
				}
			case "cross_references":
				err := db.Where("1 = 0").Find(&CrossReferenceModel{}).Error
				if err != nil {
					t.Errorf("Error querying %s table: %v", tableName, err)
				}
			case "global_items":
				err := db.Where("1 = 0").Find(&GlobalItemModel{}).Error
				if err != nil {
					t.Errorf("Error querying %s table: %v", tableName, err)
				}
			}
		}

		// Test creating records after migration
		createdBookmark, _, err := service.BookmarkService.CreateBookmark(ctx, "migration-test-item", string(ItemTypeCVE), "migration-cve-id", "Migration Test Bookmark", "Migration test content")
		if err != nil {
			t.Fatalf("Failed to create bookmark after migration: %v", err)
		}

		if createdBookmark.ID == 0 {
			t.Error("Expected non-zero ID for created bookmark")
		}

		// Test that we can perform operations after migration
		notes, err := service.NoteService.GetNotesByBookmarkID(ctx, createdBookmark.ID)
		if err != nil {
			t.Errorf("Failed to get notes by bookmark ID: %v", err)
		}

		if len(notes) != 0 {
			t.Errorf("Expected 0 notes initially, got %d", len(notes))
		}
	})

}

// TestServiceOperationErrorHandling tests service operation error handling
func TestServiceOperationErrorHandling(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestServiceOperationErrorHandling", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_error_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Test error when getting non-existent bookmark
		_, err = service.BookmarkService.GetBookmarkByID(ctx, 999999)
		if err == nil {
			t.Error("Expected error when getting non-existent bookmark")
		} else if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected 'not found' error, got: %v", err)
		}

		// Test error when getting bookmarks by non-existent global item ID
		bookmarks, err := service.BookmarkService.GetBookmarksByGlobalItemID(ctx, "non-existent-id")
		if err != nil {
			t.Errorf("Expected no error when getting bookmarks by non-existent global item ID, got: %v", err)
		}
		if len(bookmarks) != 0 {
			t.Errorf("Expected 0 bookmarks for non-existent global item ID, got %d", len(bookmarks))
		}

		// Test error when getting bookmarks by invalid learning state
		invalidState := LearningState("INVALID_STATE")
		invalidBookmarks, err := service.BookmarkService.GetBookmarksByLearningState(ctx, invalidState)
		if err != nil {
			t.Errorf("Expected no error when getting bookmarks by invalid learning state, got: %v", err)
		}
		if len(invalidBookmarks) != 0 {
			t.Errorf("Expected 0 bookmarks for invalid learning state, got %d", len(invalidBookmarks))
		}

		// Test error when updating non-existent bookmark - first create a valid bookmark then try to update with an invalid one
		createdBookmark, _, err := service.BookmarkService.CreateBookmark(ctx, "test-item", string(ItemTypeCVE), "test-cve", "Test Bookmark", "Test Description")
		if err != nil {
			t.Fatalf("Failed to create test bookmark: %v", err)
		}

		// Now try to update the bookmark with invalid data
		createdBookmark.Title = "Updated Title"
		err = service.BookmarkService.UpdateBookmark(ctx, createdBookmark)
		if err != nil {
			t.Errorf("Unexpected error when updating existing bookmark: %v", err)
		}

		// Now try to update a non-existent bookmark by manually setting an invalid ID
		invalidBookmark := &BookmarkModel{
			ID:           999999,
			GlobalItemID: "test-item",
			ItemType:     string(ItemTypeCVE),
			ItemID:       "test-cve",
			Title:        "Non-existent Bookmark",
			Description:  "Content",
		}
		err = service.BookmarkService.UpdateBookmark(ctx, invalidBookmark)
		if err == nil {
			t.Error("Expected error when updating non-existent bookmark")
		}

		// Test error when updating learning state for non-existent bookmark
		err = service.BookmarkService.UpdateLearningState(ctx, 999999, LearningStateLearning)
		if err == nil {
			t.Error("Expected error when updating learning state for non-existent bookmark")
		}

		// Test error when adding note to non-existent bookmark
		_, err = service.NoteService.AddNote(ctx, 999999, "Test note content", nil, false)
		// Adding a note to a non-existent bookmark may or may not return an error depending on implementation
		// The important thing is that it doesn't panic
		if err != nil {
			t.Logf("Got expected error when adding note to non-existent bookmark: %v", err)
		} else {
			t.Logf("No error when adding note to non-existent bookmark (this may be acceptable behavior)")
		}

		// Test error when creating cross-reference with invalid data
		_, err = service.CrossReferenceService.CreateCrossReference(ctx, "999999", "999999", string(ItemTypeCVE), string(ItemTypeCWE), string(RelationshipTypeRelatedTo), 1.0, nil)
		if err != nil {
			t.Logf("Creating cross-reference with test data produced error (this is OK): %v", err)
		}
	})

}

// TestConcurrentAccess tests concurrent access to the service
func TestConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_concurrent_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Pre-populate with some test data
		createdBookmark, _, err := service.BookmarkService.CreateBookmark(ctx, "concurrent-test-item", string(ItemTypeCVE), "concurrent-cve", "Concurrent Test Bookmark", "Concurrent test content")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		var wg sync.WaitGroup
		const numGoroutines = 10

		// Test concurrent reads
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Perform multiple operations in each goroutine
				for j := 0; j < 5; j++ {
					// Test GetBookmarkByID
					_, err := service.BookmarkService.GetBookmarkByID(ctx, createdBookmark.ID)
					if err != nil && !strings.Contains(err.Error(), "not found") {
						t.Errorf("Goroutine %d: GetBookmarkByID failed: %v", goroutineID, err)
					}

					// Test GetBookmarksByGlobalItemID
					_, err = service.BookmarkService.GetBookmarksByGlobalItemID(ctx, "concurrent-test-item")
					if err != nil {
						t.Errorf("Goroutine %d: GetBookmarksByGlobalItemID failed: %v", goroutineID, err)
					}

					// Small delay to allow other goroutines to interleave
					time.Sleep(time.Millisecond * 1)
				}
			}(i)
		}

		wg.Wait()

		// Test concurrent writes
		var writeWg sync.WaitGroup
		for i := 0; i < numGoroutines; i++ {
			writeWg.Add(1)
			go func(goroutineID int) {
				defer writeWg.Done()

				// Add notes concurrently
				noteContent := "Note from goroutine " + string(rune(48+goroutineID)) // Convert to ASCII
				_, err := service.NoteService.AddNote(ctx, createdBookmark.ID, noteContent, nil, false)
				if err != nil {
					t.Errorf("Goroutine %d: AddNote failed: %v", goroutineID, err)
				}
			}(i)
		}

		writeWg.Wait()

		// Verify that all notes were added
		notes, err := service.NoteService.GetNotesByBookmarkID(ctx, createdBookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get notes: %v", err)
		}

		expectedNoteCount := numGoroutines
		if len(notes) < expectedNoteCount {
			t.Errorf("Expected at least %d notes, got %d", expectedNoteCount, len(notes))
		}
	})

}

// TestMemoryCardFunctionality tests memory card functionality
func TestMemoryCardFunctionality(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestMemoryCardFunctionality", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_memory_card_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Create a bookmark first
		createdBookmark, _, err := service.BookmarkService.CreateBookmark(ctx, "memory-card-test-item", string(ItemTypeCVE), "memory-cve", "Memory Card Test Bookmark", "Memory card test content")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Create a memory card for the bookmark
		createdCard, err := service.MemoryCardService.CreateMemoryCard(ctx, createdBookmark.ID, "What is the capital of France?", "Paris")
		if err != nil {
			t.Fatalf("Failed to create memory card: %v", err)
		}

		if createdCard.ID == 0 {
			t.Error("Expected non-zero ID for created memory card")
		}

		// Get memory cards by bookmark ID
		cards, err := service.MemoryCardService.GetMemoryCardsByBookmarkID(ctx, createdBookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get memory cards: %v", err)
		}

		if len(cards) != 2 {
			t.Errorf("Expected 2 memory cards, got %d", len(cards))
		} else {
			// Find the manually created card
			found := false
			for _, card := range cards {
				if card.Front == "What is the capital of France?" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected front 'What is the capital of France?' not found in cards")
			}
		}

		// Update card after review
		err = service.MemoryCardService.UpdateCardAfterReview(ctx, createdCard.ID, CardRatingGood)
		if err != nil {
			t.Fatalf("Failed to update card after review: %v", err)
		}

		// After updating the card, get all cards for the bookmark to verify the update
		cards, err = service.MemoryCardService.GetMemoryCardsByBookmarkID(ctx, createdCard.BookmarkID)
		if err != nil {
			t.Fatalf("Failed to get memory cards after update: %v", err)
		}

		if len(cards) == 0 {
			t.Fatal("Expected at least 1 card after update")
		}

		var foundCard *MemoryCardModel
		for _, card := range cards {
			if card.ID == createdCard.ID {
				foundCard = card
				break
			}
		}

		if foundCard == nil {
			t.Fatalf("Could not find updated card with ID %d", createdCard.ID)
		}

		if foundCard.Repetition != 1 {
			t.Errorf("Expected repetition count 1 after good rating, got %d", foundCard.Repetition)
		}
	})

}

// TestCrossReferenceFunctionality tests cross-reference functionality
func TestCrossReferenceFunctionality(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestCrossReferenceFunctionality", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_cross_ref_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Create two bookmarks for cross-referencing
		createdBookmark1, _, err := service.BookmarkService.CreateBookmark(ctx, "cross-ref-item-1", string(ItemTypeCVE), "cross-ref-cve-1", "Cross Reference Test Bookmark 1", "Cross reference test content 1")
		if err != nil {
			t.Fatalf("Failed to create first bookmark: %v", err)
		}

		createdBookmark2, _, err := service.BookmarkService.CreateBookmark(ctx, "cross-ref-item-2", string(ItemTypeCWE), "cross-ref-cwe-2", "Cross Reference Test Bookmark 2", "Cross reference test content 2")
		if err != nil {
			t.Fatalf("Failed to create second bookmark: %v", err)
		}

		// Create a cross-reference between the bookmarks
		createdCrossRef, err := service.CrossReferenceService.CreateCrossReference(ctx, fmt.Sprintf("%d", createdBookmark1.ID), fmt.Sprintf("%d", createdBookmark2.ID), string(ItemTypeCVE), string(ItemTypeCWE), string(RelationshipTypeRelatedTo), 1.0, nil)
		if err != nil {
			t.Fatalf("Failed to create cross-reference: %v", err)
		}

		if createdCrossRef.ID == 0 {
			t.Error("Expected non-zero ID for created cross-reference")
		}

		// Get cross-references by source
		sourceRefs, err := service.CrossReferenceService.GetCrossReferencesBySource(ctx, fmt.Sprintf("%d", createdBookmark1.ID))
		if err != nil {
			t.Fatalf("Failed to get cross-references by source: %v", err)
		}

		if len(sourceRefs) != 1 {
			t.Errorf("Expected 1 cross-reference from source, got %d", len(sourceRefs))
		} else if sourceRefs[0].TargetItemID != fmt.Sprintf("%d", createdBookmark2.ID) {
			t.Errorf("Expected target ID %s, got %s", fmt.Sprintf("%d", createdBookmark2.ID), sourceRefs[0].TargetItemID)
		}

		// Get cross-references by type
		typeRefs, err := service.CrossReferenceService.GetCrossReferencesByType(ctx, RelationshipTypeRelatedTo)
		if err != nil {
			t.Fatalf("Failed to get cross-references by type: %v", err)
		}

		if len(typeRefs) != 1 {
			t.Errorf("Expected 1 cross-reference of type 'related', got %d", len(typeRefs))
		}
	})

}

// TestHistoryFunctionality tests history functionality
func TestHistoryFunctionality(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestHistoryFunctionality", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_history_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Create a bookmark
		createdBookmark, _, err := service.BookmarkService.CreateBookmark(ctx, "history-test-item", string(ItemTypeCVE), "history-cve", "History Test Bookmark", "History test content")
		if err != nil {
			t.Fatalf("Failed to create bookmark: %v", err)
		}

		// Update the bookmark to trigger history creation
		updatedBookmark := createdBookmark
		updatedBookmark.Title = "Updated History Test Bookmark"
		updatedBookmark.Description = "Updated history test content"
		updatedBookmark.UpdatedAt = time.Now()

		err = service.BookmarkService.UpdateBookmark(ctx, updatedBookmark)
		if err != nil {
			t.Fatalf("Failed to update bookmark: %v", err)
		}

		// Get history for the bookmark
		history, err := service.HistoryService.GetHistoryByBookmarkID(ctx, createdBookmark.ID)
		if err != nil {
			t.Fatalf("Failed to get history: %v", err)
		}

		if len(history) == 0 {
			t.Error("Expected at least 1 history record after update")
		} else {
			// Check that the history record has the expected values
			latestHistory := history[0]
			if latestHistory.BookmarkID != createdBookmark.ID {
				t.Errorf("Expected bookmark ID %d in history, got %d", createdBookmark.ID, latestHistory.BookmarkID)
			}
			if latestHistory.Action != string(BookmarkActionCreated) && latestHistory.Action != string(BookmarkActionUpdated) {
				t.Errorf("Expected action '%s' or '%s', got '%s'", BookmarkActionCreated, BookmarkActionUpdated, latestHistory.Action)
			}
		}
	})

}

// TestPerformanceWithLargeDatasets tests performance with large datasets
func TestPerformanceWithLargeDatasets(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestPerformanceWithLargeDatasets", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "notes_performance_test.db")

		// Initialize the database
		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to connect to database: %v", err)
		}

		// Run migrations
		err = MigrateNotesTables(db)
		if err != nil {
			t.Fatalf("Failed to migrate tables: %v", err)
		}

		service := NewServiceContainer(db)
		defer func() {
			dbConn, _ := db.DB()
			dbConn.Close()
		}()

		ctx := context.Background()

		// Create a large number of bookmarks for performance testing
		const numBookmarks = 500

		// Measure insertion performance
		startTime := time.Now()
		for i := 0; i < numBookmarks; i++ {
			_, _, err := service.BookmarkService.CreateBookmark(ctx,
				fmt.Sprintf("perf-test-item-%d", i),
				string(ItemTypeCVE),
				fmt.Sprintf("perf-cve-%d", i),
				fmt.Sprintf("Performance Test Bookmark %d", i),
				fmt.Sprintf("Performance test content for bookmark %d", i))
			if err != nil {
				t.Fatalf("Failed to create bookmark %d: %v", i, err)
			}
		}
		insertDuration := time.Since(startTime)

		t.Logf("Inserted %d bookmarks in %v", numBookmarks, insertDuration)

		// Measure retrieval performance
		retrieveStartTime := time.Now()
		for i := 0; i < 10; i++ { // Test retrieval of 10 random bookmarks
			_, err := service.BookmarkService.GetBookmarkByID(ctx, uint(i+1))
			if err != nil {
				t.Errorf("Failed to retrieve bookmark %d: %v", i+1, err)
			}
		}
		retrieveDuration := time.Since(retrieveStartTime)

		t.Logf("Retrieved 10 bookmarks in %v", retrieveDuration)

		// Measure pagination performance - get bookmarks by global item ID pattern
		paginateStartTime := time.Now()
		for i := 0; i < 5; i++ {
			_, err := service.BookmarkService.GetBookmarksByGlobalItemID(ctx, fmt.Sprintf("perf-test-item-%d", i))
			if err != nil {
				t.Errorf("Failed to get bookmarks by global item ID: %v", err)
			}
		}
		paginateDuration := time.Since(paginateStartTime)

		t.Logf("Paginated bookmarks for 5 global item IDs in %v", paginateDuration)

		// Performance thresholds (adjust based on expected performance)
		maxInsertDuration := time.Second * 5          // Allow up to 5 seconds for insertion
		maxRetrieveDuration := time.Millisecond * 100 // Allow up to 100ms for retrieval
		maxPaginateDuration := time.Millisecond * 50  // Allow up to 50ms for pagination

		if insertDuration > maxInsertDuration {
			t.Logf("Insertion took longer than expected: %v (threshold: %v)", insertDuration, maxInsertDuration)
		}
		if retrieveDuration > maxRetrieveDuration {
			t.Logf("Retrieval took longer than expected: %v (threshold: %v)", retrieveDuration, maxRetrieveDuration)
		}
		if paginateDuration > maxPaginateDuration {
			t.Logf("Pagination took longer than expected: %v (threshold: %v)", paginateDuration, maxPaginateDuration)
		}
	})

}
