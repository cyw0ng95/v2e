package repo

import (
	"encoding/json"
	"time"

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
func (d *DB) SaveCVE(cve *CVEItem) error {
	// Marshal the full CVE data to JSON
	data, err := json.Marshal(cve)
	if err != nil {
		return err
	}

	record := CVERecord{
		CVEID:        cve.ID,
		SourceID:     cve.SourceID,
		Published:    cve.Published,
		LastModified: cve.LastModified,
		VulnStatus:   cve.VulnStatus,
		Data:         string(data),
	}

	// Check if record exists
	var existing CVERecord
	result := d.db.Where("cve_id = ?", cve.ID).First(&existing)
	
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
func (d *DB) SaveCVEs(cves []CVEItem) error {
	for _, cve := range cves {
		if err := d.SaveCVE(&cve); err != nil {
			return err
		}
	}
	return nil
}

// GetCVE retrieves a CVE by ID from the database
func (d *DB) GetCVE(cveID string) (*CVEItem, error) {
	var record CVERecord
	if err := d.db.Where("cve_id = ?", cveID).First(&record).Error; err != nil {
		return nil, err
	}

	var cve CVEItem
	if err := json.Unmarshal([]byte(record.Data), &cve); err != nil {
		return nil, err
	}

	return &cve, nil
}

// ListCVEs retrieves CVEs with pagination
func (d *DB) ListCVEs(offset, limit int) ([]CVEItem, error) {
	var records []CVERecord
	if err := d.db.Offset(offset).Limit(limit).Order("published desc").Find(&records).Error; err != nil {
		return nil, err
	}

	cves := make([]CVEItem, 0, len(records))
	for _, record := range records {
		var cve CVEItem
		if err := json.Unmarshal([]byte(record.Data), &cve); err != nil {
			return nil, err
		}
		cves = append(cves, cve)
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
