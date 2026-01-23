package attack

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blevesearch/bleve/v2"
	"github.com/xuri/excelize/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// IndexOperation represents an indexing operation
type IndexOperation struct {
	DocType string
	DocID   string
	Item    interface{}
	Action  string // "index" or "delete"
}

// LocalAttackStore manages a local database of ATT&CK items
type LocalAttackStore struct {
	db            *gorm.DB
	index         bleve.Index
	indexQueue    chan IndexOperation
	closeIndexing chan struct{}
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

	// Create Bleve index
	indexPath := strings.TrimSuffix(dbPath, ".db") + "_index.bleve"
	index, err := createBleveIndex(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create bleve index: %v", err)
	}

	// Create indexing queue and start async indexer
	store := &LocalAttackStore{
		db:            db,
		index:         index,
		indexQueue:    make(chan IndexOperation, 100), // Buffered channel
		closeIndexing: make(chan struct{}),
	}

	// Start the async indexer goroutine
	go store.runAsyncIndexer()

	return store, nil
}

// createBleveIndex creates a new Bleve index for ATT&CK data
func createBleveIndex(indexPath string) (bleve.Index, error) {
	// Check if index already exists
	if _, err := os.Stat(indexPath); err == nil {
		// Open existing index
		return bleve.Open(indexPath)
	}

	// Create new index mapping
	mapping := bleve.NewIndexMapping()

	// Create the index
	index, err := bleve.New(indexPath, mapping)
	if err != nil {
		return nil, err
	}

	return index, nil
}

// indexItem adds an item to the indexing queue (async)
func (s *LocalAttackStore) indexItem(docType, docID string, item interface{}) error {
	op := IndexOperation{
		DocType: docType,
		DocID:   docID,
		Item:    item,
		Action:  "index",
	}

	// Send to indexing queue (non-blocking unless queue is full)
	select {
	case s.indexQueue <- op:
		return nil // Successfully queued
	default:
		// Queue is full, log warning but don't block
		fmt.Printf("Warning: indexing queue is full, skipping document %s_%s\n", docType, docID)
		return nil // Return nil to not affect the main operation
	}
}

// runAsyncIndexer runs the async indexing process
func (s *LocalAttackStore) runAsyncIndexer() {
	for {
		select {
		case op := <-s.indexQueue:
			docIDWithPrefix := op.DocType + "_" + op.DocID
			var err error

			switch op.Action {
			case "delete":
				err = s.index.Delete(docIDWithPrefix)
			default: // "index" or any other action defaults to indexing
				err = s.index.Index(docIDWithPrefix, op.Item)
			}

			if err != nil {
				// In a production system, we might want to log this to a persistent log
				// or have a retry mechanism for failed indexing operations
				fmt.Printf("Warning: failed to %s document %s in index: %v\n", op.Action, docIDWithPrefix, err)
			}
		case <-s.closeIndexing:
			// Close signal received, exit the goroutine
			return
		}
	}
}

// closeIndex closes the Bleve index
func (s *LocalAttackStore) CloseIndex() error {
	// Signal the indexer to stop
	close(s.closeIndexing)

	// Wait a bit for pending operations to complete
	time.Sleep(100 * time.Millisecond)

	if s.index != nil {
		return s.index.Close()
	}
	return nil
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
						// Index in Bleve
						if err := s.indexItem("technique", technique.ID, technique); err != nil {
							// Log the error but don't fail the whole import
							fmt.Printf("Warning: failed to index technique %s: %v\n", technique.ID, err)
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
						// Index in Bleve
						if err := s.indexItem("tactic", tactic.ID, tactic); err != nil {
							// Log the error but don't fail the whole import
							fmt.Printf("Warning: failed to index tactic %s: %v\n", tactic.ID, err)
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
						// Index in Bleve
						if err := s.indexItem("mitigation", mitigation.ID, mitigation); err != nil {
							// Log the error but don't fail the whole import
							fmt.Printf("Warning: failed to index mitigation %s: %v\n", mitigation.ID, err)
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
						// Index in Bleve
						if err := s.indexItem("software", software.ID, software); err != nil {
							// Log the error but don't fail the whole import
							fmt.Printf("Warning: failed to index software %s: %v\n", software.ID, err)
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
						// Index in Bleve
						if err := s.indexItem("group", group.ID, group); err != nil {
							// Log the error but don't fail the whole import
							fmt.Printf("Warning: failed to index group %s: %v\n", group.ID, err)
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

// SearchTechniques searches for ATT&CK techniques by text with enhanced intelligence
func (s *LocalAttackStore) SearchTechniques(ctx context.Context, search string, offset, limit int) ([]AttackTechnique, int64, error) {
	if search == "" {
		// Fall back to pagination if no search term
		return s.ListTechniquesPaginated(ctx, offset, limit)
	}

	// Create a more intelligent search query that handles various search patterns
	query := createIntelligentSearchQuery(search, "technique")
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.From = offset
	searchRequest.Size = limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlight()

	// Enable relevance scoring and sorting
	searchRequest.SortBy([]string{"-_score"}) // Sort by relevance score

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var techniques []AttackTechnique
	for _, hit := range searchResults.Hits {
		id := strings.TrimPrefix(hit.ID, "technique_")
		technique, err := s.GetTechniqueByID(ctx, id)
		if err != nil {
			continue // Skip if technique not found in DB
		}
		techniques = append(techniques, *technique)
	}

	return techniques, int64(searchResults.Total), nil
}

// createIntelligentSearchQuery creates a more intelligent search query with multiple matching strategies
func createIntelligentSearchQuery(searchTerm string, docType string) bleve.Query {
	// If search term is an ID (starts with T, TA, M, S, G followed by digits), do exact match
	if strings.HasPrefix(searchTerm, "T") && len(searchTerm) > 1 {
		// Could be technique (T1234) or tactic (TA0001)
		if len(searchTerm) > 2 && searchTerm[1] == 'A' {
			// Tactic ID
			return bleve.NewTermQuery(strings.ToLower(searchTerm))
		} else {
			// Technique ID
			return bleve.NewTermQuery(strings.ToLower(searchTerm))
		}
	} else if strings.HasPrefix(searchTerm, "M") || strings.HasPrefix(searchTerm, "S") || strings.HasPrefix(searchTerm, "G") {
		// Other ID types
		return bleve.NewTermQuery(strings.ToLower(searchTerm))
	}

	// For text search, use a combination of strategies:
	// 1. Phrase match for exact phrase
	phraseQuery := bleve.NewMatchPhraseQuery(searchTerm)
	
	// 2. Fuzzy match for typo tolerance
	fuzzyQuery := bleve.NewFuzzyQuery(searchTerm)
	fuzzyQuery.SetFuzziness(1) // Allow 1 character difference
	
	// 3. Prefix match for partial matches
	prefixQuery := bleve.NewPrefixQuery(searchTerm)
	
	// 4. Match query for individual terms
	matchQuery := bleve.NewMatchQuery(searchTerm)

	// Combine queries with boolean query - this gives us comprehensive matching
	bq := bleve.NewBooleanQuery()
	bq.AddShould(phraseQuery)
	bq.AddShould(fuzzyQuery)
	bq.AddShould(prefixQuery)
	bq.AddShould(matchQuery)

	// Also add field-specific boosting for better relevance
	nameBoostQuery := bleve.NewMatchQuery(searchTerm)
	nameBoostQuery.SetField("name")
	nameBoostQuery.SetBoost(2.0) // Boost name field matches

	descriptionBoostQuery := bleve.NewMatchQuery(searchTerm)
	descriptionBoostQuery.SetField("description")
	descriptionBoostQuery.SetBoost(1.5) // Boost description matches

	bq.AddShould(nameBoostQuery)
	bq.AddShould(descriptionBoostQuery)

	return bq
}

// SearchTactics searches for ATT&CK tactics by text with enhanced intelligence
func (s *LocalAttackStore) SearchTactics(ctx context.Context, search string, offset, limit int) ([]AttackTactic, int64, error) {
	if search == "" {
		// Fall back to pagination if no search term
		return s.ListTacticsPaginated(ctx, offset, limit)
	}

	// Create a more intelligent search query that handles various search patterns
	query := createIntelligentSearchQuery(search, "tactic")
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.From = offset
	searchRequest.Size = limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlight()

	// Enable relevance scoring and sorting
	searchRequest.SortBy([]string{"-_score"}) // Sort by relevance score

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var tactics []AttackTactic
	for _, hit := range searchResults.Hits {
		id := strings.TrimPrefix(hit.ID, "tactic_")
		tactic, err := s.GetTacticByID(ctx, id)
		if err != nil {
			continue // Skip if tactic not found in DB
		}
		tactics = append(tactics, *tactic)
	}

	return tactics, int64(searchResults.Total), nil
}

// SearchMitigations searches for ATT&CK mitigations by text with enhanced intelligence
func (s *LocalAttackStore) SearchMitigations(ctx context.Context, search string, offset, limit int) ([]AttackMitigation, int64, error) {
	if search == "" {
		// Fall back to pagination if no search term
		return s.ListMitigationsPaginated(ctx, offset, limit)
	}

	// Create a more intelligent search query that handles various search patterns
	query := createIntelligentSearchQuery(search, "mitigation")
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.From = offset
	searchRequest.Size = limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlight()

	// Enable relevance scoring and sorting
	searchRequest.SortBy([]string{"-_score"}) // Sort by relevance score

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var mitigations []AttackMitigation
	for _, hit := range searchResults.Hits {
		id := strings.TrimPrefix(hit.ID, "mitigation_")
		mitigation, err := s.GetMitigationByID(ctx, id)
		if err != nil {
			continue // Skip if mitigation not found in DB
		}
		mitigations = append(mitigations, *mitigation)
	}

	return mitigations, int64(searchResults.Total), nil
}

// SearchSoftware searches for ATT&CK software by text with enhanced intelligence
func (s *LocalAttackStore) SearchSoftware(ctx context.Context, search string, offset, limit int) ([]AttackSoftware, int64, error) {
	if search == "" {
		// Fall back to pagination if no search term
		return s.ListSoftwarePaginated(ctx, offset, limit)
	}

	// Create a more intelligent search query that handles various search patterns
	query := createIntelligentSearchQuery(search, "software")
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.From = offset
	searchRequest.Size = limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlight()

	// Enable relevance scoring and sorting
	searchRequest.SortBy([]string{"-_score"}) // Sort by relevance score

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var software []AttackSoftware
	for _, hit := range searchResults.Hits {
		id := strings.TrimPrefix(hit.ID, "software_")
		sw, err := s.GetSoftwareByID(ctx, id)
		if err != nil {
			continue // Skip if software not found in DB
		}
		software = append(software, *sw)
	}

	return software, int64(searchResults.Total), nil
}

// SearchGroups searches for ATT&CK groups by text with enhanced intelligence
func (s *LocalAttackStore) SearchGroups(ctx context.Context, search string, offset, limit int) ([]AttackGroup, int64, error) {
	if search == "" {
		// Fall back to pagination if no search term
		return s.ListGroupsPaginated(ctx, offset, limit)
	}

	// Create a more intelligent search query that handles various search patterns
	query := createIntelligentSearchQuery(search, "group")
	searchRequest := bleve.NewSearchRequest(query)
	searchRequest.From = offset
	searchRequest.Size = limit
	searchRequest.Fields = []string{"*"}
	searchRequest.Highlight = bleve.NewHighlight()

	// Enable relevance scoring and sorting
	searchRequest.SortBy([]string{"-_score"}) // Sort by relevance score

	searchResults, err := s.index.Search(searchRequest)
	if err != nil {
		return nil, 0, err
	}

	var groups []AttackGroup
	for _, hit := range searchResults.Hits {
		id := strings.TrimPrefix(hit.ID, "group_")
		group, err := s.GetGroupByID(ctx, id)
		if err != nil {
			continue // Skip if group not found in DB
		}
		groups = append(groups, *group)
	}

	return groups, int64(searchResults.Total), nil
}

// Helper functions for parsing Excel data
func getStringValue(row []string, colIndex int, headers []string, possibleHeaders ...string) string {
	// First, try to find the column by header name
	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))
		for _, possibleHeader := range possibleHeaders {
			if headerLower == strings.ToLower(strings.TrimSpace(possibleHeader)) {
				if i < len(row) {
					return strings.TrimSpace(row[i])
				}
				break
			}
		}
	}

	// Fallback to index if headers don't match or row doesn't have enough columns
	if colIndex < len(row) {
		return strings.TrimSpace(row[colIndex])
	}

	return ""
}

func getStringIndex(headers []string, possibleHeaders []string) int {
	for i, header := range headers {
		headerLower := strings.ToLower(strings.TrimSpace(header))
		for _, possibleHeader := range possibleHeaders {
			if headerLower == strings.ToLower(strings.TrimSpace(possibleHeader)) {
				return i
			}
		}
	}
	return -1 // Not found
}

func getBoolValue(row []string, colIndex int) bool {
	if colIndex < 0 || colIndex >= len(row) {
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
		if err := s.db.WithContext(ctx).Where("id = ?", rel.SourceRef).First(&technique).Error; err != nil {
			continue // Skip if technique not found
		}
		techniques = append(techniques, technique)
	}

	return techniques, nil
}
