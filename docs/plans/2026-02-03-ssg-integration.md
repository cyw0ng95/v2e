# SSG (SCAP Security Guide) Integration Plan

**Date:** 2026-02-03  
**Author:** AI Assistant  
**Status:** Draft - Pending Review  
**Target Version:** v0.4.0+

## Executive Summary

This document outlines the plan to integrate SCAP Security Guide (SSG) support into the v2e system. SSG provides security compliance profiles and benchmarks in SCAP format, similar to how CVE, CWE, CAPEC, and ATT&CK data are currently managed.

The implementation will follow the existing broker-first architecture pattern, adding:
1. New pkg/ssg package for SSG data handling
2. Job control in meta service for SSG ETL
3. Remote fetching capability for SSG packages from GitHub
4. Local storage and RPC handlers in local service
5. Frontend UI components for SSG data display

## Requirements

### 1. Meta Service: Job Control for SSG
- Add SSG as a new data type to `RPCStartTypedSession`
- Support session control (start, stop, pause, resume) for SSG jobs
- Track SSG fetch/import progress similar to CVE/CWE/CAPEC/ATTACK
- Store SSG session state in BoltDB

### 2. Remote Service: Fetch SSG Package from GitHub
- Download tar.gz from: `https://github.com/ComplianceAsCode/content/releases/download/v0.1.79/scap-security-guide-0.1.79.tar.gz`
- Verify with sha512: `https://github.com/ComplianceAsCode/content/releases/download/v0.1.79/scap-security-guide-0.1.79.tar.gz.sha512`
- Add new RPC method: `RPCFetchSSGPackage`
- Support version parameter for future updates

### 3. Local Service: Receive and Deploy SSG Package
- Add new environment variable: `SSG_DOCPATH` for SSG data storage
- Default path: `ssg/` (parallel to database files)
- Extract tar.gz and deploy to `SSG_DOCPATH`
- Add RPC handler: `RPCDeploySSGPackage` to receive and extract package
- Preserve directory structure from tar.gz

### 4. pkg/ssg Package: XML Analysis and Data Structure
- Create new package: `pkg/ssg/`
- Parse SSG XML files (pattern: `ssg-*-ds.xml`)
- Design data structures:
  - SSGProfile: Security profiles (e.g., NIST 800-53, PCI-DSS)
  - SSGRule: Individual security rules
  - SSGBenchmark: Complete benchmark definitions
  - SSGReference: External references and mappings
- Database schema using SQLite (if performance requires)
  - Environment variable: `SSG_DB_PATH`
  - Default: `ssg.db`
- OR: Keep as file-based read-only storage (lighter approach)

### 5. Local Service: RPC Calls for SSG
Since SSG data is read-only (no CRUD), add RPC methods:
- `RPCListSSGProfiles`: List available security profiles
- `RPCGetSSGProfile`: Get detailed profile information by ID
- `RPCListSSGRules`: List security rules (with filtering)
- `RPCGetSSGRule`: Get detailed rule information by ID
- `RPCSearchSSGContent`: Search across SSG content
- `RPCGetSSGMetadata`: Get version and catalog metadata

### 6. Website: SSG Tab in Frontend
- Add SSG tab parallel to CVE/CWE/CAPEC/ATT&CK tabs
- Components to create:
  - `components/ssg-table.tsx`: Display SSG profiles/rules in table
  - `components/ssg-detail-dialog.tsx`: Show detailed rule/profile info
  - TypeScript types in `lib/types.ts` (camelCase, mirroring Go structs)
  - React Query hooks in `lib/hooks.ts`
- UI Features:
  - List security profiles with description
  - View individual rules with compliance mappings
  - Search/filter functionality
  - Link to source SCAP files

## Technical Architecture

### Data Flow

```
User (Frontend) → Access Service → Broker → Meta Service
                                              ↓
                                    Starts SSG Job Session
                                              ↓
                  Remote Service ← Broker ← Meta Service
                       ↓
              Fetches SSG tar.gz + sha512
                       ↓
                  Broker → Local Service
                              ↓
                    Extracts to SSG_DOCPATH
                              ↓
                    Parses XML via pkg/ssg
                              ↓
                    Stores in SSG_DB_PATH (optional)
                              ↓
                  Ready for queries via RPC
```

### File Structure

```
v2e/
├── pkg/ssg/                      # New SSG package
│   ├── constants.go              # SSG version, URLs, patterns
│   ├── types.go                  # SSG data structures
│   ├── parser.go                 # XML parsing logic
│   ├── local.go                  # Local storage interface
│   ├── models.go                 # GORM models (if using DB)
│   └── *_test.go                 # Unit tests
├── cmd/meta/
│   ├── main.go                   # Add SSG job handlers
│   └── service.md                # Update with SSG RPC methods
├── cmd/remote/
│   ├── main.go                   # Add SSG fetch handlers
│   └── service.md                # Update with SSG RPC methods
├── cmd/local/
│   ├── main.go                   # Add SSG storage/query handlers
│   └── service.md                # Update with SSG RPC methods
└── website/
    ├── components/
    │   ├── ssg-table.tsx         # New component
    │   └── ssg-detail-dialog.tsx # New component
    ├── lib/
    │   ├── types.ts              # Add SSG TypeScript types
    │   ├── hooks.ts              # Add SSG React Query hooks
    │   └── rpc-client.ts         # Add SSG RPC methods
    └── app/page.tsx              # Add SSG tab to UI

```

## Implementation Phases

### Phase 1: Core Infrastructure (pkg/ssg)
**Duration:** 1-2 days

- [ ] Create pkg/ssg package structure
- [ ] Define SSG data structures (types.go)
- [ ] Implement XML parser for ssg-*-ds.xml files
- [ ] Design storage approach (file-based vs. DB)
  - Decision: Start with file-based (lighter), add DB if performance needed
- [ ] Add constants for SSG GitHub URLs
- [ ] Write unit tests for parsing logic
- [ ] Update .gitignore if needed (exclude downloaded SSG packages)

**Files to Create:**
- `pkg/ssg/constants.go`
- `pkg/ssg/types.go`
- `pkg/ssg/parser.go`
- `pkg/ssg/local.go`
- `pkg/ssg/types_test.go`
- `pkg/ssg/parser_test.go`

**Configuration:**
- `SSG_DOCPATH` env var (default: "ssg/")
- `SSG_DB_PATH` env var (default: "ssg.db", optional)

### Phase 2: Remote Service Integration
**Duration:** 1 day

- [ ] Add RPCFetchSSGPackage handler to cmd/remote/main.go
- [ ] Implement SSG package download with sha512 verification
- [ ] Add version parameter support
- [ ] Handle HTTP errors and retries
- [ ] Update cmd/remote/service.md with new RPC method
- [ ] Write unit tests for SSG fetching

**RPC Method Specification:**
```
RPCFetchSSGPackage
- Description: Downloads SSG package from GitHub and verifies integrity
- Request Parameters:
  - version (string, optional): SSG version (default: "0.1.79")
- Response:
  - package_data (bytes): The tar.gz package data
  - sha512 (string): SHA512 checksum
  - verified (bool): Whether checksum verification passed
- Errors:
  - Download failed: Failed to download SSG package
  - Verification failed: SHA512 checksum mismatch
  - Network error: HTTP request failed
```

### Phase 3: Local Service Integration
**Duration:** 1-2 days

- [ ] Add SSG deployment handler to cmd/local/main.go
- [ ] Implement tar.gz extraction to SSG_DOCPATH
- [ ] Initialize SSG store at startup (load metadata)
- [ ] Add query handlers for SSG data
  - RPCListSSGProfiles
  - RPCGetSSGProfile
  - RPCListSSGRules
  - RPCGetSSGRule
  - RPCSearchSSGContent
  - RPCGetSSGMetadata
- [ ] Update cmd/local/service.md with SSG RPC methods
- [ ] Write integration tests

**RPC Methods Specification:**
```
1. RPCDeploySSGPackage
   - Receives tar.gz and extracts to SSG_DOCPATH

2. RPCListSSGProfiles
   - Lists available security profiles
   - Supports pagination (offset, limit)

3. RPCGetSSGProfile
   - Gets detailed profile by ID
   - Returns rules, references, metadata

4. RPCListSSGRules
   - Lists security rules
   - Supports filtering by severity, profile

5. RPCGetSSGRule
   - Gets detailed rule by ID
   - Returns remediation, checks, references

6. RPCGetSSGMetadata
   - Returns SSG version, catalog info
```

### Phase 4: Meta Service Job Control
**Duration:** 1 day

- [ ] Add "ssg" to data_type enum in RPCStartTypedSession
- [ ] Implement SSG job orchestration logic
- [ ] Add job state tracking for SSG sessions
- [ ] Test job control (start, stop, pause, resume)
- [ ] Update cmd/meta/service.md

**Job Flow:**
1. User calls RPCStartTypedSession with data_type="ssg"
2. Meta service calls remote service to fetch SSG package
3. Meta service calls local service to deploy package
4. Local service extracts and parses SSG data
5. Job completes, status updated

### Phase 5: Frontend Integration
**Duration:** 2 days

- [ ] Add SSG TypeScript types to website/lib/types.ts
- [ ] Add SSG RPC client methods to website/lib/rpc-client.ts
- [ ] Add SSG React Query hooks to website/lib/hooks.ts
- [ ] Create SSGTable component (website/components/ssg-table.tsx)
- [ ] Create SSGDetailDialog component
- [ ] Add SSG tab to main page (website/app/page.tsx)
- [ ] Add mock data for development (NEXT_PUBLIC_USE_MOCK_DATA=true)
- [ ] Test static build (npm run build)
- [ ] Manual testing with running backend

**UI Components:**
```typescript
// components/ssg-table.tsx
- Displays profiles and rules in tabular format
- Pagination and search support
- Click to view details

// components/ssg-detail-dialog.tsx
- Shows full profile/rule details
- Displays compliance mappings
- Links to original SCAP files

// UI Features:
- Tab: "SSG" (parallel to CVE, CWE, CAPEC, ATT&CK)
- Table columns: ID, Title, Severity, Profile
- Filter by: Profile, Severity
- Search: Full-text across title/description
```

### Phase 6: Testing and Documentation
**Duration:** 1 day

- [ ] Run unit tests: `./build.sh -t`
- [ ] Run integration tests: `./build.sh -i`
- [ ] Test with full system: `./build.sh -r`
- [ ] Verify SSG tab in frontend (screenshot)
- [ ] Update main README.md (add SSG to feature list)
- [ ] Code review with code_review tool
- [ ] Security scan with codeql_checker
- [ ] Performance benchmarks if needed

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| SSG_DOCPATH | `ssg/` | Directory for SSG XML files |
| SSG_DB_PATH | `ssg.db` | SQLite database for SSG (optional) |
| SSG_VERSION | `0.1.79` | Default SSG version to fetch |

## Database Schema (Optional)

If using SQLite for SSG storage:

```sql
CREATE TABLE ssg_profiles (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    version TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE ssg_rules (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT,
    severity TEXT,
    rationale TEXT,
    remediation TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

CREATE TABLE ssg_profile_rules (
    profile_id TEXT,
    rule_id TEXT,
    FOREIGN KEY (profile_id) REFERENCES ssg_profiles(id),
    FOREIGN KEY (rule_id) REFERENCES ssg_rules(id),
    PRIMARY KEY (profile_id, rule_id)
);

CREATE TABLE ssg_metadata (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at DATETIME
);
```

## Risk Assessment

### Risks and Mitigations

1. **XML Parsing Complexity**
   - Risk: SSG XML files may have complex nested structures
   - Mitigation: Start with basic fields, expand incrementally
   - Fallback: Use encoding/xml standard library, well-tested

2. **Large File Size**
   - Risk: SSG tar.gz may be large (>100MB)
   - Mitigation: Stream download, progress tracking
   - Add timeout configuration

3. **Schema Changes**
   - Risk: SSG XML schema may change between versions
   - Mitigation: Version-specific parsers if needed
   - Log warnings for unknown fields

4. **Frontend Build Size**
   - Risk: Adding SSG components increases bundle size
   - Mitigation: Use dynamic imports (already in place)
   - Lazy load SSG components

5. **Integration Testing**
   - Risk: No remote API calls allowed in tests
   - Mitigation: Use fixture files from real SSG package
   - Mock remote service responses

## Success Criteria

- [ ] SSG package can be fetched and verified from GitHub
- [ ] SSG data is extracted and stored locally
- [ ] All RPC methods work correctly
- [ ] Frontend displays SSG profiles and rules
- [ ] Job control (start/stop/pause/resume) works
- [ ] All tests pass (unit, integration)
- [ ] No security vulnerabilities introduced
- [ ] Documentation updated
- [ ] Build succeeds: `./build.sh -p`
- [ ] System runs: `./build.sh -r`

## Timeline

- Phase 1 (pkg/ssg): 1-2 days
- Phase 2 (remote): 1 day
- Phase 3 (local): 1-2 days
- Phase 4 (meta): 1 day
- Phase 5 (frontend): 2 days
- Phase 6 (testing): 1 day

**Total Estimated Duration:** 7-9 days

## Next Steps

1. Review and approve this plan
2. Create GitHub issue for tracking
3. Begin Phase 1 implementation
4. Regular progress updates via report_progress
5. Conduct code reviews after each phase

## References

- SSG GitHub: https://github.com/ComplianceAsCode/content
- SSG Release: https://github.com/ComplianceAsCode/content/releases/tag/v0.1.79
- SCAP Specification: https://csrc.nist.gov/projects/security-content-automation-protocol
- Existing implementations: pkg/capec (XML parsing), pkg/attack (Excel parsing)

## Notes

- This follows the same pattern as CVE, CWE, CAPEC, and ATT&CK
- Broker-first architecture maintained
- Read-only data (no CRUD), simpler than CVE
- File-based storage is lighter than DB (start simple)
- Frontend uses existing patterns (table + dialog)
- All RPC communication routed through broker
