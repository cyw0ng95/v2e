package local

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"os"
	"testing"
)

func TestLearningDatabaseOperations(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLearningDatabaseOperations", nil, func(t *testing.T, tx *gorm.DB) {
		// Create temporary database for testing
		dbPath := "./test_learning.db"
		defer os.Remove(dbPath)

		// Create learning database
		ldb, err := NewLearningDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create learning database: %v", err)
		}
		defer ldb.Close()

		// Test creating a learning item
		item := &LearningItem{
			ItemType: LearningItemTypeCVE,
			ItemID:   "CVE-2021-1234",
			Status:   LearningStatusNew,
		}

		err = ldb.CreateLearningItem(item)
		if err != nil {
			t.Fatalf("Failed to create learning item: %v", err)
		}

		// Test retrieving the item
		retrievedItem, err := ldb.GetLearningItem(item.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve learning item: %v", err)
		}

		if retrievedItem.ItemType != LearningItemTypeCVE {
			t.Errorf("Expected item type %s, got %s", LearningItemTypeCVE, retrievedItem.ItemType)
		}

		if retrievedItem.ItemID != "CVE-2021-1234" {
			t.Errorf("Expected item ID CVE-2021-1234, got %s", retrievedItem.ItemID)
		}

		// Test updating item with rating
		err = ldb.RateLearningItem(item.ID, "good")
		if err != nil {
			t.Fatalf("Failed to rate learning item: %v", err)
		}

		// Verify the item was updated
		updatedItem, err := ldb.GetLearningItem(item.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated learning item: %v", err)
		}

		if updatedItem.Repetition != 1 {
			t.Errorf("Expected repetition 1, got %d", updatedItem.Repetition)
		}

		if updatedItem.Status != LearningStatusLearning {
			t.Errorf("Expected status learning, got %s", updatedItem.Status)
		}

		// Test listing items
		items, total, err := ldb.ListLearningItems("", 10, 0)
		if err != nil {
			t.Fatalf("Failed to list learning items: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected 1 total item, got %d", total)
		}

		if len(items) != 1 {
			t.Errorf("Expected 1 item in list, got %d", len(items))
		}

		// Test getting stats
		stats, err := ldb.GetLearningStats()
		if err != nil {
			t.Fatalf("Failed to get learning stats: %v", err)
		}

		if stats["learning"] != 1 {
			t.Errorf("Expected 1 learning item in stats, got %d", stats["learning"])
		}

		// Test deleting item
		err = ldb.DeleteLearningItem(item.ID)
		if err != nil {
			t.Fatalf("Failed to delete learning item: %v", err)
		}

		// Verify item is deleted
		_, err = ldb.GetLearningItem(item.ID)
		if err == nil {
			t.Error("Expected error when retrieving deleted item, got nil")
		}
	})

}

func TestLearningItemDefaults(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLearningItemDefaults", nil, func(t *testing.T, tx *gorm.DB) {
		item := &LearningItem{}

		// Test that defaults are set correctly
		if item.Status != "" {
			t.Errorf("Expected empty status initially, got %s", item.Status)
		}

		if item.EaseFactor != 0 {
			t.Errorf("Expected ease factor 0 initially, got %f", item.EaseFactor)
		}

		if item.Interval != 0 {
			t.Errorf("Expected interval 0 initially, got %d", item.Interval)
		}
	})

}

func TestAdaptiveSpacedRepetition(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestAdaptiveSpacedRepetition", nil, func(t *testing.T, tx *gorm.DB) {
		// Create temporary database for testing
		dbPath := "./test_adaptive_learning.db"
		defer os.Remove(dbPath)

		// Create learning database
		ldb, err := NewLearningDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create learning database: %v", err)
		}
		defer ldb.Close()

		// Test creating a learning item
		item := &LearningItem{
			ItemType: LearningItemTypeCVE,
			ItemID:   "CVE-2021-1234",
			Status:   LearningStatusNew,
		}

		err = ldb.CreateLearningItem(item)
		if err != nil {
			t.Fatalf("Failed to create learning item: %v", err)
		}

		// Test adaptive rating - good performance should increase confidence

		// Apply "easy" rating
		err = ldb.RateLearningItem(item.ID, "easy")
		if err != nil {
			t.Fatalf("Failed to rate learning item: %v", err)
		}

		// Verify confidence increased
		updatedItem, err := ldb.GetLearningItem(item.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve updated item: %v", err)
		}

		// Parse history to check confidence
		history, err := ldb.parseHistory(updatedItem.HistoryJSON)
		if err != nil {
			t.Fatalf("Failed to parse history: %v", err)
		}

		if history.Confidence <= 0.0 {
			t.Errorf("Expected confidence to increase after 'easy' rating, got %f", history.Confidence)
		}

		// For new items (repetition=0), interval remains 1 even after 'easy' rating
		// This is expected behavior in the SM-2 algorithm
		if updatedItem.Repetition != 1 {
			t.Errorf("Expected repetition to be 1 after rating, got %d", updatedItem.Repetition)
		}
	})

}

func TestBatchLearningOperations(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBatchLearningOperations", nil, func(t *testing.T, tx *gorm.DB) {
		// Create temporary database for testing
		dbPath := "./test_batch_learning.db"
		defer os.Remove(dbPath)

		// Create learning database
		ldb, err := NewLearningDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create learning database: %v", err)
		}
		defer ldb.Close()

		// Create multiple learning items
		items := []*LearningItem{
			{
				ItemType: LearningItemTypeCVE,
				ItemID:   "CVE-2021-0001",
				Status:   LearningStatusNew,
			},
			{
				ItemType: LearningItemTypeCWE,
				ItemID:   "CWE-79",
				Status:   LearningStatusLearning,
			},
			{
				ItemType: LearningItemTypeCAPEC,
				ItemID:   "CAPEC-1",
				Status:   LearningStatusNew,
			},
		}

		// Save items
		for _, item := range items {
			err = ldb.CreateLearningItem(item)
			if err != nil {
				t.Fatalf("Failed to create learning item: %v", err)
			}
		}

		// Test batch rating
		ids := []uint{items[0].ID, items[1].ID, items[2].ID}
		ratings := []string{"good", "easy", "hard"}

		err = ldb.RateLearningItems(ids, ratings)
		if err != nil {
			t.Fatalf("Failed to rate learning items in batch: %v", err)
		}

		// Verify all items were updated
		for i, id := range ids {
			item, err := ldb.GetLearningItem(id)
			if err != nil {
				t.Fatalf("Failed to retrieve item %d: %v", id, err)
			}

			// Check that status was updated appropriately
			expectedStatus := LearningStatusLearning
			if ratings[i] == "easy" {
				expectedStatus = LearningStatusMastered
			}

			if item.Status != expectedStatus {
				t.Errorf("Item %d: expected status %s, got %s", id, expectedStatus, item.Status)
			}
		}
	})

}

func TestLearningAnalytics(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLearningAnalytics", nil, func(t *testing.T, tx *gorm.DB) {
		// Create temporary database for testing
		dbPath := "./test_analytics.db"
		defer os.Remove(dbPath)

		// Create learning database
		ldb, err := NewLearningDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create learning database: %v", err)
		}
		defer ldb.Close()

		// Create and rate some items
		items := []*LearningItem{
			{
				ItemType: LearningItemTypeCVE,
				ItemID:   "CVE-2021-0001",
				Status:   LearningStatusNew,
			},
			{
				ItemType: LearningItemTypeCWE,
				ItemID:   "CWE-79",
				Status:   LearningStatusLearning,
			},
		}

		for _, item := range items {
			err = ldb.CreateLearningItem(item)
			if err != nil {
				t.Fatalf("Failed to create learning item: %v", err)
			}
			// Rate the items to generate some history
			err = ldb.RateLearningItem(item.ID, "good")
			if err != nil {
				t.Fatalf("Failed to rate item: %v", err)
			}
		}

		// Test learning trends
		trends, err := ldb.GetLearningTrends(30) // Last 30 days
		if err != nil {
			t.Fatalf("Failed to get learning trends: %v", err)
		}

		if trends["total_reviews"] != 2 {
			t.Errorf("Expected 2 total reviews, got %v", trends["total_reviews"])
		}

		// Test mastery prediction
		prediction, err := ldb.PredictMasteryTimeline(LearningItemTypeCVE, "CVE-2021-0001")
		if err != nil {
			t.Fatalf("Failed to predict mastery timeline: %v", err)
		}

		if prediction["current_confidence"] == nil {
			t.Error("Expected current confidence in prediction")
		}

		if prediction["performance_trend"] == nil {
			t.Error("Expected performance trend in prediction")
		}
	})

}

func TestLearningHistoryPersistence(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestLearningHistoryPersistence", nil, func(t *testing.T, tx *gorm.DB) {
		// Create temporary database for testing
		dbPath := "./test_history.db"
		defer os.Remove(dbPath)

		// Create learning database
		ldb, err := NewLearningDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create learning database: %v", err)
		}
		defer ldb.Close()

		// Create learning item
		item := &LearningItem{
			ItemType: LearningItemTypeCVE,
			ItemID:   "CVE-2021-HIST",
			Status:   LearningStatusNew,
		}

		err = ldb.CreateLearningItem(item)
		if err != nil {
			t.Fatalf("Failed to create learning item: %v", err)
		}

		// Apply multiple ratings to build history
		ratings := []string{"again", "hard", "good", "easy"}
		for _, rating := range ratings {
			err = ldb.RateLearningItem(item.ID, rating)
			if err != nil {
				t.Fatalf("Failed to rate item with '%s': %v", rating, err)
			}
		}

		// Retrieve item and verify history persistence
		retrievedItem, err := ldb.GetLearningItem(item.ID)
		if err != nil {
			t.Fatalf("Failed to retrieve item: %v", err)
		}

		history, err := ldb.parseHistory(retrievedItem.HistoryJSON)
		if err != nil {
			t.Fatalf("Failed to parse history: %v", err)
		}

		if len(history.Ratings) != 4 {
			t.Errorf("Expected 4 ratings in history, got %d", len(history.Ratings))
		}

		// Verify the ratings are in correct order
		expectedRatings := []string{"again", "hard", "good", "easy"}
		for i, expected := range expectedRatings {
			if history.Ratings[i] != expected {
				t.Errorf("Rating %d: expected %s, got %s", i, expected, history.Ratings[i])
			}
		}
	})

}
