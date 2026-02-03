// Package ssg provides data models for SCAP Security Guide integration.
package ssg

import (
	"time"

	"gorm.io/gorm"
)

// SSGLoadTime is a custom type for loading related associations.
// Used to eager-load relationships when querying.
type SSGLoadTime string

const (
	// LoadWithReferences loads rule references.
	LoadWithReferences SSGLoadTime = "References"
	// LoadWithChildren loads child groups and rules (for tree queries).
	LoadWithChildren SSGLoadTime = "Children"
)

// SSGGuide represents an HTML documentation guide from SSG.
// Guides contain formatted security guidance for specific products and profiles.
type SSGGuide struct {
	ID          string    `gorm:"primaryKey" json:"id"`           // e.g., "ssg-al2023-guide-cis"
	Product     string    `gorm:"index" json:"product"`           // al2023, rhel9, etc.
	ProfileID   string    `gorm:"index" json:"profile_id"`        // Profile ID from HTML (empty for index)
	ShortID     string    `gorm:"index" json:"short_id"`          // e.g., "cis", "index"
	Title       string    `json:"title"`                          // e.g., "CIS Amazon Linux 2023 Benchmark"
	HTMLContent string    `gorm:"type:text" json:"html_content"` // Full HTML content
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGGuide.
func (SSGGuide) TableName() string {
	return "ssg_guides"
}

// SSGGroup represents an XCCDF group (category) from HTML guide.
// Groups organize rules into a hierarchical tree structure.
type SSGGroup struct {
	ID          string    `gorm:"primaryKey" json:"id"`      // e.g., "xccdf_org.ssgproject.content_group_system"
	GuideID     string    `gorm:"index" json:"guide_id"`     // Parent guide
	ParentID    string    `gorm:"index" json:"parent_id"`    // Parent group (empty for top-level)
	Title       string    `json:"title"`                     // e.g., "System Settings"
	Description string    `json:"description"`
	Level       int       `json:"level"`                    // Tree depth (0, 1, 2...)
	GroupCount  int       `json:"group_count"`              // Number of child groups
	RuleCount   int       `json:"rule_count"`               // Number of child rules
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGGroup.
func (SSGGroup) TableName() string {
	return "ssg_groups"
}

// SSGRule represents an XCCDF rule from HTML guide.
// Rules define specific security requirements or recommendations.
type SSGRule struct {
	ID          string         `gorm:"primaryKey" json:"id"`       // e.g., "xccdf_org.ssgproject.content_rule_package_aide_installed"
	GuideID     string         `gorm:"index" json:"guide_id"`      // Parent guide
	GroupID     string         `gorm:"index" json:"group_id"`      // Parent group
	ShortID     string         `gorm:"index" json:"short_id"`      // e.g., "package_aide_installed"
	Title       string         `json:"title"`                      // e.g., "Install AIDE"
	Description string         `json:"description"`
	Rationale   string         `json:"rationale"`
	Severity    string         `gorm:"index" json:"severity"`     // low, medium, high
	References  []SSGReference `gorm:"foreignKey:RuleID" json:"references"`
	Level       int            `json:"level"`                      // Tree depth
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// TableName specifies the table name for SSGRule.
func (SSGRule) TableName() string {
	return "ssg_rules"
}

// SSGReference represents a rule reference (e.g., CIS, NIST, PCI-DSS).
// References provide external documentation or standards mappings.
type SSGReference struct {
	ID     uint   `gorm:"primaryKey" json:"-"`
	RuleID string `gorm:"index" json:"rule_id"` // Foreign key to SSGRule
	Href   string `json:"href"`        // e.g., "https://www.cisecurity.org/controls/"
	Label  string `json:"label"`       // e.g., "cis-csc"
	Value  string `json:"value"`       // e.g., "1, 11, 12, 13, 14, 15, 16, 2, 3, 5, 7, 8, 9"
}

// TableName specifies the table name for SSGReference.
func (SSGReference) TableName() string {
	return "ssg_references"
}

// SSGTree represents a complete tree structure for a guide.
// Used for returning the full hierarchical data structure.
type SSGTree struct {
	Guide  SSGGuide   `json:"guide"`
	Groups []SSGGroup  `json:"groups"`
	Rules  []SSGRule   `json:"rules"`
}

// TreeNode represents a node in the SSG tree (for building the tree structure).
type TreeNode struct {
	ID       string      `json:"id"`
	ParentID string      `json:"parent_id"`
	Level    int         `json:"level"`
	Type     string      `json:"type"` // "group" or "rule"
	Group    *SSGGroup   `json:"group,omitempty"`
	Rule     *SSGRule    `json:"rule,omitempty"`
	Children []*TreeNode  `json:"children,omitempty"`
}

// BeforeCreate is a GORM hook called before creating a new record.
func (g *SSGGuide) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	g.CreatedAt = now
	g.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook called before updating a record.
func (g *SSGGuide) BeforeUpdate(tx *gorm.DB) error {
	g.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate is a GORM hook called before creating a new record.
func (g *SSGGroup) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	g.CreatedAt = now
	g.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook called before updating a record.
func (g *SSGGroup) BeforeUpdate(tx *gorm.DB) error {
	g.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate is a GORM hook called before creating a new record.
func (r *SSGRule) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	r.CreatedAt = now
	r.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook called before updating a record.
func (r *SSGRule) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = time.Now()
	return nil
}

// SSGTable represents a mapping table from SSG (e.g., CCE mappings, NIST refs).
// Tables contain flat lists of rules with their mappings to security identifiers.
// The actual table data is stored in SSGTableEntry records, not as HTML.
type SSGTable struct {
	ID          string    `gorm:"primaryKey" json:"id"`           // e.g., "table-al2023-cces"
	Product     string    `gorm:"index" json:"product"`           // al2023, rhel9, etc.
	TableType   string    `gorm:"index" json:"table_type"`        // cces, nistrefs, stig, etc.
	Title       string    `json:"title"`                          // e.g., "CCE Identifiers in Guide..."
	Description string    `json:"description"`                    // Optional description
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGTable.
func (SSGTable) TableName() string {
	return "ssg_tables"
}

// SSGTableEntry represents a single row in an SSG mapping table.
// Each entry maps a rule to a security identifier (CCE, NIST, STIG, etc.).
type SSGTableEntry struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TableID     string    `gorm:"index" json:"table_id"`     // Foreign key to SSGTable
	Mapping     string    `gorm:"index" json:"mapping"`      // e.g., "CCE-80644-8", "NIST 800-53"
	RuleTitle   string    `json:"rule_title"`                // e.g., "Install the tmux Package"
	Description string    `gorm:"type:text" json:"description"` // Full description
	Rationale   string    `gorm:"type:text" json:"rationale"`   // Rationale text
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGTableEntry.
func (SSGTableEntry) TableName() string {
	return "ssg_table_entries"
}

// BeforeCreate is a GORM hook called before creating a new record.
func (t *SSGTable) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	t.CreatedAt = now
	t.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook called before updating a record.
func (t *SSGTable) BeforeUpdate(tx *gorm.DB) error {
	t.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate is a GORM hook called before creating a new record.
func (e *SSGTableEntry) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	e.CreatedAt = now
	e.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook called before updating a record.
func (e *SSGTableEntry) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}
