package cce

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// LocalCCEStore manages a local database of CCE items
type LocalCCEStore struct {
	db *gorm.DB
}

// CCEModel is the GORM model for CCE entries
type CCEModel struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:text" json:"description"`
	Owner       string    `json:"owner"`
	Status      string    `gorm:"index" json:"status"`
	Type        string    `json:"type"`
	Reference   string    `gorm:"type:text" json:"reference"`
	Metadata    string    `gorm:"type:text" json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName specifies the table name for GORM
func (CCEModel) TableName() string {
	return "cce_items"
}

// NewLocalCCEStore creates a new CCE store with a SQLite database
func NewLocalCCEStore(dbPath string, logger *common.Logger) (*LocalCCEStore, error) {
	if dbPath == "" {
		dbPath = "cce.db"
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: nil,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)

	db.Exec("PRAGMA journal_mode=WAL")
	db.Exec("PRAGMA synchronous=NORMAL")
	// Set busy_timeout to handle lock contention when multiple services access the database
	db.Exec("PRAGMA busy_timeout=30000")

	if err := db.AutoMigrate(&CCEModel{}); err != nil {
		return nil, err
	}

	return &LocalCCEStore{db: db}, nil
}

// Close closes the database connection
func (s *LocalCCEStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// CreateCCE creates a new CCE entry
func (s *LocalCCEStore) CreateCCE(ctx context.Context, entry CCE) error {
	model := CCEModel{
		ID:          entry.ID,
		Title:       entry.Title,
		Description: entry.Description,
		Owner:       entry.Owner,
		Status:      entry.Status,
		Type:        entry.Type,
		Reference:   entry.Reference,
		Metadata:    entry.Metadata,
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "description", "owner", "status", "type", "reference", "metadata", "updated_at"}),
	}).Create(&model).Error
}

// CreateCCEs creates multiple CCE entries in batch
func (s *LocalCCEStore) CreateCCEs(ctx context.Context, entries []CCE) error {
	if len(entries) == 0 {
		return nil
	}

	models := make([]CCEModel, len(entries))
	for i, entry := range entries {
		models[i] = CCEModel{
			ID:          entry.ID,
			Title:       entry.Title,
			Description: entry.Description,
			Owner:       entry.Owner,
			Status:      entry.Status,
			Type:        entry.Type,
			Reference:   entry.Reference,
			Metadata:    entry.Metadata,
		}
	}

	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"title", "description", "owner", "status", "type", "reference", "metadata", "updated_at"}),
	}).CreateInBatches(models, 100).Error
}

// GetCCEByID retrieves a CCE entry by ID
func (s *LocalCCEStore) GetCCEByID(ctx context.Context, id string) (*CCE, error) {
	var model CCEModel
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, err
	}

	return &CCE{
		ID:          model.ID,
		Title:       model.Title,
		Description: model.Description,
		Owner:       model.Owner,
		Status:      model.Status,
		Type:        model.Type,
		Reference:   model.Reference,
		Metadata:    model.Metadata,
	}, nil
}

// ListCCEs lists CCE entries with pagination
func (s *LocalCCEStore) ListCCEs(ctx context.Context, offset, limit int) ([]CCE, int64, error) {
	var models []CCEModel
	var total int64

	if err := s.db.WithContext(ctx).Model(&CCEModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := s.db.WithContext(ctx).
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	entries := make([]CCE, len(models))
	for i, model := range models {
		entries[i] = CCE{
			ID:          model.ID,
			Title:       model.Title,
			Description: model.Description,
			Owner:       model.Owner,
			Status:      model.Status,
			Type:        model.Type,
			Reference:   model.Reference,
			Metadata:    model.Metadata,
		}
	}

	return entries, total, nil
}

// SearchCCEs searches CCE entries by query
func (s *LocalCCEStore) SearchCCEs(ctx context.Context, query string, offset, limit int) ([]CCE, int64, error) {
	var models []CCEModel
	var total int64

	searchQuery := "%" + query + "%"
	countQuery := s.db.WithContext(ctx).Model(&CCEModel{}).
		Where("id LIKE ? OR title LIKE ? OR description LIKE ?", searchQuery, searchQuery, searchQuery)

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := s.db.WithContext(ctx).
		Where("id LIKE ? OR title LIKE ? OR description LIKE ?", searchQuery, searchQuery, searchQuery).
		Order("id ASC").
		Offset(offset).
		Limit(limit).
		Find(&models).Error
	if err != nil {
		return nil, 0, err
	}

	entries := make([]CCE, len(models))
	for i, model := range models {
		entries[i] = CCE{
			ID:          model.ID,
			Title:       model.Title,
			Description: model.Description,
			Owner:       model.Owner,
			Status:      model.Status,
			Type:        model.Type,
			Reference:   model.Reference,
			Metadata:    model.Metadata,
		}
	}

	return entries, total, nil
}

// CountCCEs returns the total count of CCE entries
func (s *LocalCCEStore) CountCCEs(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&CCEModel{}).Count(&count).Error
	return count, err
}

// DeleteCCE deletes a CCE entry by ID
func (s *LocalCCEStore) DeleteCCE(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Where("id = ?", id).Delete(&CCEModel{}).Error
}

// UpdateCCE updates an existing CCE entry
func (s *LocalCCEStore) UpdateCCE(ctx context.Context, entry CCE) error {
	model := CCEModel{
		ID:          entry.ID,
		Title:       entry.Title,
		Description: entry.Description,
		Owner:       entry.Owner,
		Status:      entry.Status,
		Type:        entry.Type,
		Reference:   entry.Reference,
		Metadata:    entry.Metadata,
	}

	return s.db.WithContext(ctx).
		Where("id = ?", entry.ID).
		Updates(&model).Error
}

// ImportCCEsFromExcel imports CCE entries from an Excel file
func (s *LocalCCEStore) ImportCCEsFromExcel(ctx context.Context, filePath string) (int, error) {
	parser := NewParser(filePath)
	entries, err := parser.ParseAll()
	if err != nil {
		return 0, err
	}

	if err := s.CreateCCEs(ctx, entries); err != nil {
		return 0, err
	}

	return len(entries), nil
}

// GetStats returns database statistics
func (s *LocalCCEStore) GetStats(ctx context.Context) (map[string]interface{}, error) {
	var total, active, deprecated int64

	if err := s.db.WithContext(ctx).Model(&CCEModel{}).Count(&total).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&CCEModel{}).Where("status = ?", "ACTIVE").Count(&active).Error; err != nil {
		return nil, err
	}

	if err := s.db.WithContext(ctx).Model(&CCEModel{}).Where("status = ?", "DEPRECATED").Count(&deprecated).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total":      total,
		"active":     active,
		"deprecated": deprecated,
	}, nil
}

// ParseExcel parses CCE data from Excel file
func ParseExcel(filePath string) ([]CCE, error) {
	parser := NewParser(filePath)
	return parser.ParseAll()
}

// LoadCCEData loads CCE data from a JSON file
func LoadCCEData(filePath string) ([]CCE, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var entries []CCE
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}

	return entries, nil
}
