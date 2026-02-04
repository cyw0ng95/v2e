# Test Conversion Status

## Current State

**Total Test Files**: 159
**Converted Files**: 1
**Remaining Files**: 158

## Conversion Infrastructure ✅

The following infrastructure is complete and ready for use:

### 1. Unified Wrapper API
- `testutils.Run(t, level, name, db, f)` - Single function for all tests
- Automatic parallelization via `t.Parallel()`
- Automatic transaction isolation for Level 2+ tests
- Cumulative level behavior (Level 3 runs 1+2+3)

### 2. CI Integration
- Single job instead of matrix (saves CPU time)
- `V2E_TEST_LEVEL=3` runs all tests
- No parallel matrix builds

### 3. Documentation
- `TEST_CONVERSION_GUIDE.md` - Comprehensive conversion guide
- README.md updated with cumulative behavior
- Copilot instructions updated

## Conversion Strategy

### File Categories

1. **Database Tests (Level 2)** - 4 files
   - Use transactions for isolation
   - Pass db parameter to testutils.Run
   - Replace db references with tx inside tests
   
   Files:
   - [ ] pkg/notes/service_status_test.go
   - [ ] pkg/notes/bookmark_stats_test.go
   - [x] pkg/notes/service_test.go (partially done)
   - [x] pkg/testutils/runner_test.go (done)

2. **Logic Tests (Level 1)** - ~154 files
   - Pass nil for db parameter
   - No database operations
   - Most tests fall into this category
   
   Examples:
   - pkg/cve/job/*_test.go
   - pkg/cwe/*_test.go
   - pkg/capec/*_test.go
   - cmd/v2broker/core/*_test.go
   - cmd/v2broker/routing/*_test.go
   - pkg/proc/*_test.go
   - pkg/common/*_test.go
   - etc.

## Conversion Patterns by File Type

### Pattern A: Simple Test Functions
**Files**: ~100 files
**Complexity**: Low
**Example**: pkg/cve/job/controller_test.go

```go
// Before
func TestSomething(t *testing.T) {
    result := DoSomething()
    if result != expected {
        t.Error("failed")
    }
}

// After (add import, wrap function body)
import "github.com/cyw0ng95/v2e/pkg/testutils"

func TestSomething(t *testing.T) {
    testutils.Run(t, testutils.Level1, "TestSomething", nil, func(t *testing.T, tx *gorm.DB) {
        result := DoSomething()
        if result != expected {
            t.Error("failed")
        }
    })
}
```

### Pattern B: Tests with t.Run
**Files**: ~48 files
**Complexity**: Medium
**Example**: Many broker tests

```go
// Before
func TestFeature(t *testing.T) {
    t.Run("Case1", func(t *testing.T) {
        // test code
    })
    t.Run("Case2", func(t *testing.T) {
        // test code
    })
}

// After (replace t.Run with testutils.Run)
import "github.com/cyw0ng95/v2e/pkg/testutils"

func TestFeature(t *testing.T) {
    testutils.Run(t, testutils.Level1, "Case1", nil, func(t *testing.T, tx *gorm.DB) {
        // test code (unchanged)
    })
    testutils.Run(t, testutils.Level1, "Case2", nil, func(t *testing.T, tx *gorm.DB) {
        // test code (unchanged)
    })
}
```

### Pattern C: Database Tests
**Files**: ~4 files
**Complexity**: Medium-High
**Example**: pkg/notes/*.go

```go
// Before
func TestDB(t *testing.T) {
    db := setupTestDB(t)
    t.Run("Insert", func(t *testing.T) {
        db.Create(&record)
    })
}

// After (use Level2, pass db, use tx)
import (
    "github.com/cyw0ng95/v2e/pkg/testutils"
    "gorm.io/gorm"
)

func TestDB(t *testing.T) {
    db := setupTestDB(t)
    testutils.Run(t, testutils.Level2, "Insert", db, func(t *testing.T, tx *gorm.DB) {
        tx.Create(&record)  // Use tx instead of db
    })
}
```

## Recommended Conversion Order

### Phase 1: Complete Database Tests (3 remaining)
High priority, enables full Level 2 testing.

- [ ] pkg/notes/service_status_test.go
- [ ] pkg/notes/bookmark_stats_test.go

Estimated: 1-2 hours

### Phase 2: Core Business Logic (30 files)
Medium priority, core functionality tests.

- [ ] pkg/cve/job/*_test.go (4 files)
- [ ] pkg/cve/session_edge_cases_test.go
- [ ] pkg/cwe/*_test.go (~5 files)
- [ ] pkg/capec/*_test.go (~5 files)
- [ ] pkg/attack/*_test.go (~5 files)
- [ ] pkg/ssg/*_test.go (~10 files)

Estimated: 3-5 hours

### Phase 3: Broker Tests (40 files)
Medium priority, infrastructure tests.

- [ ] cmd/v2broker/core/*_test.go (15 files)
- [ ] cmd/v2broker/routing/*_test.go (5 files)
- [ ] cmd/v2broker/transport/*_test.go (7 files)
- [ ] cmd/v2broker/mq/*_test.go (3 files)
- [ ] cmd/v2broker/perf/*_test.go (3 files)
- [ ] cmd/v2broker/*_test.go (7 files)

Estimated: 4-6 hours

### Phase 4: Support Packages (85 files)
Lower priority, utility tests.

- [ ] pkg/proc/*_test.go (~20 files)
- [ ] pkg/common/*_test.go (~10 files)
- [ ] pkg/rpc/*_test.go (~5 files)
- [ ] tool/*_test.go (~5 files)
- [ ] Other remaining files (~45 files)

Estimated: 6-8 hours

**Total Estimated Time**: 14-21 hours for complete conversion

## Automation Opportunities

### Semi-Automated Conversion
For Pattern A (simple tests), a script could:
1. Detect test functions without t.Run
2. Add testutils import
3. Wrap function body with testutils.Run
4. Detect gorm.DB to assign Level 2

### Manual Review Required
- Tests with complex setup/teardown
- Tests with shared state
- Tests with unusual patterns
- Database tests (tx vs db usage)

## Verification Process

After converting each batch:

```bash
# Test at Level 1 (logic only)
V2E_TEST_LEVEL=1 go test ./pkg/cve/...

# Test at Level 2 (logic + database)
V2E_TEST_LEVEL=2 go test ./pkg/notes/...

# Test all levels
V2E_TEST_LEVEL=3 go test ./...
```

## Benefits of Conversion

Once complete:

1. **Parallel Execution**: All tests run in parallel automatically
2. **Level Filtering**: Run only the tests you need during development
3. **Transaction Isolation**: Database tests never pollute each other
4. **Consistent Pattern**: All tests use same wrapper
5. **Cumulative Testing**: Higher levels include lower levels
6. **CI Efficiency**: Single job instead of 3 parallel matrix jobs

## Current Blockers

**None** - Infrastructure is complete and ready for conversion.

## Next Actions

1. **Immediate**: Convert remaining 3 database tests (Phase 1)
2. **Short-term**: Convert core business logic (Phase 2)
3. **Medium-term**: Convert broker tests (Phase 3)
4. **Long-term**: Convert remaining files (Phase 4)

Each phase can be done incrementally with verification between conversions.

---

**Note**: Conversion can be done gradually. The infrastructure supports both converted and unconverted tests coexisting. Tests without wrappers still run normally.
