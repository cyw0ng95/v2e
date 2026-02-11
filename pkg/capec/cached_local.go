//go:build CONFIG_USE_LIBXML2

package capec

import (
	"context"
	"regexp"
	"strconv"
	"sync"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
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
	return importCAPECFromXML(s.db, xmlPath, force, s.invalidateAllCache)
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
