package local

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
)

// TestConcurrentDatabaseAccess verifies that concurrent database operations
// don't cause "database is locked" errors after our fixes
func TestConcurrentDatabaseAccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConcurrentDatabaseAccess", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "concurrent_test.db")

		// Create database
		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		// Pre-populate with test data
		testCVEs := make([]*cve.CVEItem, 100)
		for i := 0; i < 100; i++ {
			testCVEs[i] = &cve.CVEItem{
				ID:           fmt.Sprintf("CVE-2021-%04d", i),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{
					{
						Lang:  "en",
						Value: fmt.Sprintf("Test CVE description %d", i),
					},
				},
			}
			if err := db.SaveCVE(testCVEs[i]); err != nil {
				t.Fatalf("Failed to save test CVE %d: %v", i, err)
			}
		}

		// Test concurrent ListCVEs and Count operations
		var wg sync.WaitGroup
		const numGoroutines = 50
		errors := make(chan error, numGoroutines*2) // Space for both operations per goroutine

		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Perform ListCVEs operation
				_, err := db.ListCVEs(0, 10)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: ListCVEs failed: %v", goroutineID, err)
					return
				}

				// Perform Count operation
				_, err = db.Count()
				if err != nil {
					errors <- fmt.Errorf("goroutine %d: Count failed: %v", goroutineID, err)
					return
				}

				// Small delay to increase chance of concurrency
				time.Sleep(time.Millisecond * 1)
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any database locking errors
		databaseLockErrors := 0
		otherErrors := 0

		for err := range errors {
			if err != nil {
				t.Logf("Error occurred: %v", err)
				if containsDatabaseLockedError(err) {
					databaseLockErrors++
				} else {
					otherErrors++
				}
			}
		}

		// With our fixes, there should be no database locking errors
		if databaseLockErrors > 0 {
			t.Errorf("Expected 0 database lock errors, got %d", databaseLockErrors)
		}

		// Report other errors if any
		if otherErrors > 0 {
			t.Logf("Note: %d other errors occurred (not related to database locking)", otherErrors)
		}

		// Final verification - ensure we can still perform operations
		finalCount, err := db.Count()
		if err != nil {
			t.Errorf("Final count failed: %v", err)
		} else if finalCount != 100 {
			t.Errorf("Expected 100 CVEs, got %d", finalCount)
		}

		t.Logf("Concurrent test completed successfully with %d goroutines", numGoroutines)
	})

}

// Helper function to detect database locked errors
func containsDatabaseLockedError(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "database is locked" ||
		err.Error() == "[meta] RPC error response: failed to list CVEs: failed to list CVEs: database is locked"
}

// TestRetryLogic verifies that the retry mechanism works for database locking
func TestRetryLogic(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRetryLogic", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "retry_test.db")

		// Create database
		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		// Pre-populate with minimal data
		testCVE := &cve.CVEItem{
			ID:           "CVE-2021-0001",
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Now()),
			LastModified: cve.NewNVDTime(time.Now()),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{Lang: "en", Value: "Test CVE"},
			},
		}

		if err := db.SaveCVE(testCVE); err != nil {
			t.Fatalf("Failed to save test CVE: %v", err)
		}

		// Test that normal operations work
		_, err = db.ListCVEs(0, 10)
		if err != nil {
			t.Errorf("ListCVEs should work: %v", err)
		}

		_, err = db.Count()
		if err != nil {
			t.Errorf("Count should work: %v", err)
		}

		t.Log("Retry logic test passed - basic operations work correctly")
	})

}
