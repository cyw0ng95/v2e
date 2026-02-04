# Test File Conversion Guide

## Overview

This guide explains how to convert the remaining 158 test files to use the unified `testutils.Run()` wrapper.

## Quick Reference

### Conversion Formula

**Level 1 (No Database):**
```go
// OLD
func TestExample(t *testing.T) {
    t.Run("Case1", func(t *testing.T) {
        // test code
    })
}

// NEW
import "github.com/cyw0ng95/v2e/pkg/testutils"

func TestExample(t *testing.T) {
    testutils.Run(t, testutils.Level1, "Case1", nil, func(t *testing.T, tx *gorm.DB) {
        // test code (unchanged)
    })
}
```

**Level 2 (Database):**
```go
// OLD
func TestDatabase(t *testing.T) {
    db := setupTestDB(t)
    t.Run("Insert", func(t *testing.T) {
        db.Create(&record)
    })
}

// NEW  
import (
    "github.com/cyw0ng95/v2e/pkg/testutils"
    "gorm.io/gorm"
)

func TestDatabase(t *testing.T) {
    db := setupTestDB(t)
    testutils.Run(t, testutils.Level2, "Insert", db, func(t *testing.T, tx *gorm.DB) {
        tx.Create(&record)  // Use tx instead of db!
    })
}
```

## Step-by-Step Conversion

### Step 1: Add Imports

Add testutils import:
```go
import (
    "testing"
    
    "github.com/cyw0ng95/v2e/pkg/testutils"
)
```

For database tests, also add:
```go
import "gorm.io/gorm"
```

### Step 2: Replace t.Run with testutils.Run

**Pattern:** `t.Run(name, func(t *testing.T) { ... })`

**Becomes:** `testutils.Run(t, level, name, db, func(t *testing.T, tx *gorm.DB) { ... })`

Where:
- `level` = `testutils.Level1` (no DB) or `testutils.Level2` (with DB)
- `db` = `nil` (Level 1) or database instance (Level 2)
- `tx` = transaction parameter (nil for Level 1, transaction for Level 2)

### Step 3: Update Database References

For Level 2 tests, replace `db` with `tx` inside the test function:
- `db.Create(...)` → `tx.Create(...)`
- `db.Find(...)` → `tx.Find(...)`
- etc.

## Level Assignment Rules

- **Level 1**: Pure logic, no database, no external APIs
  - Most tests in cmd/*, pkg/proc/*, pkg/common/*
  
- **Level 2**: Database operations (uses gorm.DB)
  - Tests in pkg/notes/*, pkg/cve/local*, pkg/cwe/local*, pkg/capec/local*
  
- **Level 3**: External API calls, E2E tests
  - Rare, assign manually

## Batch Conversion Strategy

### Priority 1: Database Tests (4 files)
- [x] pkg/notes/service_test.go (done)
- [ ] pkg/notes/service_status_test.go
- [ ] pkg/notes/bookmark_stats_test.go
- [ ] pkg/testutils/runner_test.go (done)

### Priority 2: Core Packages (estimated 30 files)
- [ ] pkg/cve/*_test.go
- [ ] pkg/cwe/*_test.go
- [ ] pkg/capec/*_test.go
- [ ] pkg/attack/*_test.go

### Priority 3: Broker Tests (estimated 40 files)
- [ ] cmd/v2broker/core/*_test.go
- [ ] cmd/v2broker/routing/*_test.go
- [ ] cmd/v2broker/transport/*_test.go
- [ ] cmd/v2broker/mq/*_test.go
- [ ] cmd/v2broker/perf/*_test.go

### Priority 4: Remaining (estimated 84 files)
- [ ] pkg/proc/*_test.go
- [ ] pkg/common/*_test.go
- [ ] pkg/rpc/*_test.go
- [ ] pkg/ssg/*_test.go
- [ ] tool/*_test.go
- [ ] Other cmd/*_test.go

## Common Patterns

### Pattern: Simple Test
```go
func TestAdd(t *testing.T) {
    testutils.Run(t, testutils.Level1, "Addition", nil, func(t *testing.T, tx *gorm.DB) {
        if Add(2, 3) != 5 {
            t.Error("math broken")
        }
    })
}
```

### Pattern: Table-Driven Tests
```go
func TestValidation(t *testing.T) {
    tests := []struct {
        name string
        input string
        want bool
    }{
        {"valid", "abc", true},
        {"invalid", "123", false},
    }
    
    for _, tt := range tests {
        testutils.Run(t, testutils.Level1, tt.name, nil, func(t *testing.T, tx *gorm.DB) {
            got := Validate(tt.input)
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

### Pattern: Database Test with Setup
```go
func TestUserCRUD(t *testing.T) {
    db := setupTestDB(t)
    
    testutils.Run(t, testutils.Level2, "Create", db, func(t *testing.T, tx *gorm.DB) {
        user := &User{Name: "test"}
        if err := tx.Create(user).Error; err != nil {
            t.Fatal(err)
        }
        // Transaction auto-rollbacks - no cleanup needed!
    })
}
```

## Verification

After conversion, verify:
1. Tests still pass: `V2E_TEST_LEVEL=1 go test ./...`
2. Database tests work: `V2E_TEST_LEVEL=2 go test ./...`
3. All tests run: `V2E_TEST_LEVEL=3 go test ./...`

## Automation Script

For batch conversion, use:
```bash
# Convert a single file
./scripts/convert_test.sh path/to/test_file.go

# Convert a directory
./scripts/convert_test.sh pkg/cve/
```

## Notes

- Keep changes minimal - only wrap with testutils.Run
- Don't change test logic or structure
- Transaction isolation for Level 2 prevents database pollution
- Tests run in parallel automatically
