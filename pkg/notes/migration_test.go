package notes

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestNormalizationMigration ensures the SQL migration normalizes legacy statuses
func TestNormalizationMigration(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestNormalizationMigration", db, func(t *testing.T, tx *gorm.DB) {
		// create in-memory sqlite DB and run base migrations
		db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		if err != nil {
			t.Fatalf("failed open db: %v", err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			t.Fatalf("failed to get sql.DB: %v", err)
		}

		if err := MigrateNotesTables(db); err != nil {
			t.Fatalf("migrate tables: %v", err)
		}

		// seed rows with legacy statuses
		rows := []MemoryCardModel{
			{BookmarkID: 1, Front: "f1", Back: "b1", Status: "active", Content: "{}"},
			{BookmarkID: 1, Front: "f2", Back: "b2", Status: "in-progress", Content: "{}"},
			{BookmarkID: 1, Front: "f3", Back: "b3", Status: "archive", Content: "{}"},
			{BookmarkID: 1, Front: "f4", Back: "b4", Status: "to_review", Content: "{}"},
		}
		for _, r := range rows {
			if err := db.Create(&r).Error; err != nil {
				t.Fatalf("failed seed: %v", err)
			}
		}

		// run the SQL migration script using Exec
		sqlBytes, rerr := os.ReadFile("tool/migrations/0003_normalize_memory_card_status.sql")
		if rerr != nil {
			t.Fatalf("failed read migration file: %v", rerr)
		}
		if _, err := sqlDB.Exec(string(sqlBytes)); err != nil {
			t.Fatalf("failed exec migration: %v", err)
		}

		// verify normalization happened
		var counts []struct {
			Status string
			Count  int
		}
		if err := db.Raw("SELECT status, COUNT(*) as count FROM memory_card_models GROUP BY status").Scan(&counts).Error; err != nil {
			t.Fatalf("failed counts: %v", err)
		}
		// minimal checks: ensure canonical statuses present
		found := map[string]bool{}
		for _, c := range counts {
			found[c.Status] = true
		}
		if !found["new"] {
			t.Fatalf("expected 'new' status after migration")
		}
		if !found["learning"] {
			t.Fatalf("expected 'learning' status after migration")
		}
		if !found["archived"] {
			t.Fatalf("expected 'archived' status after migration")
		}
	})

}
