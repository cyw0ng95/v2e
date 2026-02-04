// Package local provides unit tests for SSG storage operations.
package local

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/ssg"
)

func TestNewStore(t *testing.T) {
	// Use in-memory database for testing
	store, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	defer store.Close()

	if store == nil {
		t.Fatal("NewStore() returned nil")
	}

	// Verify tables were created (by attempting to query)
	// This will error if tables don't exist
	var guideCount int64
	if err := store.db.Model(&ssg.SSGGuide{}).Count(&guideCount).Error; err != nil {
		t.Errorf("Failed to query guides table: %v", err)
	}
}

func TestDefaultDBPath(t *testing.T) {
	path := DefaultDBPath()
	if path == "" {
		t.Error("DefaultDBPath returned empty string")
	}
	if path != "ssg.db" {
		t.Errorf("DefaultDBPath = %s, want ssg.db", path)
	}
}

func TestStore_SaveGuide(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	guide := &ssg.SSGGuide{
		ID:          "test-guide",
		Product:     "test-product",
		ProfileID:   "test-profile",
		ShortID:     "test",
		Title:       "Test Guide",
		HTMLContent: "<html>Test</html>",
	}

	err := store.SaveGuide(guide)
	if err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	// Verify it was saved
	retrieved, err := store.GetGuide("test-guide")
	if err != nil {
		t.Fatalf("GetGuide() error = %v", err)
	}

	if retrieved.ID != "test-guide" {
		t.Errorf("ID = %s, want test-guide", retrieved.ID)
	}
	if retrieved.Title != "Test Guide" {
		t.Errorf("Title = %s, want Test Guide", retrieved.Title)
	}
}

func TestStore_ListGuides(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Create test guides
	guides := []ssg.SSGGuide{
		{ID: "guide1", Product: "al2023", ProfileID: "cis", ShortID: "cis", Title: "CIS Guide", HTMLContent: "<html></html>"},
		{ID: "guide2", Product: "al2023", ProfileID: "pci-dss", ShortID: "pci", Title: "PCI Guide", HTMLContent: "<html></html>"},
		{ID: "guide3", Product: "rhel9", ProfileID: "cis", ShortID: "cis", Title: "RHEL CIS", HTMLContent: "<html></html>"},
	}

	for _, guide := range guides {
		if err := store.SaveGuide(&guide); err != nil {
			t.Fatalf("SaveGuide() error = %v", err)
		}
	}

	tests := []struct {
		name      string
		product   string
		profileID string
		wantCount int
	}{
		{
			name:      "all guides",
			product:   "",
			profileID: "",
			wantCount: 3,
		},
		{
			name:      "filter by product",
			product:   "al2023",
			profileID: "",
			wantCount: 2,
		},
		{
			name:      "filter by profile",
			product:   "",
			profileID: "cis",
			wantCount: 2,
		},
		{
			name:      "filter by both",
			product:   "al2023",
			profileID: "cis",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			guides, err := store.ListGuides(tt.product, tt.profileID)
			if err != nil {
				t.Fatalf("ListGuides() error = %v", err)
			}
			if len(guides) != tt.wantCount {
				t.Errorf("ListGuides() count = %d, want %d", len(guides), tt.wantCount)
			}
		})
	}
}

func TestStore_SaveGroup(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	group := &ssg.SSGGroup{
		ID:         "test-group",
		GuideID:    "test-guide",
		ParentID:   "",
		Title:      "System Settings",
		Level:      0,
		GroupCount: 2,
		RuleCount:  5,
	}

	err := store.SaveGroup(group)
	if err != nil {
		t.Fatalf("SaveGroup() error = %v", err)
	}

	// Verify it was saved
	retrieved, err := store.GetGroup("test-group")
	if err != nil {
		t.Fatalf("GetGroup() error = %v", err)
	}

	if retrieved.ID != "test-group" {
		t.Errorf("ID = %s, want test-group", retrieved.ID)
	}
	if retrieved.Title != "System Settings" {
		t.Errorf("Title = %s, want System Settings", retrieved.Title)
	}
}

func TestStore_GetChildGroups(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Create test guide
	guide := &ssg.SSGGuide{
		ID:          "test-guide",
		Product:     "test-product",
		ProfileID:   "test-profile",
		ShortID:     "test",
		Title:       "Test Guide",
		HTMLContent: "<html></html>",
	}
	if err := store.SaveGuide(guide); err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	// Create groups: root -> child -> grandchild
	groups := []*ssg.SSGGroup{
		{ID: "group1", GuideID: "test-guide", ParentID: "", Title: "Root Group", Level: 0},
		{ID: "group2", GuideID: "test-guide", ParentID: "group1", Title: "Child Group", Level: 1},
		{ID: "group3", GuideID: "test-guide", ParentID: "group2", Title: "Grandchild Group", Level: 2},
	}

	for _, group := range groups {
		if err := store.SaveGroup(group); err != nil {
			t.Fatalf("SaveGroup() error = %v", err)
		}
	}

	// Get children of group1
	children, err := store.GetChildGroups("group1")
	if err != nil {
		t.Fatalf("GetChildGroups() error = %v", err)
	}

	if len(children) != 1 {
		t.Errorf("GetChildGroups() count = %d, want 1", len(children))
	}
	if children[0].ID != "group2" {
		t.Errorf("GetChildGroups()[0].ID = %s, want group2", children[0].ID)
	}

	// Get root groups
	roots, err := store.GetRootGroups("test-guide")
	if err != nil {
		t.Fatalf("GetRootGroups() error = %v", err)
	}

	if len(roots) != 1 {
		t.Errorf("GetRootGroups() count = %d, want 1", len(roots))
	}
}

func TestStore_SaveRule(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Create test guide and group
	guide := &ssg.SSGGuide{
		ID:          "test-guide",
		Product:     "test-product",
		ProfileID:   "test-profile",
		ShortID:     "test",
		Title:       "Test Guide",
		HTMLContent: "<html></html>",
	}
	if err := store.SaveGuide(guide); err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	group := &ssg.SSGGroup{
		ID:      "test-group",
		GuideID: "test-guide",
		ParentID: "",
		Title:   "Test Group",
		Level:   0,
	}
	if err := store.SaveGroup(group); err != nil {
		t.Fatalf("SaveGroup() error = %v", err)
	}

	rule := &ssg.SSGRule{
		ID:        "test-rule",
		GuideID:   "test-guide",
		GroupID:   "test-group",
		ShortID:   "test_rule",
		Title:     "Install AIDE",
		Severity:  "medium",
		References: []ssg.SSGReference{
			{Href: "https://example.com", Label: "cis-csc", Value: "1, 2, 3"},
			{Href: "https://nist.gov", Label: "nist", Value: "CM-6"},
		},
		Level: 1,
	}

	err := store.SaveRule(rule)
	if err != nil {
		t.Fatalf("SaveRule() error = %v", err)
	}

	// Verify it was saved with references
	retrieved, err := store.GetRule("test-rule")
	if err != nil {
		t.Fatalf("GetRule() error = %v", err)
	}

	if retrieved.Title != "Install AIDE" {
		t.Errorf("Title = %s, want Install AIDE", retrieved.Title)
	}
	if len(retrieved.References) != 2 {
		t.Errorf("References count = %d, want 2", len(retrieved.References))
	}
}

func TestStore_GetChildRules(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Setup: guide, group, and two rules
	guide := &ssg.SSGGuide{
		ID:          "test-guide",
		Product:     "test-product",
		ProfileID:   "test-profile",
		ShortID:     "test",
		Title:       "Test Guide",
		HTMLContent: "<html></html>",
	}
	if err := store.SaveGuide(guide); err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	group := &ssg.SSGGroup{
		ID:      "test-group",
		GuideID: "test-guide",
		ParentID: "",
		Title:   "Test Group",
		Level:   0,
	}
	if err := store.SaveGroup(group); err != nil {
		t.Fatalf("SaveGroup() error = %v", err)
	}

	rules := []*ssg.SSGRule{
		{
			ID:        "rule1",
			GuideID:   "test-guide",
			GroupID:   "test-group",
			ShortID:   "rule1",
			Title:     "Rule 1",
			Severity:  "low",
			Level:     1,
		},
		{
			ID:        "rule2",
			GuideID:   "test-guide",
			GroupID:   "test-group",
			ShortID:   "rule2",
			Title:     "Rule 2",
			Severity:  "high",
			Level:     1,
		},
	}

	for _, rule := range rules {
		if err := store.SaveRule(rule); err != nil {
			t.Fatalf("SaveRule() error = %v", err)
		}
	}

	// Get child rules
	childRules, err := store.GetChildRules("test-group")
	if err != nil {
		t.Fatalf("GetChildRules() error = %v", err)
	}

	if len(childRules) != 2 {
		t.Errorf("GetChildRules() count = %d, want 2", len(childRules))
	}
}

func TestStore_GetTree(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Setup: guide with groups and rules
	guide := &ssg.SSGGuide{
		ID:          "test-guide",
		Product:     "test-product",
		ProfileID:   "test-profile",
		ShortID:     "test",
		Title:       "Test Guide",
		HTMLContent: "<html></html>",
	}
	if err := store.SaveGuide(guide); err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	// Create hierarchy: root group -> child group -> two rules
	groups := []*ssg.SSGGroup{
		{ID: "group-root", GuideID: "test-guide", ParentID: "", Title: "Root", Level: 0},
		{ID: "group-child", GuideID: "test-guide", ParentID: "group-root", Title: "Child", Level: 1},
	}
	for _, group := range groups {
		if err := store.SaveGroup(group); err != nil {
			t.Fatalf("SaveGroup() error = %v", err)
		}
	}

	rules := []*ssg.SSGRule{
		{ID: "rule1", GuideID: "test-guide", GroupID: "group-child", ShortID: "r1", Title: "R1", Severity: "low", Level: 2},
		{ID: "rule2", GuideID: "test-guide", GroupID: "group-child", ShortID: "r2", Title: "R2", Severity: "high", Level: 2},
	}
	for _, rule := range rules {
		if err := store.SaveRule(rule); err != nil {
			t.Fatalf("SaveRule() error = %v", err)
		}
	}

	// Get tree
	tree, err := store.GetTree("test-guide")
	if err != nil {
		t.Fatalf("GetTree() error = %v", err)
	}

	if tree.Guide.ID != "test-guide" {
		t.Errorf("Guide.ID = %s, want test-guide", tree.Guide.ID)
	}
	if len(tree.Groups) != 2 {
		t.Errorf("Groups count = %d, want 2", len(tree.Groups))
	}
	if len(tree.Rules) != 2 {
		t.Errorf("Rules count = %d, want 2", len(tree.Rules))
	}
}

func TestStore_BuildTreeNodes(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Setup: create guide, groups, rules
	guide := &ssg.SSGGuide{
		ID:          "test-guide",
		Product:     "test-product",
		ProfileID:   "test-profile",
		ShortID:     "test",
		Title:       "Test Guide",
		HTMLContent: "<html></html>",
	}
	if err := store.SaveGuide(guide); err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	// Create tree structure: root -> (child1, child2) -> rules
	groups := []*ssg.SSGGroup{
		{ID: "g-root", GuideID: "test-guide", ParentID: "", Title: "Root", Level: 0},
		{ID: "g-child1", GuideID: "test-guide", ParentID: "g-root", Title: "Child 1", Level: 1},
		{ID: "g-child2", GuideID: "test-guide", ParentID: "g-root", Title: "Child 2", Level: 1},
	}
	for _, group := range groups {
		if err := store.SaveGroup(group); err != nil {
			t.Fatalf("SaveGroup() error = %v", err)
		}
	}

	rules := []*ssg.SSGRule{
		{ID: "r1", GuideID: "test-guide", GroupID: "g-child1", ShortID: "r1", Title: "Rule 1", Severity: "low", Level: 2},
		{ID: "r2", GuideID: "test-guide", GroupID: "g-child1", ShortID: "r2", Title: "Rule 2", Severity: "high", Level: 2},
		{ID: "r3", GuideID: "test-guide", GroupID: "g-child2", ShortID: "r3", Title: "Rule 3", Severity: "medium", Level: 2},
	}
	for _, rule := range rules {
		if err := store.SaveRule(rule); err != nil {
			t.Fatalf("SaveRule() error = %v", err)
		}
	}

	// Build tree
	nodes, err := store.BuildTreeNodes("test-guide")
	if err != nil {
		t.Fatalf("BuildTreeNodes() error = %v", err)
	}

	// Debug: print all root IDs
	t.Logf("Number of roots: %d", len(nodes))
	for i, root := range nodes {
		t.Logf("Root %d: ID=%s, Type=%s, Children=%d", i, root.ID, root.Type, len(root.Children))
	}

	// Should have 1 root (g-root)
	if len(nodes) != 1 {
		t.Errorf("BuildTreeNodes() roots count = %d, want 1", len(nodes))
	}

	root := nodes[0]
	if root.ID != "g-root" {
		t.Errorf("Root ID = %s, want g-root", root.ID)
	}
	if root.Type != "group" {
		t.Errorf("Root Type = %s, want group", root.Type)
	}

	// Root should have 2 children (g-child1, g-child2)
	if len(root.Children) != 2 {
		t.Errorf("Root children count = %d, want 2", len(root.Children))
	}

	// Print all child IDs
	for i, child := range root.Children {
		t.Logf("Child %d: ID=%s, Type=%s", i, child.ID, child.Type)
	}

	// Find child by ID
	child1 := findChildPtr(root.Children, "g-child1")
	if child1 == nil {
		t.Fatal("Child g-child1 not found")
	}
	if len(child1.Children) != 2 {
		t.Errorf("Child1 children count = %d, want 2", len(child1.Children))
	}
}

func TestStore_ListRules(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Setup
	guide := &ssg.SSGGuide{ID: "test-guide", Product: "p1", ProfileID: "prof1", ShortID: "t", Title: "T", HTMLContent: "<html></html>"}
	if err := store.SaveGuide(guide); err != nil {
		t.Fatalf("SaveGuide() error = %v", err)
	}

	group := &ssg.SSGGroup{ID: "g1", GuideID: "test-guide", ParentID: "", Title: "G1", Level: 0}
	if err := store.SaveGroup(group); err != nil {
		t.Fatalf("SaveGroup() error = %v", err)
	}

	// Create rules with different severities
	rules := []*ssg.SSGRule{
		{ID: "r1", GuideID: "test-guide", GroupID: "g1", ShortID: "r1", Title: "R1", Severity: "low", Level: 1},
		{ID: "r2", GuideID: "test-guide", GroupID: "g1", ShortID: "r2", Title: "R2", Severity: "medium", Level: 1},
		{ID: "r3", GuideID: "test-guide", GroupID: "g1", ShortID: "r3", Title: "R3", Severity: "high", Level: 1},
	}
	for _, rule := range rules {
		if err := store.SaveRule(rule); err != nil {
			t.Fatalf("SaveRule() error = %v", err)
		}
	}

	tests := []struct {
		name     string
		groupID  string
		severity string
		wantCount int
	}{
		{
			name:     "all rules",
			groupID:  "",
			severity: "",
			wantCount: 3,
		},
		{
			name:     "filter by severity",
			groupID:  "",
			severity: "high",
			wantCount: 1,
		},
		{
			name:     "filter by group",
			groupID:  "g1",
			severity: "",
			wantCount: 3,
		},
		{
			name:     "filter by both",
			groupID:  "g1",
			severity: "low",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, total, err := store.ListRules(tt.groupID, tt.severity, 0, 100)
			if err != nil {
				t.Fatalf("ListRules() error = %v", err)
			}
			if total != int64(tt.wantCount) {
				t.Errorf("ListRules() total = %d, want %d", total, tt.wantCount)
			}
			if len(rules) != tt.wantCount {
				t.Errorf("ListRules() count = %d, want %d", len(rules), tt.wantCount)
			}
		})
	}
}

// setupTestStore creates a new store for testing.
func setupTestStore(t *testing.T) *Store {
	// Use in-memory database for faster tests
	store, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}
	return store
}

// findChildPtr finds a child node by ID in a slice of node pointers.
func findChildPtr(children []*ssg.TreeNode, id string) *ssg.TreeNode {
	for i := range children {
		if children[i].ID == id {
			return children[i]
		}
	}
	return nil
}

func TestStore_CrossReferences(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Create test cross-references
	refs := []ssg.SSGCrossReference{
		{
			SourceType: "guide",
			SourceID:   "guide-1",
			TargetType: "table",
			TargetID:   "table-1",
			LinkType:   "cce",
			Metadata:   `{"cce_number":"CCE-12345-6"}`,
		},
		{
			SourceType: "guide",
			SourceID:   "guide-1",
			TargetType: "datastream",
			TargetID:   "ds-1",
			LinkType:   "rule_id",
			Metadata:   `{"rule_short_id":"aide_installed"}`,
		},
		{
			SourceType: "manifest",
			SourceID:   "manifest-1",
			TargetType: "guide",
			TargetID:   "guide-1",
			LinkType:   "product",
			Metadata:   `{"product_name":"al2023"}`,
		},
	}

	// Test SaveCrossReferences
	err := store.SaveCrossReferences(refs)
	if err != nil {
		t.Fatalf("SaveCrossReferences() error = %v", err)
	}

	// Test GetCrossReferences (by source)
	sourceRefs, err := store.GetCrossReferences("guide", "guide-1", 0, 0)
	if err != nil {
		t.Fatalf("GetCrossReferences() error = %v", err)
	}
	if len(sourceRefs) != 2 {
		t.Errorf("GetCrossReferences() count = %d, want 2", len(sourceRefs))
	}

	// Test GetCrossReferencesByTarget
	targetRefs, err := store.GetCrossReferencesByTarget("guide", "guide-1", 0, 0)
	if err != nil {
		t.Fatalf("GetCrossReferencesByTarget() error = %v", err)
	}
	if len(targetRefs) != 1 {
		t.Errorf("GetCrossReferencesByTarget() count = %d, want 1", len(targetRefs))
	}

	// Test FindRelatedObjects (all)
	allRefs, err := store.FindRelatedObjects("guide", "guide-1", "", 0, 0)
	if err != nil {
		t.Fatalf("FindRelatedObjects() error = %v", err)
	}
	if len(allRefs) != 3 {
		t.Errorf("FindRelatedObjects() count = %d, want 3", len(allRefs))
	}

	// Test FindRelatedObjects with linkType filter
	ruleRefs, err := store.FindRelatedObjects("guide", "guide-1", "rule_id", 0, 0)
	if err != nil {
		t.Fatalf("FindRelatedObjects(rule_id) error = %v", err)
	}
	if len(ruleRefs) != 1 {
		t.Errorf("FindRelatedObjects(rule_id) count = %d, want 1", len(ruleRefs))
	}

	// Test pagination
	pagedRefs, err := store.GetCrossReferences("guide", "guide-1", 1, 0)
	if err != nil {
		t.Fatalf("GetCrossReferences(limit=1) error = %v", err)
	}
	if len(pagedRefs) != 1 {
		t.Errorf("GetCrossReferences(limit=1) count = %d, want 1", len(pagedRefs))
	}
}

func TestStore_SaveCrossReferences_Empty(t *testing.T) {
	store := setupTestStore(t)
	defer store.Close()

	// Test with empty slice - should not error
	err := store.SaveCrossReferences([]ssg.SSGCrossReference{})
	if err != nil {
		t.Errorf("SaveCrossReferences(empty) error = %v, want nil", err)
	}
}
