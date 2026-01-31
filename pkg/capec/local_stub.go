//go:build !CONFIG_USE_LIBXML2

package capec

import (
	"context"
	"encoding/xml"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LocalCAPECStore manages a local database of CAPEC items (stubbed without libxml2).
type LocalCAPECStore struct {
	db *gorm.DB
}

// NewLocalCAPECStore creates or opens a local CAPEC database at dbPath.
// This stub implementation mirrors the DB setup but does not perform XML validation.
func NewLocalCAPECStore(dbPath string) (*LocalCAPECStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	// AutoMigrate minimal tables to allow app to run; detailed imports require libxml2 build tag.
	if err := db.AutoMigrate(&CAPECItemModel{}, &CAPECRelatedWeaknessModel{}, &CAPECExampleModel{}, &CAPECMitigationModel{}, &CAPECReferenceModel{}, &CAPECCatalogMeta{}); err != nil {
		return nil, err
	}
	return &LocalCAPECStore{db: db}, nil
}

// ImportFromXML imports CAPEC items from XML into DB without XSD validation.
func (s *LocalCAPECStore) ImportFromXML(xmlPath string, force bool) error {
	// Permissive importer: parse the CAPEC XML without XSD validation
	f, err := os.Open(xmlPath)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := xml.NewDecoder(f)
	// The CAPEC XML uses a default namespace; we'll match elements by local name.
	// First, check for catalog version in root element
	catalogVersion := ""
	firstToken := true

	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		se, ok := t.(xml.StartElement)
		if !ok {
			continue
		}
		// Check if this is the root element and extract version if present
		if firstToken {
			firstToken = false
			for _, attr := range se.Attr {
				if attr.Name.Local == "Version" {
					catalogVersion = attr.Value
					break
				}
			}
			// Now that we have the version, check if we should skip import
			if !force && catalogVersion != "" {
				var meta CAPECCatalogMeta
				if err := s.db.First(&meta).Error; err == nil {
					if meta.Version == catalogVersion {
						return nil
					}
				}
			}
		}
		if se.Name.Local != "Attack_Pattern" {
			continue
		}
		// parse attributes
		var capecID int
		var nameAttr string
		for _, a := range se.Attr {
			if a.Name.Local == "ID" {
				if n, err := strconv.Atoi(a.Value); err == nil {
					capecID = n
				}
			}
			if a.Name.Local == "Name" {
				nameAttr = a.Value
			}
		}

		// defaults
		description := ""
		summary := ""
		likelihood := ""
		typicalSeverity := ""
		var weaknesses []string
		var examples []string
		var mitigations []string
		var references []string

		// read inner tokens until end of Attack_Pattern
		for {
			it, err := dec.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			switch tt := it.(type) {
			case xml.StartElement:
				switch tt.Name.Local {
				case "Description":
					// Decode the Description element content
					var descContent string
					if err := dec.DecodeElement(&descContent, &tt); err == nil {
						description = descContent
						// Check if the description content contains a nested Summary element
						trimmed := strings.TrimSpace(descContent)
						if strings.Contains(trimmed, "<Summary>") && strings.Contains(trimmed, "</Summary>") {
							startIndex := strings.Index(trimmed, "<Summary>")
							endIndex := strings.Index(trimmed, "</Summary>")
							if startIndex >= 0 && endIndex > startIndex {
								innerStart := startIndex + len("<Summary>")
								summary = trimmed[innerStart:endIndex]
							}
						}
					}
				case "Likelihood_Of_Attack":
					var stext string
					if err := dec.DecodeElement(&stext, &tt); err == nil {
						likelihood = stext
					}
				case "Typical_Severity":
					var stext string
					if err := dec.DecodeElement(&stext, &tt); err == nil {
						typicalSeverity = stext
					}
				case "Related_Weaknesses":
					// read until end Related_Weaknesses
					for {
						inner, err := dec.Token()
						if err != nil {
							return err
						}
						if end, ok := inner.(xml.EndElement); ok && end.Name.Local == "Related_Weaknesses" {
							break
						}
						if re, ok := inner.(xml.StartElement); ok && re.Name.Local == "Related_Weakness" {
							for _, a := range re.Attr {
								if a.Name.Local == "CWE_ID" {
									weaknesses = append(weaknesses, a.Value)
								}
							}
						}
					}
				case "Example_Instances":
					for {
						inner, err := dec.Token()
						if err != nil {
							return err
						}
						if end, ok := inner.(xml.EndElement); ok && end.Name.Local == "Example_Instances" {
							break
						}
						if ex, ok := inner.(xml.StartElement); ok && ex.Name.Local == "Example" {
							// decode inner content as raw string; examples often use xhtml:p
							var buf string
							if err := dec.DecodeElement(&buf, &ex); err == nil {
								examples = append(examples, strings.TrimSpace(buf))
							}
						}
					}
				case "Mitigations":
					for {
						inner, err := dec.Token()
						if err != nil {
							return err
						}
						if end, ok := inner.(xml.EndElement); ok && end.Name.Local == "Mitigations" {
							break
						}
						if me, ok := inner.(xml.StartElement); ok && me.Name.Local == "Mitigation" {
							var buf string
							if err := dec.DecodeElement(&buf, &me); err == nil {
								mitigations = append(mitigations, strings.TrimSpace(buf))
							}
						}
					}
				case "References":
					for {
						inner, err := dec.Token()
						if err != nil {
							return err
						}
						if end, ok := inner.(xml.EndElement); ok && end.Name.Local == "References" {
							break
						}
						if ref, ok := inner.(xml.StartElement); ok && ref.Name.Local == "Reference" {
							for _, a := range ref.Attr {
								if a.Name.Local == "External_Reference_ID" {
									references = append(references, a.Value)
								}
							}
						}
					}
				}
			case xml.EndElement:
				if tt.Name.Local == "Attack_Pattern" {
					// commit to DB
					item := CAPECItemModel{
						CAPECID: capecID,
						Name:    firstNonEmpty(nameAttr, ""),
						Summary: func() string {
							if summary != "" {
								return summary
							}
							return truncateString(description, 200)
						}(),
						Description:     description,
						Status:          "",
						Likelihood:      likelihood,
						TypicalSeverity: typicalSeverity,
					}
					// use transaction
					tx := s.db.Begin()
					if err := tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&item).Error; err != nil {
						tx.Rollback()
						return err
					}
					// replace related tables
					if err := tx.Where("capec_id = ?", capecID).Delete(&CAPECRelatedWeaknessModel{}).Error; err != nil {
						tx.Rollback()
						return err
					}
					// Deduplicate related weaknesses to avoid unique constraint violations
					seenCWEs := make(map[string]bool)
					for _, w := range weaknesses {
						if w != "" && !seenCWEs[w] {
							seenCWEs[w] = true
							rw := CAPECRelatedWeaknessModel{CAPECID: capecID, CWEID: w}
							if err := tx.Create(&rw).Error; err != nil {
								tx.Rollback()
								return err
							}
						}
					}
					if err := tx.Where("capec_id = ?", capecID).Delete(&CAPECExampleModel{}).Error; err != nil {
						tx.Rollback()
						return err
					}
					for _, e := range examples {
						exm := CAPECExampleModel{CAPECID: capecID, ExampleText: e}
						if err := tx.Create(&exm).Error; err != nil {
							tx.Rollback()
							return err
						}
					}
					if err := tx.Where("capec_id = ?", capecID).Delete(&CAPECMitigationModel{}).Error; err != nil {
						tx.Rollback()
						return err
					}
					for _, m := range mitigations {
						mm := CAPECMitigationModel{CAPECID: capecID, MitigationText: m}
						if err := tx.Create(&mm).Error; err != nil {
							tx.Rollback()
							return err
						}
					}
					if err := tx.Where("capec_id = ?", capecID).Delete(&CAPECReferenceModel{}).Error; err != nil {
						tx.Rollback()
						return err
					}
					// Deduplicate references to avoid unique constraint violations
					seenRefs := make(map[string]bool)
					for _, r := range references {
						if r != "" && !seenRefs[r] {
							seenRefs[r] = true
							rr := CAPECReferenceModel{CAPECID: capecID, ExternalReference: r, URL: ""}
							if err := tx.Create(&rr).Error; err != nil {
								tx.Rollback()
								return err
							}
						}
					}
					if err := tx.Commit().Error; err != nil {
						return err
					}
				}
			}
		}
	}
	// persist catalog metadata
	if catalogVersion != "" {
		// Use a fixed primary key to ensure a single-row metadata table.
		meta := CAPECCatalogMeta{ID: 1, Version: catalogVersion, Source: xmlPath, ImportedAtUTC: time.Now().Unix()}
		// upsert single-row meta by primary key
		if err := s.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, UpdateAll: true}).Create(&meta).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetByID returns a CAPEC item by its textual ID (e.g. "CAPEC-123" or "123").
func (s *LocalCAPECStore) GetByID(ctx context.Context, id string) (*CAPECItemModel, error) {
	// Try to extract digits from the id
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
