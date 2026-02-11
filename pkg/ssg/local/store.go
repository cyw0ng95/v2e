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
	dsn := fmt.Sprintf("%s?_pragma=journal_mode=WAL&_pragma=synchronous=NORMAL&_pragma=foreign_keys=on&_pragma=busy_timeout=30000", dbPath)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto-migrate schemas
	if err := db.AutoMigrate(
		&ssg.SSGGuide{},
		&ssg.SSGGroup{},
		&ssg.SSGRule{},
		&ssg.SSGReference{},
		&ssg.SSGTable{},
		&ssg.SSGTableEntry{},
		&ssg.SSGManifest{},
		&ssg.SSGProfile{},
		&ssg.SSGProfileRule{},
		&ssg.SSGDataStream{},
		&ssg.SSGBenchmark{},
		&ssg.SSGDSProfile{},
		&ssg.SSGDSProfileRule{},
		&ssg.SSGDSGroup{},
		&ssg.SSGDSRule{},
		&ssg.SSGDSRuleReference{},
		&ssg.SSGDSRuleIdentifier{},
		&ssg.SSGCrossReference{},
	); err != nil {
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

// SaveTable saves or updates a table in the database.
func (s *Store) SaveTable(table *ssg.SSGTable) error {
	return s.db.Save(table).Error
}

// GetTable retrieves a table by ID.
func (s *Store) GetTable(id string) (*ssg.SSGTable, error) {
	var table ssg.SSGTable
	result := s.db.First(&table, "id = ?", id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &table, nil
}

// ListTables lists tables with optional filters.
// product and tableType are optional filters; pass empty string to ignore.
func (s *Store) ListTables(product, tableType string) ([]ssg.SSGTable, error) {
	var tables []ssg.SSGTable
	query := s.db.Model(&ssg.SSGTable{})

	if product != "" {
		query = query.Where("product = ?", product)
	}
	if tableType != "" {
		query = query.Where("table_type = ?", tableType)
	}

	result := query.Order("created_at DESC").Find(&tables)
	if result.Error != nil {
		return nil, result.Error
	}
	return tables, nil
}

// SaveTableEntry saves or updates a table entry in the database.
func (s *Store) SaveTableEntry(entry *ssg.SSGTableEntry) error {
	return s.db.Save(entry).Error
}

// GetTableEntries retrieves all entries for a table.
// offset and limit control pagination.
func (s *Store) GetTableEntries(tableID string, offset, limit int) ([]ssg.SSGTableEntry, int64, error) {
	var entries []ssg.SSGTableEntry
	var total int64

	query := s.db.Model(&ssg.SSGTableEntry{}).Where("table_id = ?", tableID)

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

	result := query.Offset(offset).Limit(limit).Order("id ASC").Find(&entries)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	return entries, total, nil
}

// SaveManifest saves a manifest and its associated profiles and profile rules.
// This is an atomic operation - all or nothing.
func (s *Store) SaveManifest(manifest *ssg.SSGManifest, profiles []ssg.SSGProfile, profileRules []ssg.SSGProfileRule) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Save manifest
		if err := tx.Save(manifest).Error; err != nil {
			return fmt.Errorf("failed to save manifest: %w", err)
		}

		// Delete existing profiles for this manifest
		if err := tx.Where("manifest_id = ?", manifest.ID).Delete(&ssg.SSGProfile{}).Error; err != nil {
			return fmt.Errorf("failed to delete old profiles: %w", err)
		}

		// Delete existing profile rules for profiles of this manifest
		if err := tx.Where("profile_id LIKE ?", manifest.Product+":%").Delete(&ssg.SSGProfileRule{}).Error; err != nil {
			return fmt.Errorf("failed to delete old profile rules: %w", err)
		}

		// Save new profiles
		if len(profiles) > 0 {
			if err := tx.Create(&profiles).Error; err != nil {
				return fmt.Errorf("failed to save profiles: %w", err)
			}
		}

		// Save profile rules in batches
		if len(profileRules) > 0 {
			if err := tx.CreateInBatches(&profileRules, 100).Error; err != nil {
				return fmt.Errorf("failed to save profile rules: %w", err)
			}
		}

		return nil
	})
}

// GetManifest retrieves a manifest by ID.
func (s *Store) GetManifest(id string) (*ssg.SSGManifest, error) {
	var manifest ssg.SSGManifest
	if err := s.db.First(&manifest, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &manifest, nil
}

// ListManifests retrieves all manifests, optionally filtered by product.
func (s *Store) ListManifests(product string, limit, offset int) ([]ssg.SSGManifest, error) {
	var manifests []ssg.SSGManifest
	query := s.db.Order("product ASC, id ASC")

	if product != "" {
		query = query.Where("product = ?", product)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&manifests).Error; err != nil {
		return nil, err
	}
	return manifests, nil
}

// ListProfiles retrieves profiles, optionally filtered by product or profile ID.
func (s *Store) ListProfiles(product, profileID string, limit, offset int) ([]ssg.SSGProfile, error) {
	var profiles []ssg.SSGProfile
	query := s.db.Order("product ASC, profile_id ASC")

	if product != "" {
		query = query.Where("product = ?", product)
	}
	if profileID != "" {
		query = query.Where("profile_id = ?", profileID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetProfile retrieves a profile by ID.
func (s *Store) GetProfile(id string) (*ssg.SSGProfile, error) {
	var profile ssg.SSGProfile
	if err := s.db.First(&profile, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetProfileRules retrieves all rule short IDs for a given profile.
func (s *Store) GetProfileRules(profileID string, limit, offset int) ([]ssg.SSGProfileRule, error) {
	var profileRules []ssg.SSGProfileRule
	query := s.db.Where("profile_id = ?", profileID).Order("rule_short_id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&profileRules).Error; err != nil {
		return nil, err
	}
	return profileRules, nil
}

// SaveDataStream saves a data stream and all its associated components.
// This is an atomic operation that saves: data stream, benchmark, profiles, profile rules, groups, rules, references, and identifiers.
func (s *Store) SaveDataStream(
	ds *ssg.SSGDataStream,
	benchmark *ssg.SSGBenchmark,
	profiles []ssg.SSGDSProfile,
	profileRules []ssg.SSGDSProfileRule,
	groups []ssg.SSGDSGroup,
	rules []ssg.SSGDSRule,
	references []ssg.SSGDSRuleReference,
	identifiers []ssg.SSGDSRuleIdentifier,
) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Save data stream
		if err := tx.Save(ds).Error; err != nil {
			return fmt.Errorf("failed to save data stream: %w", err)
		}

		// Delete existing components for this data stream
		if err := tx.Where("data_stream_id = ?", ds.ID).Delete(&ssg.SSGBenchmark{}).Error; err != nil {
			return fmt.Errorf("failed to delete old benchmark: %w", err)
		}
		if err := tx.Where("data_stream_id = ?", ds.ID).Delete(&ssg.SSGDSProfile{}).Error; err != nil {
			return fmt.Errorf("failed to delete old profiles: %w", err)
		}
		if err := tx.Where("data_stream_id = ?", ds.ID).Delete(&ssg.SSGDSGroup{}).Error; err != nil {
			return fmt.Errorf("failed to delete old groups: %w", err)
		}
		if err := tx.Where("data_stream_id = ?", ds.ID).Delete(&ssg.SSGDSRule{}).Error; err != nil {
			return fmt.Errorf("failed to delete old rules: %w", err)
		}

		// Save benchmark
		if benchmark != nil {
			if err := tx.Save(benchmark).Error; err != nil {
				return fmt.Errorf("failed to save benchmark: %w", err)
			}
		}

		// Save profiles
		if len(profiles) > 0 {
			if err := tx.Create(&profiles).Error; err != nil {
				return fmt.Errorf("failed to save profiles: %w", err)
			}
		}

		// Save profile rules in batches
		if len(profileRules) > 0 {
			if err := tx.CreateInBatches(&profileRules, 100).Error; err != nil {
				return fmt.Errorf("failed to save profile rules: %w", err)
			}
		}

		// Save groups in batches
		if len(groups) > 0 {
			if err := tx.CreateInBatches(&groups, 100).Error; err != nil {
				return fmt.Errorf("failed to save groups: %w", err)
			}
		}

		// Save rules in batches
		if len(rules) > 0 {
			if err := tx.CreateInBatches(&rules, 100).Error; err != nil {
				return fmt.Errorf("failed to save rules: %w", err)
			}
		}

		// Save references in batches
		if len(references) > 0 {
			if err := tx.CreateInBatches(&references, 500).Error; err != nil {
				return fmt.Errorf("failed to save references: %w", err)
			}
		}

		// Save identifiers in batches
		if len(identifiers) > 0 {
			if err := tx.CreateInBatches(&identifiers, 100).Error; err != nil {
				return fmt.Errorf("failed to save identifiers: %w", err)
			}
		}

		return nil
	})
}

// GetDataStream retrieves a data stream by ID.
func (s *Store) GetDataStream(id string) (*ssg.SSGDataStream, error) {
	var ds ssg.SSGDataStream
	if err := s.db.First(&ds, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

// ListDataStreams retrieves all data streams, optionally filtered by product.
func (s *Store) ListDataStreams(product string, limit, offset int) ([]ssg.SSGDataStream, error) {
	var dataStreams []ssg.SSGDataStream
	query := s.db.Order("product ASC, id ASC")

	if product != "" {
		query = query.Where("product = ?", product)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&dataStreams).Error; err != nil {
		return nil, err
	}
	return dataStreams, nil
}

// GetBenchmark retrieves the benchmark for a data stream.
func (s *Store) GetBenchmark(dataStreamID string) (*ssg.SSGBenchmark, error) {
	var benchmark ssg.SSGBenchmark
	if err := s.db.Where("data_stream_id = ?", dataStreamID).First(&benchmark).Error; err != nil {
		return nil, err
	}
	return &benchmark, nil
}

// ListDSProfiles retrieves all profiles for a data stream.
func (s *Store) ListDSProfiles(dataStreamID string, limit, offset int) ([]ssg.SSGDSProfile, error) {
	var profiles []ssg.SSGDSProfile
	query := s.db.Where("data_stream_id = ?", dataStreamID).Order("profile_id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// GetDSProfile retrieves a specific profile from a data stream.
func (s *Store) GetDSProfile(id string) (*ssg.SSGDSProfile, error) {
	var profile ssg.SSGDSProfile
	if err := s.db.First(&profile, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &profile, nil
}

// GetDSProfileRules retrieves all rule selections for a profile.
func (s *Store) GetDSProfileRules(profileID string, limit, offset int) ([]ssg.SSGDSProfileRule, error) {
	var profileRules []ssg.SSGDSProfileRule
	query := s.db.Where("profile_id = ?", profileID).Order("rule_id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&profileRules).Error; err != nil {
		return nil, err
	}
	return profileRules, nil
}

// ListDSGroups retrieves all groups for a data stream.
func (s *Store) ListDSGroups(dataStreamID string, limit, offset int) ([]ssg.SSGDSGroup, error) {
	var groups []ssg.SSGDSGroup
	query := s.db.Where("data_stream_id = ?", dataStreamID).Order("level ASC, title ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil
}

// ListDSRules retrieves all rules for a data stream with optional filters.
func (s *Store) ListDSRules(dataStreamID string, severity string, limit, offset int) ([]ssg.SSGDSRule, int64, error) {
	var rules []ssg.SSGDSRule
	var total int64

	query := s.db.Model(&ssg.SSGDSRule{}).Where("data_stream_id = ?", dataStreamID)

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	if err := query.Offset(offset).Limit(limit).Order("title ASC").Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// GetDSRule retrieves a specific rule by ID.
func (s *Store) GetDSRule(id string) (*ssg.SSGDSRule, error) {
	var rule ssg.SSGDSRule
	if err := s.db.First(&rule, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &rule, nil
}

// GetDSRuleReferences retrieves all references for a rule.
func (s *Store) GetDSRuleReferences(ruleID string, limit, offset int) ([]ssg.SSGDSRuleReference, error) {
	var references []ssg.SSGDSRuleReference
	query := s.db.Where("rule_id = ?", ruleID).Order("href ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&references).Error; err != nil {
		return nil, err
	}
	return references, nil
}

// GetDSRuleIdentifiers retrieves all identifiers for a rule.
func (s *Store) GetDSRuleIdentifiers(ruleID string) ([]ssg.SSGDSRuleIdentifier, error) {
	var identifiers []ssg.SSGDSRuleIdentifier
	if err := s.db.Where("rule_id = ?", ruleID).Order("system ASC").Find(&identifiers).Error; err != nil {
		return nil, err
	}
	return identifiers, nil
}

// SaveCrossReferences saves cross-references in batches.
// This is used during import to create links between SSG objects.
func (s *Store) SaveCrossReferences(refs []ssg.SSGCrossReference) error {
	if len(refs) == 0 {
		return nil
	}

	// Use batch insert for performance
	if err := s.db.CreateInBatches(&refs, 500).Error; err != nil {
		return fmt.Errorf("failed to save cross-references: %w", err)
	}
	return nil
}

// GetCrossReferences retrieves cross-references for a given source object.
// Returns all links where the object is the source.
func (s *Store) GetCrossReferences(sourceType, sourceID string, limit, offset int) ([]ssg.SSGCrossReference, error) {
	var refs []ssg.SSGCrossReference
	query := s.db.Where("source_type = ? AND source_id = ?", sourceType, sourceID).
		Order("link_type ASC, target_type ASC, target_id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

// GetCrossReferencesByTarget retrieves cross-references where the object is the target.
// This finds all objects that reference the given object.
func (s *Store) GetCrossReferencesByTarget(targetType, targetID string, limit, offset int) ([]ssg.SSGCrossReference, error) {
	var refs []ssg.SSGCrossReference
	query := s.db.Where("target_type = ? AND target_id = ?", targetType, targetID).
		Order("link_type ASC, source_type ASC, source_id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

// FindRelatedObjects finds all objects related to the given object via cross-references.
// This includes both outgoing (source) and incoming (target) references.
func (s *Store) FindRelatedObjects(objectType, objectID string, linkType string, limit, offset int) ([]ssg.SSGCrossReference, error) {
	var refs []ssg.SSGCrossReference
	query := s.db.Where(
		"((source_type = ? AND source_id = ?) OR (target_type = ? AND target_id = ?))",
		objectType, objectID, objectType, objectID,
	)

	if linkType != "" {
		query = query.Where("link_type = ?", linkType)
	}

	query = query.Order("link_type ASC, source_type ASC, source_id ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}
