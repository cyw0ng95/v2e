package attack

import (
	"context"
	"path/filepath"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/xuri/excelize/v2"
)

func TestImportFromXLSX_MissingFile(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestImportFromXLSX_MissingFile", nil, func(t *testing.T, tx *gorm.DB) {
		store, err := NewLocalAttackStore(":memory:")
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		err = store.ImportFromXLSX("/tmp/does-not-exist.xlsx", false)
		if err == nil || err.Error() == "" {
			t.Fatalf("expected missing file error, got %v", err)
		}
	})

}

func TestImportFromXLSX_AndQueries(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestImportFromXLSX_AndQueries", nil, func(t *testing.T, tx *gorm.DB) {
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
	})
}

func TestImportFromXLSX_MultipleTypes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ImportFromXLSX_MultipleTypes", nil, func(t *testing.T, tx *gorm.DB) {
		store, err := NewLocalAttackStore(":memory:")
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		tmpDir := t.TempDir()
		xlsxPath := filepath.Join(tmpDir, "attack_multiple.xlsx")

		f := excelize.NewFile()
		f.DeleteSheet("Sheet1")

		sheetName := "Tactics"
		f.NewSheet(sheetName)
		row := []interface{}{"ID", "Name", "Description", "Domain", "Created", "Modified"}
		if err := f.SetSheetRow(sheetName, "A1", &row); err != nil {
			t.Fatalf("failed to set header row: %v", err)
		}
		dataRow := []interface{}{"TA9999", "Test Tactic", "Desc", "enterprise", "2020-01-01", "2021-01-01"}
		if err := f.SetSheetRow(sheetName, "A2", &dataRow); err != nil {
			t.Fatalf("failed to set data row: %v", err)
		}

		sheetName2 := "Mitigations"
		f.NewSheet(sheetName2)
		row2 := []interface{}{"ID", "Name", "Description", "Domain", "Created", "Modified"}
		if err := f.SetSheetRow(sheetName2, "A1", &row2); err != nil {
			t.Fatalf("failed to set header row: %v", err)
		}
		dataRow2 := []interface{}{"M9999", "Test Mitigation", "Desc", "enterprise", "2020-01-01", "2021-01-01"}
		if err := f.SetSheetRow(sheetName2, "A2", &dataRow2); err != nil {
			t.Fatalf("failed to set data row: %v", err)
		}

		sheetName3 := "Software"
		f.NewSheet(sheetName3)
		row3 := []interface{}{"ID", "Name", "Description", "Type", "Domain", "Created", "Modified"}
		if err := f.SetSheetRow(sheetName3, "A1", &row3); err != nil {
			t.Fatalf("failed to set header row: %v", err)
		}
		dataRow3 := []interface{}{"S9999", "Test Software", "Desc", "malware", "enterprise", "2020-01-01", "2021-01-01"}
		if err := f.SetSheetRow(sheetName3, "A2", &dataRow3); err != nil {
			t.Fatalf("failed to set data row: %v", err)
		}

		sheetName4 := "Groups"
		f.NewSheet(sheetName4)
		row4 := []interface{}{"ID", "Name", "Description", "Domain", "Created", "Modified"}
		if err := f.SetSheetRow(sheetName4, "A1", &row4); err != nil {
			t.Fatalf("failed to set header row: %v", err)
		}
		dataRow4 := []interface{}{"G9999", "Test Group", "Desc", "enterprise", "2020-01-01", "2021-01-01"}
		if err := f.SetSheetRow(sheetName4, "A2", &dataRow4); err != nil {
			t.Fatalf("failed to set data row: %v", err)
		}

		if err := f.SaveAs(xlsxPath); err != nil {
			t.Fatalf("failed to save xlsx: %v", err)
		}

		ctx := context.Background()
		if err := store.ImportFromXLSX(xlsxPath, true); err != nil {
			t.Fatalf("ImportFromXLSX returned error: %v", err)
		}

		meta, err := store.GetImportMetadata(ctx)
		if err != nil {
			t.Fatalf("GetImportMetadata error: %v", err)
		}
		t.Logf("Import metadata: total records = %d", meta.TotalRecords)

		tac, err := store.GetTacticByID(ctx, "TA9999")
		if err != nil {
			t.Fatalf("GetTacticByID error: %v", err)
		}
		if tac.Name != "Test Tactic" {
			t.Fatalf("unexpected tactic data: %+v", tac)
		}

		mit, err := store.GetMitigationByID(ctx, "M9999")
		if err != nil {
			t.Fatalf("GetMitigationByID error: %v", err)
		}
		if mit.Name != "Test Mitigation" {
			t.Fatalf("unexpected mitigation data: %+v", mit)
		}

		soft, err := store.GetSoftwareByID(ctx, "S9999")
		if err != nil {
			t.Fatalf("GetSoftwareByID error: %v", err)
		}
		if soft.Name != "Test Software" || soft.Type != "malware" {
			t.Fatalf("unexpected software data: %+v", soft)
		}

		grp, err := store.GetGroupByID(ctx, "G9999")
		if err != nil {
			t.Fatalf("GetGroupByID error: %v", err)
		}
		if grp.Name != "Test Group" {
			t.Fatalf("unexpected group data: %+v", grp)
		}
	})
}
