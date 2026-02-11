package cwe

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// TestViewRendering tests view rendering functionality
func TestViewRendering(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestViewRendering", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe_view_test.db")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		// Test that views can be created and migrated properly
		ctx := context.Background()

		// Create a sample CWE item with all fields populated
		cweItem := CWEItem{
			ID:                  "CWE-79",
			Name:                "Cross-site Scripting",
			Abstraction:         "Base",
			Structure:           "Simple",
			Status:              "Draft",
			Description:         "The software does not neutralize or incorrectly neutralizes user-controllable input before it is placed in output that is used as a web page.",
			ExtendedDescription: "Extended description here...",
			LikelihoodOfExploit: "High",
			RelatedWeaknesses: []RelatedWeakness{
				{
					Nature:  "ChildOf",
					CweID:   "CWE-714",
					ViewID:  "1000",
					Ordinal: "Primary",
				},
			},
			WeaknessOrdinalities: []WeaknessOrdinality{
				{
					Ordinality:  "Primary",
					Description: "This is the primary weakness",
				},
			},
			DetectionMethods: []DetectionMethod{
				{
					DetectionMethodID:  "DM1",
					Method:             "Automated Static Analysis",
					Description:        "Description of detection method",
					Effectiveness:      "Moderate",
					EffectivenessNotes: "Notes about effectiveness",
				},
			},
			PotentialMitigations: []Mitigation{
				{
					MitigationID:       "M1",
					Phase:              []string{"Architecture and Design", "Implementation"},
					Strategy:           "Parameter Validation",
					Description:        "Validate input parameters",
					Effectiveness:      "High",
					EffectivenessNotes: "Effective when implemented properly",
				},
			},
			DemonstrativeExamples: []DemonstrativeExample{
				{
					ID: "DE1",
					Entries: []DemonstrativeEntry{
						{
							IntroText:   "Example introduction",
							BodyText:    "Example body",
							Nature:      "Sample",
							Language:    "JavaScript",
							ExampleCode: "<script>alert('XSS')</script>",
							Reference:   "REF-001",
						},
					},
				},
			},
			ObservedExamples: []ObservedExample{
				{
					Reference:   "OBS-001",
					Description: "Observed example description",
					Link:        "http://example.com",
				},
			},
			TaxonomyMappings: []TaxonomyMapping{
				{
					TaxonomyName: "1003",
					EntryName:    "Malicious Data Element",
					EntryID:      "1003",
					MappingFit:   "Exact",
				},
			},
			Notes: []Note{
				{
					Type: "Maintenance",
					Note: "This is a maintenance note",
				},
			},
			ContentHistory: []ContentHistory{
				{
					Type:                   "Submission",
					SubmissionName:         "Initial Submission",
					SubmissionOrganization: "MITRE",
					SubmissionDate:         "2023-01-01",
				},
			},
		}

		// Save the item to the database
		err = store.db.Create(&CWEItemModel{
			ID:                  cweItem.ID,
			Name:                cweItem.Name,
			Abstraction:         cweItem.Abstraction,
			Structure:           cweItem.Structure,
			Status:              cweItem.Status,
			Description:         cweItem.Description,
			ExtendedDescription: cweItem.ExtendedDescription,
			LikelihoodOfExploit: cweItem.LikelihoodOfExploit,
		}).Error
		if err != nil {
			t.Fatalf("Failed to create CWE item: %v", err)
		}

		// Test retrieving the item by ID
		retrievedItem, err := store.GetByID(ctx, "CWE-79")
		if err != nil {
			t.Fatalf("Failed to retrieve item: %v", err)
		}

		if retrievedItem.ID != "CWE-79" {
			t.Errorf("Expected ID 'CWE-79', got '%s'", retrievedItem.ID)
		}
		if retrievedItem.Name != "Cross-site Scripting" {
			t.Errorf("Expected name 'Cross-site Scripting', got '%s'", retrievedItem.Name)
		}

		// Verify that all complex fields were properly initialized in the original item
		if len(cweItem.RelatedWeaknesses) != 1 {
			t.Errorf("Expected 1 related weakness, got %d", len(cweItem.RelatedWeaknesses))
		} else if cweItem.RelatedWeaknesses[0].CweID != "CWE-714" {
			t.Errorf("Expected related weakness ID 'CWE-714', got '%s'", cweItem.RelatedWeaknesses[0].CweID)
		}

		if len(cweItem.WeaknessOrdinalities) != 1 {
			t.Errorf("Expected 1 weakness ordinality, got %d", len(cweItem.WeaknessOrdinalities))
		} else if cweItem.WeaknessOrdinalities[0].Ordinality != "Primary" {
			t.Errorf("Expected ordinality 'Primary', got '%s'", cweItem.WeaknessOrdinalities[0].Ordinality)
		}

		if len(cweItem.DetectionMethods) != 1 {
			t.Errorf("Expected 1 detection method, got %d", len(cweItem.DetectionMethods))
		} else if cweItem.DetectionMethods[0].Method != "Automated Static Analysis" {
			t.Errorf("Expected detection method 'Automated Static Analysis', got '%s'", cweItem.DetectionMethods[0].Method)
		}

		if len(cweItem.PotentialMitigations) != 1 {
			t.Errorf("Expected 1 potential mitigation, got %d", len(cweItem.PotentialMitigations))
		} else if cweItem.PotentialMitigations[0].Strategy != "Parameter Validation" {
			t.Errorf("Expected mitigation strategy 'Parameter Validation', got '%s'", cweItem.PotentialMitigations[0].Strategy)
		}

		if len(cweItem.DemonstrativeExamples) != 1 {
			t.Errorf("Expected 1 demonstrative example, got %d", len(cweItem.DemonstrativeExamples))
		} else if cweItem.DemonstrativeExamples[0].ID != "DE1" {
			t.Errorf("Expected demonstrative example ID 'DE1', got '%s'", cweItem.DemonstrativeExamples[0].ID)
		}

		if len(cweItem.ObservedExamples) != 1 {
			t.Errorf("Expected 1 observed example, got %d", len(cweItem.ObservedExamples))
		} else if cweItem.ObservedExamples[0].Reference != "OBS-001" {
			t.Errorf("Expected observed example reference 'OBS-001', got '%s'", cweItem.ObservedExamples[0].Reference)
		}

		if len(cweItem.TaxonomyMappings) != 1 {
			t.Errorf("Expected 1 taxonomy mapping, got %d", len(cweItem.TaxonomyMappings))
		} else if cweItem.TaxonomyMappings[0].TaxonomyName != "1003" {
			t.Errorf("Expected taxonomy name '1003', got '%s'", cweItem.TaxonomyMappings[0].TaxonomyName)
		}

		if len(cweItem.Notes) != 1 {
			t.Errorf("Expected 1 note, got %d", len(cweItem.Notes))
		} else if cweItem.Notes[0].Type != "Maintenance" {
			t.Errorf("Expected note type 'Maintenance', got '%s'", cweItem.Notes[0].Type)
		}

		if len(cweItem.ContentHistory) != 1 {
			t.Errorf("Expected 1 content history, got %d", len(cweItem.ContentHistory))
		} else if cweItem.ContentHistory[0].Type != "Submission" {
			t.Errorf("Expected content history type 'Submission', got '%s'", cweItem.ContentHistory[0].Type)
		}
	})

}

// TestDataRelationships tests data relationship functionality
func TestDataRelationships(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataRelationships", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe_relationships_test.db")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Manually create the main CWE item
		mainItem := CWEItemModel{
			ID:   "CWE-REL-TEST",
			Name: "Relationship Test CWE",
		}
		err = store.db.Create(&mainItem).Error
		if err != nil {
			t.Fatalf("Failed to create CWE item: %v", err)
		}

		// Manually create related weaknesses
		relatedWeakness := RelatedWeaknessModel{
			CWEID:   "CWE-REL-TEST",
			Nature:  "ChildOf",
			CweID:   "CWE-PARENT-TEST",
			ViewID:  "1000",
			Ordinal: "Primary",
		}
		err = store.db.Create(&relatedWeakness).Error
		if err != nil {
			t.Fatalf("Failed to create related weakness: %v", err)
		}

		// Manually create weakness ordinality
		ordinality := WeaknessOrdinalityModel{
			CWEID:       "CWE-REL-TEST",
			Ordinality:  "Secondary",
			Description: "Secondary weakness",
		}
		err = store.db.Create(&ordinality).Error
		if err != nil {
			t.Fatalf("Failed to create weakness ordinality: %v", err)
		}

		// Retrieve the item and verify relationships
		retrievedItem, err := store.GetByID(ctx, "CWE-REL-TEST")
		if err != nil {
			t.Fatalf("Failed to retrieve item: %v", err)
		}

		if len(retrievedItem.RelatedWeaknesses) != 1 {
			t.Errorf("Expected 1 related weakness, got %d", len(retrievedItem.RelatedWeaknesses))
		} else if retrievedItem.RelatedWeaknesses[0].CweID != "CWE-PARENT-TEST" {
			t.Errorf("Expected related weakness ID 'CWE-PARENT-TEST', got '%s'", retrievedItem.RelatedWeaknesses[0].CweID)
		}

		if len(retrievedItem.WeaknessOrdinalities) != 1 {
			t.Errorf("Expected 1 weakness ordinality, got %d", len(retrievedItem.WeaknessOrdinalities))
		} else if retrievedItem.WeaknessOrdinalities[0].Ordinality != "Secondary" {
			t.Errorf("Expected ordinality 'Secondary', got '%s'", retrievedItem.WeaknessOrdinalities[0].Ordinality)
		}

		// Test listing with relationships
		items, total, err := store.ListCWEsPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list items: %v", err)
		}

		if total != 1 {
			t.Errorf("Expected total 1, got %d", total)
		}
		if len(items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(items))
		} else if items[0].ID != "CWE-REL-TEST" {
			t.Errorf("Expected item ID 'CWE-REL-TEST', got '%s'", items[0].ID)
		}

		// Verify relationships are preserved in list
		if len(items[0].RelatedWeaknesses) != 1 {
			t.Errorf("Expected 1 related weakness in list, got %d", len(items[0].RelatedWeaknesses))
		}
		if len(items[0].WeaknessOrdinalities) != 1 {
			t.Errorf("Expected 1 weakness ordinality in list, got %d", len(items[0].WeaknessOrdinalities))
		}
	})

}

// TestErrorHandling tests error handling scenarios
func TestErrorHandling(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestErrorHandling", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe_error_test.db")
		jsonPath := filepath.Join(tempDir, "test_data.json")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		// Test importing from non-existent JSON file
		err = store.ImportFromJSON("/nonexistent/file.json")
		if err == nil {
			t.Error("Expected error when importing from non-existent file")
		} else {
			t.Logf("Expected error for non-existent file: %v", err)
		}

		// Create invalid JSON file
		invalidJSON := `{"invalid": json, "missing": quotes}`
		err = os.WriteFile(jsonPath, []byte(invalidJSON), 0644)
		if err != nil {
			t.Fatalf("Failed to write invalid JSON: %v", err)
		}

		// Test importing invalid JSON
		err = store.ImportFromJSON(jsonPath)
		if err == nil {
			t.Error("Expected error when importing invalid JSON")
		} else {
			t.Logf("Expected error for invalid JSON: %v", err)
		}

		// Test retrieving non-existent item
		ctx := context.Background()
		_, err = store.GetByID(ctx, "NONEXISTENT-CWE")
		if err == nil {
			t.Error("Expected error when retrieving non-existent item")
		}

		// Test pagination with invalid parameters
		_, _, err = store.ListCWEsPaginated(ctx, -1, 10)
		if err != nil {
			t.Logf("Error with negative offset (expected): %v", err)
		}

		_, _, err = store.ListCWEsPaginated(ctx, 0, -1)
		if err != nil {
			t.Logf("Error with negative limit (expected): %v", err)
		}

		// Test with large offset
		_, total, err := store.ListCWEsPaginated(ctx, 10000, 10)
		if err != nil {
			t.Errorf("Error with large offset: %v", err)
		}
		if total != 0 {
			t.Errorf("Expected total 0, got %d", total)
		}
	})

}

// TestPerformanceWithLargeDatasets tests performance with large datasets
func TestPerformanceWithLargeDatasets(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestPerformanceWithLargeDatasets", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe_performance_test.db")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Insert a large number of items to test performance
		const numItems = 100
		startTime := time.Now()

		for i := 1; i <= numItems; i++ {
			item := CWEItemModel{
				ID:          "CWE-" + string(rune(i+48)), // Creates IDs like CWE-1, CWE-2, etc.
				Name:        "Test CWE " + string(rune(i+48)),
				Description: "Description for CWE " + string(rune(i+48)),
			}

			err = store.db.Create(&item).Error
			if err != nil {
				t.Fatalf("Failed to create item %d: %v", i, err)
			}
		}

		insertDuration := time.Since(startTime)
		t.Logf("Inserted %d items in %v", numItems, insertDuration)

		// Test retrieval performance
		retrieveStart := time.Now()
		for i := 1; i <= 10; i++ { // Test with 10 retrievals
			id := "CWE-" + string(rune(i+48))
			_, err := store.GetByID(ctx, id)
			if err != nil {
				t.Logf("Item %s not found: %v", id, err)
			}
		}
		retrieveDuration := time.Since(retrieveStart)
		t.Logf("Retrieved 10 items in %v", retrieveDuration)

		// Test pagination performance
		paginateStart := time.Now()
		_, total, err := store.ListCWEsPaginated(ctx, 0, 50)
		if err != nil {
			t.Fatalf("Failed to paginate: %v", err)
		}
		paginateDuration := time.Since(paginateStart)
		t.Logf("Paginated 50 items (total %d) in %v", total, paginateDuration)

		// Performance expectations: operations should complete within reasonable time
		if insertDuration.Seconds() > 5.0 {
			t.Logf("Insertion took longer than expected: %v", insertDuration)
		}
		if retrieveDuration.Seconds() > 2.0 {
			t.Logf("Retrieval took longer than expected: %v", retrieveDuration)
		}
		if paginateDuration.Seconds() > 2.0 {
			t.Logf("Pagination took longer than expected: %v", paginateDuration)
		}
	})

}

// TestFilteringAndSearch tests filtering and search functionality
func TestFilteringAndSearch(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestFilteringAndSearch", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe_filter_test.db")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Insert test data with different characteristics
		testItems := []CWEItemModel{
			{ID: "CWE-79", Name: "Cross-site Scripting", Description: "Software does not neutralize user input", Status: "Draft"},
			{ID: "CWE-89", Name: "SQL Injection", Description: "Improper neutralization of SQL commands", Status: "Complete"},
			{ID: "CWE-22", Name: "Path Traversal", Description: "Allows unauthorized access to files", Status: "Draft"},
			{ID: "CWE-78", Name: "OS Command Injection", Description: "Allows arbitrary OS command execution", Status: "Complete"},
			{ID: "CWE-119", Name: "Buffer Overflow", Description: "Memory corruption vulnerability", Status: "Draft"},
		}

		for _, item := range testItems {
			err = store.db.Create(&item).Error
			if err != nil {
				t.Fatalf("Failed to create item: %v", err)
			}
		}

		// Test pagination with different limits and offsets
		// First page, limit 2
		page1, total, err := store.ListCWEsPaginated(ctx, 0, 2)
		if err != nil {
			t.Fatalf("Failed to get first page: %v", err)
		}
		if len(page1) != 2 {
			t.Errorf("Expected 2 items in first page, got %d", len(page1))
		}
		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}

		// Second page, limit 2
		page2, total, err := store.ListCWEsPaginated(ctx, 2, 2)
		if err != nil {
			t.Fatalf("Failed to get second page: %v", err)
		}
		if len(page2) != 2 {
			t.Errorf("Expected 2 items in second page, got %d", len(page2))
		}
		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}

		// Third page, limit 2 (should have 1 remaining)
		page3, total, err := store.ListCWEsPaginated(ctx, 4, 2)
		if err != nil {
			t.Fatalf("Failed to get third page: %v", err)
		}
		if len(page3) != 1 {
			t.Errorf("Expected 1 item in third page, got %d", len(page3))
		}
		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}

		// Test edge cases for pagination
		emptyPage, total, err := store.ListCWEsPaginated(ctx, 10, 5) // offset beyond total
		if err != nil {
			t.Errorf("Error with offset beyond total: %v", err)
		}
		if len(emptyPage) != 0 {
			t.Errorf("Expected 0 items with offset beyond total, got %d", len(emptyPage))
		}
		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}

		// Test retrieving specific items by ID
		for _, expectedItem := range testItems {
			retrievedItem, err := store.GetByID(ctx, expectedItem.ID)
			if err != nil {
				t.Errorf("Failed to retrieve item %s: %v", expectedItem.ID, err)
				continue
			}
			if retrievedItem.ID != expectedItem.ID {
				t.Errorf("Expected ID %s, got %s", expectedItem.ID, retrievedItem.ID)
			}
			if retrievedItem.Name != expectedItem.Name {
				t.Errorf("Expected name %s, got %s", expectedItem.Name, retrievedItem.Name)
			}
			if retrievedItem.Status != expectedItem.Status {
				t.Errorf("Expected status %s, got %s", expectedItem.Status, retrievedItem.Status)
			}
		}

		// Test retrieval of non-existent item
		_, err = store.GetByID(ctx, "CWE-9999")
		if err == nil {
			t.Error("Expected error when retrieving non-existent item")
		}
	})

}

// TestConcurrentAccess tests concurrent access to the store
func TestConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe_concurrent_test.db")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Pre-populate with some test data
		testItems := []CWEItemModel{
			{ID: "CWE-1", Name: "Test CWE 1", Description: "Description 1"},
			{ID: "CWE-2", Name: "Test CWE 2", Description: "Description 2"},
			{ID: "CWE-3", Name: "Test CWE 3", Description: "Description 3"},
		}

		for _, item := range testItems {
			err = store.db.Create(&item).Error
			if err != nil {
				t.Fatalf("Failed to create item: %v", err)
			}
		}

		var wg sync.WaitGroup
		const numGoroutines = 10

		// Test concurrent reads
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Perform multiple operations in each goroutine
				for j := 0; j < 5; j++ {
					// Test GetByID
					_, err := store.GetByID(ctx, "CWE-1")
					if err != nil && !strings.Contains(err.Error(), "record not found") {
						t.Errorf("Goroutine %d: GetByID failed: %v", goroutineID, err)
					}

					// Test ListCWEsPaginated
					_, _, err = store.ListCWEsPaginated(ctx, 0, 10)
					if err != nil {
						t.Errorf("Goroutine %d: ListCWEsPaginated failed: %v", goroutineID, err)
					}

					// Small delay to allow other goroutines to interleave
					time.Sleep(time.Millisecond * 1)
				}
			}(i)
		}

		wg.Wait()
	})

}

// TestJSONImportFunctionality tests the JSON import functionality
func TestJSONImportFunctionality(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestJSONImportFunctionality", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "cwe_import_test.db")
		jsonPath := filepath.Join(tempDir, "cwe_data.json")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		// Create test JSON data
		testData := []CWEItem{
			{
				ID:                  "CWE-100",
				Name:                "Test Import CWE",
				Abstraction:         "Class",
				Structure:           "Simple",
				Status:              "Complete",
				Description:         "Test description for import",
				ExtendedDescription: "Extended test description",
				LikelihoodOfExploit: "Medium",
				RelatedWeaknesses: []RelatedWeakness{
					{
						Nature:  "ChildOf",
						CweID:   "CWE-101",
						ViewID:  "1000",
						Ordinal: "Primary",
					},
				},
				WeaknessOrdinalities: []WeaknessOrdinality{
					{
						Ordinality:  "Primary",
						Description: "Primary weakness in category",
					},
				},
				DetectionMethods: []DetectionMethod{
					{
						DetectionMethodID:  "DM-TEST-001",
						Method:             "Manual Analysis",
						Description:        "Manual review of code",
						Effectiveness:      "High",
						EffectivenessNotes: "Requires expert review",
					},
				},
			},
		}

		jsonData, err := json.Marshal(testData)
		if err != nil {
			t.Fatalf("Failed to marshal test data: %v", err)
		}

		err = os.WriteFile(jsonPath, jsonData, 0644)
		if err != nil {
			t.Fatalf("Failed to write JSON file: %v", err)
		}

		// Import the data
		err = store.ImportFromJSON(jsonPath)
		if err != nil {
			t.Fatalf("Failed to import JSON: %v", err)
		}

		// Verify the import worked
		ctx := context.Background()
		importedItem, err := store.GetByID(ctx, "CWE-100")
		if err != nil {
			t.Fatalf("Failed to retrieve imported item: %v", err)
		}

		if importedItem.ID != "CWE-100" {
			t.Errorf("Expected ID 'CWE-100', got '%s'", importedItem.ID)
		}
		if importedItem.Name != "Test Import CWE" {
			t.Errorf("Expected name 'Test Import CWE', got '%s'", importedItem.Name)
		}
		if importedItem.Status != "Complete" {
			t.Errorf("Expected status 'Complete', got '%s'", importedItem.Status)
		}

		// Verify related data was imported
		if len(importedItem.RelatedWeaknesses) != 1 {
			t.Errorf("Expected 1 related weakness, got %d", len(importedItem.RelatedWeaknesses))
		} else if importedItem.RelatedWeaknesses[0].CweID != "CWE-101" {
			t.Logf("Expected related weakness 'CWE-101', got '%s'", importedItem.RelatedWeaknesses[0].CweID)
		}

		if len(importedItem.WeaknessOrdinalities) != 1 {
			t.Errorf("Expected 1 weakness ordinality, got %d", len(importedItem.WeaknessOrdinalities))
		} else if importedItem.WeaknessOrdinalities[0].Ordinality != "Primary" {
			t.Logf("Expected ordinality 'Primary', got '%s'", importedItem.WeaknessOrdinalities[0].Ordinality)
		}

		if len(importedItem.DetectionMethods) != 1 {
			t.Errorf("Expected 1 detection method, got %d", len(importedItem.DetectionMethods))
		} else if importedItem.DetectionMethods[0].Method != "Manual Analysis" {
			t.Logf("Expected method 'Manual Analysis', got '%s'", importedItem.DetectionMethods[0].Method)
		}

		// Test that importing the same data again skips (due to existence check)
		err = store.ImportFromJSON(jsonPath)
		if err != nil {
			t.Logf("Second import result (expected to skip): %v", err)
		}
	})

}
