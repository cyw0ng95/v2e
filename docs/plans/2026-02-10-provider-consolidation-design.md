# Provider Consolidation Design

**Date:** 2026-02-10
**Status:** Approved
**Updated:** 2026-02-10 (Verified current state)

## Problem

4 of 7 provider implementations (CAPEC, CCE, ASVS, SSG) still contain duplicated code that already exists in `BaseProviderFSM`:
- Same struct fields: `batchSize`, `maxRetries`, `retryDelay`, `mu sync.RWMutex`
- Same methods: `Initialize`, `Fetch`, `Store`, `Cleanup`, `GetStats`, `GetBatchSize`, `SetBatchSize`
- `BaseProviderFSM` already provides all these with proper locking

**Already Consolidated (3 providers):**
- CVE Provider (127 lines) - Clean
- CWE Provider (129 lines) - Clean
- ATTACK Provider (120 lines) - Clean

## Solution

Remove redundant code from the 4 providers that still have duplication. Use `BaseProviderFSM` methods directly.

## Current State Analysis

### BaseProviderFSM (pkg/meta/fsm/provider.go)

Already provides:
- Fields: `batchSize`, `maxRetries`, `retryDelay` (with defaults: 100, 3, 5s)
- `Initialize(ctx context.Context) error` - no-op base implementation
- `GetBatchSize() int` / `SetBatchSize(size int)`
- `GetMaxRetries() int` / `SetMaxRetries(retries int)`
- `GetRetryDelay() time.Duration` / `SetRetryDelay(delay time.Duration)`
- `GetStats() map[string]interface{}` - returns id, state, counts, etc.
- `Execute() error` - calls the executor function

### Providers Needing Cleanup

| Provider | File | Current Lines | Lines to Remove | Final Lines |
|----------|------|---------------|-----------------|-------------|
| CAPEC | `pkg/capec/provider/capec_provider.go` | 179 | ~70 | ~109 |
| CCE | `pkg/cce/provider/cce_provider.go` | 180 | ~75 | ~105 |
| ASVS | `pkg/asvs/provider/asvs_provider.go` | 148 | ~55 | ~93 |
| SSG | `pkg/ssg/provider/git_provider.go` | 150 | ~60 | ~90 |

**Total:** ~260 lines removed

## Detailed Changes per Provider

### 1. CAPEC Provider (pkg/capec/provider/capec_provider.go)

**Remove struct fields (lines 22-25):**
```go
// DELETE:
batchSize  int
maxRetries int
retryDelay time.Duration
mu         sync.RWMutex
```

**Update NewCAPECProvider to pass config:**
```go
base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
    ID:           "capec",
    ProviderType: "capec",
    Storage:      store,
    Executor:     provider.execute,
    BatchSize:    50,  // CAPEC-specific default
})
```

**Remove/update methods:**
- Delete `Initialize()` (lines 57-60) - use BaseProviderFSM's
- Delete `Fetch()` (lines 161-164) - just calls Execute()
- Delete `Store()` (lines 166-169) - just calls Execute()
- Delete `Cleanup()` (lines 156-159) - empty implementation
- Delete `GetStats()` (lines 171-178) - use BaseProviderFSM's
- Delete `GetBatchSize()` (lines 142-147) - use BaseProviderFSM's
- Delete `SetBatchSize()` (lines 149-154) - use BaseProviderFSM's
- Update `GetLocalPath()`/`SetLocalPath()` to not use `mu` (localPath is provider-specific, not shared)

**Update execute() to use BaseProviderFSM.GetBatchSize():**
```go
batchSize := p.GetBatchSize()  // instead of p.mu.RLock(); batchSize := p.batchSize; p.mu.RUnlock()
```

### 2. CCE Provider (pkg/cce/provider/cce_provider.go)

**Remove struct fields (lines 21-25):**
```go
// DELETE:
batchSize  int
maxRetries int
retryDelay time.Duration
mu         sync.RWMutex
```

**Update NewCCEProvider to pass config:**
```go
base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
    ID:           "cce",
    ProviderType: "cce",
    Storage:      store,
    Executor:     provider.execute,
    BatchSize:    100,  // CCE-specific default
})
```

**Remove methods:**
- `Initialize()` (lines 57-60)
- `Fetch()` (lines 129-132)
- `Store()` (lines 134-137)
- `Cleanup()` (lines 148-151)
- `GetStats()` (lines 139-146)
- `GetBatchSize()` (lines 174-179)
- `SetBatchSize()` (lines 153-158)

**Keep with simplified locking:**
- `GetLocalPath()`/`SetLocalPath()` - localPath is provider-specific

### 3. ASVS Provider (pkg/asvs/provider/asvs_provider.go)

**Remove struct fields (lines 19-22):**
```go
// DELETE:
batchSize  int
maxRetries int
retryDelay time.Duration
mu         sync.RWMutex
```

**Update NewASVSProvider:**
```go
base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
    ID:           "asvs",
    ProviderType: "asvs",
    Storage:      store,
    Executor:     provider.execute,
    BatchSize:    100,
})
```

**Remove methods:**
- `Initialize()` (lines 54-57)
- `Cleanup()` (lines 96-99)
- `Fetch()` (lines 101-104)
- `Store()` (lines 106-109)
- `GetStats()` (lines 111-119) - but merge `csv_url` into a custom GetStats if needed
- `GetBatchSize()` (lines 128-133)
- `SetBatchSize()` (lines 121-126)

**Keep with simplified locking:**
- `GetCSVURL()`/`SetCSVURL()` - csvURL is provider-specific

### 4. SSG Provider (pkg/ssg/provider/git_provider.go)

**Remove struct fields (lines 18-21):**
```go
// DELETE:
batchSize  int
maxRetries int
retryDelay time.Duration
mu         sync.RWMutex
```

**Update NewSSGProvider:**
```go
base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
    ID:           "ssg",
    ProviderType: "ssg",
    Storage:      store,
    Executor:     provider.execute,
    BatchSize:    10,   // SSG-specific default (smaller batches)
    RetryDelay:   10 * time.Second,  // SSG-specific retry delay
})
```

**Remove methods:**
- `Initialize()` (lines 55-58)
- `Cleanup()` (lines 96-99)
- `Fetch()` (lines 101-104)
- `Store()` (lines 106-109)
- `GetStats()` (lines 111-119) - but merge `repo_url` into custom GetStats if needed
- `GetBatchSize()` (lines 128-133)
- `SetBatchSize()` (lines 121-126)

**Keep with simplified locking:**
- `GetRepoURL()`/`SetRepoURL()` - repoURL is provider-specific

## Example: CAPEC Provider After Refactor

```go
package provider

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/meta/fsm"
	"github.com/cyw0ng95/v2e/pkg/meta/storage"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

// CAPECProvider implements ProviderFSM for CAPEC data
type CAPECProvider struct {
	*fsm.BaseProviderFSM
	localPath string
}

// NewCAPECProvider creates a new CAPEC provider with FSM support
func NewCAPECProvider(localPath string, store *storage.Store) (*CAPECProvider, error) {
	if localPath == "" {
		localPath = "assets/capec_contents_latest.xml"
	}

	provider := &CAPECProvider{
		localPath: localPath,
	}

	base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
		ID:           "capec",
		ProviderType: "capec",
		Storage:      store,
		Executor:     provider.execute,
		BatchSize:    50,
	})
	if err != nil {
		return nil, err
	}

	provider.BaseProviderFSM = base
	return provider, nil
}

// execute performs CAPEC fetch and store operations
func (p *CAPECProvider) execute() error {
	if p.GetState() != fsm.ProviderRunning {
		return fmt.Errorf("cannot execute in state %s, must be RUNNING", p.GetState())
	}

	data, err := os.ReadFile(p.localPath)
	if err != nil {
		return fmt.Errorf("failed to read CAPEC file: %w", err)
	}

	capecData := capec.Root{}
	if err := xml.Unmarshal(data, &capecData); err != nil {
		return fmt.Errorf("failed to unmarshal CAPEC data: %w", err)
	}

	attackPatterns := capecData.AttackPatterns.AttackPattern
	batchSize := p.GetBatchSize()

	for i, attackPattern := range attackPatterns {
		if p.GetState() != fsm.ProviderRunning {
			break
		}

		itemURN := urn.MustParse(fmt.Sprintf("v2e::mitre::capec::%d", attackPattern.ID))

		_, err := json.Marshal(attackPattern)
		if err != nil {
			return fmt.Errorf("failed to marshal CAPEC item: %w", err)
		}

		if i > 0 && i%batchSize == 0 {
			time.Sleep(1 * time.Second)
		}

		if err := p.SaveCheckpoint(itemURN, true, ""); err != nil {
			return fmt.Errorf("failed to save checkpoint for %d: %w", attackPattern.ID, err)
		}
	}

	return nil
}

// GetLocalPath returns the local file path
func (p *CAPECProvider) GetLocalPath() string {
	return p.localPath
}

// SetLocalPath sets the local file path
func (p *CAPECProvider) SetLocalPath(path string) {
	p.localPath = path
}
```

## Implementation Steps

1. **CAPEC** - Remove fields and methods, pass BatchSize in config
2. **CCE** - Remove fields and methods, pass BatchSize in config
3. **ASVS** - Remove fields and methods, pass BatchSize in config
4. **SSG** - Remove fields and methods, pass BatchSize and RetryDelay in config
5. Run `./build.sh -t` to verify

## Verification

After changes:
```bash
./build.sh -t  # All tests pass
```

## Commit

Single commit:
```
refactor(providers): remove redundant code from CAPEC, CCE, ASVS, SSG

- Remove duplicated batchSize, maxRetries, retryDelay, mu fields
- Remove redundant GetBatchSize/SetBatchSize methods (use BaseProviderFSM)
- Remove redundant Initialize, Fetch, Store, Cleanup methods
- Pass provider-specific config via ProviderConfig
- ~260 lines removed

CVE, CWE, and ATTACK providers already use BaseProviderFSM correctly.
```
