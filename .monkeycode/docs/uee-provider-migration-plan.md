# UEE Provider Migration Plan

## Overview

This document outlines the plan to migrate all data sources (CVE, CWE, CAPEC, ATT&CK, SSG, ASVS) to the Unified ETL Engine (UEE) Provider framework.

## Current State

### ✅ UEE Framework (Implemented)
- `pkg/meta/fsm/provider.go` - BaseProviderFSM implementation
- `pkg/meta/fsm/macro.go` - MacroFSM for managing multiple providers
- `pkg/meta/storage/` - BoltDB persistence for provider states
- Provider states: IDLE, ACQUIRING, RUNNING, WAITING_QUOTA, WAITING_BACKOFF, PAUSED, TERMINATED
- Event-driven architecture with state transitions
- Permit management system

### ❌ Data Sources (Not Integrated)
1. **pkg/cve** - Has independent `remote/` package with Fetcher, AdaptiveRetry
2. **pkg/cwe** - Has independent `job/` package with Controller
3. **pkg/capec** - No remote fetching, only local storage
4. **pkg/attack** - No remote fetching, only local storage
5. **pkg/ssg** - Has independent `remote/` package with GitClient
6. **pkg/asvs** - No remote fetching, only local storage

## Migration Strategy

### Phase 1: Provider Interface & Base Implementation (Priority 1)

#### Task 1.1: Create Data Source Provider Interface
**File**: `pkg/meta/provider/types.go` (new)

```go
package provider

import (
    "context"
    "github.com/cyw0ng95/v2e/pkg/meta/fsm"
)

// DataSourceProvider defines the interface for all data source providers
type DataSourceProvider interface {
    fsm.ProviderFSM
    
    // Initialize provider-specific resources
    Initialize(ctx context.Context) error
    
    // Fetch data from remote source
    Fetch(ctx context.Context) error
    
    // Process and store fetched data
    Store(ctx context.Context) error
    
    // Get current progress metrics
    GetProgress() *ProviderProgress
    
    // Get configuration
    GetConfig() *ProviderConfig
}

// ProviderProgress represents progress metrics
type ProviderProgress struct {
    Fetched      int64
    Stored       int64
    Failed       int64
    LastFetchAt  time.Time
    LastStoreAt  time.Time
}
```

**Estimated**: 150 lines

---

#### Task 1.2: Create Base Provider Implementation
**File**: `pkg/meta/provider/base_provider.go` (new)

```go
package provider

import (
    "context"
    "fmt"
    "sync"
    "time"
    
    "github.com/cyw0ng95/v2e/pkg/meta/fsm"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// BaseProvider implements DataSourceProvider with common functionality
type BaseProvider struct {
    *fsm.BaseProviderFSM
    
    config     *ProviderConfig
    progress   *ProviderProgress
    rateLimiter *RateLimiter
}

// ProviderConfig holds provider-specific configuration
type ProviderConfig struct {
    Name              string
    DataType          string
    BaseURL           string
    APIKey           string
    BatchSize         int
    MaxRetries        int
    RetryDelay        time.Duration
    RateLimitPermits  int
}

func NewBaseProvider(config *ProviderConfig, storage *storage.Store) (*BaseProvider, error) {
    baseFSM, err := fsm.NewBaseProviderFSM(fsm.ProviderConfig{
        ID:           config.DataType,
        ProviderType: config.DataType,
        Storage:      storage,
    })
    if err != nil {
        return nil, err
    }
    
    return &BaseProvider{
        BaseProviderFSM: baseFSM,
        config:         config,
        progress:       &ProviderProgress{},
        rateLimiter:   NewRateLimiter(config.RateLimitPermits),
    }, nil
}

// Implement DataSourceProvider interface methods...
```

**Estimated**: 300 lines

---

### Phase 2: CVE Provider Implementation (Priority 1)

#### Task 2.1: Create CVE Provider
**File**: `pkg/cve/provider/cve_provider.go` (new)

```go
package provider

import (
    "context"
    "fmt"
    
    "github.com/cyw0ng95/v2e/pkg/cve/remote"
    "github.com/cyw0ng95/v2e/pkg/meta/provider"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// CVEProvider implements DataSourceProvider for CVE data
type CVEProvider struct {
    *provider.BaseProvider
    fetcher *remote.Fetcher
}

func NewCVEProvider(storage *storage.Store, apiKey string) (*CVEProvider, error) {
    base, err := provider.NewBaseProvider(&provider.ProviderConfig{
        Name:             "CVE",
        DataType:          "CVE",
        BaseURL:          "https://services.nvd.nist.gov/rest/json/cves/2.0",
        APIKey:           apiKey,
        BatchSize:        2000,
        MaxRetries:       3,
        RetryDelay:        5 * time.Second,
        RateLimitPermits:  10,
    }, storage)
    if err != nil {
        return nil, err
    }
    
    fetcher, err := remote.NewFetcher(apiKey)
    if err != nil {
        return nil, err
    }
    
    return &CVEProvider{
        BaseProvider: base,
        fetcher:      fetcher,
    }, nil
}

func (p *CVEProvider) Fetch(ctx context.Context) error {
    startIndex := 0
    pageSize := p.config.BatchSize
    
    for {
        resp, err := p.fetcher.FetchCVEs(startIndex, pageSize)
        if err != nil {
            if err == remote.ErrRateLimited {
                // Transition to WAITING_QUOTA
                if err := p.OnQuotaRevoked(10); err != nil {
                    return err
                }
                time.Sleep(30 * time.Second)
                continue
            }
            return err
        }
        
        // Store fetched CVEs in buffer for later processing
        p.progress.Fetched += int64(len(resp.Vulnerabilities))
        startIndex += len(resp.Vulnerabilities)
        
        // Check for completion
        if startIndex >= resp.TotalResults {
            break
        }
        
        // Rate limiting
        if err := p.rateLimiter.Wait(ctx); err != nil {
            return err
        }
    }
    
    return nil
}

func (p *CVEProvider) Store(ctx context.Context) error {
    // Implement batch storage using existing local RPC
    // ...
    return nil
}
```

**Estimated**: 400 lines

---

#### Task 2.2: Integrate CVE Provider into Meta Service
**File**: `cmd/v2meta/provider_registry.go` (new)

```go
package main

import (
    "github.com/cyw0ng95/v2e/pkg/cve/provider"
    "github.com/cyw0ng95/v2e/pkg/meta/fsm"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// ProviderRegistry manages all data source providers
type ProviderRegistry struct {
    providers map[string]fsm.ProviderFSM
    macroFSM  *fsm.MacroFSM
    storage   *storage.Store
}

func NewProviderRegistry(storage *storage.Store) (*ProviderRegistry, error) {
    // Create CVE provider
    cveProvider, err := provider.NewCVEProvider(storage, getAPIKey("CVE_API_KEY"))
    if err != nil {
        return nil, err
    }
    
    macroFSM, err := fsm.NewMacroFSM(storage)
    if err != nil {
        return nil, err
    }
    
    // Register providers with MacroFSM
    macroFSM.RegisterProvider(cveProvider)
    
    return &ProviderRegistry{
        providers: map[string]fsm.ProviderFSM{
            "CVE": cveProvider,
        },
        macroFSM:  macroFSM,
        storage:    storage,
    }, nil
}

func (r *ProviderRegistry) StartProvider(providerID string) error {
    provider, ok := r.providers[providerID]
    if !ok {
        return fmt.Errorf("provider not found: %s", providerID)
    }
    
    return provider.Start()
}
```

**Estimated**: 250 lines

---

### Phase 3: CWE Provider Implementation (Priority 1)

#### Task 3.1: Create CWE Provider
**File**: `pkg/cwe/provider/cwe_provider.go` (new)

```go
package provider

import (
    "context"
    "os"
    
    "github.com/cyw0ng95/v2e/pkg/cwe"
    "github.com/cyw0ng95/v2e/pkg/meta/provider"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// CWEProvider implements DataSourceProvider for CWE data
type CWEProvider struct {
    *provider.BaseProvider
    localPath string
}

func NewCWEProvider(storage *storage.Store, localPath string) (*CWEProvider, error) {
    if localPath == "" {
        localPath = "assets/cwe-raw.json" // default
    }
    
    base, err := provider.NewBaseProvider(&provider.ProviderConfig{
        Name:             "CWE",
        DataType:          "CWE",
        BatchSize:        100,
        MaxRetries:       3,
        RetryDelay:        5 * time.Second,
        RateLimitPermits:  10,
    }, storage)
    if err != nil {
        return nil, err
    }
    
    return &CWEProvider{
        BaseProvider: base,
        localPath:    localPath,
    }, nil
}

func (p *CWEProvider) Fetch(ctx context.Context) error {
    // Read CWE data from local file
    data, err := os.ReadFile(p.localPath)
    if err != nil {
        return err
    }
    
    // Parse and count items
    var cweData []cwe.CWE
    if err := json.Unmarshal(data, &cweData); err != nil {
        return err
    }
    
    p.progress.Fetched = int64(len(cweData))
    return nil
}

func (p *CWEProvider) Store(ctx context.Context) error {
    // Use existing local RPC to store CWE views
    // ...
    return nil
}
```

**Estimated**: 300 lines

---

### Phase 4: CAPEC Provider Implementation (Priority 2)

#### Task 4.1: Create CAPEC Provider
**File**: `pkg/capec/provider/capec_provider.go` (new)

```go
package provider

import (
    "context"
    "fmt"
    "os"
    
    "github.com/cyw0ng95/v2e/pkg/capec"
    "github.com/cyw0ng95/v2e/pkg/meta/provider"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// CAPECProvider implements DataSourceProvider for CAPEC data
type CAPECProvider struct {
    *provider.BaseProvider
    localPath string
}

func NewCAPECProvider(storage *storage.Store, localPath string) (*CAPECProvider, error) {
    if localPath == "" {
        localPath = "assets/capec_contents_latest.xml" // default
    }
    
    base, err := provider.NewBaseProvider(&provider.ProviderConfig{
        Name:             "CAPEC",
        DataType:          "CAPEC",
        BatchSize:        50,
        MaxRetries:       3,
        RetryDelay:        5 * time.Second,
        RateLimitPermits:  10,
    }, storage)
    if err != nil {
        return nil, err
    }
    
    return &CAPECProvider{
        BaseProvider: base,
        localPath:    localPath,
    }, nil
}

func (p *CAPECProvider) Fetch(ctx context.Context) error {
    // Parse CAPEC XML file
    file, err := os.Open(p.localPath)
    if err != nil {
        return err
    }
    defer file.Close()
    
    // Use existing CAPEC parsing logic
    // ...
    return nil
}

func (p *CAPECProvider) Store(ctx context.Context) error {
    // Use existing local RPC to store CAPEC
    // ...
    return nil
}
```

**Estimated**: 350 lines

---

### Phase 5: ATT&CK Provider Implementation (Priority 2)

#### Task 5.1: Create ATT&CK Provider
**File**: `pkg/attack/provider/attack_provider.go` (new)

```go
package provider

import (
    "context"
    "os"
    
    "github.com/cyw0ng95/v2e/pkg/attack"
    "github.com/cyw0ng95/v2e/pkg/meta/provider"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// ATTACKProvider implements DataSourceProvider for ATT&CK data
type ATTACKProvider struct {
    *provider.BaseProvider
    localPath string
}

func NewATTACKProvider(storage *storage.Store, localPath string) (*ATTACKProvider, error) {
    if localPath == "" {
        localPath = "assets/enterprise-attack.xlsx" // default
    }
    
    base, err := provider.NewBaseProvider(&provider.ProviderConfig{
        Name:             "ATT&CK",
        DataType:          "ATTACK",
        BatchSize:        100,
        MaxRetries:       3,
        RetryDelay:        5 * time.Second,
        RateLimitPermits:  10,
    }, storage)
    if err != nil {
        return nil, err
    }
    
    return &ATTACKProvider{
        BaseProvider: base,
        localPath:    localPath,
    }, nil
}

func (p *ATTACKProvider) Fetch(ctx context.Context) error {
    // Parse ATT&CK Excel file
    // Use existing attack parsing logic
    // ...
    return nil
}

func (p *ATTACKProvider) Store(ctx context.Context) error {
    // Use existing local RPC to store ATT&CK
    // ...
    return nil
}
```

**Estimated**: 350 lines

---

### Phase 6: SSG Provider Implementation (Priority 2)

#### Task 6.1: Create SSG Git Provider
**File**: `pkg/ssg/provider/git_provider.go` (new)

```go
package provider

import (
    "context"
    
    "github.com/cyw0ng95/v2e/pkg/ssg/remote"
    "github.com/cyw0ng95/v2e/pkg/meta/provider"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// SSGProvider implements DataSourceProvider for SSG data
type SSGProvider struct {
    *provider.BaseProvider
    gitClient *remote.GitClient
}

func NewSSGProvider(storage *storage.Store, repoURL string) (*SSGProvider, error) {
    if repoURL == "" {
        repoURL = "https://github.com/OWASP/wg-ssg" // default
    }
    
    base, err := provider.NewBaseProvider(&provider.ProviderConfig{
        Name:             "SSG",
        DataType:          "SSG",
        BatchSize:        10,
        MaxRetries:       3,
        RetryDelay:        10 * time.Second,
        RateLimitPermits:  5,
    }, storage)
    if err != nil {
        return nil, err
    }
    
    gitClient := remote.NewGitClient(repoURL, "")
    
    return &SSGProvider{
        BaseProvider: base,
        gitClient:    gitClient,
    }, nil
}

func (p *SSGProvider) Fetch(ctx context.Context) error {
    // Clone or pull Git repository
    if err := p.gitClient.Pull(); err != nil {
        return err
    }
    
    // List data files
    guideFiles, err := p.gitClient.ListGuideFiles()
    if err != nil {
        return err
    }
    
    p.progress.Fetched = int64(len(guideFiles))
    return nil
}

func (p *SSGProvider) Store(ctx context.Context) error {
    // Use existing SSG importer to parse and store
    // ...
    return nil
}
```

**Estimated**: 350 lines

---

### Phase 7: ASVS Provider Implementation (Priority 3)

#### Task 7.1: Create ASVS Provider
**File**: `pkg/asvs/provider/asvs_provider.go` (new)

```go
package provider

import (
    "context"
    "os"
    
    "github.com/cyw0ng95/v2e/pkg/asvs"
    "github.com/cyw0ng95/v2e/pkg/meta/provider"
    "github.com/cyw0ng95/v2e/pkg/meta/storage"
)

// ASVSProvider implements DataSourceProvider for ASVS data
type ASVSProvider struct {
    *provider.BaseProvider
    csvURL   string
    localPath string
}

func NewASVSProvider(storage *storage.Store, csvURL string) (*ASVSProvider, error) {
    if csvURL == "" {
        csvURL = "https://raw.githubusercontent.com/OWASP/ASVS/v5.0.0/5.0/docs_en/OWASP_Application_Security_Verification_Standard_5.0.0_en.csv"
    }
    
    base, err := provider.NewBaseProvider(&provider.ProviderConfig{
        Name:             "ASVS",
        DataType:          "ASVS",
        BatchSize:        100,
        MaxRetries:       3,
        RetryDelay:        5 * time.Second,
        RateLimitPermits:  10,
    }, storage)
    if err != nil {
        return nil, err
    }
    
    return &ASVSProvider{
        BaseProvider: base,
        csvURL:       csvURL,
    }, nil
}

func (p *ASVSProvider) Fetch(ctx context.Context) error {
    // Download CSV file
    // Use existing ASVS CSV parsing logic
    // ...
    return nil
}

func (p *ASVSProvider) Store(ctx context.Context) error {
    // Use existing local RPC to store ASVS
    // ...
    return nil
}
```

**Estimated**: 300 lines

---

### Phase 8: Meta Service Integration (Priority 1)

#### Task 8.1: Update Meta Service to Use Provider Registry
**File**: `cmd/v2meta/main.go` (modify)

**Changes**:
1. Replace `DataPopulationController` with `ProviderRegistry`
2. Remove individual import controllers for CVE, CWE, CAPEC, ATT&CK
3. Use `ProviderRegistry.StartProvider()` instead
4. Add RPC handlers for provider lifecycle:
   - `RPCStartProvider` - Already exists, update to use registry
   - `RPCPauseProvider` - Already exists, update to use registry
   - `RPCStopProvider` - Already exists, update to use registry
   - `RPCGetProviderStatus` - New handler for provider status
   - `RPCListProviders` - New handler for listing all providers

**Estimated**: 150 lines modified

---

#### Task 8.2: Add Provider Status RPC Handler
**File**: `cmd/v2meta/handlers.go` (new)

```go
package main

// RPCGetProviderStatus returns the status of a specific provider
func createGetProviderStatusHandler(registry *ProviderRegistry, logger *common.Logger) subprocess.Handler {
    return func(params json.RawMessage) (interface{}, error) {
        var req struct {
            ProviderID string `json:"providerId"`
        }
        
        if err := json.Unmarshal(params, &req); err != nil {
            return nil, err
        }
        
        provider := registry.GetProvider(req.ProviderID)
        if provider == nil {
            return nil, fmt.Errorf("provider not found: %s", req.ProviderID)
        }
        
        return map[string]interface{}{
            "id":       provider.GetID(),
            "type":     provider.GetType(),
            "state":    provider.GetState(),
            "progress": provider.GetProgress(),
        }, nil
    }
}

// RPCListProviders returns all registered providers and their statuses
func createListProvidersHandler(registry *ProviderRegistry, logger *common.Logger) subprocess.Handler {
    return func(params json.RawMessage) (interface{}, error) {
        providers := registry.GetAllProviders()
        
        var result []map[string]interface{}
        for _, p := range providers {
            result = append(result, map[string]interface{}{
                "id":       p.GetID(),
                "type":     p.GetType(),
                "state":    p.GetState(),
                "progress": p.GetProgress(),
            })
        }
        
        return map[string]interface{}{
            "providers": result,
        }, nil
    }
}
```

**Estimated**: 150 lines

---

### Phase 9: Testing & Documentation (Priority 1)

#### Task 9.1: Add Provider Tests
**Files**: 
- `pkg/meta/provider/base_provider_test.go`
- `pkg/cve/provider/cve_provider_test.go`
- `pkg/cwe/provider/cwe_provider_test.go`

**Tests to implement**:
1. Provider state transitions
2. Rate limiting behavior
3. Error handling and retries
4. Progress tracking
5. Concurrent access

**Estimated**: 500 lines

---

#### Task 9.2: Update Documentation
**Files**: 
- `cmd/v2meta/service.md` - Update RPC API documentation
- `README.md` - Add Provider architecture section
- `pkg/meta/fsm/README.md` - Provider FSM documentation

**Content to add**:
1. Provider lifecycle
2. State machine transitions
3. Error handling strategies
4. Configuration examples
5. Migration guide for existing code

**Estimated**: 300 lines

---

## Migration Timeline

### Phase 1: Foundation (Week 1)
- Task 1.1: Create Data Source Provider Interface
- Task 1.2: Create Base Provider Implementation
- Task 8.1: Update Meta Service to Use Provider Registry

### Phase 2: CVE Provider (Week 2)
- Task 2.1: Create CVE Provider
- Task 9.1: Add CVE Provider Tests

### Phase 3: CWE Provider (Week 3)
- Task 3.1: Create CWE Provider
- Task 9.1: Add CWE Provider Tests

### Phase 4: CAPEC Provider (Week 4)
- Task 4.1: Create CAPEC Provider
- Task 9.1: Add CAPEC Provider Tests

### Phase 5: ATT&CK Provider (Week 5)
- Task 5.1: Create ATT&CK Provider
- Task 9.1: Add ATT&CK Provider Tests

### Phase 6: SSG Provider (Week 6)
- Task 6.1: Create SSG Git Provider
- Task 9.1: Add SSG Provider Tests

### Phase 7: ASVS Provider (Week 7)
- Task 7.1: Create ASVS Provider
- Task 9.1: Add ASVS Provider Tests

### Phase 8: Integration & Testing (Week 8)
- Task 8.2: Add Provider Status RPC Handler
- Task 9.1: Full integration tests
- Task 9.2: Update Documentation

---

## Risk Assessment

### High Risk
1. **Breaking Changes**: Existing taskflow-based job management will be replaced
   - Mitigation: Keep old RPC handlers for backward compatibility during migration
   - Rollback plan: Maintain feature flag to switch between old and new systems

2. **State Persistence**: BoltDB state management must be robust
   - Mitigation: Add comprehensive tests for state recovery
   - Rollback plan: Provide state reset functionality

### Medium Risk
3. **Rate Limiting**: Multiple providers may compete for permits
   - Mitigation: Implement per-provider permit allocation
   - Monitoring: Add metrics for permit usage

4. **Error Recovery**: Providers must handle network errors gracefully
   - Mitigation: Use existing retry logic (AdaptiveRetry)
   - Testing: Simulate network failures in tests

### Low Risk
5. **Performance**: Provider framework may add overhead
   - Mitigation: Benchmark and optimize hot paths
   - Monitoring: Track FSM transition latency

---

## Success Criteria

1. ✅ All data sources implement `DataSourceProvider` interface
2. ✅ All providers registered with `MacroFSM`
3. ✅ Provider lifecycle (start/pause/stop/resume) works correctly
4. ✅ State transitions follow UEE FSM rules
5. ✅ Rate limiting and permit management functional
6. ✅ Progress tracking accurate for all providers
7. ✅ Error handling and retry logic robust
8. ✅ Integration tests pass for all providers
9. ✅ Documentation updated and complete
10. ✅ No performance regression (maintain current throughput)

---

## Rollback Plan

If migration fails at any phase:

1. **Immediate**: Revert to previous commit before migration
2. **Service**: Keep existing `DataPopulationController` active
3. **Data**: No data loss - providers only fetch/store, don't modify existing data
4. **Communication**: Document rollback reason and next steps

---

## Next Steps

1. **Phase 1 Start**: Implement `DataSourceProvider` interface and `BaseProvider`
2. **CVE Integration**: Migrate CVE provider first (highest priority)
3. **Testing**: Add comprehensive tests for each provider
4. **Documentation**: Update service.md and README as we go
5. **Monitoring**: Add provider-specific metrics and logging

---

## References

- UEE Architecture: `pkg/meta/fsm/README.md`
- Provider FSM: `pkg/meta/fsm/provider.go`
- Macro FSM: `pkg/meta/fsm/macro.go`
- Storage: `pkg/meta/storage/storage.go`
- CVE Fetcher: `pkg/cve/remote/fetcher.go`
- CWE Controller: `pkg/cwe/job/controller.go`
- SSG Git Client: `pkg/ssg/remote/git.go`
