package cce

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestNewParser(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewParser", nil, func(t *testing.T, tx *gorm.DB) {
		parser := NewParser("test.xlsx")
		if parser == nil {
			t.Fatal("NewParser returned nil")
		}
		if parser.filePath != "test.xlsx" {
			t.Errorf("Expected filePath 'test.xlsx', got %s", parser.filePath)
		}
	})
}

func TestParseAll(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseAll", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_parse_all.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test Title 1", "Test Description 1", "NIST", "ACTIVE", "OS", "Ref1"},
			{"CCE-00001-0", "Test Title 2", "Test Description 2", "DISA", "DEPRECATED", "Application", "Ref2"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		entries, err := parser.ParseAll()
		if err != nil {
			t.Fatalf("ParseAll failed: %v", err)
		}

		if len(entries) != 2 {
			t.Errorf("Expected 2 entries, got %d", len(entries))
		}

		if entries[0].ID != "CCE-00000-0" {
			t.Errorf("Expected ID 'CCE-00000-0', got %s", entries[0].ID)
		}
		if entries[0].Title != "Test Title 1" {
			t.Errorf("Expected Title 'Test Title 1', got %s", entries[0].Title)
		}
		if entries[0].Owner != "NIST" {
			t.Errorf("Expected Owner 'NIST', got %s", entries[0].Owner)
		}
	})
}

func TestParseAll_EmptyFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseAll_EmptyFile", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_empty.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		entries, err := parser.ParseAll()
		if err != nil {
			t.Fatalf("ParseAll failed: %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("Expected 0 entries, got %d", len(entries))
		}
	})
}

func TestParseAll_InvalidFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseAll_InvalidFile", nil, func(t *testing.T, tx *gorm.DB) {
		parser := NewParser("/non/existent/file.xlsx")
		_, err := parser.ParseAll()
		if err == nil {
			t.Error("Expected error for non-existent file, got nil")
		}
	})
}

func TestParseAll_SkipEmptyRows(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseAll_SkipEmptyRows", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_skip_rows.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"", "", "", "", "", "", ""}, // Empty row
			{"CCE-00000-0", "Test Title", "Test Description", "NIST", "ACTIVE", "OS", "Ref1"},
			{"", "", "", "", "", "", ""}, // Another empty row
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		entries, err := parser.ParseAll()
		if err != nil {
			t.Fatalf("ParseAll failed: %v", err)
		}

		if len(entries) != 1 {
			t.Errorf("Expected 1 entry (empty rows skipped), got %d", len(entries))
		}
	})
}

func TestParseBatch(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseBatch", nil, func(t *testing.T, tx *gorm.DB) {
		rows := [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test 1", "Description 1", "NIST", "ACTIVE", "OS", "Ref1"},
			{"CCE-00001-0", "Test 2", "Description 2", "DISA", "ACTIVE", "Application", "Ref2"},
			{"CCE-00002-0", "Test 3", "Description 3", "NIST", "ACTIVE", "OS", "Ref3"},
			{"CCE-00003-0", "Test 4", "Description 4", "DISA", "ACTIVE", "Application", "Ref4"},
			{"CCE-00004-0", "Test 5", "Description 5", "NIST", "ACTIVE", "OS", "Ref5"},
		}
		tmpFile := createTestExcelFile(t, "test_batch.xlsx", rows)
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)

		entries, total, err := parser.ParseBatch(0, 2)
		if err != nil {
			t.Fatalf("ParseBatch failed: %v", err)
		}

		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}

		if len(entries) != 2 {
			t.Errorf("Expected 2 entries, got %d", len(entries))
		}

		if entries[0].ID != "CCE-00000-0" {
			t.Errorf("Expected first entry ID 'CCE-00000-0', got %s", entries[0].ID)
		}

		secondBatch, _, err := parser.ParseBatch(2, 2)
		if err != nil {
			t.Fatalf("Second ParseBatch failed: %v", err)
		}

		if len(secondBatch) != 2 {
			t.Errorf("Expected 2 entries in second batch, got %d", len(secondBatch))
		}

		if secondBatch[0].ID != "CCE-00002-0" {
			t.Errorf("Expected first entry ID 'CCE-00002-0', got %s", secondBatch[0].ID)
		}
	})
}

func TestParseBatch_OffsetBeyondEnd(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseBatch_OffsetBeyondEnd", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_offset.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test", "Description", "NIST", "ACTIVE", "OS", "Ref1"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)

		entries, total, err := parser.ParseBatch(10, 2)
		if err != nil {
			t.Fatalf("ParseBatch failed: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}

		if entries != nil {
			t.Errorf("Expected nil entries when offset beyond end, got %v", entries)
		}
	})
}

func TestParseRowCount(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseRowCount", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_row_count.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test 1", "Description 1", "NIST", "ACTIVE", "OS", "Ref1"},
			{"CCE-00001-0", "Test 2", "Description 2", "DISA", "ACTIVE", "Application", "Ref2"},
			{"CCE-00002-0", "Test 3", "Description 3", "NIST", "ACTIVE", "OS", "Ref3"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		count, err := parser.ParseRowCount()
		if err != nil {
			t.Fatalf("ParseRowCount failed: %v", err)
		}

		if count != 3 {
			t.Errorf("Expected 3 rows, got %d", count)
		}
	})
}

func TestParseRowCount_EmptyFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseRowCount_EmptyFile", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_empty_rows.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		count, err := parser.ParseRowCount()
		if err != nil {
			t.Fatalf("ParseRowCount failed: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected 0 rows, got %d", count)
		}
	})
}

func TestParseByCCEID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseByCCEID", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_by_id.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test Title 1", "Test Description 1", "NIST", "ACTIVE", "OS", "Ref1"},
			{"CCE-00001-0", "Test Title 2", "Test Description 2", "DISA", "ACTIVE", "Application", "Ref2"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		entry, err := parser.ParseByCCEID("CCE-00001-0")
		if err != nil {
			t.Fatalf("ParseByCCEID failed: %v", err)
		}

		if entry == nil {
			t.Fatal("ParseByCCEID returned nil")
		}

		if entry.ID != "CCE-00001-0" {
			t.Errorf("Expected ID 'CCE-00001-0', got %s", entry.ID)
		}

		if entry.Title != "Test Title 2" {
			t.Errorf("Expected Title 'Test Title 2', got %s", entry.Title)
		}

		if entry.Owner != "DISA" {
			t.Errorf("Expected Owner 'DISA', got %s", entry.Owner)
		}
	})
}

func TestParseByCCEID_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseByCCEID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_not_found.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test Title", "Test Description", "NIST", "ACTIVE", "OS", "Ref1"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		_, err := parser.ParseByCCEID("CCE-99999-9")
		if err == nil {
			t.Error("Expected error for non-existent CCE ID, got nil")
		}
	})
}

func TestGetOffsetFromCCEID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetOffsetFromCCEID", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_offset_id.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test 1", "Description 1", "NIST", "ACTIVE", "OS", "Ref1"},
			{"CCE-00001-0", "Test 2", "Description 2", "DISA", "ACTIVE", "Application", "Ref2"},
			{"CCE-00002-0", "Test 3", "Description 3", "NIST", "ACTIVE", "OS", "Ref3"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)

		offset, err := parser.GetOffsetFromCCEID("CCE-00001-0")
		if err != nil {
			t.Fatalf("GetOffsetFromCCEID failed: %v", err)
		}

		if offset != 1 {
			t.Errorf("Expected offset 1, got %d", offset)
		}

		offset, err = parser.GetOffsetFromCCEID("CCE-00000-0")
		if err != nil {
			t.Fatalf("GetOffsetFromCCEID for first entry failed: %v", err)
		}

		if offset != 0 {
			t.Errorf("Expected offset 0 for first entry, got %d", offset)
		}
	})
}

func TestGetOffsetFromCCEID_NotFound(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetOffsetFromCCEID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_offset_not_found.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test", "Description", "NIST", "ACTIVE", "OS", "Ref1"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		_, err := parser.GetOffsetFromCCEID("CCE-99999-9")
		if err == nil {
			t.Error("Expected error for non-existent CCE ID, got nil")
		}
	})
}

func TestParseNumericOffset(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseNumericOffset", nil, func(t *testing.T, tx *gorm.DB) {
		offset, err := ParseNumericOffset("123")
		if err != nil {
			t.Fatalf("ParseNumericOffset failed: %v", err)
		}

		if offset != 123 {
			t.Errorf("Expected offset 123, got %d", offset)
		}

		offset, err = ParseNumericOffset("0")
		if err != nil {
			t.Fatalf("ParseNumericOffset for zero failed: %v", err)
		}

		if offset != 0 {
			t.Errorf("Expected offset 0, got %d", offset)
		}

		_, err = ParseNumericOffset("invalid")
		if err == nil {
			t.Error("Expected error for invalid numeric offset, got nil")
		}
	})
}

func TestGetCellValue(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetCellValue", nil, func(t *testing.T, tx *gorm.DB) {
		row := []string{"val1", "val2", "val3", "val4"}

		result := getCellValue(row, 0)
		if result != "val1" {
			t.Errorf("Expected 'val1', got '%s'", result)
		}

		result = getCellValue(row, 3)
		if result != "val4" {
			t.Errorf("Expected 'val4', got '%s'", result)
		}

		result = getCellValue(row, -1)
		if result != "" {
			t.Errorf("Expected empty string for negative index, got '%s'", result)
		}

		result = getCellValue(row, 10)
		if result != "" {
			t.Errorf("Expected empty string for out-of-bounds index, got '%s'", result)
		}
	})
}

func TestParseAll_InsufficientColumns(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseAll_InsufficientColumns", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_insufficient_cols.xlsx", [][]string{
			{"CCE ID", "Title", "Description"},
			{"CCE-00000-0", "Test Title"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)
		entries, err := parser.ParseAll()
		if err != nil {
			t.Fatalf("ParseAll failed: %v", err)
		}

		if len(entries) != 0 {
			t.Errorf("Expected 0 entries (insufficient columns), got %d", len(entries))
		}
	})
}

func TestParseBatch_ZeroBatchSize(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseBatch_ZeroBatchSize", nil, func(t *testing.T, tx *gorm.DB) {
		tmpFile := createTestExcelFile(t, "test_zero_batch.xlsx", [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Test", "Description", "NIST", "ACTIVE", "OS", "Ref1"},
		})
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)

		entries, total, err := parser.ParseBatch(0, 0)
		if err != nil {
			t.Fatalf("ParseBatch failed: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}

		if len(entries) != 0 {
			t.Errorf("Expected 0 entries for zero batch size, got %d", len(entries))
		}
	})
}

func TestParserIntegration(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParserIntegration", nil, func(t *testing.T, tx *gorm.DB) {
		rows := [][]string{
			{"CCE ID", "Title", "Description", "Owner", "Status", "Type", "Reference"},
			{"CCE-00000-0", "Integration Test 1", "Description 1", "NIST", "ACTIVE", "OS", "Ref1"},
			{"CCE-00001-0", "Integration Test 2", "Description 2", "DISA", "DEPRECATED", "Application", "Ref2"},
			{"CCE-00002-0", "Integration Test 3", "Description 3", "NIST", "ACTIVE", "OS", "Ref3"},
		}
		tmpFile := createTestExcelFile(t, "test_integration.xlsx", rows)
		defer os.Remove(tmpFile)

		parser := NewParser(tmpFile)

		count, err := parser.ParseRowCount()
		if err != nil {
			t.Fatalf("ParseRowCount failed: %v", err)
		}

		if count != 3 {
			t.Errorf("Expected 3 rows, got %d", count)
		}

		entry, err := parser.ParseByCCEID("CCE-00001-0")
		if err != nil {
			t.Fatalf("ParseByCCEID failed: %v", err)
		}

		if entry.Title != "Integration Test 2" {
			t.Errorf("Expected 'Integration Test 2', got %s", entry.Title)
		}

		entries, total, err := parser.ParseBatch(0, 2)
		if err != nil {
			t.Fatalf("ParseBatch failed: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}

		if len(entries) != 2 {
			t.Errorf("Expected 2 entries, got %d", len(entries))
		}

		allEntries, err := parser.ParseAll()
		if err != nil {
			t.Fatalf("ParseAll failed: %v", err)
		}

		if len(allEntries) != 3 {
			t.Errorf("Expected 3 entries, got %d", len(allEntries))
		}
	})
}

func createTestExcelFile(t *testing.T, filename string, data [][]string) string {
	t.Helper()

	f := excelize.NewFile()
	sheetName := "Sheet1"

	for rowIndex, row := range data {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+1)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, filename)

	if err := f.SaveAs(filePath); err != nil {
		t.Fatalf("Failed to save test Excel file: %v", err)
	}

	f.Close()

	return filePath
}

func TestConstants(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestConstants", nil, func(t *testing.T, tx *gorm.DB) {
		if CCEIDIndex != 0 {
			t.Errorf("Expected CCEIDIndex 0, got %d", CCEIDIndex)
		}
		if TitleIndex != 1 {
			t.Errorf("Expected TitleIndex 1, got %d", TitleIndex)
		}
		if DescriptionIndex != 2 {
			t.Errorf("Expected DescriptionIndex 2, got %d", DescriptionIndex)
		}
		if OwnerIndex != 3 {
			t.Errorf("Expected OwnerIndex 3, got %d", OwnerIndex)
		}
		if StatusIndex != 4 {
			t.Errorf("Expected StatusIndex 4, got %d", StatusIndex)
		}
		if TypeIndex != 5 {
			t.Errorf("Expected TypeIndex 5, got %d", TypeIndex)
		}
		if ReferenceIndex != 6 {
			t.Errorf("Expected ReferenceIndex 6, got %d", ReferenceIndex)
		}
	})
}
