package notes

import (
	"gorm.io/gorm"
)

// MigrateNotesTables migrates all notes-related tables into the provided DB.
// This includes bookmarks, notes, history, memory cards, learning sessions, 
// cross-references, and global items tables.
func MigrateNotesTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&BookmarkModel{},
		&BookmarkHistoryModel{},
		&NoteModel{},
		&MemoryCardModel{},
		&LearningSessionModel{},
		&CrossReferenceModel{},
		&GlobalItemModel{},
	)
}

// MigrateNotesTablesRollback removes all notes-related tables from the database.
// NOTE: This will permanently delete all bookmark and note data.
func MigrateNotesTablesRollback(db *gorm.DB) error {
	// Drop tables in reverse order to respect foreign key constraints
	if err := db.Migrator().DropTable(&CrossReferenceModel{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&LearningSessionModel{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&MemoryCardModel{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&NoteModel{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&BookmarkHistoryModel{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&BookmarkModel{}); err != nil {
		return err
	}
	if err := db.Migrator().DropTable(&GlobalItemModel{}); err != nil {
		return err
	}
	return nil
}