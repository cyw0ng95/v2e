package local

import (
	"github.com/bytedance/sonic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB represents the database connection
type DB struct {
	db *gorm.DB
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

// NewDB creates a new database connection
// dbPath is the path to the SQLite database file (e.g., "cve.db")
func NewDB(dbPath string) (*DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate the schema
	if err := db.AutoMigrate(&CVERecord{}); err != nil {
		return nil, err
	}

	return &DB{db: db}, nil
}

// SaveCVE saves a CVE item to the database
func (d *DB) SaveCVE(cveItem *cve.CVEItem) error {
	// Marshal the full CVE data to JSON
	data, err := sonic.Marshal(cveItem)
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
	result := d.db.Where("cve_id = ?", cveItem.ID).First(&existing)
	
	if result.Error == nil {
		// Record exists, update it
		record.ID = existing.ID
		record.CreatedAt = existing.CreatedAt
		return d.db.Save(&record).Error
	} else if result.Error == gorm.ErrRecordNotFound {
		// Record doesn't exist, create it
		return d.db.Create(&record).Error
	}
	
	return result.Error
}

// SaveCVEs saves multiple CVE items to the database
func (d *DB) SaveCVEs(cves []cve.CVEItem) error {
	for _, cveItem := range cves {
		if err := d.SaveCVE(&cveItem); err != nil {
			return err
		}
	}
	return nil
}

// GetCVE retrieves a CVE by ID from the database
func (d *DB) GetCVE(cveID string) (*cve.CVEItem, error) {
	var record CVERecord
	if err := d.db.Where("cve_id = ?", cveID).First(&record).Error; err != nil {
		return nil, err
	}

	var cveItem cve.CVEItem
	if err := sonic.Unmarshal([]byte(record.Data), &cveItem); err != nil {
		return nil, err
	}

	return &cveItem, nil
}

// ListCVEs retrieves CVEs with pagination
func (d *DB) ListCVEs(offset, limit int) ([]cve.CVEItem, error) {
	var records []CVERecord
	if err := d.db.Offset(offset).Limit(limit).Order("published desc").Find(&records).Error; err != nil {
		return nil, err
	}

	cves := make([]cve.CVEItem, 0, len(records))
	for _, record := range records {
		var cveItem cve.CVEItem
		if err := sonic.Unmarshal([]byte(record.Data), &cveItem); err != nil {
			return nil, err
		}
		cves = append(cves, cveItem)
	}

	return cves, nil
}

// Count returns the total number of CVEs in the database
func (d *DB) Count() (int64, error) {
	var count int64
	if err := d.db.Model(&CVERecord{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

// Close closes the database connection
func (d *DB) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
