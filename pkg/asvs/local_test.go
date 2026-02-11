package asvs

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestNewLocalASVSStore(t *testing.T) {
	testutils.Run(t, testutils.Level2, "NewLocalASVSStore_CreatesStore", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		if store == nil {
			t.Fatal("Store is nil")
		}

		count, err := store.Count(context.Background())
		if err != nil {
			t.Fatalf("Failed to count records: %v", err)
		}

		if count != 0 {
			t.Errorf("Expected 0 records, got %d", count)
		}
	})
}

func TestParseBoolColumn(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"x", true},
		{"X", true},
		{"✓", true},
		{"true", true},
		{"TRUE", true},
		{"yes", true},
		{"YES", true},
		{"1", true},
		{"", false},
		{"false", false},
		{"no", false},
		{"0", false},
		{"abc", false},
	}

	for _, tt := range tests {
		testutils.Run(t, testutils.Level1, "ParseBoolColumn_"+tt.input, nil, func(t *testing.T, tx *gorm.DB) {
			result := parseBoolColumn(tt.input)
			if result != tt.expected {
				t.Errorf("parseBoolColumn(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestListASVSPaginated(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ListASVSPaginated_AllRecords", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_list.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		testRecords := []ASVSRequirementModel{
			{
				RequirementID: "1.1.1",
				Chapter:       "V1",
				Section:       "Architecture",
				Description:   "Test requirement 1",
				Level1:        true,
				Level2:        true,
				Level3:        true,
				CWE:           "CWE-79",
			},
			{
				RequirementID: "1.1.2",
				Chapter:       "V1",
				Section:       "Architecture",
				Description:   "Test requirement 2",
				Level1:        false,
				Level2:        true,
				Level3:        true,
				CWE:           "CWE-89",
			},
			{
				RequirementID: "2.1.1",
				Chapter:       "V2",
				Section:       "Authentication",
				Description:   "Test requirement 3",
				Level1:        true,
				Level2:        true,
				Level3:        false,
				CWE:           "",
			},
		}

		for _, rec := range testRecords {
			if err := store.db.Create(&rec).Error; err != nil {
				t.Fatalf("Failed to insert test record: %v", err)
			}
		}

		requirements, total, err := store.ListASVSPaginated(ctx, 0, 10, "", 0)
		if err != nil {
			t.Fatalf("Failed to list requirements: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}

		if len(requirements) != 3 {
			t.Errorf("Expected 3 requirements, got %d", len(requirements))
		}
	})

	testutils.Run(t, testutils.Level2, "ListASVSPaginated_FilterByChapter", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_list_chapter.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		testRecords := []ASVSRequirementModel{
			{
				RequirementID: "1.1.1",
				Chapter:       "V1",
				Description:   "Test 1",
			},
			{
				RequirementID: "1.1.2",
				Chapter:       "V1",
				Description:   "Test 2",
			},
			{
				RequirementID: "2.1.1",
				Chapter:       "V2",
				Description:   "Test 3",
			},
		}

		for _, rec := range testRecords {
			if err := store.db.Create(&rec).Error; err != nil {
				t.Fatalf("Failed to insert test record: %v", err)
			}
		}

		_, total, err := store.ListASVSPaginated(ctx, 0, 10, "V1", 0)
		if err != nil {
			t.Fatalf("Failed to list requirements by chapter: %v", err)
		}

		if total != 2 {
			t.Errorf("Expected total 2 for V1, got %d", total)
		}
	})

	testutils.Run(t, testutils.Level2, "ListASVSPaginated_FilterByLevel", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_list_level.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		testRecords := []ASVSRequirementModel{
			{
				RequirementID: "1.1.1",
				Level1:        true,
			},
			{
				RequirementID: "1.1.2",
				Level1:        true,
			},
			{
				RequirementID: "2.1.1",
				Level1:        false,
			},
		}

		for _, rec := range testRecords {
			if err := store.db.Create(&rec).Error; err != nil {
				t.Fatalf("Failed to insert test record: %v", err)
			}
		}

		_, total, err := store.ListASVSPaginated(ctx, 0, 10, "", 1)
		if err != nil {
			t.Fatalf("Failed to list requirements by level: %v", err)
		}

		if total != 2 {
			t.Errorf("Expected total 2 for Level 1, got %d", total)
		}
	})
}

func TestGetByID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetByID_ExistingRecord", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_getbyid.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		testRecord := ASVSRequirementModel{
			RequirementID: "1.1.1",
			Chapter:       "V1",
			Section:       "Architecture",
			Description:   "Test requirement",
			Level1:        true,
			Level2:        true,
			Level3:        true,
			CWE:           "CWE-79",
		}

		if err := store.db.Create(&testRecord).Error; err != nil {
			t.Fatalf("Failed to insert test record: %v", err)
		}

		requirement, err := store.GetByID(ctx, "1.1.1")
		if err != nil {
			t.Fatalf("Failed to get requirement: %v", err)
		}

		if requirement.RequirementID != "1.1.1" {
			t.Errorf("Expected requirement ID 1.1.1, got %s", requirement.RequirementID)
		}

		if requirement.Description != "Test requirement" {
			t.Errorf("Expected description 'Test requirement', got %s", requirement.Description)
		}
	})

	testutils.Run(t, testutils.Level2, "GetByID_NonExistentRecord", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_getbyid_notfound.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		_, err = store.GetByID(ctx, "999.999.999")
		if err == nil {
			t.Error("Expected error for non-existent ID, got nil")
		}
	})
}

func TestGetByCWE(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetByCWE_MultipleMatches", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_getbycwe.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		testRecords := []ASVSRequirementModel{
			{
				RequirementID: "1.1.1",
				Chapter:       "V1",
				Description:   "Test requirement 1",
				CWE:           "CWE-79",
			},
			{
				RequirementID: "1.1.2",
				Chapter:       "V1",
				Description:   "Test requirement 2",
				CWE:           "CWE-79, CWE-89",
			},
			{
				RequirementID: "2.1.1",
				Chapter:       "V2",
				Description:   "Test requirement 3",
				CWE:           "CWE-89",
			},
		}

		for _, rec := range testRecords {
			if err := store.db.Create(&rec).Error; err != nil {
				t.Fatalf("Failed to insert test record: %v", err)
			}
		}

		reqs, err := store.GetByCWE(ctx, "CWE-79")
		if err != nil {
			t.Fatalf("Failed to get requirements by CWE: %v", err)
		}

		if len(reqs) != 2 {
			t.Errorf("Expected 2 requirements for CWE-79, got %d", len(reqs))
		}
	})
}

func TestImportFromCSV(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ImportFromCSV_Success", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_import.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		csvContent := `Requirement ID,Chapter,Section,Description,L1,L2,L3,CWE
1.1.1,V1,Architecture,Test requirement 1,x,x,x,CWE-79
1.1.2,V1,Architecture,Test requirement 2,x,x,,CWE-89
2.1.1,V2,Authentication,Test requirement 3,x,x,,
`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, csvContent)
		}))
		defer server.Close()

		err = store.ImportFromCSV(ctx, server.URL)
		if err != nil {
			t.Fatalf("Failed to import from CSV: %v", err)
		}

		count, err := store.Count(ctx)
		if err != nil {
			t.Fatalf("Failed to count records: %v", err)
		}

		if count != 3 {
			t.Errorf("Expected 3 records after import, got %d", count)
		}

		req, err := store.GetByID(ctx, "1.1.1")
		if err != nil {
			t.Fatalf("Failed to get imported record: %v", err)
		}

		if req.Description != "Test requirement 1" {
			t.Errorf("Expected description 'Test requirement 1', got %s", req.Description)
		}

		if !req.Level1 {
			t.Error("Expected Level1 to be true")
		}
	})

	testutils.Run(t, testutils.Level2, "ImportFromCSV_ErrorStatus", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_import_error.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		err = store.ImportFromCSV(ctx, server.URL)
		if err == nil {
			t.Error("Expected error for non-200 status, got nil")
		}

		if !strings.Contains(err.Error(), "unexpected status code") {
			t.Errorf("Expected error about status code, got: %v", err)
		}
	})

	testutils.Run(t, testutils.Level2, "ImportFromCSV_MissingColumns", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_import_missing.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		csvContent := `Chapter,Section
V1,Architecture
`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, csvContent)
		}))
		defer server.Close()

		err = store.ImportFromCSV(ctx, server.URL)
		if err == nil {
			t.Error("Expected error for missing required columns, got nil")
		}

		if !strings.Contains(err.Error(), "required columns not found") {
			t.Errorf("Expected error about missing columns, got: %v", err)
		}
	})

	testutils.Run(t, testutils.Level2, "ImportFromCSV_EmptyRecords", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_import_empty.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		csvContent := `Requirement ID,Chapter,Section,Description
`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, csvContent)
		}))
		defer server.Close()

		err = store.ImportFromCSV(ctx, server.URL)
		if err == nil {
			t.Error("Expected error for no valid records, got nil")
		}

		if !strings.Contains(err.Error(), "no valid records") {
			t.Errorf("Expected error about no records, got: %v", err)
		}
	})

	testutils.Run(t, testutils.Level2, "ImportFromCSV_UpdateExisting", nil, func(t *testing.T, tx *gorm.DB) {
		tmpDB := "/tmp/test_asvs_import_update.db"
		defer os.Remove(tmpDB)

		store, err := NewLocalASVSStore(tmpDB)
		if err != nil {
			t.Fatalf("Failed to create ASVS store: %v", err)
		}

		ctx := context.Background()

		testRecord := ASVSRequirementModel{
			RequirementID: "1.1.1",
			Chapter:       "V1",
			Description:   "Old description",
			Level1:        false,
		}

		if err := store.db.Create(&testRecord).Error; err != nil {
			t.Fatalf("Failed to insert test record: %v", err)
		}

		csvContent := `Requirement ID,Chapter,Section,Description,L1
1.1.1,V1,Architecture,New description,x
`

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/csv")
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, csvContent)
		}))
		defer server.Close()

		err = store.ImportFromCSV(ctx, server.URL)
		if err != nil {
			t.Fatalf("Failed to import from CSV: %v", err)
		}

		req, err := store.GetByID(ctx, "1.1.1")
		if err != nil {
			t.Fatalf("Failed to get updated record: %v", err)
		}

		if req.Description != "New description" {
			t.Errorf("Expected updated description 'New description', got %s", req.Description)
		}

		if !req.Level1 {
			t.Error("Expected Level1 to be true after update")
		}
	})
}
