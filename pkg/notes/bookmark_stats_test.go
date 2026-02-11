package notes

import (
	"context"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestBookmarkStatistics(t *testing.T) {
	// Create a temporary database for testing
	db := setupTestDB(t)
	defer cleanupTestDB(db)

	// Create bookmark service
	bookmarkService := NewBookmarkService(db)
	require.NotNil(t, bookmarkService)

	ctx := context.Background()

	// Test 1: Create a bookmark and verify initial stats
	testutils.Run(t, testutils.Level2, "CreateBookmarkWithInitialStats", db, func(t *testing.T, tx *gorm.DB) {
		bookmark, _, err := bookmarkService.CreateBookmark(
			ctx,
			"test-global-id-1",
			"CVE",
			"CVE-2024-0001",
			"Test CVE Item",
			"A test CVE for statistics",
		)
		require.NoError(t, err)
		require.NotNil(t, bookmark)

		// Check initial stats
		stats, err := bookmarkService.GetBookmarkStats(ctx, bookmark.ID)
		require.NoError(t, err)
		require.NotNil(t, stats)

		// Verify initial values
		assert.Equal(t, float64(0), stats["view_count"])
		assert.Equal(t, float64(0), stats["study_sessions"])
		assert.NotEmpty(t, stats["first_bookmarked"])
		assert.NotEmpty(t, stats["last_viewed"])

		// Parse timestamps to verify they're valid
		firstBookmarked, ok := stats["first_bookmarked"].(string)
		assert.True(t, ok)
		lastViewed, ok := stats["last_viewed"].(string)
		assert.True(t, ok)

		// Both timestamps should be the same initially
		assert.Equal(t, firstBookmarked, lastViewed)
	})

	// Test 2: Update bookmark stats with positive deltas
	testutils.Run(t, testutils.Level2, "UpdateBookmarkStatsPositiveDeltas", db, func(t *testing.T, tx *gorm.DB) {
		bookmark, _, err := bookmarkService.CreateBookmark(
			ctx,
			"test-global-id-2",
			"CWE",
			"CWE-123",
			"Test CWE Item",
			"A test CWE for statistics",
		)
		require.NoError(t, err)
		require.NotNil(t, bookmark)

		// Get initial stats
		initialStats, err := bookmarkService.GetBookmarkStats(ctx, bookmark.ID)
		require.NoError(t, err)

		// Longer delay to ensure timestamp difference (RFC3339 has second precision)
		time.Sleep(1 * time.Second)

		// Update stats (increment view count by 3, study sessions by 2)
		err = bookmarkService.UpdateBookmarkStats(ctx, bookmark.ID, 3, 2)
		require.NoError(t, err)

		// Get updated stats
		updatedStats, err := bookmarkService.GetBookmarkStats(ctx, bookmark.ID)
		require.NoError(t, err)

		// Verify increments
		assert.Equal(t, float64(3), updatedStats["view_count"])
		assert.Equal(t, float64(2), updatedStats["study_sessions"])
		assert.Equal(t, initialStats["first_bookmarked"], updatedStats["first_bookmarked"])
		// last_viewed should be updated (different from initial)
		assert.NotEqual(t, initialStats["last_viewed"], updatedStats["last_viewed"])

		// Verify last_viewed was updated to current time
		lastViewedStr, ok := updatedStats["last_viewed"].(string)
		assert.True(t, ok)

		lastViewed, err := time.Parse(time.RFC3339, lastViewedStr)
		require.NoError(t, err)

		// Should be within a reasonable time window of now
		now := time.Now()
		timeDiff := now.Sub(lastViewed)
		assert.True(t, timeDiff >= 0 && timeDiff < time.Second*5, "last_viewed should be recent")
	})

	// Test 3: Update bookmark stats with zero deltas (should still update timestamp)
	testutils.Run(t, testutils.Level2, "UpdateBookmarkStatsZeroDeltas", db, func(t *testing.T, tx *gorm.DB) {
		bookmark, _, err := bookmarkService.CreateBookmark(
			ctx,
			"test-global-id-3",
			"CAPEC",
			"CAPEC-456",
			"Test CAPEC Item",
			"A test CAPEC for statistics",
		)
		require.NoError(t, err)
		require.NotNil(t, bookmark)

		// Get initial stats
		initialStats, err := bookmarkService.GetBookmarkStats(ctx, bookmark.ID)
		require.NoError(t, err)

		// Longer delay to ensure timestamp difference (RFC3339 has second precision)
		time.Sleep(1 * time.Second)

		// Update stats with zero deltas
		err = bookmarkService.UpdateBookmarkStats(ctx, bookmark.ID, 0, 0)
		require.NoError(t, err)

		// Get updated stats
		updatedStats, err := bookmarkService.GetBookmarkStats(ctx, bookmark.ID)
		require.NoError(t, err)

		// Values should remain the same
		assert.Equal(t, initialStats["view_count"], updatedStats["view_count"])
		assert.Equal(t, initialStats["study_sessions"], updatedStats["study_sessions"])
		assert.Equal(t, initialStats["first_bookmarked"], updatedStats["first_bookmarked"])

		// But last_viewed should be updated (different from initial)
		assert.NotEqual(t, initialStats["last_viewed"], updatedStats["last_viewed"])

		// Verify that the new last_viewed is more recent than first_bookmarked
		firstBookmarkedStr, ok := updatedStats["first_bookmarked"].(string)
		assert.True(t, ok)
		lastViewedStr, ok := updatedStats["last_viewed"].(string)
		assert.True(t, ok)

		firstTime, err := time.Parse(time.RFC3339, firstBookmarkedStr)
		require.NoError(t, err)
		lastTime, err := time.Parse(time.RFC3339, lastViewedStr)
		require.NoError(t, err)

		// last_viewed should be equal to or after first_bookmarked
		assert.True(t, lastTime.After(firstTime) || lastTime.Equal(firstTime))
	})

	// Test 4: Verify single bookmark constraint with stats
	testutils.Run(t, testutils.Level2, "SingleBookmarkConstraintWithStats", db, func(t *testing.T, tx *gorm.DB) {
		// Create first bookmark
		bookmark1, _, err := bookmarkService.CreateBookmark(
			ctx,
			"unique-global-id",
			"CVE",
			"CVE-2024-9999",
			"Unique CVE Item",
			"This should be the only bookmark for this item",
		)
		require.NoError(t, err)
		require.NotNil(t, bookmark1)

		// Try to create duplicate bookmark - should return existing
		bookmark2, _, err := bookmarkService.CreateBookmark(
			ctx,
			"unique-global-id",
			"CVE",
			"CVE-2024-9999",
			"Duplicate CVE Item",
			"This should not be created",
		)
		require.NoError(t, err)
		require.NotNil(t, bookmark2)

		// Should be the same bookmark
		assert.Equal(t, bookmark1.ID, bookmark2.ID)
		assert.Equal(t, bookmark1.GlobalItemID, bookmark2.GlobalItemID)

		// Stats should be preserved from the original bookmark
		stats1, err := bookmarkService.GetBookmarkStats(ctx, bookmark1.ID)
		require.NoError(t, err)

		stats2, err := bookmarkService.GetBookmarkStats(ctx, bookmark2.ID)
		require.NoError(t, err)

		assert.Equal(t, stats1, stats2)
	})

	// Test 5: Update stats on existing bookmark multiple times
	testutils.Run(t, testutils.Level2, "MultipleStatUpdates", db, func(t *testing.T, tx *gorm.DB) {
		bookmark, _, err := bookmarkService.CreateBookmark(
			ctx,
			"test-global-id-4",
			"ATT&CK",
			"T1001",
			"Test ATT&CK Item",
			"A test ATT&CK technique",
		)
		require.NoError(t, err)
		require.NotNil(t, bookmark)

		// Perform multiple updates
		updates := []struct {
			viewDelta   int
			studyDelta  int
			description string
		}{
			{1, 0, "First view"},
			{2, 1, "Second view with study session"},
			{0, 1, "Additional study session"},
			{3, 0, "Three more views"},
		}

		expectedViewCount := 0
		expectedStudyCount := 0

		for _, update := range updates {
			err = bookmarkService.UpdateBookmarkStats(ctx, bookmark.ID, update.viewDelta, update.studyDelta)
			require.NoError(t, err, update.description)

			expectedViewCount += update.viewDelta
			expectedStudyCount += update.studyDelta

			// Verify cumulative stats
			stats, err := bookmarkService.GetBookmarkStats(ctx, bookmark.ID)
			require.NoError(t, err, update.description)

			assert.Equal(t, float64(expectedViewCount), stats["view_count"], update.description)
			assert.Equal(t, float64(expectedStudyCount), stats["study_sessions"], update.description)
		}
	})
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(db *gorm.DB) {
	// Close database connection
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}
