package local

import (
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// cacheItem holds cached CVE data with timestamp for cache invalidation
type cacheItem struct {
	data      *cve.CVEItem
	timestamp time.Time
}

// CachedDB represents the database connection with caching capabilities
type CachedDB struct {
	db    *gorm.DB
	cache map[string]*cacheItem
	mu    sync.RWMutex  // Protects the cache
	ttl   time.Duration // Time-to-live for cache entries
}

// NewCachedDB creates a new database connection with caching
// dbPath is the path to the SQLite database file (e.g., "cve.db")
func NewCachedDB(dbPath string) (*CachedDB, error) {
	// Disable GORM logging to prevent interference with RPC message parsing
	// When running as a subprocess, stdout is used for RPC messages only
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// Enable prepared statement caching for better performance
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&CVERecord{}); err != nil {
		return nil, err
	}

	// Configure connection pool for better performance
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Set connection pool parameters
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Enable WAL mode for better concurrent access (Principle 10)
	// WAL mode allows readers and writers to work simultaneously
	if _, err := sqlDB.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, err
	}

	// Optimize synchronous mode for better performance (Principle 10)
	// NORMAL is faster than FULL while still being safe
	if _, err := sqlDB.Exec("PRAGMA synchronous=NORMAL"); err != nil {
		return nil, err
	}

	// Increase cache size for better query performance (Principle 10)
	// Default is 2000 pages, we set to 10000 (about 40MB with 4KB pages)
	if _, err := sqlDB.Exec("PRAGMA cache_size=-40000"); err != nil {
		return nil, err
	}

	// Create the cached DB instance with a 5-minute TTL for cache entries
	cachedDB := &CachedDB{
		db:    db,
		cache: make(map[string]*cacheItem),
		ttl:   5 * time.Minute, // Cache entries for 5 minutes
	}

	return cachedDB, nil
}

// getCachedCVE retrieves a CVE from cache if available and not expired
func (cdb *CachedDB) getCachedCVE(cveID string) (*cve.CVEItem, bool) {
	cdb.mu.RLock()
	defer cdb.mu.RUnlock()

	item, exists := cdb.cache[cveID]
	if !exists {
		return nil, false
	}

	// Check if cache entry is still valid (not expired)
	if time.Since(item.timestamp) > cdb.ttl {
		// Entry is expired, remove it from cache
		delete(cdb.cache, cveID)
		return nil, false
	}

	return item.data, true
}

// setCachedCVE stores a CVE in cache
func (cdb *CachedDB) setCachedCVE(cveID string, item *cve.CVEItem) {
	cdb.mu.Lock()
	defer cdb.mu.Unlock()

	cdb.cache[cveID] = &cacheItem{
		data:      item,
		timestamp: time.Now(),
	}
}

// invalidateCachedCVE removes a CVE from cache (useful after updates/deletes)
func (cdb *CachedDB) invalidateCachedCVE(cveID string) {
	cdb.mu.Lock()
	defer cdb.mu.Unlock()

	delete(cdb.cache, cveID)
}

// GetCVE retrieves a CVE by ID from the database with caching
func (cdb *CachedDB) GetCVE(cveID string) (*cve.CVEItem, error) {
	// First, check the cache
	if cached, found := cdb.getCachedCVE(cveID); found {
		return cached, nil
	}

	// Cache miss, get from database
	var record CVERecord
	if err := cdb.db.Where("cve_id = ?", cveID).First(&record).Error; err != nil {
		return nil, err
	}

	var cveItem cve.CVEItem
	if err := jsonutil.Unmarshal([]byte(record.Data), &cveItem); err != nil {
		return nil, err
	}

	// Store in cache for future requests
	cdb.setCachedCVE(cveID, &cveItem)

	return &cveItem, nil
}

// SaveCVE saves a CVE item to the database and updates cache
func (cdb *CachedDB) SaveCVE(cveItem *cve.CVEItem) error {
	// Marshal the full CVE data to JSON
	data, err := jsonutil.Marshal(cveItem)
	if err != nil {
		return err
	}

	record := CVERecord{
		CVEID:        cveItem.ID,
		SourceID:     cveItem.SourceID,
		Published:    cveItem.Published.Time,
		LastModified: cveItem.LastModified.Time,
		VulnStatus:   cveItem.VulnStatus,
		Data:         string(data),
	}

	// Check if record exists
	var existing CVERecord
	result := cdb.db.Unscoped().Where("cve_id = ?", cveItem.ID).First(&existing)

	if result.Error == nil {
		// Record exists, update it
		record.ID = existing.ID
		record.CreatedAt = existing.CreatedAt
		record.DeletedAt = gorm.DeletedAt{} // Clear soft delete flag
		if err := cdb.db.Unscoped().Save(&record).Error; err != nil {
			return err
		}
	} else if result.Error == gorm.ErrRecordNotFound {
		// Record doesn't exist, create it
		if err := cdb.db.Create(&record).Error; err != nil {
			return err
		}
	} else {
		return result.Error
	}

	// Update cache with the new/updated item
	cdb.setCachedCVE(cveItem.ID, cveItem)

	return nil
}

// DeleteCVE deletes a CVE from the database and removes from cache
func (cdb *CachedDB) DeleteCVE(cveID string) error {
	result := cdb.db.Where("cve_id = ?", cveID).Delete(&CVERecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	// Invalidate cache entry
	cdb.invalidateCachedCVE(cveID)

	return nil
}

// Close closes the database connection
func (cdb *CachedDB) Close() error {
	sqlDB, err := cdb.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
