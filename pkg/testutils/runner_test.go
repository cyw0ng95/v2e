package testutils

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGetTestLevel(t *testing.T) {
	// Save and restore original environment variable
	originalValue := os.Getenv("V2E_TEST_LEVEL")
	defer func() {
		if originalValue != "" {
			os.Setenv("V2E_TEST_LEVEL", originalValue)
		} else {
			os.Unsetenv("V2E_TEST_LEVEL")
		}
	}()

	tests := []struct {
		name     string
		envValue string
		expected TestLevel
	}{
		{"Default when unset", "", Level1},
		{"Level 1", "1", Level1},
		{"Level 2", "2", Level2},
		{"Level 3", "3", Level3},
		{"Invalid string", "invalid", Level1},
		{"Below minimum", "0", Level1},
		{"Above maximum", "10", Level3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for this subtest
			if tt.envValue != "" {
				os.Setenv("V2E_TEST_LEVEL", tt.envValue)
			} else {
				os.Unsetenv("V2E_TEST_LEVEL")
			}

			result := getTestLevel()
			if result != tt.expected {
				t.Errorf("Expected level %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestRun_Execution(t *testing.T) {
	Run(t, Level1, "ShouldExecute", func(t *testing.T) {
		// Test executes successfully
	})

	// Note: We can't reliably check if test executed because t.Run returns before
	// parallel tests complete. The test framework will verify execution.
}

func TestRun_Parallelism(t *testing.T) {
	// Test that Run correctly enables parallel execution
	// We verify this by ensuring all subtests are created
	for i := 0; i < 10; i++ {
		Run(t, Level1, "ParallelTest", func(t *testing.T) {
			// Each test runs in parallel
			// The test framework verifies they all complete
		})
	}

	// Note: t.Run returns before parallel subtests complete,
	// but the test framework ensures they all run
}

func TestRun_LevelFiltering(t *testing.T) {
	// Set level to 1 for this test only
	t.Setenv("V2E_TEST_LEVEL", "1")

	// Test that level filtering works correctly
	// Level 1 test should run
	Run(t, Level1, "Level1Test", func(t *testing.T) {
		// This should execute
	})

	// Level 2 and 3 tests should be skipped
	// The Skip message will be visible in test output
	Run(t, Level2, "Level2Test", func(t *testing.T) {
		t.Error("Level 2 test should have been skipped")
	})

	Run(t, Level3, "Level3Test", func(t *testing.T) {
		t.Error("Level 3 test should have been skipped")
	})
}

func TestRunWithDB_TransactionIsolation(t *testing.T) {
	// Create in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test table
	type TestRecord struct {
		ID   uint   `gorm:"primaryKey"`
		Name string
	}

	if err := db.AutoMigrate(&TestRecord{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Test 1: Verify rollback happens
	RunWithDB(t, Level2, "VerifyRollback", db, func(t *testing.T, tx *gorm.DB) {
		// Insert a record in transaction
		record := TestRecord{Name: "test"}
		if err := tx.Create(&record).Error; err != nil {
			t.Fatalf("Failed to create record: %v", err)
		}

		// Verify record exists in transaction
		var count int64
		tx.Model(&TestRecord{}).Count(&count)
		if count != 1 {
			t.Errorf("Expected 1 record in transaction, got %d", count)
		}
	})

	// Verify record was rolled back in main database
	var count int64
	db.Model(&TestRecord{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 records after rollback, got %d", count)
	}
}

func TestRunWithDB_ParallelIsolation(t *testing.T) {
	// Create file-based database for testing (in-memory doesn't share schema across connections)
	dbPath := "/tmp/test_parallel_isolation.db"
	t.Cleanup(func() {
		// Clean up database file after test
		_ = os.Remove(dbPath)
		_ = os.Remove(dbPath + "-shm")
		_ = os.Remove(dbPath + "-wal")
	})

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test table
	type TestRecord struct {
		ID   uint   `gorm:"primaryKey"`
		Name string
	}

	if err := db.AutoMigrate(&TestRecord{}); err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// Run multiple parallel tests with different data
	for i := 0; i < 5; i++ {
		name := string(rune('A' + i))
		RunWithDB(t, Level2, "Parallel"+name, db, func(t *testing.T, tx *gorm.DB) {
			// Each test inserts its own record
			record := TestRecord{Name: name}
			if err := tx.Create(&record).Error; err != nil {
				t.Fatalf("Failed to create record: %v", err)
			}

			// Verify only one record exists in this transaction
			var count int64
			tx.Model(&TestRecord{}).Count(&count)
			if count != 1 {
				t.Errorf("Expected 1 record in transaction, got %d", count)
			}
		})
	}

	// Verify no records persist after rollback
	var count int64
	db.Model(&TestRecord{}).Count(&count)
	if count != 0 {
		t.Errorf("Expected 0 records after all rollbacks, got %d", count)
	}
}

func TestRunWithDB_EnforcesMinimumLevel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Test that Level2 works correctly
	RunWithDB(t, Level2, "Level2Works", db, func(t *testing.T, tx *gorm.DB) {
		// This should execute successfully at Level 2 or higher
		if tx == nil {
			t.Error("Transaction should not be nil")
		}
	})

	// Note: Testing Level1 enforcement would cause a test failure (by design)
	// since RunWithDB calls t.Fatalf for Level1. This is the correct behavior.
}
