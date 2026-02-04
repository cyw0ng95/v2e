package local

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"fmt"
	"os"
	"testing"
)

func BenchmarkLearningDatabaseOperations(b *testing.B) {
	// Create temporary database for benchmarking
	dbPath := "./benchmark_learning.db"
	defer os.Remove(dbPath)

	ldb, err := NewLearningDB(dbPath)
	if err != nil {
		b.Fatalf("Failed to create learning database: %v", err)
	}
	defer ldb.Close()

	b.Run("CreateLearningItem", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			item := &LearningItem{
				ItemType: LearningItemTypeCVE,
				ItemID:   "CVE-2021-1234",
				Status:   LearningStatusNew,
			}
			err := ldb.CreateLearningItem(item)
			if err != nil {
				b.Fatalf("Failed to create learning item: %v", err)
			}
		}
	})

	b.Run("GetLearningItem", func(b *testing.B) {
		// Create one item to retrieve
		item := &LearningItem{
			ItemType: LearningItemTypeCWE,
			ItemID:   "CWE-79",
			Status:   LearningStatusLearning,
		}
		err := ldb.CreateLearningItem(item)
		if err != nil {
			b.Fatalf("Failed to create test item: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := ldb.GetLearningItem(item.ID)
			if err != nil {
				b.Fatalf("Failed to get learning item: %v", err)
			}
		}
	})

	b.Run("RateLearningItem", func(b *testing.B) {
		// Create one item to rate
		item := &LearningItem{
			ItemType: LearningItemTypeCAPEC,
			ItemID:   "CAPEC-1",
			Status:   LearningStatusNew,
		}
		err := ldb.CreateLearningItem(item)
		if err != nil {
			b.Fatalf("Failed to create test item: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := ldb.RateLearningItem(item.ID, "good")
			if err != nil {
				b.Fatalf("Failed to rate learning item: %v", err)
			}
		}
	})

	b.Run("ListLearningItems", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _, err := ldb.ListLearningItems("", 10, 0)
			if err != nil {
				b.Fatalf("Failed to list learning items: %v", err)
			}
		}
	})
}

func BenchmarkBatchLearningOperations(b *testing.B) {
	// Create temporary database for benchmarking
	dbPath := "./benchmark_batch.db"
	defer os.Remove(dbPath)

	ldb, err := NewLearningDB(dbPath)
	if err != nil {
		b.Fatalf("Failed to create learning database: %v", err)
	}
	defer ldb.Close()

	b.Run("RateLearningItems_Single", func(b *testing.B) {
		// Create items for testing
		var items []*LearningItem
		for i := 0; i < 10; i++ {
			item := &LearningItem{
				ItemType: LearningItemTypeCVE,
				ItemID:   fmt.Sprintf("CVE-2021-%04d", i),
				Status:   LearningStatusNew,
			}
			err := ldb.CreateLearningItem(item)
			if err != nil {
				b.Fatalf("Failed to create test item: %v", err)
			}
			items = append(items, item)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Rate items individually (traditional approach)
			for _, item := range items {
				err := ldb.RateLearningItem(item.ID, "good")
				if err != nil {
					b.Fatalf("Failed to rate item: %v", err)
				}
			}
		}
	})

	b.Run("RateLearningItems_Batch", func(b *testing.B) {
		// Create fresh items for each iteration
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			// Create new items for this iteration
			var ids []uint
			var ratings []string
			for j := 0; j < 10; j++ {
				item := &LearningItem{
					ItemType: LearningItemTypeCVE,
					ItemID:   fmt.Sprintf("CVE-2021-BATCH-%d-%04d", i, j),
					Status:   LearningStatusNew,
				}
				err := ldb.CreateLearningItem(item)
				if err != nil {
					b.Fatalf("Failed to create test item: %v", err)
				}
				ids = append(ids, item.ID)
				ratings = append(ratings, "good")
			}
			b.StartTimer()

			// Rate items in batch
			err := ldb.RateLearningItems(ids, ratings)
			if err != nil {
				b.Fatalf("Failed to rate items in batch: %v", err)
			}
		}
	})
}

func BenchmarkAdaptiveAlgorithm(b *testing.B) {
	// Create temporary database for benchmarking
	dbPath := "./benchmark_adaptive.db"
	defer os.Remove(dbPath)

	ldb, err := NewLearningDB(dbPath)
	if err != nil {
		b.Fatalf("Failed to create learning database: %v", err)
	}
	defer ldb.Close()

	// Create item with substantial history
	item := &LearningItem{
		ItemType: LearningItemTypeCVE,
		ItemID:   "CVE-2021-ADAPTIVE",
		Status:   LearningStatusLearning,
	}
	err = ldb.CreateLearningItem(item)
	if err != nil {
		b.Fatalf("Failed to create test item: %v", err)
	}

	// Build up history
	for i := 0; i < 20; i++ {
		rating := "good"
		if i%3 == 0 {
			rating = "easy"
		} else if i%5 == 0 {
			rating = "hard"
		}
		err = ldb.RateLearningItem(item.ID, rating)
		if err != nil {
			b.Fatalf("Failed to build history: %v", err)
		}
	}

	b.Run("CalculatePerformanceTrend", func(b *testing.B) {
		retrievedItem, err := ldb.GetLearningItem(item.ID)
		if err != nil {
			b.Fatalf("Failed to retrieve item: %v", err)
		}

		history, err := ldb.parseHistory(retrievedItem.HistoryJSON)
		if err != nil {
			b.Fatalf("Failed to parse history: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = ldb.calculatePerformanceTrend(history)
		}
	})

	b.Run("ApplyAdaptiveRating", func(b *testing.B) {
		retrievedItem, err := ldb.GetLearningItem(item.ID)
		if err != nil {
			b.Fatalf("Failed to retrieve item: %v", err)
		}

		history, err := ldb.parseHistory(retrievedItem.HistoryJSON)
		if err != nil {
			b.Fatalf("Failed to parse history: %v", err)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ldb.applyAdaptiveRating(retrievedItem, "good", history)
		}
	})
}
