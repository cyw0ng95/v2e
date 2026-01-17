package cwe

import (
	"context"
	"encoding/json"
	"os"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// LocalCWEStore manages a local database of CWE items.
type LocalCWEStore struct {
	db *gorm.DB
}

// cweItemModel is the GORM model for CWEItem.
type cweItemModel struct {
	ID   string `gorm:"primaryKey"`
	Data []byte `gorm:"type:json"`
}

// NewLocalCWEStore creates or opens a local CWE database at dbPath.
func NewLocalCWEStore(dbPath string) (*LocalCWEStore, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&cweItemModel{}); err != nil {
		return nil, err
	}
	return &LocalCWEStore{db: db}, nil
}

// ImportFromJSON imports CWE records from a JSON file (array of CWEItem).
func (s *LocalCWEStore) ImportFromJSON(jsonPath string) error {
	f, err := os.Open(jsonPath)
	if err != nil {
		return err
	}
	defer f.Close()
	var items []CWEItem
	if err := json.NewDecoder(f).Decode(&items); err != nil {
		return err
	}
	for _, item := range items {
		data, err := json.Marshal(item)
		if err != nil {
			return err
		}
		m := cweItemModel{ID: item.ID, Data: data}
		if err := s.db.Clauses(clause.OnConflict{UpdateAll: true}).Create(&m).Error; err != nil {
			return err
		}
	}
	return nil
}

// GetByID retrieves a CWEItem by ID.
func (s *LocalCWEStore) GetByID(ctx context.Context, id string) (*CWEItem, error) {
	var m cweItemModel
	if err := s.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	var item CWEItem
	if err := json.Unmarshal(m.Data, &item); err != nil {
		return nil, err
	}
	return &item, nil
}

// ListAll returns all CWEItems.
func (s *LocalCWEStore) ListAll(ctx context.Context) ([]CWEItem, error) {
	var models []cweItemModel
	if err := s.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	items := make([]CWEItem, 0, len(models))
	for _, m := range models {
		var item CWEItem
		if err := json.Unmarshal(m.Data, &item); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}
