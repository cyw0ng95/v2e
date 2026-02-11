package cwe

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/cyw0ng95/v2e/pkg/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
		if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&m).Error; err != nil {
			return err
		}
		// RelatedWeaknesses
		s.db.Where("cwe_id = ?", item.ID).Delete(&RelatedWeaknessModel{})
		for _, rw := range item.RelatedWeaknesses {
			rwm := RelatedWeaknessModel{
				CWEID:   item.ID,
				Nature:  rw.Nature,
				CweID:   rw.CweID,
				ViewID:  rw.ViewID,
				Ordinal: rw.Ordinal,
			}
			if err := s.db.Create(&rwm).Error; err != nil {
				return err
			}
		}
		// WeaknessOrdinalities
		s.db.Where("cwe_id = ?", item.ID).Delete(&WeaknessOrdinalityModel{})
		for _, wo := range item.WeaknessOrdinalities {
			wom := WeaknessOrdinalityModel{
				CWEID:       item.ID,
				Ordinality:  wo.Ordinality,
				Description: wo.Description,
			}
			if err := s.db.Create(&wom).Error; err != nil {
				return err
			}
		}
		// DetectionMethods
		s.db.Where("cwe_id = ?", item.ID).Delete(&DetectionMethodModel{})
		for _, dm := range item.DetectionMethods {
			dmm := DetectionMethodModel{
				CWEID:              item.ID,
				DetectionMethodID:  dm.DetectionMethodID,
				Method:             dm.Method,
				Description:        dm.Description,
				Effectiveness:      dm.Effectiveness,
				EffectivenessNotes: dm.EffectivenessNotes,
			}
			if err := s.db.Create(&dmm).Error; err != nil {
				return err
			}
		}
		// Mitigations
		s.db.Where("cwe_id = ?", item.ID).Delete(&MitigationModel{})
		for _, mt := range item.PotentialMitigations {
			mtm := MitigationModel{
				CWEID:              item.ID,
				MitigationID:       mt.MitigationID,
				Phase:              "", // flatten []string as needed
				Strategy:           mt.Strategy,
				Description:        mt.Description,
				Effectiveness:      mt.Effectiveness,
				EffectivenessNotes: mt.EffectivenessNotes,
			}
			if err := s.db.Create(&mtm).Error; err != nil {
				return err
			}
		}
		// DemonstrativeExamples
		s.db.Where("cwe_id = ?", item.ID).Delete(&DemonstrativeExampleModel{})
		for _, de := range item.DemonstrativeExamples {
			for _, entry := range de.Entries {
				dem := DemonstrativeExampleModel{
					CWEID:       item.ID,
					EntryID:     de.ID,
					IntroText:   entry.IntroText,
					BodyText:    entry.BodyText,
					Nature:      entry.Nature,
					Language:    entry.Language,
					ExampleCode: entry.ExampleCode,
					Reference:   entry.Reference,
				}
				if err := s.db.Create(&dem).Error; err != nil {
					return err
				}
			}
		}
		// ObservedExamples
		s.db.Where("cwe_id = ?", item.ID).Delete(&ObservedExampleModel{})
		for _, oe := range item.ObservedExamples {
			oem := ObservedExampleModel{
				CWEID:       item.ID,
				Reference:   oe.Reference,
				Description: oe.Description,
				Link:        oe.Link,
			}
			if err := s.db.Create(&oem).Error; err != nil {
				return err
			}
		}
		// TaxonomyMappings
		s.db.Where("cwe_id = ?", item.ID).Delete(&TaxonomyMappingModel{})
		for _, tm := range item.TaxonomyMappings {
			tmm := TaxonomyMappingModel{
				CWEID:        item.ID,
				TaxonomyName: tm.TaxonomyName,
				EntryName:    tm.EntryName,
				EntryID:      tm.EntryID,
				MappingFit:   tm.MappingFit,
			}
			if err := s.db.Create(&tmm).Error; err != nil {
				return err
			}
		}
		// Notes
		s.db.Where("cwe_id = ?", item.ID).Delete(&NoteModel{})
		for _, n := range item.Notes {
			nm := NoteModel{
				CWEID: item.ID,
				Type:  n.Type,
				Note:  n.Note,
			}
			if err := s.db.Create(&nm).Error; err != nil {
				return err
			}
		}
		// ContentHistory
		s.db.Where("cwe_id = ?", item.ID).Delete(&ContentHistoryModel{})
		for _, ch := range item.ContentHistory {
			chm := ContentHistoryModel{
				CWEID:                    item.ID,
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
			if err := s.db.Create(&chm).Error; err != nil {
				return err
			}
		}
		// Add similar logic for other nested fields as needed
	}
	return nil
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
	// RelatedWeaknesses
	var rws []RelatedWeaknessModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&rws)
	for _, rw := range rws {
		item.RelatedWeaknesses = append(item.RelatedWeaknesses, RelatedWeakness{
			Nature:  rw.Nature,
			CweID:   rw.CweID,
			ViewID:  rw.ViewID,
			Ordinal: rw.Ordinal,
		})
	}
	// WeaknessOrdinalities
	var wos []WeaknessOrdinalityModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&wos)
	for _, wo := range wos {
		item.WeaknessOrdinalities = append(item.WeaknessOrdinalities, WeaknessOrdinality{
			Ordinality:  wo.Ordinality,
			Description: wo.Description,
		})
	}
	// DetectionMethods
	var dms []DetectionMethodModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&dms)
	for _, dm := range dms {
		item.DetectionMethods = append(item.DetectionMethods, DetectionMethod{
			DetectionMethodID:  dm.DetectionMethodID,
			Method:             dm.Method,
			Description:        dm.Description,
			Effectiveness:      dm.Effectiveness,
			EffectivenessNotes: dm.EffectivenessNotes,
		})
	}
	// Mitigations
	var mts []MitigationModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&mts)
	for _, mt := range mts {
		phase := []string{}
		if mt.Phase != "" {
			phase = strings.Split(strings.TrimSpace(mt.Phase), ",")
		}
		item.PotentialMitigations = append(item.PotentialMitigations, Mitigation{
			MitigationID:       mt.MitigationID,
			Phase:              phase,
			Strategy:           mt.Strategy,
			Description:        mt.Description,
			Effectiveness:      mt.Effectiveness,
			EffectivenessNotes: mt.EffectivenessNotes,
		})
	}
	// DemonstrativeExamples
	var des []DemonstrativeExampleModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&des)
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
	var oes []ObservedExampleModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&oes)
	for _, oe := range oes {
		item.ObservedExamples = append(item.ObservedExamples, ObservedExample{
			Reference:   oe.Reference,
			Description: oe.Description,
			Link:        oe.Link,
		})
	}
	// TaxonomyMappings
	var tms []TaxonomyMappingModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&tms)
	for _, tm := range tms {
		item.TaxonomyMappings = append(item.TaxonomyMappings, TaxonomyMapping{
			TaxonomyName: tm.TaxonomyName,
			EntryName:    tm.EntryName,
			EntryID:      tm.EntryID,
			MappingFit:   tm.MappingFit,
		})
	}
	// Notes
	var ns []NoteModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&ns)
	for _, n := range ns {
		item.Notes = append(item.Notes, Note{
			Type: n.Type,
			Note: n.Note,
		})
	}
	// ContentHistory
	var chs []ContentHistoryModel
	s.db.WithContext(ctx).Where("cwe_id = ?", id).Find(&chs)
	for _, ch := range chs {
		item.ContentHistory = append(item.ContentHistory, ContentHistory{
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
		})
	}
	// Add similar logic for other nested fields as needed
	return item, nil
}

// ListCWEsPaginated returns a paginated list of CWEItems.
func (s *LocalCWEStore) ListCWEsPaginated(ctx context.Context, offset, limit int) ([]CWEItem, int64, error) {
	var models []CWEItemModel
	var total int64
	s.db.Model(&CWEItemModel{}).Count(&total)
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
		// RelatedWeaknesses
		var rws []RelatedWeaknessModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&rws)
		for _, rw := range rws {
			item.RelatedWeaknesses = append(item.RelatedWeaknesses, RelatedWeakness{
				Nature:  rw.Nature,
				CweID:   rw.CweID,
				ViewID:  rw.ViewID,
				Ordinal: rw.Ordinal,
			})
		}
		// WeaknessOrdinalities
		var wos []WeaknessOrdinalityModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&wos)
		for _, wo := range wos {
			item.WeaknessOrdinalities = append(item.WeaknessOrdinalities, WeaknessOrdinality{
				Ordinality:  wo.Ordinality,
				Description: wo.Description,
			})
		}
		// DetectionMethods
		var dms []DetectionMethodModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&dms)
		for _, dm := range dms {
			item.DetectionMethods = append(item.DetectionMethods, DetectionMethod{
				DetectionMethodID:  dm.DetectionMethodID,
				Method:             dm.Method,
				Description:        dm.Description,
				Effectiveness:      dm.Effectiveness,
				EffectivenessNotes: dm.EffectivenessNotes,
			})
		}
		// Mitigations
		var mts []MitigationModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&mts)
		for _, mt := range mts {
			phase := []string{}
			if mt.Phase != "" {
				phase = strings.Split(strings.TrimSpace(mt.Phase), ",")
			}
			item.PotentialMitigations = append(item.PotentialMitigations, Mitigation{
				MitigationID:       mt.MitigationID,
				Phase:              phase,
				Strategy:           mt.Strategy,
				Description:        mt.Description,
				Effectiveness:      mt.Effectiveness,
				EffectivenessNotes: mt.EffectivenessNotes,
			})
		}
		// DemonstrativeExamples (group by EntryID)
		var des []DemonstrativeExampleModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&des)
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
		var oes []ObservedExampleModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&oes)
		for _, oe := range oes {
			item.ObservedExamples = append(item.ObservedExamples, ObservedExample{
				Reference:   oe.Reference,
				Description: oe.Description,
				Link:        oe.Link,
			})
		}
		// TaxonomyMappings
		var tms []TaxonomyMappingModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&tms)
		for _, tm := range tms {
			item.TaxonomyMappings = append(item.TaxonomyMappings, TaxonomyMapping{
				TaxonomyName: tm.TaxonomyName,
				EntryName:    tm.EntryName,
				EntryID:      tm.EntryID,
				MappingFit:   tm.MappingFit,
			})
		}
		// Notes
		var ns []NoteModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&ns)
		for _, n := range ns {
			item.Notes = append(item.Notes, Note{
				Type: n.Type,
				Note: n.Note,
			})
		}
		// ContentHistory
		var chs []ContentHistoryModel
		s.db.WithContext(ctx).Where("cwe_id = ?", m.ID).Find(&chs)
		for _, ch := range chs {
			item.ContentHistory = append(item.ContentHistory, ContentHistory{
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
			})
		}
		// Add similar logic for other nested fields as needed
		items = append(items, item)
	}
	return items, total, nil
}
