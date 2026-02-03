// Package local provides SQLite storage operations for SSG data.
package local

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cyw0ng95/v2e/pkg/ssg"
)

// Store manages SSG data storage in SQLite.
type Store struct {
	db *gorm.DB
}

// NewStore creates a new SSG store with the given database path.
// If the database file doesn't exist, it will be created.
func NewStore(dbPath string) (*Store, error) {
	if dbPath == "" {
		dbPath = DefaultDBPath()
	}

	// Open database with WAL mode for better performance
	dsn := fmt.Sprintf("%s?_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL&_pragma=foreign_keys=on", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-migrate schemas
	if err := db.AutoMigrate(&ssg.SSGGuide{}, &ssg.SSGGroup{}, &ssg.SSGRule{}, &ssg.SSGReference{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// SaveGuide saves or updates a guide in the database.
func (s *Store) SaveGuide(guide *ssg.SSGGuide) error {
	return s.db.Save(guide).Error
}

// GetGuide retrieves a guide by ID.
func (s *Store) GetGuide(id string) (*ssg.SSGGuide, error) {
	var guide ssg.SSGGuide
	result := s.db.First(&guide, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &guide, nil
}

// ListGuides lists guides with optional filters.
// product and profileID are optional filters; pass empty string to ignore.
func (s *Store) ListGuides(product, profileID string) ([]ssg.SSGGuide, error) {
	var guides []ssg.SSGGuide
	query := s.db.Model(&ssg.SSGGuide{})

	if product != "" {
		query = query.Where("product = ?", product)
	}
	if profileID != "" {
		query = query.Where("profile_id = ?", profileID)
	}

	result := query.Order("created_at DESC").Find(&guides)
	if result.Error != nil {
		return nil, result.Error
	}
	return guides, nil
}

// SaveGroup saves or updates a group in the database.
func (s *Store) SaveGroup(group *ssg.SSGGroup) error {
	return s.db.Save(group).Error
}

// GetGroup retrieves a group by ID.
func (s *Store) GetGroup(id string) (*ssg.SSGGroup, error) {
	var group ssg.SSGGroup
	result := s.db.First(&group, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &group, nil
}

// GetChildGroups retrieves direct child groups of a parent group.
// Use empty parentID to get top-level groups.
func (s *Store) GetChildGroups(parentID string) ([]ssg.SSGGroup, error) {
	var groups []ssg.SSGGroup
	query := s.db.Where("parent_id = ?", parentID)
	result := query.Order("title ASC").Find(&groups)
	if result.Error != nil {
		return nil, result.Error
	}
	return groups, nil
}

// GetRootGroups retrieves top-level groups for a guide (groups with no parent).
func (s *Store) GetRootGroups(guideID string) ([]ssg.SSGGroup, error) {
	var groups []ssg.SSGGroup
	result := s.db.Where("guide_id = ? AND parent_id = ?", guideID, "").Order("title ASC").Find(&groups)
	if result.Error != nil {
		return nil, result.Error
	}
	return groups, nil
}

// GetChildRules retrieves rules that are direct children of a group.
func (s *Store) GetChildRules(groupID string) ([]ssg.SSGRule, error) {
	var rules []ssg.SSGRule
	result := s.db.Where("group_id = ?", groupID).Preload("References").Order("title ASC").Find(&rules)
	if result.Error != nil {
		return nil, result.Error
	}
	return rules, nil
}

// SaveRule saves or updates a rule in the database.
// This will also save associated references.
func (s *Store) SaveRule(rule *ssg.SSGRule) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Save the rule
		if err := tx.Save(rule).Error; err != nil {
			return err
		}

		// Delete old references
		if err := tx.Where("rule_id = ?", rule.ID).Delete(&ssg.SSGReference{}).Error; err != nil {
			return err
		}

		// Save new references
		for _, ref := range rule.References {
			ref.RuleID = rule.ID
			if err := tx.Create(&ref).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// GetRule retrieves a rule by ID with references preloaded.
func (s *Store) GetRule(id string) (*ssg.SSGRule, error) {
	var rule ssg.SSGRule
	result := s.db.Preload("References").First(&rule, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &rule, nil
}

// ListRules lists rules with optional filters.
// groupID and severity are optional filters; pass empty string to ignore.
// offset and limit control pagination.
func (s *Store) ListRules(groupID, severity string, offset, limit int) ([]ssg.SSGRule, int64, error) {
	var rules []ssg.SSGRule
	var total int64

	query := s.db.Model(&ssg.SSGRule{})

	if groupID != "" {
		query = query.Where("group_id = ?", groupID)
	}
	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = 100 // Default limit
	}

	result := query.Preload("References").Offset(offset).Limit(limit).Order("title ASC").Find(&rules)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return rules, total, nil
}

// GetTree retrieves the complete tree structure for a guide.
// Returns all groups and rules for hierarchical display.
func (s *Store) GetTree(guideID string) (*ssg.SSGTree, error) {
	// Get the guide
	var guide ssg.SSGGuide
	if result := s.db.First(&guide, "id = ?", guideID); result.Error != nil {
		return nil, result.Error
	}

	// Get all groups for this guide
	var groups []ssg.SSGGroup
	if result := s.db.Where("guide_id = ?", guideID).Find(&groups); result.Error != nil {
		return nil, result.Error
	}

	// Get all rules for this guide
	var rules []ssg.SSGRule
	if result := s.db.Where("guide_id = ?", guideID).Preload("References").Find(&rules); result.Error != nil {
		return nil, result.Error
	}

	return &ssg.SSGTree{
		Guide:  guide,
		Groups: groups,
		Rules:  rules,
	}, nil
}

// BuildTreeNodes builds a tree structure from flat groups and rules.
// This is useful for constructing a hierarchical view for the frontend.
func (s *Store) BuildTreeNodes(guideID string) ([]*ssg.TreeNode, error) {
	tree, err := s.GetTree(guideID)
	if err != nil {
		return nil, err
	}

	// Create a map of ID to TreeNode pointers
	nodeMap := make(map[string]*ssg.TreeNode)

	// Add all group nodes
	for i := range tree.Groups {
		node := &ssg.TreeNode{
			ID:       tree.Groups[i].ID,
			ParentID: tree.Groups[i].ParentID,
			Level:    tree.Groups[i].Level,
			Type:     "group",
			Group:    &tree.Groups[i],
		}
		nodeMap[node.ID] = node
	}

	// Add all rule nodes
	for i := range tree.Rules {
		node := &ssg.TreeNode{
			ID:       tree.Rules[i].ID,
			ParentID: tree.Rules[i].GroupID,
			Level:    tree.Rules[i].Level,
			Type:     "rule",
			Rule:     &tree.Rules[i],
		}
		nodeMap[node.ID] = node
	}

	// Build tree structure using pointers
	var roots []*ssg.TreeNode
	for _, node := range nodeMap {
		if node.ParentID == "" {
			// This is a root node (top-level group)
			roots = append(roots, node)
		} else if parent, ok := nodeMap[node.ParentID]; ok {
			// Add to parent's children using pointer
			parent.Children = append(parent.Children, node)
		}
	}

	return roots, nil
}

// DeleteGuide deletes a guide and all associated groups and rules.
func (s *Store) DeleteGuide(id string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete references associated with rules in this guide
		if err := tx.Exec("DELETE FROM ssg_references WHERE rule_id IN (SELECT id FROM ssg_rules WHERE guide_id = ?)", id).Error; err != nil {
			return err
		}

		// Delete rules
		if err := tx.Where("guide_id = ?", id).Delete(&ssg.SSGRule{}).Error; err != nil {
			return err
		}

		// Delete groups
		if err := tx.Where("guide_id = ?", id).Delete(&ssg.SSGGroup{}).Error; err != nil {
			return err
		}

		// Delete guide
		if err := tx.Delete(&ssg.SSGGuide{}, "id = ?", id).Error; err != nil {
			return err
		}

		return nil
	})
}
