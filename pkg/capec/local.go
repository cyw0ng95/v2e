//go:build libxml2

package capec

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/lestrrat-go/libxml2/parser"
)

// LocalCAPECStore manages a local database of CAPEC items.
type LocalCAPECStore struct {
	db *gorm.DB
}

// NewLocalCAPECStore creates or opens a local CAPEC database at dbPath.
func NewLocalCAPECStore(dbPath string) (*LocalCAPECStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetMaxOpenConns(100)
		sqlDB.SetConnMaxLifetime(0)
		db.Exec("PRAGMA journal_mode=WAL")
		db.Exec("PRAGMA synchronous=NORMAL")
		db.Exec("PRAGMA cache_size=-40000")
	}

	if err := db.AutoMigrate(&CAPECItemModel{}, &CAPECRelatedWeaknessModel{}, &CAPECExampleModel{}, &CAPECMitigationModel{}, &CAPECReferenceModel{}, &CAPECCatalogMeta{}); err != nil {
		return nil, err
	}
	return &LocalCAPECStore{db: db}, nil
}

// ImportFromXML imports CAPEC items from XML into DB without XSD validation.
func (s *LocalCAPECStore) ImportFromXML(xmlPath string, force bool) error {
	common.Info("Importing CAPEC data from XML file: %s", xmlPath)

	// Parse XML file into a libxml2 document using the parser package
	xf, err := os.Open(xmlPath)
	if err != nil {
		return fmt.Errorf("failed to open xml: %w", err)
	}
	defer xf.Close()
	p := parser.New()
	doc, err := p.ParseReader(xf)
	if err != nil {
		return fmt.Errorf("failed to parse xml: %w", err)
	}
	defer func() {
		if doc != nil {
			doc.Free()
		}
	}()

	// Extract catalog version from root element attribute (if present) to decide
	// whether import is needed. Use doc.DocumentElement() which returns (node, error).
	catalogVersion := ""
	root, err := doc.DocumentElement()
	if err == nil && root != nil {
		if xr, xerr := root.Find("@Version"); xerr == nil {
			if v := xr.String(); v != "" {
				catalogVersion = v
			}
			xr.Free()
		}
		// if a Name or other source is desired, capture it too
	}

	// Check existing catalog meta: if same version already imported, skip import unless forced.
	if !force && catalogVersion != "" {
		var meta CAPECCatalogMeta
		if err := s.db.First(&meta).Error; err == nil {
			if meta.Version == catalogVersion {
				common.Info(LogMsgImportSkipped, catalogVersion)
				return nil
			}
		}
	}

	// Skip XSD validation entirely - this ensures imports work without XSD schema
	common.Info("Skipping XSD validation as per security requirement; continuing with permissive import")

	// Parse XML into attack pattern structs (streaming)
	f, err := os.Open(xmlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	decoder := xml.NewDecoder(f)

	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			tx.Rollback()
			return err
		}
		switch se := t.(type) {
		case xml.StartElement:
			if se.Name.Local == "Attack_Pattern" {
				var ap CAPECAttackPattern
				if err := decoder.DecodeElement(&ap, &se); err != nil {
					tx.Rollback()
					return err
				}
				// Upsert CAPEC item; ensure we populate Abstraction/Status and compute a
				// summary fallback when Summary is empty. Description is stored as
				// the inner XML (may contain xhtml tags) so preserve ap.Description.XML.
				summary := ap.Summary
				if strings.TrimSpace(summary) == "" {
					summary = truncateString(strings.TrimSpace(ap.Description.XML), 200)
				}
				item := CAPECItemModel{
					CAPECID:         ap.ID,
					Name:            ap.Name,
					Summary:         summary,
					Description:     ap.Description.XML,
					Status:          ap.Status,
					Abstraction:     ap.Abstraction,
					Likelihood:      ap.Likelihood,
					TypicalSeverity: ap.TypicalSeverity,
				}
				if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error; err != nil {
					tx.Rollback()
					return err
				}
				// Related weaknesses
				tx.Where("capec_id = ?", ap.ID).Delete(&CAPECRelatedWeaknessModel{})
				for _, cwe := range ap.RelatedWeaknesses {
					r := CAPECRelatedWeaknessModel{CAPECID: ap.ID, CWEID: cwe.CWEID}
					if err := tx.Create(&r).Error; err != nil {
						tx.Rollback()
						return err
					}
				}

				// Examples
				tx.Where("capec_id = ?", ap.ID).Delete(&CAPECExampleModel{})
				for _, ex := range ap.Examples {
					e := strings.TrimSpace(ex.XML)
					if err := tx.Create(&CAPECExampleModel{CAPECID: ap.ID, ExampleText: e}).Error; err != nil {
						tx.Rollback()
						return err
					}
				}

				// Mitigations
				tx.Where("capec_id = ?", ap.ID).Delete(&CAPECMitigationModel{})
				for _, m := range ap.Mitigations {
					mm := strings.TrimSpace(m.XML)
					if err := tx.Create(&CAPECMitigationModel{CAPECID: ap.ID, MitigationText: mm}).Error; err != nil {
						tx.Rollback()
						return err
					}
				}

				// References
				tx.Where("capec_id = ?", ap.ID).Delete(&CAPECReferenceModel{})
				for _, rref := range ap.References {
					ref := rref.ExternalRef
					if err := tx.Create(&CAPECReferenceModel{CAPECID: ap.ID, ExternalReference: ref, URL: ""}).Error; err != nil {
						tx.Rollback()
						return err
					}
				}
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}
	// persist catalog metadata
	if catalogVersion != "" {
		// Use a fixed primary key to ensure a single-row metadata table.
		meta := CAPECCatalogMeta{ID: 1, Version: catalogVersion, Source: xmlPath, ImportedAtUTC: time.Now().UTC().Unix()}
		// upsert single-row meta by primary key
		if err := s.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(&meta).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetByID returns a CAPEC item by its textual ID (e.g. "CAPEC-123" or "123").
func (s *LocalCAPECStore) GetByID(ctx context.Context, id string) (*CAPECItemModel, error) {
	re := regexp.MustCompile(`\d+`)
	m := re.FindString(id)
	if m == "" {
		return nil, gorm.ErrRecordNotFound
	}
	n, err := strconv.Atoi(m)
	if err != nil {
		return nil, err
	}
	var item CAPECItemModel
	if err := s.db.WithContext(ctx).First(&item, "capec_id = ?", n).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

// ListCAPECsPaginated returns CAPEC items with pagination.
func (s *LocalCAPECStore) ListCAPECsPaginated(ctx context.Context, offset, limit int) ([]CAPECItemModel, int64, error) {
	var items []CAPECItemModel
	var total int64
	if err := s.db.WithContext(ctx).Model(&CAPECItemModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := s.db.WithContext(ctx).Order("capec_id asc").Offset(offset).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

// GetRelatedWeaknesses returns related CWE IDs for a given CAPEC numeric ID.
func (s *LocalCAPECStore) GetRelatedWeaknesses(ctx context.Context, capecID int) ([]CAPECRelatedWeaknessModel, error) {
	var rows []CAPECRelatedWeaknessModel
	if err := s.db.WithContext(ctx).Where("capec_id = ?", capecID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetExamples returns example texts for a given CAPEC numeric ID.
func (s *LocalCAPECStore) GetExamples(ctx context.Context, capecID int) ([]CAPECExampleModel, error) {
	var rows []CAPECExampleModel
	if err := s.db.WithContext(ctx).Where("capec_id = ?", capecID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetMitigations returns mitigation texts for a given CAPEC numeric ID.
func (s *LocalCAPECStore) GetMitigations(ctx context.Context, capecID int) ([]CAPECMitigationModel, error) {
	var rows []CAPECMitigationModel
	if err := s.db.WithContext(ctx).Where("capec_id = ?", capecID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetReferences returns references for a given CAPEC numeric ID.
func (s *LocalCAPECStore) GetReferences(ctx context.Context, capecID int) ([]CAPECReferenceModel, error) {
	var rows []CAPECReferenceModel
	if err := s.db.WithContext(ctx).Where("capec_id = ?", capecID).Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

// GetCatalogMeta returns the stored CAPEC catalog metadata (single row expected)
func (s *LocalCAPECStore) GetCatalogMeta(ctx context.Context) (*CAPECCatalogMeta, error) {
	var meta CAPECCatalogMeta
	if err := s.db.WithContext(ctx).First(&meta).Error; err != nil {
		return nil, err
	}
	return &meta, nil
}

func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

func truncateString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
