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
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
"github.com/lestrrat-go/libxml2/parser"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// cacheItem holds cached CAPEC data with timestamp for cache invalidation
type capecCacheItem struct {
	data      *CAPECItemModel
	timestamp time.Time
}

// CachedLocalCAPECStore manages a local database of CAPEC items with caching.
type CachedLocalCAPECStore struct {
	db    *gorm.DB
	cache map[int]*capecCacheItem
	mu    sync.RWMutex  // Protects the cache
	ttl   time.Duration // Time-to-live for cache entries
}

// NewCachedLocalCAPECStore creates or opens a local CAPEC database at dbPath with caching.
func NewCachedLocalCAPECStore(dbPath string) (*CachedLocalCAPECStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		// Enable prepared statement caching for better performance
		PrepareStmt: true,
	})
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

	store := &CachedLocalCAPECStore{
		db:    db,
		cache: make(map[int]*capecCacheItem),
		ttl:   10 * time.Minute, // Cache for 10 minutes since CAPEC data changes rarely
	}

	return store, nil
}

// getCachedCAPEC retrieves a CAPEC from cache if available and not expired
func (s *CachedLocalCAPECStore) getCachedCAPEC(capecID int) (*CAPECItemModel, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, exists := s.cache[capecID]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid (not expired)
	if time.Since(item.timestamp) > s.ttl {
		// Entry is expired, remove it from cache
		delete(s.cache, capecID)
		return nil, false
	}

	return item.data, true
}

// setCachedCAPEC stores a CAPEC in cache
func (s *CachedLocalCAPECStore) setCachedCAPEC(capecID int, item *CAPECItemModel) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache[capecID] = &capecCacheItem{
		data:      item,
		timestamp: time.Now(),
	}
}

// GetByID returns a CAPEC item by its numeric ID with caching.
func (s *CachedLocalCAPECStore) GetByID(ctx context.Context, id string) (*CAPECItemModel, error) {
	// Parse the ID to get the numeric ID
	re := regexp.MustCompile(`\d+`)
	m := re.FindString(id)
	if m == "" {
		return nil, gorm.ErrRecordNotFound
	}
	n, err := strconv.Atoi(m)
	if err != nil {
		return nil, err
	}

	// First, check the cache
	if cached, found := s.getCachedCAPEC(n); found {
		return cached, nil
	}

	// Cache miss, get from database
	var item CAPECItemModel
	if err := s.db.WithContext(ctx).First(&item, "capec_id = ?", n).Error; err != nil {
		return nil, err
	}

	// Store in cache for future requests
	s.setCachedCAPEC(n, &item)

	return &item, nil
}

// invalidateCachedCAPEC removes a CAPEC from cache (useful after updates)
func (s *CachedLocalCAPECStore) invalidateCachedCAPEC(capecID int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.cache, capecID)
}

// ImportFromXML imports CAPEC items from XML into DB without XSD validation.
// This method invalidates the cache after import since data has changed.
func (s *CachedLocalCAPECStore) ImportFromXML(xmlPath string, force bool) error {
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
				common.Info("CAPEC catalog version %s already imported; skipping import", catalogVersion)
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
				// summary fallback when Summary is empty.
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

	// Invalidate entire cache after import since data has changed
	s.invalidateAllCache()

	return nil
}

// invalidateAllCache removes all entries from the cache
func (s *CachedLocalCAPECStore) invalidateAllCache() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear the entire cache
	s.cache = make(map[int]*capecCacheItem)
}

// Close closes the database connection
func (s *CachedLocalCAPECStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
