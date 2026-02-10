package glc

import (
	"gorm.io/gorm"
)

// MigrateTables runs database migrations for GLC tables
func MigrateTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&GraphModel{},
		&GraphVersionModel{},
		&UserPresetModel{},
		&ShareLinkModel{},
	)
}
