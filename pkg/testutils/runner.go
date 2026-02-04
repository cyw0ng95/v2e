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

// Run executes a test at the specified level with automatic parallelization
// It respects the V2E_TEST_LEVEL environment variable to filter test execution
//
// level: The minimum test level required to run this test (1, 2, or 3)
// name: The test name (used in t.Run)
// f: The test function to execute
//
// Example usage:
//
//	testutils.Run(t, testutils.Level1, "BasicLogic", func(t *testing.T) {
//	    // Test implementation
//	})
func Run(t *testing.T, level TestLevel, name string, f func(t *testing.T)) {
	t.Helper()

	currentLevel := getTestLevel()

	// Skip test if current level is lower than required level
	if currentLevel < level {
		t.Skipf("Skipping test %s (requires level %d, current level %d)", name, level, currentLevel)
		return
	}

	t.Run(name, func(t *testing.T) {
		t.Parallel() // Auto-enable parallel execution
		f(t)
	})
}

// RunWithDB executes a Level 2+ test with automatic transaction isolation
// The test runs in a transaction that is automatically rolled back on cleanup
//
// level: The minimum test level (must be >= 2 for database tests)
// name: The test name
// db: The GORM database instance
// f: The test function that receives a transaction-wrapped database
//
// Example usage:
//
//	testutils.RunWithDB(t, testutils.Level2, "DatabaseOperation", db, func(t *testing.T, tx *gorm.DB) {
//	    // Test implementation using tx instead of db
//	})
func RunWithDB(t *testing.T, level TestLevel, name string, db *gorm.DB, f func(t *testing.T, tx *gorm.DB)) {
	t.Helper()

	// Enforce minimum level for database tests
	if level < Level2 {
		t.Fatalf("RunWithDB requires level >= 2 (got level %d)", level)
		return
	}

	currentLevel := getTestLevel()

	// Skip test if current level is lower than required level
	if currentLevel < level {
		t.Skipf("Skipping test %s (requires level %d, current level %d)", name, level, currentLevel)
		return
	}

	t.Run(name, func(t *testing.T) {
		t.Parallel() // Auto-enable parallel execution

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
	})
}
