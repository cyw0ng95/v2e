package local

import (
	"github.com/cyw0ng95/v2e/pkg/ssg"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"
)

func TestStore_SaveTable(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStore_SaveTable", nil, func(t *testing.T, _ interface{}) {
		store := setupTestStore(t)
		defer store.Close()

		table := &ssg.SSGTable{
			ID:          "test-table",
			Product:     "al2023",
			TableType:   "profile",
			Title:       "Test Table",
			Description: "A test table",
		}

		err := store.SaveTable(table)
		if err != nil {
			t.Fatalf("SaveTable() error = %v", err)
		}

		retrieved, err := store.GetTable("test-table")
		if err != nil {
			t.Fatalf("GetTable() error = %v", err)
		}

		if retrieved.Title != "Test Table" {
			t.Errorf("Title = %s, want Test Table", retrieved.Title)
		}
	})
}

func TestStore_ListTables(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStore_ListTables", nil, func(t *testing.T, _ interface{}) {
		store := setupTestStore(t)
		defer store.Close()

		tables := []ssg.SSGTable{
			{ID: "table1", Product: "al2023", TableType: "profile", Title: "Table 1"},
			{ID: "table2", Product: "al2023", TableType: "xccdf", Title: "Table 2"},
			{ID: "table3", Product: "rhel9", TableType: "profile", Title: "Table 3"},
		}

		for _, table := range tables {
			if err := store.SaveTable(&table); err != nil {
				t.Fatalf("SaveTable() error = %v", err)
			}
		}

		allTables, err := store.ListTables("", "")
		if err != nil {
			t.Fatalf("ListTables() error = %v", err)
		}

		if len(allTables) != 3 {
			t.Errorf("ListTables() count = %d, want 3", len(allTables))
		}

		filteredTables, err := store.ListTables("al2023", "")
		if err != nil {
			t.Fatalf("ListTables(product) error = %v", err)
		}

		if len(filteredTables) != 2 {
			t.Errorf("ListTables(product) count = %d, want 2", len(filteredTables))
		}
	})
}

func TestStore_SaveTableEntry(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStore_SaveTableEntry", nil, func(t *testing.T, _ interface{}) {
		store := setupTestStore(t)
		defer store.Close()

		table := &ssg.SSGTable{
			ID:        "test-table",
			Product:   "al2023",
			TableType: "profile",
			Title:     "Test Table",
		}
		if err := store.SaveTable(table); err != nil {
			t.Fatalf("SaveTable() error = %v", err)
		}

		entry := &ssg.SSGTableEntry{
			TableID:     "test-table",
			Mapping:     "cce-12345-6",
			RuleTitle:   "Install AIDE",
			Rationale:   "For system integrity",
			Description: "Test description",
		}

		err := store.SaveTableEntry(entry)
		if err != nil {
			t.Fatalf("SaveTableEntry() error = %v", err)
		}

		entries, total, err := store.GetTableEntries("test-table", 0, 100)
		if err != nil {
			t.Fatalf("GetTableEntries() error = %v", err)
		}

		if total != 1 {
			t.Errorf("GetTableEntries() total = %d, want 1", total)
		}

		if len(entries) != 1 {
			t.Errorf("GetTableEntries() count = %d, want 1", len(entries))
		}
	})
}
