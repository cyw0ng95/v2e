//go:build CONFIG_USE_LIBXML2

package capec

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestImportGatingAndForce(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestImportGatingAndForce", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_test.db")
		xmlPath := filepath.Join(dir, "capec_sample.xml")

		// Create a minimal CAPEC catalog with Version attribute and one Attack_Pattern
		xmlContent := `<?xml version="1.0"?>
	    <Attack_Pattern_Catalog Version="v-test" xmlns="http://capec.mitre.org/capec-3">
	      <Attack_Patterns>
	        <Attack_Pattern ID="1" Name="Test" Abstraction="Detailed" Status="Stable">
	          <Description>desc</Description>
	        </Attack_Pattern>
	      </Attack_Patterns>
	    </Attack_Pattern_Catalog>`
		if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
			t.Fatalf("failed to write xml: %v", err)
		}

		store, err := NewLocalCAPECStore(dbPath)
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		// First import should succeed
		if err := store.ImportFromXML(xmlPath, false); err != nil {
			t.Fatalf("first import failed: %v", err)
		}
		meta, err := store.GetCatalogMeta(nil)
		if err != nil {
			t.Fatalf("failed to get meta after import: %v", err)
		}
		if meta.Version != "v-test" {
			t.Fatalf("meta version mismatch: %s", meta.Version)
		}
		prevImported := meta.ImportedAtUTC

		// Second import without force should be skipped (ImportedAtUTC unchanged)
		time.Sleep(1 * time.Second)
		if err := store.ImportFromXML(xmlPath, false); err != nil {
			t.Fatalf("second import (should skip) returned error: %v", err)
		}
		meta2, err := store.GetCatalogMeta(nil)
		if err != nil {
			t.Fatalf("failed to get meta after second import: %v", err)
		}
		if meta2.ImportedAtUTC != prevImported {
			t.Fatalf("expected ImportedAtUTC unchanged but it changed: %d -> %d", prevImported, meta2.ImportedAtUTC)
		}

		// Force import should update ImportedAtUTC
		time.Sleep(1 * time.Second)
		if err := store.ImportFromXML(xmlPath, true); err != nil {
			t.Fatalf("force import failed: %v", err)
		}
		meta3, err := store.GetCatalogMeta(nil)
		if err != nil {
			t.Fatalf("failed to get meta after force import: %v", err)
		}
		if meta3.ImportedAtUTC == prevImported {
			t.Fatalf("expected ImportedAtUTC to be updated on force import")
		}
	})

}
