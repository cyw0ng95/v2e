package local

import (
	"strings"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB represents the database connection
type DB struct {
	db *gorm.DB
}

// retryOnLocked executes a database operation with retry logic for lock errors.
// It will retry up to 2 times (3 total attempts) with exponential backoff
// (10ms, 20ms) when encountering "database is locked" errors.
func retryOnLocked(fn func() error) error {
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		// If it's a database lock error, wait and retry
		if strings.Contains(err.Error(), "database is locked") && attempt < 2 {
			time.Sleep(time.Millisecond * time.Duration(10*(attempt+1))) // Exponential backoff
			continue
		}
		// For other errors or final attempt, return immediately
		return err
	}
	return nil
}

// GormDB returns the underlying GORM database instance
func (d *DB) GormDB() *gorm.DB {
	return d.db
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CVERecord represents a CVE record in the database
type CVERecord struct {
	gorm.Model
	CVEID        string    `gorm:"uniqueIndex;not null"`
	SourceID     string    `gorm:"index"`
	Published    time.Time `gorm:"index"`
	LastModified time.Time `gorm:"index"`
	VulnStatus   string    `gorm:"index"`
	Data         string    `gorm:"type:text"` // JSON representation of full CVEItem
}

// NewOptimizedDB creates an optimized database connection
func NewOptimizedDB(dbPath string) (*DB, error) {
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
	sqlDB.SetMaxIdleConns(20) // Increased idle connections for concurrent requests
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Enhanced SQLite PRAGMAs for performance
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-40000",
		"PRAGMA mmap_size=268435456", // 256MB
		"PRAGMA temp_store=memory",
	}
	for _, p := range pragmas {
		if _, err := sqlDB.Exec(p); err != nil {
			return nil, err
		}
	}

	return &DB{db: db}, nil
}

// NewDB creates a new database connection
// dbPath is the path to the SQLite database file (e.g., "cve.db")
func NewDB(dbPath string) (*DB, error) {
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
	sqlDB.SetMaxIdleConns(20) // Increased idle connections for concurrent requests
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

	// Additional PRAGMAs for enhanced performance
	pragmas := []string{
		"PRAGMA mmap_size=268435456", // 256MB memory mapping
		"PRAGMA temp_store=memory",   // Store temp tables in memory
		"PRAGMA foreign_keys=OFF",    // Disable FK constraints for speed
		"PRAGMA busy_timeout=30000",  // Wait up to 30 seconds for locks
	}
	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return nil, err
		}
	}

	return &DB{db: db}, nil
}

// SaveCVE saves a CVE item to the database
func (d *DB) SaveCVE(cveItem *cve.CVEItem) error {
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
	result := d.db.Unscoped().Where("cve_id = ?", cveItem.ID).First(&existing)

	switch {
	case result.Error == nil:
		// Record exists, update it
		record.ID = existing.ID
		record.CreatedAt = existing.CreatedAt
		record.DeletedAt = gorm.DeletedAt{} // Clear soft delete flag
		return d.db.Unscoped().Save(&record).Error
	case result.Error == gorm.ErrRecordNotFound:
		// Record doesn't exist, create it
		return d.db.Create(&record).Error
	default:
		return result.Error
	}
}

// BulkInsertRecords efficiently inserts multiple records in a single transaction
func (d *DB) BulkInsert(records []CVERecord, batchSize int) error {
	tx := d.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i := 0; i < len(records); i += batchSize {
		end := min(i+batchSize, len(records))
		if err := tx.Create(records[i:end]).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

// SaveCVEs saves multiple CVE items to the database using batch insert for better performance
func (d *DB) SaveCVEs(cves []cve.CVEItem) error {
	if len(cves) == 0 {
		return nil
	}

	// Pre-allocate records slice with exact capacity
	records := make([]CVERecord, len(cves))

	for i := range cves {
		// Marshal the full CVE data to JSON
		// Use value type instead of pointer to avoid unnecessary allocation
		data, err := jsonutil.Marshal(cves[i])
		if err != nil {
			return err
		}

		// Direct assignment instead of append since we pre-allocated
		records[i] = CVERecord{
			CVEID:        cves[i].ID,
			SourceID:     cves[i].SourceID,
			Published:    cves[i].Published.Time,
			LastModified: cves[i].LastModified.Time,
			VulnStatus:   cves[i].VulnStatus,
			Data:         string(data),
		}
	}

	// Use CreateInBatches for better performance
	// Process 100 records at a time to balance memory and performance
	return d.db.CreateInBatches(records, 100).Error
}

// GetCVE retrieves a CVE by ID from the database
func (d *DB) GetCVE(cveID string) (*cve.CVEItem, error) {
	var record CVERecord
	if err := d.db.Where("cve_id = ?", cveID).First(&record).Error; err != nil {
		return nil, err
	}

	var cveItem cve.CVEItem
	if err := jsonutil.Unmarshal([]byte(record.Data), &cveItem); err != nil {
		return nil, err
	}

	return &cveItem, nil
}

// ListCVEs retrieves CVEs with pagination
func (d *DB) ListCVEs(offset, limit int) ([]cve.CVEItem, error) {
	var records []CVERecord

	err := retryOnLocked(func() error {
		return d.db.Offset(offset).Limit(limit).Order("published desc").Find(&records).Error
	})
	if err != nil {
		return nil, err
	}

	// Pre-allocate with exact capacity to avoid re-allocations
	cves := make([]cve.CVEItem, len(records))
	for i, record := range records {
		if err := jsonutil.Unmarshal([]byte(record.Data), &cves[i]); err != nil {
			return nil, err
		}
	}

	return cves, nil
}

// Count returns the total number of CVEs in the database
func (d *DB) Count() (int64, error) {
	var count int64

	err := retryOnLocked(func() error {
		return d.db.Model(&CVERecord{}).Count(&count).Error
	})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// LazyCVERecord provides lazy loading for CVE data
type LazyCVERecord struct {
	ID         string
	*CVERecord // Loaded on demand
	loaded     bool
	db         *DB
}

// NewLazyCVERecord creates a new lazy-loaded CVE record
func (d *DB) NewLazyCVERecord(id string) *LazyCVERecord {
	return &LazyCVERecord{
		ID:     id,
		loaded: false,
		db:     d,
	}
}

// Load ensures the CVE record is loaded from the database
func (l *LazyCVERecord) Load() error {
	if !l.loaded {
		record, err := l.db.GetCVERaw(l.ID)
		if err != nil {
			return err
		}
		l.CVERecord = record
		l.loaded = true
	}
	return nil
}

// GetCVERaw retrieves raw CVE record without unmarshaling the data field
func (d *DB) GetCVERaw(cveID string) (*CVERecord, error) {
	var record CVERecord
	if err := d.db.Where("cve_id = ?", cveID).First(&record).Error; err != nil {
		return nil, err
	}
	return &record, nil
}

// DeleteCVE deletes a CVE from the database by ID
func (d *DB) DeleteCVE(cveID string) error {
	result := d.db.Where("cve_id = ?", cveID).Delete(&CVERecord{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// Close closes the database connection
func (d *DB) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
