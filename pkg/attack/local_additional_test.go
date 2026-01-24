package attack

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestImportFromXLSX_MissingFile(t *testing.T) {
	store, err := NewLocalAttackStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	err = store.ImportFromXLSX("/tmp/does-not-exist.xlsx", false)
	if err == nil || err.Error() == "" {
		t.Fatalf("expected missing file error, got %v", err)
	}
}

func TestImportFromXLSX_AndQueries(t *testing.T) {
	store, err := NewLocalAttackStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	tmpDir := t.TempDir()
	xlsxPath := filepath.Join(tmpDir, "attack.xlsx")

	f := excelize.NewFile()
	// Create Techniques sheet with headers and a single row
	sheetName := "Techniques"
	f.SetSheetName("Sheet1", sheetName)
	row := []interface{}{"ID", "Name", "Description", "Domain", "Platform", "Created", "Modified", "Revoked", "Deprecated"}
	if err := f.SetSheetRow(sheetName, "A1", &row); err != nil {
		t.Fatalf("failed to set header row: %v", err)
	}
	dataRow := []interface{}{"T9999", "Test Technique", "Desc", "enterprise", "linux", "2020-01-01", "2021-01-01", "false", "true"}
	if err := f.SetSheetRow(sheetName, "A2", &dataRow); err != nil {
		t.Fatalf("failed to set data row: %v", err)
	}
	if err := f.SaveAs(xlsxPath); err != nil {
		t.Fatalf("failed to save xlsx: %v", err)
	}

	if err := store.ImportFromXLSX(xlsxPath, true); err != nil {
		t.Fatalf("ImportFromXLSX returned error: %v", err)
	}

	ctx := context.Background()
	teq, err := store.GetTechniqueByID(ctx, "T9999")
	if err != nil {
		t.Fatalf("GetTechniqueByID error: %v", err)
	}
	if teq.Name != "Test Technique" || teq.Domain != "enterprise" {
		t.Fatalf("unexpected technique data: %+v", teq)
	}
	if teq.Deprecated != true || teq.Revoked != false {
		t.Fatalf("unexpected bool flags: revoked=%v deprecated=%v", teq.Revoked, teq.Deprecated)
	}

	list, total, err := store.ListTechniquesPaginated(ctx, 0, 10)
	if err != nil {
		t.Fatalf("ListTechniquesPaginated error: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("unexpected pagination results total=%d len=%d", total, len(list))
	}

	meta, err := store.GetImportMetadata(ctx)
	if err != nil {
		t.Fatalf("GetImportMetadata error: %v", err)
	}
	if meta.TotalRecords == 0 {
		t.Fatalf("expected metadata TotalRecords to be recorded")
	}
}
