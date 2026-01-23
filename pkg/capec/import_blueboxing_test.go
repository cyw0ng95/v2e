//go:build libxml2

package capec

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportBlueBoxingDescription(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "capec_test.db")
	xmlPath := filepath.Join(dir, "capec_blueboxing.xml")

	xmlContent := `<?xml version="1.0"?>
    <Attack_Pattern_Catalog Version="v-test" xmlns="http://capec.mitre.org/capec-3" xmlns:xhtml="http://www.w3.org/1999/xhtml">
      <Attack_Patterns>
        <Attack_Pattern ID="5" Name="Blue Boxing" Abstraction="Detailed" Status="Obsolete">
          <Description>
            <xhtml:p>This attack targets telephone switches and trunks by sending supervisory tones to usurp control of the line.</xhtml:p>
            <xhtml:p><xhtml:b>Historical</xhtml:b> example text preserved.</xhtml:p>
          </Description>
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

	if err := store.ImportFromXML(xmlPath, "", true); err != nil {
		t.Fatalf("import failed: %v", err)
	}

	item, err := store.GetByID(context.Background(), "CAPEC-5")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if item == nil {
		t.Fatalf("expected CAPEC item but got nil")
	}
	if strings.TrimSpace(item.Description) == "" {
		t.Fatalf("expected non-empty description for CAPEC-5")
	}
	// The stored description should contain xhtml markup or expected words
	if !(strings.Contains(strings.ToLower(item.Description), "telephone") || strings.Contains(strings.ToLower(item.Description), "blue boxing") || strings.Contains(strings.ToLower(item.Description), "supervisory")) {
		t.Fatalf("description did not contain expected words: %q", item.Description)
	}
}
