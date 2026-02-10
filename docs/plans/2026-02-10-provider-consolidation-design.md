# Provider Consolidation Design

**Date:** 2026-02-10
**Status:** Approved

## Problem

7 provider implementations (asvs, attack, capec, cce, cve, cwe, ssg) contain ~500 lines of duplicated code:
- Same struct fields: `batchSize`, `maxRetries`, `retryDelay`, `mu sync.RWMutex`
- Same methods: `Initialize`, `Fetch`, `Store`, `Cleanup`, `GetStats`, `GetBatchSize`, `SetBatchSize`
- `BaseProviderFSM` already provides `Initialize`, `GetStats`, and has its own `mu`

## Solution

Embed common configuration in `BaseProviderFSM` and remove redundant code from concrete providers.

## Changes

### 1. BaseProviderFSM (pkg/meta/fsm/provider.go)

Add to `ProviderConfig`:
```go
type ProviderConfig struct {
    ID           string
    ProviderType string
    Storage      *storage.Store
    Executor     func() error

    // Common configuration
    BatchSize    int           // Default: 100
    MaxRetries   int           // Default: 3
    RetryDelay   time.Duration // Default: 5 * time.Second
}
```

Add to `BaseProviderFSM`:
```go
type BaseProviderFSM struct {
    // ... existing fields ...
    batchSize    int
    maxRetries   int
    retryDelay   time.Duration
}
```

Add methods:
- `GetBatchSize() int`
- `SetBatchSize(size int)`
- `GetMaxRetries() int`
- `SetMaxRetries(retries int)`
- `GetRetryDelay() time.Duration`
- `SetRetryDelay(delay time.Duration)`

### 2. Concrete Providers (7 files)

Each provider becomes ~60 lines instead of ~180:

| Provider | File | Lines Removed |
|----------|------|---------------|
| CVE | `pkg/cve/provider/cve_provider.go` | ~70 |
| CWE | `pkg/cwe/provider/cwe_provider.go` | ~80 |
| ATTACK | `pkg/attack/provider/attack_provider.go` | ~80 |
| CAPEC | `pkg/capec/provider/capec_provider.go` | ~70 |
| CCE | `pkg/cce/provider/cce_provider.go` | ~80 |
| ASVS | `pkg/asvs/provider/asvs_provider.go` | ~60 |
| SSG | `pkg/ssg/provider/git_provider.go` | ~70 |

**Total:** ~510 lines removed

### Removed from each provider:
- Struct fields: `batchSize`, `maxRetries`, `retryDelay`, `mu sync.RWMutex`
- Methods: `Initialize()`, `Fetch()`, `Store()`, `Cleanup()`, `GetStats()`, `GetBatchSize()`, `SetBatchSize()`

### Example: CWEProvider After Refactor

```go
type CWEProvider struct {
    *fsm.BaseProviderFSM
    localPath string
    rpcClient *rpc.Client
}

func NewCWEProvider(localPath string, store *storage.Store) (*CWEProvider, error) {
    provider := &CWEProvider{
        localPath: localPath,
        rpcClient: &rpc.Client{},
    }

    base, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
        ID:           "cwe",
        ProviderType: "cwe",
        Storage:      store,
        Executor:     provider.execute,
    })
    if err != nil {
        return nil, err
    }

    provider.BaseProviderFSM = base
    return provider, nil
}

func (p *CWEProvider) execute() error { /* existing logic */ }

func (p *CWEProvider) GetLocalPath() string { return p.localPath }
func (p *CWEProvider) SetLocalPath(path string) { p.localPath = path }
```

## Implementation

1. Update `pkg/meta/fsm/provider.go` with new fields and methods
2. Update all 7 provider files to remove redundant code
3. Run `./build.sh -t` to verify

## Commit

Single commit:
```
refactor(providers): consolidate common config in BaseProviderFSM

- Add BatchSize, MaxRetries, RetryDelay to ProviderConfig and BaseProviderFSM
- Remove ~500 lines of redundant code from 7 provider implementations
- CVE, CWE, ATTACK, CAPEC, CCE, ASVS, SSG providers simplified
```
