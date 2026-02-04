# SSG Manifests and Data Streams Integration Plan

## Overview

Extend SSG support to include:
1. **Manifests** (JSON files) - Profile and rule metadata
2. **Data Streams** (*-ds.xml files) - Comprehensive SCAP XML with XCCDF, OVAL, OCIL

## Current State

✅ **Implemented:**
- Guides (HTML) - Hierarchical XCCDF from HTML
- Tables (HTML) - Flat mapping tables (CCE, NIST, STIG)

⚠️ **To Implement:**
- Manifests (JSON) - Profile/rule metadata
- Data Streams (XML) - SCAP data streams

## Data Analysis

### 1. Manifest Structure (JSON)

**Location:** `assets/ssg-static/manifests/manifest-{product}.json`

**Structure:**
```json
{
  "product_name": "al2023",
  "rules": {},
  "profiles": {
    "cis": {
      "rules": ["account_disable_post_pw_expiration", "aide_build_database", ...]
    },
    "anssi_bp28_enhanced": {
      "rules": [...]
    }
  }
}
```

**Key Data:**
- Product name (al2023, rhel8, etc.)
- Profiles (CIS, ANSSI, STIG, etc.)
- Rules list per profile
- Empty `rules` object (potentially for future rule metadata)

**Use Cases:**
- List all profiles for a product
- Get rules included in a specific profile
- Cross-reference: Profile → Rules → Guides/Tables

### 2. Data Stream Structure (XML)

**Location:** `assets/ssg-static/ssg-{product}-ds.xml`

**XML Namespaces:**
```xml
xmlns:ds="http://scap.nist.gov/schema/scap/source/1.2"
xmlns:xccdf-1.2="http://checklists.nist.gov/xccdf/1.2"
xmlns:oval-def="http://oval.mitre.org/XMLSchema/oval-definitions-5"
xmlns:ocil="http://scap.nist.gov/schema/ocil/2.0"
```

**Components:**
1. **Data Stream Collection** (`ds:data-stream-collection`)
   - Contains multiple data streams
   - Timestamp and version info

2. **XCCDF Benchmark** (embedded in `ds:component`)
   - Complete benchmark definition
   - Groups and rules (similar to HTML guides but more detailed)
   - Profiles with rule selections
   - Check references (OVAL, OCIL, SCE)

3. **OVAL Definitions** (embedded in `ds:component`)
   - System checks and tests
   - Platform definitions
   - Criteria and criteria operators

4. **OCIL Questionnaires** (embedded in `ds:component`)
   - Interactive compliance checks
   - Questions and answer choices

5. **CPE Dictionary** (embedded in `ds:component`)
   - Platform identifiers
   - System detection rules

**Key Data to Extract:**
- Benchmark metadata (ID, title, version)
- Profiles (ID, title, description, extends)
- Groups (hierarchical structure)
- Rules (ID, title, severity, references, checks)
- Rule → OVAL/OCIL references (for cross-linking)

**Use Cases:**
- Full XCCDF benchmark parsing
- Profile definitions with rule selections
- Check references (what OVAL/OCIL checks apply to each rule)
- Platform applicability (CPE)

## Cross-Reference Bridges

### Identified Conjunction Points

1. **Rule ID** (Primary Key)
   - Format: `xccdf_org.ssgproject.content_rule_{rulename}`
   - Present in: Guides (HTML), Data Streams (XCCDF), Manifests (profile rules)
   - **Bridge:** Rule ID → Guide Rules, DS Rules, Manifest Profile Rules

2. **Product Name** (Category)
   - Present in: All objects (guides, tables, manifests, data streams)
   - **Bridge:** Product → All related objects

3. **Profile ID** (Category)
   - Present in: Guides (HTML), Data Streams (XCCDF), Manifests
   - **Bridge:** Profile → Rules in that profile

4. **CCE Identifier** (External Reference)
   - Present in: Tables (mapping column), Data Streams (rule references)
   - **Bridge:** CCE → Rules that implement it

5. **OVAL/OCIL Check IDs** (Check References)
   - Present in: Data Streams only
   - **Bridge:** Rule → OVAL Definitions, OCIL Questionnaires

### Cross-Reference Schema

```
SSGCrossReference Table:
- ID (auto-increment)
- FromType (enum: guide, table, manifest, datastream)
- FromID (string: object ID)
- ToType (enum: guide, table, manifest, datastream, oval, ocil, cce)
- ToID (string: target ID)
- LinkType (enum: rule_id, product, profile, cce, oval_check, ocil_check)
- CreatedAt

Examples:
- (guide, "ssg-al2023-guide-cis", datastream, "ssg-al2023-ds", rule_id)
- (table, "table-al2023-cces", datastream_rule, "xccdf_..._rule_aide_build_database", cce)
- (manifest_profile, "al2023:cis", datastream_profile, "xccdf_..._profile_cis", profile)
```

## Implementation Plan

### Phase 1: Manifest Support

**Step 1.1: Data Models** (2-3 hours)
```go
// SSGManifest - Product manifest metadata
type SSGManifest struct {
    ID          string    `gorm:"primaryKey"` // e.g., "manifest-al2023"
    Product     string    `gorm:"index"`      // al2023, rhel8, etc.
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// SSGProfile - Profile definition from manifest
type SSGProfile struct {
    ID          string    `gorm:"primaryKey"` // e.g., "al2023:cis"
    ManifestID  string    `gorm:"index"`      // Foreign key to manifest
    Product     string    `gorm:"index"`      // Denormalized for queries
    ProfileID   string    `gorm:"index"`      // e.g., "cis"
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// SSGProfileRule - Many-to-many relationship: Profile → Rules
type SSGProfileRule struct {
    ID          uint      `gorm:"primaryKey"`
    ProfileID   string    `gorm:"index"`      // Foreign key to SSGProfile
    RuleShortID string    `gorm:"index"`      // e.g., "aide_build_database"
    CreatedAt   time.Time
}
```

**Files:**
- `pkg/ssg/models.go` - Add manifest models
- `pkg/ssg/local/store.go` - Add manifest CRUD operations

**Step 1.2: Parser** (3-4 hours)
```go
// ParseManifestFile parses a JSON manifest file
func ParseManifestFile(path string) (*ssg.SSGManifest, []ssg.SSGProfile, []ssg.SSGProfileRule, error)
```

**Files:**
- `pkg/ssg/parser/manifest.go` - JSON parsing
- `pkg/ssg/parser/manifest_test.go` - Unit tests

**Step 1.3: Remote Service** (1-2 hours)
- Add `ListManifestFiles()` to GitClient
- Add `RPCSSGListManifestFiles` handler

**Files:**
- `pkg/ssg/remote/git.go`
- `pkg/ssg/remote/handlers.go`

**Step 1.4: Local Service** (2-3 hours)
- RPC handlers: `RPCSSGImportManifest`, `RPCSSGListManifests`, `RPCSSGGetManifest`
- RPC handlers: `RPCSSGListProfiles`, `RPCSSGGetProfile`, `RPCSSGGetProfileRules`

**Files:**
- `cmd/v2local/ssg_handlers.go`

**Step 1.5: Meta Service** (2-3 hours)
- Update tick-tock importer to include manifests
- Pattern: Table → Guide → Manifest → (repeat)
- Or: Table → Guide → Manifest → DataStream → (repeat)

**Files:**
- `pkg/ssg/job/importer.go`

### Phase 2: Data Stream Support

**Step 2.1: Data Models** (4-6 hours)

This is complex - need models for:

```go
// SSGDataStream - Root data stream object
type SSGDataStream struct {
    ID          string    `gorm:"primaryKey"` // e.g., "ssg-al2023-ds"
    Product     string    `gorm:"index"`
    Version     string    // SCAP version (1.2, 1.3)
    Timestamp   time.Time // Data stream generation timestamp
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// SSGBenchmark - XCCDF Benchmark (one per data stream)
type SSGBenchmark struct {
    ID             string    `gorm:"primaryKey"` // XCCDF benchmark ID
    DataStreamID   string    `gorm:"index"`      // Foreign key
    Product        string    `gorm:"index"`
    Title          string
    Description    string    `gorm:"type:text"`
    Version        string
    CreatedAt      time.Time
    UpdatedAt      time.Time
}

// SSGDSProfile - Profile from data stream (more complete than manifest)
type SSGDSProfile struct {
    ID           string    `gorm:"primaryKey"` // XCCDF profile ID
    BenchmarkID  string    `gorm:"index"`
    ProfileID    string    `gorm:"index"`      // Short ID (e.g., "cis")
    Title        string
    Description  string    `gorm:"type:text"`
    Extends      string    // Parent profile ID if extends another
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// SSGDSGroup - Group from data stream
type SSGDSGroup struct {
    ID           string    `gorm:"primaryKey"` // XCCDF group ID
    BenchmarkID  string    `gorm:"index"`
    ParentID     string    `gorm:"index"`
    Title        string
    Description  string    `gorm:"type:text"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// SSGDSRule - Rule from data stream (most comprehensive)
type SSGDSRule struct {
    ID           string         `gorm:"primaryKey"` // XCCDF rule ID
    BenchmarkID  string         `gorm:"index"`
    GroupID      string         `gorm:"index"`
    ShortID      string         `gorm:"index"`      // e.g., "aide_build_database"
    Title        string
    Description  string         `gorm:"type:text"`
    Rationale    string         `gorm:"type:text"`
    Severity     string         `gorm:"index"`
    Checks       []SSGDSCheck   `gorm:"foreignKey:RuleID"`
    Identifiers  []SSGDSIdent   `gorm:"foreignKey:RuleID"`
    References   []SSGDSRef     `gorm:"foreignKey:RuleID"`
    CreatedAt    time.Time
    UpdatedAt    time.Time
}

// SSGDSCheck - Check reference (OVAL, OCIL, SCE)
type SSGDSCheck struct {
    ID       uint   `gorm:"primaryKey"`
    RuleID   string `gorm:"index"`
    System   string // oval, ocil, sce
    CheckID  string // Reference to OVAL definition, OCIL questionnaire, etc.
}

// SSGDSIdent - Identifier (CCE, CVE, etc.)
type SSGDSIdent struct {
    ID       uint   `gorm:"primaryKey"`
    RuleID   string `gorm:"index"`
    System   string // cce, cve, etc.
    Value    string `gorm:"index"` // e.g., "CCE-80644-8"
}

// SSGDSRef - External reference
type SSGDSRef struct {
    ID       uint   `gorm:"primaryKey"`
    RuleID   string `gorm:"index"`
    Href     string
    Label    string
}

// SSGOVALDefinition - OVAL definition (optional - can be very large)
type SSGOVALDefinition struct {
    ID           string    `gorm:"primaryKey"` // OVAL definition ID
    DataStreamID string    `gorm:"index"`
    Class        string    // compliance, inventory, patch, etc.
    Title        string
    Description  string    `gorm:"type:text"`
    // Criteria stored as JSON blob to avoid deep modeling
    Criteria     string    `gorm:"type:text"` // JSON
    CreatedAt    time.Time
}
```

**Files:**
- `pkg/ssg/models.go`

**Step 2.2: XML Parser** (12-20 hours - COMPLEX)

This is the most challenging part. Need to:
1. Parse multi-namespace XML
2. Extract embedded components
3. Handle XCCDF 1.2 schema
4. Parse OVAL definitions
5. Parse OCIL questionnaires (optional)

**Dependencies:**
- Use Go's `encoding/xml` with custom unmarshaling
- OR use `github.com/beevik/etree` for easier XML navigation

**Approach:**
- Parse data stream collection
- Extract XCCDF benchmark component
- Parse benchmark → profiles, groups, rules
- Extract check references
- Optionally parse OVAL definitions (simplified storage as JSON)

**Files:**
- `pkg/ssg/parser/datastream.go` - Main parser
- `pkg/ssg/parser/xccdf.go` - XCCDF parsing
- `pkg/ssg/parser/oval.go` - OVAL parsing (simplified)
- `pkg/ssg/parser/datastream_test.go` - Unit tests

**Step 2.3: Remote Service** (1-2 hours)
- Add `ListDataStreamFiles()` to GitClient
- Add `RPCSSGListDataStreamFiles` handler

**Step 2.4: Local Service** (3-4 hours)
- RPC handlers for data streams, benchmarks, profiles, groups, rules
- Complex due to many related entities

**Step 2.5: Meta Service** (3-4 hours)
- Update importer for 4-way tick-tock
- Pattern: Table → Guide → Manifest → DataStream → (repeat)

### Phase 3: Cross-Reference Implementation

**Step 3.1: Cross-Reference Table** (2-3 hours)
- Add `SSGCrossReference` model
- CRUD operations
- Build cross-references during import

**Step 3.2: Cross-Reference Builder** (4-6 hours)
- After importing all objects, build cross-references
- Match rule IDs across guides, manifests, data streams
- Match CCE IDs from tables to data stream identifiers
- Match profile IDs

**Files:**
- `pkg/ssg/crossref/builder.go`
- `pkg/ssg/local/store.go` - Add cross-ref queries

### Phase 4: Frontend Updates

**Step 4.1: Types and RPC Client** (2-3 hours)
- Add TypeScript types for manifests, data streams
- Add RPC methods

**Step 4.2: UI Components** (6-8 hours)
- Add "Manifests" and "Data Streams" tabs
- Manifest viewer: Show profiles and their rules
- Data Stream viewer: Show benchmark, profiles, groups, rules
- Cross-reference panel: Show related objects

**Step 4.3: Cross-Reference UI** (4-6 hours)
- When viewing a rule, show:
  - Related guide rules
  - Related table entries (via CCE)
  - Related manifest profiles that include this rule
  - Related data stream rules
  - OVAL checks that apply

## Estimated Timeline

| Phase | Task | Hours |
|-------|------|-------|
| 1 | Manifests | 10-15 hours |
| 2 | Data Streams | 25-35 hours |
| 3 | Cross-References | 6-9 hours |
| 4 | Frontend | 12-17 hours |
| **Total** | | **53-76 hours** |

## Priority Recommendation

Given complexity:

1. **Phase 1: Manifests** (simpler, high value)
   - JSON parsing is straightforward
   - Provides profile → rule mappings
   - Quick win for cross-referencing

2. **Phase 3: Cross-References** (after manifests)
   - Build bridges between existing data (guides, tables) and manifests
   - Demonstrate value before tackling data streams

3. **Phase 2: Data Streams** (most complex)
   - Large scope, complex XML parsing
   - Consider simplified approach: parse only profiles and rules, skip OVAL/OCIL initially
   - Can iterate and add more detail later

4. **Phase 4: Frontend** (after backend complete)
   - Build UI once data is available

## Simplified Data Stream Approach (Alternative)

If full data stream parsing is too complex, consider:

**Minimal Viable Parsing:**
- Extract only: Benchmark ID, Profiles, Groups, Rules
- Store check references as simple strings (not full OVAL)
- Skip CPE, detailed OVAL definitions, OCIL

This reduces complexity by ~60% while still providing:
- Profile definitions
- Rule hierarchy
- Cross-reference capability (rule IDs, CCE identifiers)

## Next Steps

1. ✅ Create this design document
2. Review with team for priorities
3. Start with Phase 1 (Manifests)
4. Build cross-references with existing data
5. Tackle data streams with simplified or full approach based on requirements
