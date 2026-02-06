package cwe

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"path/filepath"
	"testing"
)

func TestSaveGetDeleteView(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSaveGetDeleteView", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "test_cwe_views.db")
		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("NewLocalCWEStore error: %v", err)
		}
		// ensure view tables exist
		if err := AutoMigrateViews(store.db); err != nil { // store.db is accessible in same package
			t.Fatalf("AutoMigrateViews error: %v", err)
		}
		ctx := context.Background()

		v := &CWEView{
			ID:   "V-TEST-1",
			Name: "Test View",
			Type: "Example",
			Members: []ViewMember{
				{CWEID: "100", Role: "include"},
			},
			Audience: []Stakeholder{
				{Type: "Dev", Description: "Developers"},
			},
		}
		if err := store.SaveView(ctx, v); err != nil {
			t.Fatalf("SaveView error: %v", err)
		}
		got, err := store.GetViewByID(ctx, v.ID)
		if err != nil {
			t.Fatalf("GetViewByID error: %v", err)
		}
		if got.ID != v.ID || got.Name != v.Name {
			t.Fatalf("unexpected view returned: %+v", got)
		}
		list, total, err := store.ListViewsPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("ListViewsPaginated error: %v", err)
		}
		if total == 0 || len(list) == 0 {
			t.Fatalf("expected at least one view, got total=%d len=%d", total, len(list))
		}
		if err := store.DeleteView(ctx, v.ID); err != nil {
			t.Fatalf("DeleteView error: %v", err)
		}
		// ensure deleted
		if _, err := store.GetViewByID(ctx, v.ID); err == nil {
			t.Fatalf("expected error after delete, got nil")
		}
	})

}
