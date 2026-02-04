// Package parser provides HTML parsing tests for SSG table files.
package parser

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractTableIDFromPath(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestExtractTableIDFromPath", nil, func(t *testing.T, tx *gorm.DB) {
		tests := []struct {
			name          string
			path          string
			wantID        string
			wantProduct   string
			wantTableType string
		}{
			{
				name:          "AL2023 CCE table",
				path:          "tables/table-al2023-cces.html",
				wantID:        "table-al2023-cces",
				wantProduct:   "al2023",
				wantTableType: "cces",
			},
			{
				name:          "RHEL8 NIST refs table",
				path:          "tables/table-rhel8-nistrefs.html",
				wantID:        "table-rhel8-nistrefs",
				wantProduct:   "rhel8",
				wantTableType: "nistrefs",
			},
			{
				name:          "RHEL8 STIG table",
				path:          "tables/table-rhel8-stig.html",
				wantID:        "table-rhel8-stig",
				wantProduct:   "rhel8",
				wantTableType: "stig",
			},
			{
				name:          "Complex type with underscores",
				path:          "tables/table-rhel8-nistrefs-ospp.html",
				wantID:        "table-rhel8-nistrefs-ospp",
				wantProduct:   "rhel8",
				wantTableType: "nistrefs-ospp",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				id, product, tableType := extractTableIDFromPath(tt.path)
				if id != tt.wantID {
					t.Errorf("extractTableIDFromPath() id = %v, want %v", id, tt.wantID)
				}
				if product != tt.wantProduct {
					t.Errorf("extractTableIDFromPath() product = %v, want %v", product, tt.wantProduct)
				}
				if tableType != tt.wantTableType {
					t.Errorf("extractTableIDFromPath() tableType = %v, want %v", tableType, tt.wantTableType)
				}
			})
		}
	})

}

func TestParseTableFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseTableFile", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a minimal test table HTML
		tempDir := t.TempDir()
		tablePath := filepath.Join(tempDir, "table-test-cces.html")

		htmlContent := `<!DOCTYPE html>
	<html>
	<head lang="en">
	<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">
	<title>CCE Identifiers in Test Guide</title>
	</head>
	<body>
		<div style="text-align: center; font-size: x-large; font-weight:bold">CCE Identifiers in Test Guide</div>
		<table>
			<thead>
				<th>Mapping</th>
				<th>Rule Title</th>
				<th>Description</th>
				<th>Rationale</th>
			</thead>
			<tbody>
				<tr>
					<td>CCE-80644-8</td>
					<td>Install the tmux Package</td>
					<td>To enable console screen locking, install the tmux package.</td>
					<td>The tmux package allows for a session lock to be implemented.</td>
				</tr>
				<tr>
					<td>CCE-80647-1</td>
					<td>Set Password Maximum Age</td>
					<td>To specify password maximum age for new accounts...</td>
					<td>Any password, no matter how complex, can eventually be cracked.</td>
				</tr>
			</tbody>
		</table>
	</body>
	</html>`

		if err := os.WriteFile(tablePath, []byte(htmlContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// Parse the table
		table, entries, err := ParseTableFile(tablePath)
		if err != nil {
			t.Fatalf("ParseTableFile() error = %v", err)
		}

		// Check table metadata
		if table.ID != "table-test-cces" {
			t.Errorf("table.ID = %v, want table-test-cces", table.ID)
		}
		if table.Product != "test" {
			t.Errorf("table.Product = %v, want test", table.Product)
		}
		if table.TableType != "cces" {
			t.Errorf("table.TableType = %v, want cces", table.TableType)
		}
		if !strings.Contains(table.Title, "CCE Identifiers") {
			t.Errorf("table.Title = %v, should contain 'CCE Identifiers'", table.Title)
		}
		// Table no longer stores HTML content - data is in individual entries

		// Check entries
		if len(entries) != 2 {
			t.Fatalf("expected 2 entries, got %d", len(entries))
		}

		// Check first entry
		if entries[0].TableID != "table-test-cces" {
			t.Errorf("entries[0].TableID = %v, want table-test-cces", entries[0].TableID)
		}
		if entries[0].Mapping != "CCE-80644-8" {
			t.Errorf("entries[0].Mapping = %v, want CCE-80644-8", entries[0].Mapping)
		}
		if entries[0].RuleTitle != "Install the tmux Package" {
			t.Errorf("entries[0].RuleTitle = %v, want 'Install the tmux Package'", entries[0].RuleTitle)
		}
		if entries[0].Description == "" {
			t.Error("entries[0].Description should not be empty")
		}
		if entries[0].Rationale == "" {
			t.Error("entries[0].Rationale should not be empty")
		}

		// Check second entry
		if entries[1].Mapping != "CCE-80647-1" {
			t.Errorf("entries[1].Mapping = %v, want CCE-80647-1", entries[1].Mapping)
		}
		if entries[1].RuleTitle != "Set Password Maximum Age" {
			t.Errorf("entries[1].RuleTitle = %v, want 'Set Password Maximum Age'", entries[1].RuleTitle)
		}
	})

}

func TestParseTableFile_EmptyTable(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseTableFile_EmptyTable", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a table with no entries
		tempDir := t.TempDir()
		tablePath := filepath.Join(tempDir, "table-empty-test.html")

		htmlContent := `<!DOCTYPE html>
	<html>
	<head><title>Empty Table</title></head>
	<body>
		<table>
			<thead>
				<th>Mapping</th>
				<th>Rule Title</th>
				<th>Description</th>
				<th>Rationale</th>
			</thead>
			<tbody>
			</tbody>
		</table>
	</body>
	</html>`

		if err := os.WriteFile(tablePath, []byte(htmlContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		table, entries, err := ParseTableFile(tablePath)
		if err != nil {
			t.Fatalf("ParseTableFile() error = %v", err)
		}

		if table == nil {
			t.Fatal("table should not be nil")
		}
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

}

func TestParseTableFile_InvalidFile(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseTableFile_InvalidFile", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with non-existent file
		_, _, err := ParseTableFile("/nonexistent/table.html")
		if err == nil {
			t.Error("expected error for non-existent file, got nil")
		}
	})

}

func TestParseTableFile_MalformedHTML(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseTableFile_MalformedHTML", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a malformed HTML file
		tempDir := t.TempDir()
		tablePath := filepath.Join(tempDir, "table-malformed.html")

		htmlContent := `<html><body><table><tr><td>incomplete`

		if err := os.WriteFile(tablePath, []byte(htmlContent), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		// goquery is forgiving, so this should still parse without error
		table, entries, err := ParseTableFile(tablePath)
		if err != nil {
			t.Errorf("ParseTableFile() unexpected error = %v", err)
		}
		if table == nil {
			t.Error("table should not be nil even for malformed HTML")
		}
		// Entries might be empty or incomplete
		t.Logf("Parsed %d entries from malformed HTML", len(entries))
	})

}
