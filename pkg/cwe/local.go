package cwe

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// associationLoader defines the interface for loading nested associations.
type associationLoader interface {
	// load queries the database for records associated with parentID and populates target.
	load(ctx context.Context, db *gorm.DB, parentID string) error
}

// associationSaver defines the interface for saving nested associations.
type associationSaver interface {
	// deleteExisting removes all existing records for the parent.
	deleteExisting(tx *gorm.DB, parentID string) error
	// save inserts new records into the transaction.
	save(tx *gorm.DB, parentID string) error
}

// associationSavers manages multiple association savers for a parent entity.
type associationSavers []associationSaver

// deleteAllAndSave deletes all existing associations and saves new ones in a transaction.
func (a associationSavers) deleteAllAndSave(tx *gorm.DB, parentID string) error {
	for _, saver := range a {
		if err := saver.deleteExisting(tx, parentID); err != nil {
			return err
		}
		if err := saver.save(tx, parentID); err != nil {
			return err
		}
	}
	return nil
}

// sliceAssociationSaver is a generic saver for slice-based associations.
// It handles the common pattern of deleting existing records by parent ID and inserting new ones.
type sliceAssociationSaver[T any, S any] struct {
	modelSlice      []S
	foreignKeyCol   string
	modelToDBFn     func(S, string) T
}

func (s *sliceAssociationSaver[T, S]) deleteExisting(tx *gorm.DB, parentID string) error {
	return tx.Where(s.foreignKeyCol+" = ?", parentID).Delete(new(T)).Error
}

func (s *sliceAssociationSaver[T, S]) save(tx *gorm.DB, parentID string) error {
	for _, item := range s.modelSlice {
		model := s.modelToDBFn(item, parentID)
		if err := tx.Create(&model).Error; err != nil {
			return err
		}
	}
	return nil
}

// DemonstrativeExamplesAssociationSaver handles the special case for DemonstrativeExamples
// where entries are nested within example groups.
type DemonstrativeExamplesAssociationSaver struct {
	examples []DemonstrativeExample
}

func (s *DemonstrativeExamplesAssociationSaver) deleteExisting(tx *gorm.DB, parentID string) error {
	return tx.Where("cwe_id = ?", parentID).Delete(&DemonstrativeExampleModel{}).Error
}

func (s *DemonstrativeExamplesAssociationSaver) save(tx *gorm.DB, parentID string) error {
	for _, de := range s.examples {
		for _, entry := range de.Entries {
			dem := DemonstrativeExampleModel{
				CWEID:       parentID,
				EntryID:     de.ID,
				IntroText:   entry.IntroText,
				BodyText:    entry.BodyText,
				Nature:      entry.Nature,
				Language:    entry.Language,
				ExampleCode: entry.ExampleCode,
				Reference:   entry.Reference,
			}
			if err := tx.Create(&dem).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

// buildCWEAssociationSavers creates association savers for all nested CWE fields.
func buildCWEAssociationSavers(item *CWEItem) associationSavers {
	return associationSavers{
		// RelatedWeaknesses
		&sliceAssociationSaver[RelatedWeaknessModel, RelatedWeakness]{
			modelSlice:    item.RelatedWeaknesses,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(rw RelatedWeakness, cweID string) RelatedWeaknessModel {
				return RelatedWeaknessModel{
					CWEID:  cweID,
					Nature:  rw.Nature,
					CweID:   rw.CweID,
					ViewID:  rw.ViewID,
					Ordinal: rw.Ordinal,
				}
			},
		},
		// WeaknessOrdinalities
		&sliceAssociationSaver[WeaknessOrdinalityModel, WeaknessOrdinality]{
			modelSlice:    item.WeaknessOrdinalities,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(wo WeaknessOrdinality, cweID string) WeaknessOrdinalityModel {
				return WeaknessOrdinalityModel{
					CWEID:       cweID,
					Ordinality:  wo.Ordinality,
					Description: wo.Description,
				}
			},
		},
		// DetectionMethods
		&sliceAssociationSaver[DetectionMethodModel, DetectionMethod]{
			modelSlice:    item.DetectionMethods,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(dm DetectionMethod, cweID string) DetectionMethodModel {
				return DetectionMethodModel{
					CWEID:              cweID,
					DetectionMethodID:  dm.DetectionMethodID,
					Method:             dm.Method,
					Description:        dm.Description,
					Effectiveness:      dm.Effectiveness,
					EffectivenessNotes: dm.EffectivenessNotes,
				}
			},
		},
		// Mitigations
		&sliceAssociationSaver[MitigationModel, Mitigation]{
			modelSlice:    item.PotentialMitigations,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(mt Mitigation, cweID string) MitigationModel {
				return MitigationModel{
					CWEID:              cweID,
					MitigationID:       mt.MitigationID,
					Phase:              "", // flatten []string as needed
					Strategy:           mt.Strategy,
					Description:        mt.Description,
					Effectiveness:      mt.Effectiveness,
					EffectivenessNotes: mt.EffectivenessNotes,
				}
			},
		},
		// DemonstrativeExamples - special handling
		&DemonstrativeExamplesAssociationSaver{examples: item.DemonstrativeExamples},
		// ObservedExamples
		&sliceAssociationSaver[ObservedExampleModel, ObservedExample]{
			modelSlice:    item.ObservedExamples,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(oe ObservedExample, cweID string) ObservedExampleModel {
				return ObservedExampleModel{
					CWEID:       cweID,
					Reference:   oe.Reference,
					Description: oe.Description,
					Link:        oe.Link,
				}
			},
		},
		// TaxonomyMappings
		&sliceAssociationSaver[TaxonomyMappingModel, TaxonomyMapping]{
			modelSlice:    item.TaxonomyMappings,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(tm TaxonomyMapping, cweID string) TaxonomyMappingModel {
				return TaxonomyMappingModel{
					CWEID:        cweID,
					TaxonomyName: tm.TaxonomyName,
					EntryName:    tm.EntryName,
					EntryID:      tm.EntryID,
					MappingFit:   tm.MappingFit,
				}
			},
		},
		// Notes
		&sliceAssociationSaver[NoteModel, Note]{
			modelSlice:    item.Notes,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(n Note, cweID string) NoteModel {
				return NoteModel{
					CWEID: cweID,
					Type:  n.Type,
					Note:  n.Note,
				}
			},
		},
		// ContentHistory
		&sliceAssociationSaver[ContentHistoryModel, ContentHistory]{
			modelSlice:    item.ContentHistory,
			foreignKeyCol: "cwe_id",
			modelToDBFn: func(ch ContentHistory, cweID string) ContentHistoryModel {
				return ContentHistoryModel{
					CWEID:                    cweID,
					Type:                     ch.Type,
					SubmissionName:           ch.SubmissionName,
					SubmissionOrganization:   ch.SubmissionOrganization,
					SubmissionDate:           ch.SubmissionDate,
					SubmissionVersion:        ch.SubmissionVersion,
					SubmissionReleaseDate:    ch.SubmissionReleaseDate,
					SubmissionComment:        ch.SubmissionComment,
					ModificationName:         ch.ModificationName,
					ModificationOrganization: ch.ModificationOrganization,
					ModificationDate:         ch.ModificationDate,
					ModificationVersion:      ch.ModificationVersion,
					ModificationReleaseDate:  ch.ModificationReleaseDate,
					ModificationComment:      ch.ModificationComment,
					ContributionName:         ch.ContributionName,
					ContributionOrganization: ch.ContributionOrganization,
					ContributionDate:         ch.ContributionDate,
					ContributionVersion:      ch.ContributionVersion,
					ContributionReleaseDate:  ch.ContributionReleaseDate,
					ContributionComment:      ch.ContributionComment,
					ContributionType:         ch.ContributionType,
					PreviousEntryName:        ch.PreviousEntryName,
					Date:                     ch.Date,
					Version:                  ch.Version,
				}
			},
		},
	}
}

// LocalCWEStore manages a local database of CWE items.
type LocalCWEStore struct {
	db *gorm.DB
}

// CWEItemModel is the GORM model for flat CWE fields.
type CWEItemModel struct {
	ID                  string `gorm:"primaryKey"`
	Name                string `gorm:"index"`
	Abstraction         string `gorm:"index"`
	Structure           string
	Status              string `gorm:"index"`
	Description         string
	ExtendedDescription string
	LikelihoodOfExploit string
}

// RelatedWeaknessModel is the GORM model for related weaknesses.
type RelatedWeaknessModel struct {
	ID      uint   `gorm:"primaryKey"`
	CWEID   string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	Nature  string `gorm:"column:nature;index"`
	CweID   string `gorm:"column:related_cwe_id;index"` // ID of the related CWE
	ViewID  string `gorm:"column:view_id"`
	Ordinal string `gorm:"column:ordinal"`
}

// WeaknessOrdinalityModel is the GORM model for weakness ordinalities.
type WeaknessOrdinalityModel struct {
	ID          uint   `gorm:"primaryKey"`
	CWEID       string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	Ordinality  string `gorm:"column:ordinality"`
	Description string `gorm:"column:description"`
}

// DetectionMethodModel is the GORM model for detection methods.
type DetectionMethodModel struct {
	ID                 uint   `gorm:"primaryKey"`
	CWEID              string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	DetectionMethodID  string `gorm:"column:detection_method_id;index"`
	Method             string `gorm:"column:method;index"`
	Description        string `gorm:"column:description"`
	Effectiveness      string `gorm:"column:effectiveness;index"`
	EffectivenessNotes string `gorm:"column:effectiveness_notes"`
}

// MitigationModel is the GORM model for potential mitigations.
type MitigationModel struct {
	ID                 uint   `gorm:"primaryKey"`
	CWEID              string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	MitigationID       string `gorm:"column:mitigation_id;index"`
	Phase              string `gorm:"column:phase"` // store as comma-separated string for []string
	Strategy           string `gorm:"column:strategy;index"`
	Description        string `gorm:"column:description"`
	Effectiveness      string `gorm:"column:effectiveness;index"`
	EffectivenessNotes string `gorm:"column:effectiveness_notes"`
}

// DemonstrativeExampleModel is the GORM model for demonstrative examples.
type DemonstrativeExampleModel struct {
	ID          uint   `gorm:"primaryKey"`
	CWEID       string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	EntryID     string `gorm:"column:entry_id"`
	IntroText   string `gorm:"column:intro_text"`
	BodyText    string `gorm:"column:body_text"`
	Nature      string `gorm:"column:nature"`
	Language    string `gorm:"column:language"`
	ExampleCode string `gorm:"column:example_code"`
	Reference   string `gorm:"column:reference"`
}

// ObservedExampleModel is the GORM model for observed examples.
type ObservedExampleModel struct {
	ID          uint   `gorm:"primaryKey"`
	CWEID       string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	Reference   string `gorm:"column:reference"`
	Description string `gorm:"column:description"`
	Link        string `gorm:"column:link"`
}

// TaxonomyMappingModel is the GORM model for taxonomy mappings.
type TaxonomyMappingModel struct {
	ID           uint   `gorm:"primaryKey"`
	CWEID        string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	TaxonomyName string `gorm:"column:taxonomy_name"`
	EntryName    string `gorm:"column:entry_name"`
	EntryID      string `gorm:"column:entry_id"`
	MappingFit   string `gorm:"column:mapping_fit"`
}

// NoteModel is the GORM model for notes.
type NoteModel struct {
	ID    uint   `gorm:"primaryKey"`
	CWEID string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	Type  string `gorm:"column:type"`
	Note  string `gorm:"column:note"`
}

// ContentHistoryModel is the GORM model for content history.
type ContentHistoryModel struct {
	ID                       uint   `gorm:"primaryKey"`
	CWEID                    string `gorm:"column:cwe_id;index"` // Foreign key to parent CWE item
	Type                     string `gorm:"column:type"`
	SubmissionName           string `gorm:"column:submission_name"`
	SubmissionOrganization   string `gorm:"column:submission_organization"`
	SubmissionDate           string `gorm:"column:submission_date"`
	SubmissionVersion        string `gorm:"column:submission_version"`
	SubmissionReleaseDate    string `gorm:"column:submission_release_date"`
	SubmissionComment        string `gorm:"column:submission_comment"`
	ModificationName         string `gorm:"column:modification_name"`
	ModificationOrganization string `gorm:"column:modification_organization"`
	ModificationDate         string `gorm:"column:modification_date"`
	ModificationVersion      string `gorm:"column:modification_version"`
	ModificationReleaseDate  string `gorm:"column:modification_release_date"`
	ModificationComment      string `gorm:"column:modification_comment"`
	ContributionName         string `gorm:"column:contribution_name"`
	ContributionOrganization string `gorm:"column:contribution_organization"`
	ContributionDate         string `gorm:"column:contribution_date"`
	ContributionVersion      string `gorm:"column:contribution_version"`
	ContributionReleaseDate  string `gorm:"column:contribution_release_date"`
	ContributionComment      string `gorm:"column:contribution_comment"`
	ContributionType         string `gorm:"column:contribution_type"`
	PreviousEntryName        string `gorm:"column:previous_entry_name"`
	Date                     string `gorm:"column:date"`
	Version                  string `gorm:"column:version"`
}

// NewLocalCWEStore creates or opens a local CWE database at dbPath.
func NewLocalCWEStore(dbPath string) (*LocalCWEStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: false,
	})
	if err != nil {
		return nil, err
	}
	// Set SQLite PRAGMAs for WAL mode and better concurrency
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(0)
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA synchronous=NORMAL")
		db.Exec("PRAGMA cache_size=-40000")
		// Set busy_timeout to handle lock contention when multiple services access the database
		db.Exec("PRAGMA busy_timeout=30000")
	}
	if err := db.AutoMigrate(
		&CWEItemModel{},
		&RelatedWeaknessModel{},
		&WeaknessOrdinalityModel{},
		&DetectionMethodModel{},
		&MitigationModel{},
		&DemonstrativeExampleModel{},
		&ObservedExampleModel{},
		&TaxonomyMappingModel{},
		&NoteModel{},
		&ContentHistoryModel{},
	); err != nil {
		return nil, err
	}

	// Migrate view-related tables
	if err := AutoMigrateViews(db); err != nil {
		return nil, err
	}
	return &LocalCWEStore{db: db}, nil
}

// ImportFromJSON imports CWE records from a JSON file (array of CWEItem).
// Wraps the entire import in a database transaction for atomicity.
func (s *LocalCWEStore) ImportFromJSON(jsonPath string) error {
	common.Info(LogMsgImportingJSON, jsonPath)
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer f.Close()
	var items []CWEItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return err
	}
	if len(items) == 0 {
		return nil // nothing to import
	}
	// Check if first and last CWE already exist
	var first, last CWEItemModel
	if err := s.db.First(&first, "id = ?", items[0].ID).Error; err == nil {
		if err := s.db.First(&last, "id = ?", items[len(items)-1].ID).Error; err == nil {
			common.Info(LogMsgImportSkipped, items[0].ID, items[len(items)-1].ID)
			return nil
		}
	}
	// Wrap import in transaction for atomicity - if any part fails, all changes are rolled back
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			m := CWEItemModel{
				ID:                  item.ID,
				Name:                item.Name,
				Abstraction:         item.Abstraction,
				Structure:           item.Structure,
				Status:              item.Status,
				Description:         item.Description,
				ExtendedDescription: item.ExtendedDescription,
				LikelihoodOfExploit: item.LikelihoodOfExploit,
			}
			if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&m).Error; err != nil {
				return err
			}
			// Save all nested associations using the generic repository pattern
			if err := buildCWEAssociationSavers(&item).deleteAllAndSave(tx, item.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

// loadAndMap is a generic helper for querying associations and mapping them.
// It queries models using whereClause, then applies mapFn to each result.
func loadAndMap[T any, R any](
	ctx context.Context,
	db *gorm.DB,
	whereClause string,
	whereArgs interface{},
	mapFn func(T) R,
) []R {
	var models []T
	db.WithContext(ctx).Where(whereClause, whereArgs).Find(&models)
	result := make([]R, 0, len(models))
	for _, m := range models {
		result = append(result, mapFn(m))
	}
	return result
}

// loadCWEFields loads all nested fields for a CWE item using generic pattern.
func (s *LocalCWEStore) loadCWEFields(ctx context.Context, cweID string, item *CWEItem) {
	// RelatedWeaknesses
	item.RelatedWeaknesses = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m RelatedWeaknessModel) RelatedWeakness {
		return RelatedWeakness{
			Nature:  m.Nature,
			CweID:   m.CweID,
			ViewID:  m.ViewID,
			Ordinal: m.Ordinal,
		}
	})

	// WeaknessOrdinalities
	item.WeaknessOrdinalities = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m WeaknessOrdinalityModel) WeaknessOrdinality {
		return WeaknessOrdinality{
			Ordinality:  m.Ordinality,
			Description: m.Description,
		}
	})

	// DetectionMethods
	item.DetectionMethods = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m DetectionMethodModel) DetectionMethod {
		return DetectionMethod{
			DetectionMethodID:  m.DetectionMethodID,
			Method:             m.Method,
			Description:        m.Description,
			Effectiveness:      m.Effectiveness,
			EffectivenessNotes: m.EffectivenessNotes,
		}
	})

	// Mitigations
	item.PotentialMitigations = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m MitigationModel) Mitigation {
		phase := []string{}
		if m.Phase != "" {
			phase = strings.Split(strings.TrimSpace(m.Phase), ",")
		}
		return Mitigation{
			MitigationID:       m.MitigationID,
			Phase:              phase,
			Strategy:           m.Strategy,
			Description:        m.Description,
			Effectiveness:      m.Effectiveness,
			EffectivenessNotes: m.EffectivenessNotes,
		}
	})

	// DemonstrativeExamples - special handling for grouped entries
	var des []DemonstrativeExampleModel
	s.db.WithContext(ctx).Where("cwe_id = ?", cweID).Find(&des)
	entriesByID := make(map[string][]DemonstrativeEntry)
	for _, de := range des {
		entriesByID[de.EntryID] = append(entriesByID[de.EntryID], DemonstrativeEntry{
			IntroText:   de.IntroText,
			BodyText:    de.BodyText,
			Nature:      de.Nature,
			Language:    de.Language,
			ExampleCode: de.ExampleCode,
			Reference:   de.Reference,
		})
	}
	for entryID, entries := range entriesByID {
		item.DemonstrativeExamples = append(item.DemonstrativeExamples, DemonstrativeExample{
			ID:      entryID,
			Entries: entries,
		})
	}

	// ObservedExamples
	item.ObservedExamples = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m ObservedExampleModel) ObservedExample {
		return ObservedExample{
			Reference:   m.Reference,
			Description: m.Description,
			Link:        m.Link,
		}
	})

	// TaxonomyMappings
	item.TaxonomyMappings = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m TaxonomyMappingModel) TaxonomyMapping {
		return TaxonomyMapping{
			TaxonomyName: m.TaxonomyName,
			EntryName:    m.EntryName,
			EntryID:      m.EntryID,
			MappingFit:   m.MappingFit,
		}
	})

	// Notes
	item.Notes = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m NoteModel) Note {
		return Note{
			Type: m.Type,
			Note: m.Note,
		}
	})

	// ContentHistory
	item.ContentHistory = loadAndMap(ctx, s.db, "cwe_id = ?", cweID, func(m ContentHistoryModel) ContentHistory {
		return ContentHistory{
			Type:                     m.Type,
			SubmissionName:           m.SubmissionName,
			SubmissionOrganization:   m.SubmissionOrganization,
			SubmissionDate:           m.SubmissionDate,
			SubmissionVersion:        m.SubmissionVersion,
			SubmissionReleaseDate:    m.SubmissionReleaseDate,
			SubmissionComment:        m.SubmissionComment,
			ModificationName:         m.ModificationName,
			ModificationOrganization: m.ModificationOrganization,
			ModificationDate:         m.ModificationDate,
			ModificationVersion:      m.ModificationVersion,
			ModificationReleaseDate:  m.ModificationReleaseDate,
			ModificationComment:      m.ModificationComment,
			ContributionName:         m.ContributionName,
			ContributionOrganization: m.ContributionOrganization,
			ContributionDate:         m.ContributionDate,
			ContributionVersion:      m.ContributionVersion,
			ContributionReleaseDate:  m.ContributionReleaseDate,
			ContributionComment:      m.ContributionComment,
			ContributionType:         m.ContributionType,
			PreviousEntryName:        m.PreviousEntryName,
			Date:                     m.Date,
			Version:                  m.Version,
		}
	})
}

// GetByID retrieves a CWEItem by ID.
func (s *LocalCWEStore) GetByID(ctx context.Context, id string) (*CWEItem, error) {
	var m CWEItemModel
	if err := s.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	item := &CWEItem{
		ID:                  m.ID,
		Name:                m.Name,
		Abstraction:         m.Abstraction,
		Structure:           m.Structure,
		Status:              m.Status,
		Description:         m.Description,
		ExtendedDescription: m.ExtendedDescription,
		LikelihoodOfExploit: m.LikelihoodOfExploit,
	}
	s.loadCWEFields(ctx, id, item)
	return item, nil
}

// ListCWEsPaginated returns a paginated list of CWEItems.
func (s *LocalCWEStore) ListCWEsPaginated(ctx context.Context, offset, limit int) ([]CWEItem, int64, error) {
	var models []CWEItemModel
	var total int64
	// Check Count error before proceeding with Find
	if err := s.db.Model(&CWEItemModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := s.db.WithContext(ctx).Offset(offset).Limit(limit).Find(&models).Error; err != nil {
		return nil, 0, err
	}
	items := make([]CWEItem, 0, len(models))
	for _, m := range models {
		item := CWEItem{
			ID:                  m.ID,
			Name:                m.Name,
			Abstraction:         m.Abstraction,
			Structure:           m.Structure,
			Status:              m.Status,
			Description:         m.Description,
			ExtendedDescription: m.ExtendedDescription,
			LikelihoodOfExploit: m.LikelihoodOfExploit,
		}
		s.loadCWEFields(ctx, m.ID, &item)
		items = append(items, item)
	}
	return items, total, nil
}
