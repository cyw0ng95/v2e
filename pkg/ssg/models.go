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
	ID          string    `gorm:"primaryKey" json:"id"`          // e.g., "ssg-al2023-guide-cis"
	Product     string    `gorm:"index" json:"product"`          // al2023, rhel9, etc.
	ProfileID   string    `gorm:"index" json:"profile_id"`       // Profile ID from HTML (empty for index)
	ShortID     string    `gorm:"index" json:"short_id"`         // e.g., "cis", "index"
	Title       string    `json:"title"`                         // e.g., "CIS Amazon Linux 2023 Benchmark"
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
	ID          string    `gorm:"primaryKey" json:"id"`   // e.g., "xccdf_org.ssgproject.content_group_system"
	GuideID     string    `gorm:"index" json:"guide_id"`  // Parent guide
	ParentID    string    `gorm:"index" json:"parent_id"` // Parent group (empty for top-level)
	Title       string    `json:"title"`                  // e.g., "System Settings"
	Description string    `json:"description"`
	Level       int       `json:"level"`       // Tree depth (0, 1, 2...)
	GroupCount  int       `json:"group_count"` // Number of child groups
	RuleCount   int       `json:"rule_count"`  // Number of child rules
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
	ID          string         `gorm:"primaryKey" json:"id"`  // e.g., "xccdf_org.ssgproject.content_rule_package_aide_installed"
	GuideID     string         `gorm:"index" json:"guide_id"` // Parent guide
	GroupID     string         `gorm:"index" json:"group_id"` // Parent group
	ShortID     string         `gorm:"index" json:"short_id"` // e.g., "package_aide_installed"
	Title       string         `json:"title"`                 // e.g., "Install AIDE"
	Description string         `json:"description"`
	Rationale   string         `json:"rationale"`
	Severity    string         `gorm:"index" json:"severity"` // low, medium, high
	References  []SSGReference `gorm:"foreignKey:RuleID" json:"references"`
	Level       int            `json:"level"` // Tree depth
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
	Href   string `json:"href"`                 // e.g., "https://www.cisecurity.org/controls/"
	Label  string `json:"label"`                // e.g., "cis-csc"
	Value  string `json:"value"`                // e.g., "1, 11, 12, 13, 14, 15, 16, 2, 3, 5, 7, 8, 9"
}

// TableName specifies the table name for SSGReference.
func (SSGReference) TableName() string {
	return "ssg_references"
}

// SSGTree represents a complete tree structure for a guide.
// Used for returning the full hierarchical data structure.
type SSGTree struct {
	Guide  SSGGuide   `json:"guide"`
	Groups []SSGGroup `json:"groups"`
	Rules  []SSGRule  `json:"rules"`
}

// TreeNode represents a node in the SSG tree (for building the tree structure).
type TreeNode struct {
	ID       string      `json:"id"`
	ParentID string      `json:"parent_id"`
	Level    int         `json:"level"`
	Type     string      `json:"type"` // "group" or "rule"
	Group    *SSGGroup   `json:"group,omitempty"`
	Rule     *SSGRule    `json:"rule,omitempty"`
	Children []*TreeNode `json:"children,omitempty"`
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
	ID          string    `gorm:"primaryKey" json:"id"`    // e.g., "table-al2023-cces"
	Product     string    `gorm:"index" json:"product"`    // al2023, rhel9, etc.
	TableType   string    `gorm:"index" json:"table_type"` // cces, nistrefs, stig, etc.
	Title       string    `json:"title"`                   // e.g., "CCE Identifiers in Guide..."
	Description string    `json:"description"`             // Optional description
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
	TableID     string    `gorm:"index" json:"table_id"`        // Foreign key to SSGTable
	Mapping     string    `gorm:"index" json:"mapping"`         // e.g., "CCE-80644-8", "NIST 800-53"
	RuleTitle   string    `json:"rule_title"`                   // e.g., "Install the tmux Package"
	Description string    `gorm:"type:text" json:"description"` // Full description
	Rationale   string    `gorm:"type:text" json:"rationale"`   // Rationale text
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGTableEntry.
func (SSGTableEntry) TableName() string {
	return "ssg_table_entries"
}

// SSGManifest represents a product manifest from SSG containing profile definitions.
// Manifests are JSON files that list available profiles and their associated rules.
type SSGManifest struct {
	ID        string    `gorm:"primaryKey" json:"id"` // e.g., "manifest-al2023"
	Product   string    `gorm:"index" json:"product"` // al2023, rhel8, etc.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGManifest.
func (SSGManifest) TableName() string {
	return "ssg_manifests"
}

// SSGProfile represents a security profile from a manifest.
// Profiles define sets of rules for specific compliance frameworks (CIS, STIG, etc.).
type SSGProfile struct {
	ID         string    `gorm:"primaryKey" json:"id"`     // e.g., "al2023:cis"
	ManifestID string    `gorm:"index" json:"manifest_id"` // Foreign key to manifest
	Product    string    `gorm:"index" json:"product"`     // Denormalized for queries
	ProfileID  string    `gorm:"index" json:"profile_id"`  // e.g., "cis"
	RuleCount  int       `json:"rule_count"`               // Number of rules in this profile
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGProfile.
func (SSGProfile) TableName() string {
	return "ssg_profiles"
}

// SSGProfileRule represents a many-to-many relationship between profiles and rules.
// Links profiles to their constituent rules by rule short ID.
type SSGProfileRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProfileID   string    `gorm:"index" json:"profile_id"`    // Foreign key to SSGProfile
	RuleShortID string    `gorm:"index" json:"rule_short_id"` // e.g., "aide_build_database"
	CreatedAt   time.Time `json:"created_at"`
}

// TableName specifies the table name for SSGProfileRule.
func (SSGProfileRule) TableName() string {
	return "ssg_profile_rules"
}

// ============================================================================
// Data Stream Models (SCAP XML)
// ============================================================================

// SSGDataStream represents a SCAP data stream collection file (ssg-*-ds.xml).
// Data streams contain comprehensive XCCDF benchmarks, OVAL definitions, and OCIL questionnaires.
type SSGDataStream struct {
	ID          string    `gorm:"primaryKey" json:"id"` // e.g., "scap_org.open-scap_datastream_from_xccdf_ssg-al2023-xccdf.xml"
	Product     string    `gorm:"index" json:"product"` // al2023, rhel9, etc.
	ScapVersion string    `json:"scap_version"`         // e.g., "1.3"
	Timestamp   string    `json:"timestamp"`            // ISO 8601 timestamp from data stream
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGDataStream.
func (SSGDataStream) TableName() string {
	return "ssg_data_streams"
}

// SSGBenchmark represents an XCCDF Benchmark from a data stream.
// Benchmarks contain the complete security guidance hierarchy (profiles, groups, rules).
type SSGBenchmark struct {
	ID           string    `gorm:"primaryKey" json:"id"` // e.g., "xccdf_org.ssgproject.content_benchmark_AL-2023"
	DataStreamID string    `gorm:"index" json:"data_stream_id"`
	Title        string    `json:"title"` // e.g., "Guide to the Secure Configuration of Amazon Linux 2023"
	Description  string    `gorm:"type:text" json:"description"`
	Version      string    `json:"version"`       // Benchmark version
	Status       string    `json:"status"`        // draft, accepted, etc.
	StatusDate   string    `json:"status_date"`   // ISO 8601 date
	ProfileCount int       `json:"profile_count"` // Number of profiles in benchmark
	GroupCount   int       `json:"group_count"`   // Number of groups
	RuleCount    int       `json:"rule_count"`    // Number of rules
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGBenchmark.
func (SSGBenchmark) TableName() string {
	return "ssg_benchmarks"
}

// SSGDSProfile represents an XCCDF Profile from a data stream.
// Profiles define specific security baselines (CIS, STIG, ANSSI, etc.) by selecting rules.
type SSGDSProfile struct {
	ID          string    `gorm:"primaryKey" json:"id"` // e.g., "xccdf_org.ssgproject.content_profile_cis"
	BenchmarkID string    `gorm:"index" json:"benchmark_id"`
	Title       string    `json:"title"` // e.g., "CIS Amazon Linux 2023 Benchmark for Level 2 - Server"
	Description string    `gorm:"type:text" json:"description"`
	Version     string    `json:"version"`    // Profile version (e.g., "1.0.0")
	RuleCount   int       `json:"rule_count"` // Number of selected rules
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	SelectedRules []SSGDSProfileRule `gorm:"foreignKey:ProfileID" json:"selected_rules,omitempty"`
}

// TableName specifies the table name for SSGDSProfile.
func (SSGDSProfile) TableName() string {
	return "ssg_ds_profiles"
}

// SSGDSProfileRule represents a rule selection in a data stream profile.
// Maps profiles to their selected rules with selection status.
type SSGDSProfileRule struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProfileID string    `gorm:"index" json:"profile_id"`
	RuleID    string    `gorm:"index" json:"rule_id"` // Full XCCDF rule ID
	Selected  bool      `json:"selected"`             // true if selected, false if deselected
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for SSGDSProfileRule.
func (SSGDSProfileRule) TableName() string {
	return "ssg_ds_profile_rules"
}

// SSGDSGroup represents an XCCDF Group from a data stream.
// Groups organize rules into a hierarchical structure (similar to HTML guide groups but from XML).
type SSGDSGroup struct {
	ID          string    `gorm:"primaryKey" json:"id"` // e.g., "xccdf_org.ssgproject.content_group_system"
	BenchmarkID string    `gorm:"index" json:"benchmark_id"`
	ParentID    string    `gorm:"index" json:"parent_id"` // Parent group (empty for top-level)
	Title       string    `json:"title"`                  // e.g., "System Settings"
	Description string    `gorm:"type:text" json:"description"`
	Level       int       `json:"level"` // Tree depth (0, 1, 2...)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for SSGDSGroup.
func (SSGDSGroup) TableName() string {
	return "ssg_ds_groups"
}

// SSGDSRule represents an XCCDF Rule from a data stream.
// Rules are the atomic security configuration items with detailed metadata.
type SSGDSRule struct {
	ID          string    `gorm:"primaryKey" json:"id"` // e.g., "xccdf_org.ssgproject.content_rule_package_aide_installed"
	BenchmarkID string    `gorm:"index" json:"benchmark_id"`
	GroupID     string    `gorm:"index" json:"group_id"` // Parent group
	Title       string    `json:"title"`                 // e.g., "Install AIDE"
	Description string    `gorm:"type:text" json:"description"`
	Rationale   string    `gorm:"type:text" json:"rationale"`
	Severity    string    `json:"severity"` // low, medium, high, unknown
	Selected    bool      `json:"selected"` // Default selection state
	Weight      string    `json:"weight"`   // Rule weight (importance)
	Version     string    `json:"version"`  // Rule version
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	References  []SSGDSRuleReference  `gorm:"foreignKey:RuleID" json:"references,omitempty"`
	Identifiers []SSGDSRuleIdentifier `gorm:"foreignKey:RuleID" json:"identifiers,omitempty"`
}

// TableName specifies the table name for SSGDSRule.
func (SSGDSRule) TableName() string {
	return "ssg_ds_rules"
}

// SSGDSRuleReference represents a reference/citation for a data stream rule.
// Rules typically reference multiple standards (CIS, NIST, PCI-DSS, etc.).
type SSGDSRuleReference struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID    string    `gorm:"index" json:"rule_id"`
	Href      string    `json:"href"`   // URL to standard
	RefID     string    `json:"ref_id"` // Specific section/control (e.g., "1.3.1", "CM-6(a)")
	CreatedAt time.Time `json:"created_at"`
}

// TableName specifies the table name for SSGDSRuleReference.
func (SSGDSRuleReference) TableName() string {
	return "ssg_ds_rule_references"
}

// SSGDSRuleIdentifier represents an external identifier for a data stream rule.
// Common identifiers: CCE, CVE, OVAL check references.
type SSGDSRuleIdentifier struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RuleID     string    `gorm:"index" json:"rule_id"`
	System     string    `json:"system"`                  // e.g., "http://cce.mitre.org"
	Identifier string    `gorm:"index" json:"identifier"` // e.g., "CCE-80644-8"
	CreatedAt  time.Time `json:"created_at"`
}

// TableName specifies the table name for SSGDSRuleIdentifier.
func (SSGDSRuleIdentifier) TableName() string {
	return "ssg_ds_rule_identifiers"
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

// BeforeCreate is a GORM hook called before creating a new manifest record.
func (m *SSGManifest) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook called before updating a manifest record.
func (m *SSGManifest) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate is a GORM hook called before creating a new profile record.
func (p *SSGProfile) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	p.CreatedAt = now
	p.UpdatedAt = now
	return nil
}

// BeforeUpdate is a GORM hook called before updating a profile record.
func (p *SSGProfile) BeforeUpdate(tx *gorm.DB) error {
	p.UpdatedAt = time.Now()
	return nil
}

// BeforeCreate is a GORM hook called before creating a new profile rule record.
func (pr *SSGProfileRule) BeforeCreate(tx *gorm.DB) error {
	pr.CreatedAt = time.Now()
	return nil
}

// SSGCrossReference represents a link between SSG objects.
// Enables navigation between guides, tables, manifests, and data streams
// based on common identifiers (Rule IDs, CCE, Products, Profiles).
type SSGCrossReference struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	SourceType string    `gorm:"index:idx_source;not null" json:"source_type"` // "guide", "table", "manifest", "datastream"
	SourceID   string    `gorm:"index:idx_source;not null" json:"source_id"`   // UUID of source object
	TargetType string    `gorm:"index:idx_target;not null" json:"target_type"` // "guide", "table", "manifest", "datastream"
	TargetID   string    `gorm:"index:idx_target;not null" json:"target_id"`   // UUID of target object
	LinkType   string    `gorm:"index;not null" json:"link_type"`              // "rule_id", "cce", "product", "profile_id"
	Metadata   string    `gorm:"type:text" json:"metadata"`                    // JSON with additional context (e.g., rule short ID, CCE number)
	CreatedAt  time.Time `json:"created_at"`
}

// TableName specifies the table name for SSGCrossReference.
func (SSGCrossReference) TableName() string {
	return "ssg_cross_references"
}

// BeforeCreate is a GORM hook called before creating a new cross-reference record.
func (cr *SSGCrossReference) BeforeCreate(tx *gorm.DB) error {
	cr.CreatedAt = time.Now()
	return nil
}
