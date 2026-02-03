# OpenScap SSG Integration - Step-by-Step Iteration Plan

## Overview
Integrate SCAP Security Guide (SSG) data into v2e by parsing HTML guide files to extract groups, rules, and tree relationships. Uses meta/local/remote services with go-git for runtime repository fetching.

**Scope**: Parse `*-guide-*.html` files to extract XCCDF groups and rules with tree structure. NOT importing XCCDF `*-ds.xml` files or remediation scripts initially.

**Architecture**:
```
Frontend → Access → Broker → Meta (orchestrator) → Remote (git fetch) + Local (parse/store)
```

---

## Data Model (from HTML Guide Parsing)

### Core Models
```go
// SSGGuide - HTML documentation guide (container)
type SSGGuide struct {
    ID          string    `gorm:"primaryKey" json:"id"`           // e.g., "ssg-al2023-guide-cis"
    Product     string    `gorm:"index" json:"product"`           // al2023, rhel9, etc.
    ProfileID   string    `gorm:"index" json:"profile_id"`        // Profile ID from HTML
    ShortID     string    `gorm:"index" json:"short_id"`          // e.g., "cis", "index"
    Title       string    `json:"title"`                          // e.g., "CIS Amazon Linux 2023 Benchmark"
    HTMLContent string    `json:"html_content"`                   // Full HTML content
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// SSGGroup - XCCDF group (category) from HTML guide
type SSGGroup struct {
    ID          string    `gorm:"primaryKey" json:"id"`           // e.g., "xccdf_org.ssgproject.content_group_system"
    GuideID     string    `gorm:"index" json:"guide_id"`          // Parent guide
    ParentID    string    `gorm:"index" json:"parent_id"`         // Parent group (empty for top-level)
    Title       string    `json:"title"`                          // e.g., "System Settings"
    Description string    `json:"description"`
    Level       int       `json:"level"`                          // Tree depth (0, 1, 2...)
    GroupCount  int       `json:"group_count"`                    // Number of child groups
    RuleCount   int       `json:"rule_count"`                     // Number of child rules
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// SSGRule - XCCDF rule from HTML guide
type SSGRule struct {
    ID          string         `gorm:"primaryKey" json:"id"`      // e.g., "xccdf_org.ssgproject.content_rule_package_aide_installed"
    GuideID     string         `gorm:"index" json:"guide_id"`     // Parent guide
    GroupID     string         `gorm:"index" json:"group_id"`     // Parent group
    ShortID     string         `gorm:"index" json:"short_id"`     // e.g., "package_aide_installed"
    Title       string         `json:"title"`                     // e.g., "Install AIDE"
    Description string         `json:"description"`
    Rationale   string         `json:"rationale"`
    Severity    string         `gorm:"index" json:"severity"`    // low, medium, high
    References  []SSGReference `gorm:"foreignKey:RuleID" json:"references"`
    Level       int            `json:"level"`                    // Tree depth
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
}

// SSGReference - Rule reference (from HTML)
type SSGReference struct {
    ID        uint   `gorm:"primaryKey" json:"-"`
    RuleID    string `gorm:"index" json:"rule_id"`
    Href      string `json:"href"`        // e.g., "https://www.cisecurity.org/controls/"
    Label     string `json:"label"`       // e.g., "cis-csc"
    Value     string `json:"value"`       // e.g., "1, 11, 12, 13, 14, 15, 16, 2, 3, 5, 7, 8, 9"
}
```

### Tree Structure
The HTML guide contains a tree where:
- **Root**: Benchmark (e.g., `content_benchmark_AL-2023`)
- **Groups**: Categories (e.g., `content_group_system`, `content_group_software`)
- **Rules**: Security rules (e.g., `content_rule_package_aide_installed`)

Tree reconstruction:
```
Benchmark (root)
  └── Group: System Settings
      ├── Group: Installing and Maintaining Software
      │   └── Rule: Install AIDE
      └── Group: Account and Access Control
          └── Rule: Account Disable Post PW Expiration
```

---

## Configuration

Add to `config_spec.json`:

| Variable | Default | Target | Description |
|----------|---------|--------|-------------|
| `CONFIG_SSG_DBPATH` | `ssg.db` | `github.com/cyw0ng95/v2e/pkg/ssg/local.buildSSGDBPath` | SQLite database path |
| `CONFIG_SSG_REPO_URL` | `https://github.com/cyw0ng95/scap-security-guide-0.1.79` | `github.com/cyw0ng95/v2e/pkg/ssg/remote/buildRepoURL` | Git repository |
| `CONFIG_SSG_REPO_PATH` | `assets/ssg-git` | `github.com/cyw0ng95/v2e/pkg/ssg/remote/buildRepoPath` | Local checkout path |

---

## Step 1: Remote Service - Git Operations

### Target
Remote service clones/pulls SSG repository and lists guide files.

### RPC Methods (all named `RPCSSG*`):
| Method | Description |
|--------|-------------|
| `RPCSSGCloneRepo` | Clone repository |
| `RPCSSGPullRepo` | Pull latest changes |
| `RPCSSGGetRepoStatus` | Get commit/branch/status |
| `RPCSSGListGuideFiles` | List `*-guide-*.html` files in `guides/` |
| `RPCSSGGetFilePath` | Get absolute path for file |

### Build Variables (`pkg/ssg/remote/vars.go`):
```go
var (
    buildRepoURL  string
    buildRepoPath string
)

func buildRepoURL() string {
    if buildRepoURL != "" { return buildRepoURL }
    return "https://github.com/cyw0ng95/scap-security-guide-0.1.79"
}

func buildRepoPath() string {
    if buildRepoPath != "" { return buildRepoPath }
    return "assets/ssg-git"
}
```

### Files:
- `pkg/ssg/remote/vars.go` - Build variables
- `pkg/ssg/remote/git.go` - Git operations
- `pkg/ssg/remote/handlers.go` - RPC handlers
- Update `cmd/remote/service.md`

### Cost: 3-5 hours

---

## Step 2: Local Service - Data Models

### Target
Define SSG data models for guide, group, rule with tree relationships.

### Files:
- `pkg/ssg/models.go` - All models
- `pkg/ssg/local/vars.go` - Build variables

### Build Variables:
```go
var buildSSGDBPath string

func buildSSGDBPath() string {
    if buildSSGDBPath != "" { return buildSSGDBPath }
    return "ssg.db"
}
```

### Cost: 1.5-2.5 hours

---

## Step 3: Local Service - SQLite Store

### Target
Implement SQLite storage with tree queries.

### Methods:
| Method | Description |
|--------|-------------|
| `SaveGuide(guide *SSGGuide) error` | Save guide |
| `GetGuide(id string) (*SSGGuide, error)` | Get by ID |
| `ListGuides(product, profileID string) ([]SSGGuide, error)` | List with filters |
| `SaveGroup(group *SSGGroup) error` | Save group |
| `GetGroup(id string) (*SSGGroup, error)` | Get by ID |
| `SaveRule(rule *SSGRule) error` | Save rule with references |
| `GetRule(id string) (*SSGRule, error)` | Get by ID |
| `GetTree(guideID string) (*SSGTree, error)` | Get full tree for guide |
| `GetChildGroups(groupID string) ([]SSGGroup, error)` | Get child groups |
| `GetChildRules(groupID string) ([]SSGRule, error)` | Get child rules |
| `GetRootGroups(guideID string) ([]SSGGroup, error)` | Get top-level groups |

### Files:
- `pkg/ssg/local/store.go` - Store implementation
- `pkg/ssg/local/store_test.go` - Unit tests

### Cost: 5-7 hours

---

## Step 4: Local Service - RPC Handlers

### RPC Methods (all named `RPCSSG*`):
| Method | Description |
|--------|-------------|
| `RPCSSGImportGuide` | Import guide from path |
| `RPCSSGGetGuide` | Get guide by ID |
| `RPCSSGListGuides` | List guides |
| `RPCSSGGetTree` | Get full tree for guide |
| `RPCSSGGetGroup` | Get group by ID |
| `RPCSSGGetRule` | Get rule by ID |
| `RPCSSGListRules` | List rules by group/guide |

### Files:
- `pkg/ssg/local/handlers.go` - RPC handlers
- Update `cmd/local/main.go`
- Update `cmd/local/service.md`

### Cost: 3-5 hours

---

## Step 5: HTML Guide Parser

### Target
Parse HTML guides to extract groups, rules, and tree structure.

### HTML Structure to Parse:
```html
<tr data-tt-id="children-xccdf_org.ssgproject.content_group_system"
    data-tt-parent-id="children-xccdf_org.ssgproject.content_benchmark_AL-2023">
  <td id="xccdf_org.ssgproject.content_group_system">
    <span class="label label-default">Group</span>
    System Settings
    <small>Group contains 50 groups and 200 rules</small>
  </td>
</tr>

<tr data-tt-id="xccdf_org.ssgproject.content_rule_package_aide_installed"
    data-tt-parent-id="children-xccdf_org.ssgproject.content_group_aide">
  <td id="xccdf_org.ssgproject.content_rule_package_aide_installed">
    <span class="label label-default">Rule</span> Install AIDE
    <div class="description">...</div>
    <div class="rationale">...</div>
    <div class="severity">medium</div>
    <table class="identifiers">...</table>
  </td>
</tr>
```

### Parser Functions:
```go
// ParseGuideFile parses HTML guide and extracts models
func ParseGuideFile(path string) (*SSGGuide, []SSGGroup, []SSGRule, error)

// parseHTMLTree extracts tree structure from data-tt-id/data-tt-parent-id
func parseHTMLTree(html *goquery.Document) ([]TreeNode, error)

// parseGroup extracts group data from <tr> element
func parseGroup(selection *goquery.Selection) (*SSGGroup, error)

// parseRule extracts rule data from <tr> element
func parseRule(selection *goquery.Selection) (*SSGRule, error)

// parseReferences extracts reference table
func parseReferences(selection *goquery.Selection) ([]SSGReference, error)

// extractMetadata gets title, profile_id from HTML head
func extractMetadata(html *goquery.Document) (title, profileID string)

// TreeNode represents tree node during parsing
type TreeNode struct {
    ID       string
    ParentID string
    Level    int
    Type     string  // "group" or "rule"
}
```

### Dependencies:
- `github.com/PuerkitoBio/goquery` - HTML parsing (jQuery-like API)

### Files:
- `pkg/ssg/parser/guide.go` - HTML parsing
- `pkg/ssg/parser/tree.go` - Tree extraction
- `pkg/ssg/parser/parser_test.go` - Unit tests

### Cost: 8-12 hours (HTML parsing is complex)

---

## Step 6: Meta Service - Import Job

### Target
Orchestrate SSG import: pull git, import all guides.

### Workflow:
1. Call `remote.RPCSSGPullRepo()`
2. Call `remote.RPCSSGListGuideFiles()`
3. For each guide: Call `local.RPCSSGImportGuide(path)`

### RPC Methods (all named `RPCSSG*`):
| Method | Description |
|--------|-------------|
| `RPCSSGStartImportJob` | Start import job |
| `RPCSSGStopImportJob` | Stop running job |
| `RPCSSGGetImportStatus` | Get job status |

### Files:
- `pkg/ssg/job/importer.go` - Import logic
- `pkg/ssg/job/session.go` - Session management
- Update `cmd/meta/main.go`
- Update `cmd/meta/service.md`

### Cost: 5-6 hours

---

## Step 7: Access Service - Passthrough

### Passthrough RPCs (all named `RPCSSG*`):
| Method | Target |
|--------|--------|
| `RPCSSGGetGuide` | local |
| `RPCSSGListGuides` | local |
| `RPCSSGGetTree` | local |
| `RPCSSGGetGroup` | local |
| `RPCSSGGetRule` | local |
| `RPCSSGStartImportJob` | meta |
| `RPCSSGGetImportStatus` | meta |

### Files:
- Update `cmd/access/main.go`
- Update `cmd/access/service.md`

### Cost: 1-2 hours

---

## Step 8: Frontend - TypeScript Types

```typescript
interface SSGGuide {
  id: string;
  product: string;
  profile_id: string;
  short_id: string;
  title: string;
  html_content: string;
  created_at: string;
  updated_at: string;
}

interface SSGGroup {
  id: string;
  guide_id: string;
  parent_id: string;
  title: string;
  description: string;
  level: number;
  group_count: number;
  rule_count: number;
}

interface SSGRule {
  id: string;
  guide_id: string;
  group_id: string;
  short_id: string;
  title: string;
  description: string;
  rationale: string;
  severity: 'low' | 'medium' | 'high';
  references: SSGReference[];
  level: number;
}

interface SSGReference {
  href: string;
  label: string;
  value: string;
}

interface SSGTree {
  guide: SSGGuide;
  groups: SSGGroup[];
  rules: SSGRule[];
}
```

### Files:
- Update `website/lib/types.ts`

### Cost: 0.5-1 hours

---

## Step 9: Frontend - SSG Pages

### Pages:
| Path | Description |
|------|-------------|
| `/ssg` | SSG dashboard (list guides) |
| `/ssg/guide/[id]` | Guide tree viewer |
| `/ssg/group/[id]` | Group details |
| `/ssg/rule/[id]` | Rule details |

### Components:
- `SSGGuideList` - List guides
- `SSGTreeViewer` - Tree view with expand/collapse
- `SSGGroupCard` - Group preview
- `SSGRuleCard` - Rule preview

### Files:
- `website/app/ssg/page.tsx`
- `website/app/ssg/guide/[id]/page.tsx`
- `website/app/ssg/group/[id]/page.tsx`
- `website/app/ssg/rule/[id]/page.tsx`
- `website/components/ssg/*.tsx`

### Cost: 6-8 hours (tree viewer is complex)

---

## Summary

### Total Cost: 33.5-48.5 hours

| Step | Description | Time |
|------|-------------|------|
| 1 | Remote Service - Git | 3-5 hours |
| 2 | Local Service - Models | 1.5-2.5 hours |
| 3 | Local Service - Store | 5-7 hours |
| 4 | Local Service - RPC | 3-5 hours |
| 5 | HTML Guide Parser | 8-12 hours |
| 6 | Meta Service - Job | 5-6 hours |
| 7 | Access Service - RPC | 1-2 hours |
| 8 | Frontend - Types | 0.5-1 hours |
| 9 | Frontend - Pages | 6-8 hours |

### Total Changes:
- ~3000 lines new code
- 15-20 new files
- ~8 modified files
- `config_spec.json` updated with 3 entries

### Risk Assessment:
- **Low Risk**: Steps 1, 2, 3, 4, 7, 8 (follow existing patterns)
- **Medium Risk**: Steps 5, 6, 9 (HTML parsing complexity, tree structure, frontend tree viewer)

---

## Future Enhancements (Out of Scope)

- Parse XCCDF `*-ds.xml` files for additional data
- Remediation scripts (bash/ansible/kickstart)
- OVAL/OCIL check definitions
- Search functionality
- Compliance reports
- Export profiles
