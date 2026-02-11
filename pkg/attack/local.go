package attack

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LocalAttackStore manages a local database of ATT&CK items
type LocalAttackStore struct {
	db *gorm.DB
}

// NewLocalAttackStore creates or opens a local ATT&CK database at dbPath
func NewLocalAttackStore(dbPath string) (*LocalAttackStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Configure database connection pool
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(time.Hour)
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA synchronous=NORMAL")
		db.Exec("PRAGMA cache_size=-40000")
	}

	// AutoMigrate all ATT&CK tables
	err = db.AutoMigrate(
		&AttackTechnique{},
		&AttackTactic{},
		&AttackMitigation{},
		&AttackSoftware{},
		&AttackGroup{},
		&AttackRelationship{},
		&AttackMetadata{},
	)
	if err != nil {
		return nil, err
	}

	return &LocalAttackStore{db: db}, nil
}

// ImportFromXLSX reads ATT&CK data from an Excel file and imports it into the database
func (s *LocalAttackStore) ImportFromXLSX(xlsxPath string, force bool) error {
	// Check if file exists
	if _, err := os.Stat(xlsxPath); os.IsNotExist(err) {
		return fmt.Errorf("XLSX file does not exist: %s", xlsxPath)
	}

	file, err := excelize.OpenFile(xlsxPath)
	if err != nil {
		return fmt.Errorf("failed to open XLSX file: %v", err)
	}
	defer file.Close()

	// Start a transaction for atomic import
	tx := s.db.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// Clean up existing data if force flag is true
	if force {
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AttackTechnique{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear techniques: %v", err)
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AttackTactic{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear tactics: %v", err)
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AttackMitigation{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear mitigations: %v", err)
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AttackSoftware{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear software: %v", err)
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AttackGroup{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear groups: %v", err)
		}
		if err := tx.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&AttackRelationship{}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to clear relationships: %v", err)
		}
	}

	// Get sheet names
	sheetNames := file.GetSheetMap()
	totalRecords := 0

	// Process each sheet in the XLSX file
	for sheetIndex, sheetName := range sheetNames {
		rows, err := file.GetRows(sheetName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to read sheet '%s': %v", sheetName, err)
		}

		if len(rows) == 0 {
			continue // Skip empty sheets
		}

		// Assuming first row contains headers
		headers := rows[0]
		for i := 1; i < len(rows); i++ {
			row := rows[i]

			// Determine the sheet type based on sheet name
			// Common ATT&CK sheet names: Techniques, Tactics, Mitigations, Software, Groups, Relationships
			switch strings.ToLower(sheetName) {
			case "techniques", "technique", "attack_techniques", "attacks":
				if len(row) >= 6 { // Ensure row has enough columns
					technique := &AttackTechnique{
						ID:          getStringValue(row, 0, headers, "ID"),
						Name:        getStringValue(row, 1, headers, "Name"),
						Description: getStringValue(row, 2, headers, "Description"),
						Domain:      getStringValue(row, 3, headers, "Domain"),
						Platform:    getStringValue(row, 4, headers, "Platform"),
						Created:     getStringValue(row, 5, headers, "Created"),
						Modified:    getStringValue(row, 6, headers, "Modified", "Last Modified"),
						Revoked:     getBoolValue(row, getStringIndex(headers, []string{"Revoked", "Is Revoked", "Revoked?"})),
						Deprecated:  getBoolValue(row, getStringIndex(headers, []string{"Deprecated", "Is Deprecated", "Deprecated?"})),
					}

					// Validate required fields
					if technique.ID != "" {
						if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(technique).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert technique: %v", err)
						}
						totalRecords++
					}
				}
			case "tactics", "tactic", "attack_tactics":
				if len(row) >= 5 { // Ensure row has enough columns
					tactic := &AttackTactic{
						ID:          getStringValue(row, 0, headers, "ID"),
						Name:        getStringValue(row, 1, headers, "Name"),
						Description: getStringValue(row, 2, headers, "Description"),
						Domain:      getStringValue(row, 3, headers, "Domain"),
						Created:     getStringValue(row, 4, headers, "Created"),
						Modified:    getStringValue(row, 5, headers, "Modified", "Last Modified"),
					}

					// Validate required fields
					if tactic.ID != "" {
						if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(tactic).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert tactic: %v", err)
						}
						totalRecords++
					}
				}
			case "mitigations", "mitigation", "attack_mitigations":
				if len(row) >= 5 { // Ensure row has enough columns
					mitigation := &AttackMitigation{
						ID:          getStringValue(row, 0, headers, "ID"),
						Name:        getStringValue(row, 1, headers, "Name"),
						Description: getStringValue(row, 2, headers, "Description"),
						Domain:      getStringValue(row, 3, headers, "Domain"),
						Created:     getStringValue(row, 4, headers, "Created"),
						Modified:    getStringValue(row, 5, headers, "Modified", "Last Modified"),
					}

					// Validate required fields
					if mitigation.ID != "" {
						if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(mitigation).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert mitigation: %v", err)
						}
						totalRecords++
					}
				}
			case "software", "attack_software":
				if len(row) >= 6 { // Ensure row has enough columns
					software := &AttackSoftware{
						ID:          getStringValue(row, 0, headers, "ID"),
						Name:        getStringValue(row, 1, headers, "Name"),
						Description: getStringValue(row, 2, headers, "Description"),
						Type:        getStringValue(row, 3, headers, "Type"),
						Domain:      getStringValue(row, 4, headers, "Domain"),
						Created:     getStringValue(row, 5, headers, "Created"),
						Modified:    getStringValue(row, 6, headers, "Modified", "Last Modified"),
					}

					// Validate required fields
					if software.ID != "" {
						if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(software).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert software: %v", err)
						}
						totalRecords++
					}
				}
			case "groups", "attack_groups", "adversary_groups":
				if len(row) >= 5 { // Ensure row has enough columns
					group := &AttackGroup{
						ID:          getStringValue(row, 0, headers, "ID"),
						Name:        getStringValue(row, 1, headers, "Name"),
						Description: getStringValue(row, 2, headers, "Description"),
						Domain:      getStringValue(row, 3, headers, "Domain"),
						Created:     getStringValue(row, 4, headers, "Created"),
						Modified:    getStringValue(row, 5, headers, "Modified", "Last Modified"),
					}

					// Validate required fields
					if group.ID != "" {
						if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(group).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert group: %v", err)
						}
						totalRecords++
					}
				}
			case "relationships", "attack_relationships", "relations":
				if len(row) >= 7 { // Ensure row has enough columns
					relationship := &AttackRelationship{
						ID:               fmt.Sprintf("%d_%s_%s", sheetIndex, getStringValue(row, 0, headers, "SourceRef"), getStringValue(row, 1, headers, "TargetRef")),
						SourceRef:        getStringValue(row, 0, headers, "SourceRef"),
						TargetRef:        getStringValue(row, 1, headers, "TargetRef"),
						RelationshipType: getStringValue(row, 2, headers, "RelationshipType"),
						SourceObjectType: getStringValue(row, 3, headers, "SourceObjectType"),
						TargetObjectType: getStringValue(row, 4, headers, "TargetObjectType"),
						Description:      getStringValue(row, 5, headers, "Description"),
						Domain:           getStringValue(row, 6, headers, "Domain"),
						Created:          getStringValue(row, 7, headers, "Created"),
						Modified:         getStringValue(row, 8, headers, "Modified", "Last Modified"),
					}

					// Validate required fields
					if relationship.SourceRef != "" && relationship.TargetRef != "" {
						if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(relationship).Error; err != nil {
							tx.Rollback()
							return fmt.Errorf("failed to insert relationship: %v", err)
						}
						totalRecords++
					}
				}
			}
		}
	}

	// Record import metadata
	meta := &AttackMetadata{
		ImportedAt:    time.Now().Unix(),
		SourceFile:    xlsxPath,
		TotalRecords:  totalRecords,
		ImportVersion: "1.0", // Could be derived from XLSX metadata
	}

	// Insert or update import metadata
	if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(meta).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to record import metadata: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetTechniqueByID returns an ATT&CK technique by its ID (e.g. "T1001")
func (s *LocalAttackStore) GetTechniqueByID(ctx context.Context, id string) (*AttackTechnique, error) {
	var technique AttackTechnique
	if err := s.db.WithContext(ctx).First(&technique, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &technique, nil
}

// GetTacticByID returns an ATT&CK tactic by its ID (e.g. "TA0001")
func (s *LocalAttackStore) GetTacticByID(ctx context.Context, id string) (*AttackTactic, error) {
	var tactic AttackTactic
	if err := s.db.WithContext(ctx).First(&tactic, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &tactic, nil
}

// GetMitigationByID returns an ATT&CK mitigation by its ID (e.g. "M1001")
func (s *LocalAttackStore) GetMitigationByID(ctx context.Context, id string) (*AttackMitigation, error) {
	var mitigation AttackMitigation
	if err := s.db.WithContext(ctx).First(&mitigation, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &mitigation, nil
}

// GetSoftwareByID returns an ATT&CK software by its ID (e.g. "S0001")
func (s *LocalAttackStore) GetSoftwareByID(ctx context.Context, id string) (*AttackSoftware, error) {
	var software AttackSoftware
	if err := s.db.WithContext(ctx).First(&software, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &software, nil
}

// GetGroupByID returns an ATT&CK group by its ID (e.g. "G0001")
func (s *LocalAttackStore) GetGroupByID(ctx context.Context, id string) (*AttackGroup, error) {
	var group AttackGroup
	if err := s.db.WithContext(ctx).First(&group, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &group, nil
}

// ListTechniquesPaginated returns ATT&CK techniques with pagination
func (s *LocalAttackStore) ListTechniquesPaginated(ctx context.Context, offset, limit int) ([]AttackTechnique, int64, error) {
	var techniques []AttackTechnique
	var total int64

	if err := s.db.WithContext(ctx).Model(&AttackTechnique{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).Order("id asc").Offset(offset).Limit(limit).Find(&techniques).Error; err != nil {
		return nil, 0, err
	}

	return techniques, total, nil
}

// ListTacticsPaginated returns ATT&CK tactics with pagination
func (s *LocalAttackStore) ListTacticsPaginated(ctx context.Context, offset, limit int) ([]AttackTactic, int64, error) {
	var tactics []AttackTactic
	var total int64

	if err := s.db.WithContext(ctx).Model(&AttackTactic{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).Order("id asc").Offset(offset).Limit(limit).Find(&tactics).Error; err != nil {
		return nil, 0, err
	}

	return tactics, total, nil
}

// ListMitigationsPaginated returns ATT&CK mitigations with pagination
func (s *LocalAttackStore) ListMitigationsPaginated(ctx context.Context, offset, limit int) ([]AttackMitigation, int64, error) {
	var mitigations []AttackMitigation
	var total int64

	if err := s.db.WithContext(ctx).Model(&AttackMitigation{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).Order("id asc").Offset(offset).Limit(limit).Find(&mitigations).Error; err != nil {
		return nil, 0, err
	}

	return mitigations, total, nil
}

// ListSoftwarePaginated returns ATT&CK software with pagination
func (s *LocalAttackStore) ListSoftwarePaginated(ctx context.Context, offset, limit int) ([]AttackSoftware, int64, error) {
	var software []AttackSoftware
	var total int64

	if err := s.db.WithContext(ctx).Model(&AttackSoftware{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).Order("id asc").Offset(offset).Limit(limit).Find(&software).Error; err != nil {
		return nil, 0, err
	}

	return software, total, nil
}

// ListGroupsPaginated returns ATT&CK groups with pagination
func (s *LocalAttackStore) ListGroupsPaginated(ctx context.Context, offset, limit int) ([]AttackGroup, int64, error) {
	var groups []AttackGroup
	var total int64

	if err := s.db.WithContext(ctx).Model(&AttackGroup{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := s.db.WithContext(ctx).Order("id asc").Offset(offset).Limit(limit).Find(&groups).Error; err != nil {
		return nil, 0, err
	}

	return groups, total, nil
}

// GetImportMetadata returns the stored ATT&CK import metadata
func (s *LocalAttackStore) GetImportMetadata(ctx context.Context) (*AttackMetadata, error) {
	var meta AttackMetadata
	if err := s.db.WithContext(ctx).Order("id desc").First(&meta).Error; err != nil {
		return nil, err
	}
	return &meta, nil
}

// Helper functions for parsing Excel data

// getStringValue safely extracts a string value from a row by column index or header name.
// It handles nil slices and out-of-bounds access gracefully to prevent panics.
func getStringValue(row []string, colIndex int, headers []string, possibleHeaders ...string) string {
	// Guard against nil or empty headers
	if headers == nil || len(headers) == 0 {
		// Fallback to index if headers are not available
		if row != nil && colIndex >= 0 && colIndex < len(row) {
			return strings.TrimSpace(row[colIndex])
		}
		return ""
	}

	// First, try to find the column by header name
	for i, header := range headers {
		// Guard against nil header entries
		if header == "" {
			continue
		}
		headerLower := strings.ToLower(strings.TrimSpace(header))
		for _, possibleHeader := range possibleHeaders {
			if possibleHeader == "" {
				continue
			}
			if headerLower == strings.ToLower(strings.TrimSpace(possibleHeader)) {
				// Safely access row with bounds checking
				if row != nil && i >= 0 && i < len(row) {
					return strings.TrimSpace(row[i])
				}
				return ""
			}
		}
	}

	// Fallback to index if headers don't match or row doesn't have enough columns
	if row != nil && colIndex >= 0 && colIndex < len(row) {
		return strings.TrimSpace(row[colIndex])
	}

	return ""
}

// getStringIndex finds the index of a header in the headers slice.
// It returns -1 if not found or if headers is nil/empty.
func getStringIndex(headers []string, possibleHeaders []string) int {
	// Guard against nil or empty headers
	if headers == nil || len(headers) == 0 {
		return -1
	}
	if possibleHeaders == nil || len(possibleHeaders) == 0 {
		return -1
	}

	for i, header := range headers {
		// Guard against nil header entries
		if header == "" {
			continue
		}
		headerLower := strings.ToLower(strings.TrimSpace(header))
		for _, possibleHeader := range possibleHeaders {
			if possibleHeader == "" {
				continue
			}
			if headerLower == strings.ToLower(strings.TrimSpace(possibleHeader)) {
				return i
			}
		}
	}
	return -1 // Not found
}

// getBoolValue safely extracts a boolean value from a row by column index.
// It handles nil slices and out-of-bounds access gracefully to prevent panics.
func getBoolValue(row []string, colIndex int) bool {
	// Guard against nil row and invalid indices
	if row == nil || colIndex < 0 || colIndex >= len(row) {
		return false
	}

	value := strings.TrimSpace(strings.ToLower(row[colIndex]))
	switch value {
	case "true", "1", "yes", "y", "t":
		return true
	default:
		return false
	}
}

// GetRelatedTechniquesByTactic returns techniques associated with a specific tactic
func (s *LocalAttackStore) GetRelatedTechniquesByTactic(ctx context.Context, tacticID string) ([]AttackTechnique, error) {
	var relationships []AttackRelationship
	if err := s.db.WithContext(ctx).Where("target_ref = ? AND relationship_type = ?", tacticID, "mitigates").Or("source_ref = ? AND relationship_type = ?", tacticID, "has-subtechnique").Find(&relationships).Error; err != nil {
		return nil, err
	}

	var techniques []AttackTechnique
	for _, rel := range relationships {
		var technique AttackTechnique
		// For "mitigates" relationships, technique ID is in SourceRef
		// For "has-subtechnique" relationships, technique ID is in TargetRef
		techniqueID := rel.SourceRef
		if rel.RelationshipType == "has-subtechnique" {
			techniqueID = rel.TargetRef
		}

		if err := s.db.WithContext(ctx).Where("id = ?", techniqueID).First(&technique).Error; err != nil {
			continue // Skip if technique not found
		}
		techniques = append(techniques, technique)
	}

	return techniques, nil
}
