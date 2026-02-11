package notes

import (
	"fmt"

	"gorm.io/gorm"
)

// MigrateNotesTables migrates all notes-related tables into the provided DB.
// This includes bookmarks, notes, history, memory cards, learning sessions,
// cross-references, global items, and URN links tables.
func MigrateNotesTables(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&BookmarkModel{},
		&BookmarkHistoryModel{},
		&NoteModel{},
		&MemoryCardModel{},
		&LearningSessionModel{},
		&CrossReferenceModel{},
		&GlobalItemModel{},
		&URNLink{},
	); err != nil {
		return err
	}

	// Run data migration for existing records
	return MigrateExistingData(db)
}

// MigrateExistingData updates existing records with URNs and FSM states
func MigrateExistingData(db *gorm.DB) error {
	// Migrate existing notes to generate URNs
	if err := migrateNotesURN(db); err != nil {
		return fmt.Errorf("failed to migrate notes URNs: %w", err)
	}

	// Migrate existing memory cards to generate URNs
	if err := migrateCardsURN(db); err != nil {
		return fmt.Errorf("failed to migrate cards URNs: %w", err)
	}

	return nil
}

// migrateNotesURN generates URNs for existing notes without them
func migrateNotesURN(db *gorm.DB) error {
	// Get notes without URN
	var notes []NoteModel
	if err := db.Where("urn IS NULL OR urn = ''").Find(&notes).Error; err != nil {
		return err
	}

	for _, note := range notes {
		note.URN = GetNoteURN(note.ID)
		// Set default FSM state if not set
		if note.FSMState == "" {
			note.FSMState = "draft"
		}
		if err := db.Save(&note).Error; err != nil {
			return fmt.Errorf("failed to update note %d: %w", note.ID, err)
		}
	}

	return nil
}

// migrateCardsURN generates URNs for existing memory cards without them
func migrateCardsURN(db *gorm.DB) error {
	// Get cards without URN
	var cards []MemoryCardModel
	if err := db.Where("urn IS NULL OR urn = ''").Find(&cards).Error; err != nil {
		return err
	}

	for _, card := range cards {
		card.URN = GetCardURN(card.ID)
		// Set default FSM state if not set
		if card.FSMState == "" {
			card.FSMState = "new"
		}
		if err := db.Save(&card).Error; err != nil {
			return fmt.Errorf("failed to update card %d: %w", card.ID, err)
		}
	}

	return nil
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
