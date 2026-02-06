package testutils

import (
	"os"
	"strconv"
	"testing"

	"gorm.io/gorm"
)

// TestLevel represents the test isolation level
type TestLevel int

const (
	// Level1 - Fundamental: Pure logic, mock-based, no external dependencies with minimal database related
	Level1 TestLevel = 1
	// Level2 - Integration: Database (GORM) involved
	Level2 TestLevel = 2
	// Level3 - Comprehensive: External APIs, E2E, and heavy integration
	Level3 TestLevel = 3
)

// getTestLevel returns the test level from V2E_TEST_LEVEL environment variable
// Default is Level1 if not set
func getTestLevel() TestLevel {
	levelStr := os.Getenv("V2E_TEST_LEVEL")
	if levelStr == "" {
		return Level1
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return Level1
	}

	if level < 1 {
		return Level1
	}
	if level > 3 {
		return Level3
	}

	return TestLevel(level)
}

// Run executes a test at the specified level with automatic parallelization.
// It respects the V2E_TEST_LEVEL environment variable with CUMULATIVE behavior:
//   - V2E_TEST_LEVEL=3: runs tests tagged as 1, 2, OR 3
//   - V2E_TEST_LEVEL=2: runs tests tagged as 1 OR 2
//   - V2E_TEST_LEVEL=1: runs tests tagged as 1 only
//
// Parameters:
//   - t: The testing.T instance
//   - level: The test level tag (1, 2, or 3)
//   - name: The test name (used in t.Run)
//   - db: Optional GORM database for Level 2+ tests (use nil for Level 1)
//   - f: The test function to execute
//
// For Level 1 tests (no database):
//
//	testutils.Run(t, testutils.Level1, "BasicLogic", nil, func(t *testing.T, tx *gorm.DB) {
//	    // Test implementation (tx will be nil)
//	})
//
// For Level 2+ tests (with database and automatic transaction isolation):
//
//	testutils.Run(t, testutils.Level2, "DatabaseOperation", db, func(t *testing.T, tx *gorm.DB) {
//	    // Test implementation using tx (transaction will auto-rollback)
//	})
func Run(t *testing.T, level TestLevel, name string, db *gorm.DB, f func(t *testing.T, tx *gorm.DB)) {
	t.Helper()

	currentLevel := getTestLevel()

	// Skip test if test level is higher than current level (cumulative behavior)
	// E.g., if V2E_TEST_LEVEL=2, skip tests tagged as level 3
	if level > currentLevel {
		t.Skipf("Skipping test %s (test level %d, current level %d)", name, level, currentLevel)
		return
	}

	t.Run(name, func(t *testing.T) {
		t.Parallel() // Auto-enable parallel execution

		// If database is provided, wrap in transaction for isolation
		if db != nil {
			// Begin transaction for isolation
			tx := db.Begin()
			if tx.Error != nil {
				t.Fatalf("Failed to begin transaction: %v", tx.Error)
				return
			}

			// Ensure rollback on cleanup (even if test fails)
			t.Cleanup(func() {
				if err := tx.Rollback().Error; err != nil {
					t.Logf("Warning: Failed to rollback transaction: %v", err)
				}
			})

			// Execute test with transaction-wrapped database
			f(t, tx)
		} else {
			// Execute test without database
			f(t, nil)
		}
	})
}
