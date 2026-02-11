package attack

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// TestAttackPatternValidation tests attack pattern validation functionality
func TestAttackPatternValidation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestAttackPatternValidation", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_validation_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Test valid attack technique
		validTechnique := AttackTechnique{
			ID:          "T1001",
			Name:        "Data Obfuscation",
			Description: "Adversaries may obfuscate commands",
			Domain:      "enterprise-attack",
			Platform:    "Windows",
			Created:     "2023-01-01",
			Modified:    "2023-06-01",
			Revoked:     false,
			Deprecated:  false,
		}

		// Insert the technique using the import mechanism
		dbTx := store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&validTechnique).Error; err != nil {
			t.Fatalf("Failed to create technique: %v", err)
		}
		dbTx.Commit()

		// Test retrieval
		retrievedTechnique, err := store.GetTechniqueByID(ctx, "T1001")
		if err != nil {
			t.Fatalf("Failed to retrieve technique: %v", err)
		}

		if retrievedTechnique.ID != validTechnique.ID {
			t.Errorf("Expected ID %s, got %s", validTechnique.ID, retrievedTechnique.ID)
		}
		if retrievedTechnique.Name != validTechnique.Name {
			t.Errorf("Expected name %s, got %s", validTechnique.Name, retrievedTechnique.Name)
		}

		// Test valid attack tactic
		validTactic := AttackTactic{
			ID:          "TA0001",
			Name:        "Initial Access",
			Description: "The adversary is trying to get into your network",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-06-01",
		}

		dbTx = store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&validTactic).Error; err != nil {
			t.Fatalf("Failed to create tactic: %v", err)
		}
		dbTx.Commit()

		// Test retrieval
		retrievedTactic, err := store.GetTacticByID(ctx, "TA0001")
		if err != nil {
			t.Fatalf("Failed to retrieve tactic: %v", err)
		}

		if retrievedTactic.ID != validTactic.ID {
			t.Errorf("Expected ID %s, got %s", validTactic.ID, retrievedTactic.ID)
		}
		if retrievedTactic.Name != validTactic.Name {
			t.Errorf("Expected name %s, got %s", validTactic.Name, retrievedTactic.Name)
		}

		// Test valid attack mitigation
		validMitigation := AttackMitigation{
			ID:          "M1001",
			Name:        "Access Management",
			Description: "Manage access to administrative environments",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-06-01",
		}

		dbTx = store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&validMitigation).Error; err != nil {
			t.Fatalf("Failed to create mitigation: %v", err)
		}
		dbTx.Commit()

		// Test retrieval
		retrievedMitigation, err := store.GetMitigationByID(ctx, "M1001")
		if err != nil {
			t.Fatalf("Failed to retrieve mitigation: %v", err)
		}

		if retrievedMitigation.ID != validMitigation.ID {
			t.Errorf("Expected ID %s, got %s", validMitigation.ID, retrievedMitigation.ID)
		}
		if retrievedMitigation.Name != validMitigation.Name {
			t.Errorf("Expected name %s, got %s", validMitigation.Name, retrievedMitigation.Name)
		}
	})

}

// TestDataTransformation tests data transformation functionality
func TestDataTransformation(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataTransformation", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_transform_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Test complex data transformation scenario
		techniques := []AttackTechnique{
			{
				ID:          "T1001.001",
				Name:        "Data Obfuscation: Steganography",
				Description: "Sub-technique of Data Obfuscation",
				Domain:      "enterprise-attack",
				Platform:    "Windows,Linux",
				Created:     "2023-01-01",
				Modified:    "2023-06-01",
				Revoked:     false,
				Deprecated:  false,
			},
			{
				ID:          "T1002",
				Name:        "Encrypted Channel",
				Description: "Use encrypted channels for C2 communication",
				Domain:      "enterprise-attack",
				Platform:    "Windows",
				Created:     "2023-02-01",
				Modified:    "2023-07-01",
				Revoked:     true, // Revoked technique
				Deprecated:  false,
			},
			{
				ID:          "T1003",
				Name:        "OS Credential Dumping",
				Description: "Dumping credentials from OS",
				Domain:      "enterprise-attack",
				Platform:    "Windows,Linux,macOS",
				Created:     "2023-03-01",
				Modified:    "2023-08-01",
				Revoked:     false,
				Deprecated:  true, // Deprecated technique
			},
		}

		// Insert multiple techniques
		dbTx := store.db.WithContext(ctx).Begin()
		for _, tech := range techniques {
			if err := dbTx.Create(&tech).Error; err != nil {
				t.Fatalf("Failed to create technique %s: %v", tech.ID, err)
			}
		}
		dbTx.Commit()

		// Test pagination with multiple records
		page1, total, err := store.ListTechniquesPaginated(ctx, 0, 2)
		if err != nil {
			t.Fatalf("Failed to list techniques: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
		if len(page1) != 2 {
			t.Errorf("Expected page size 2, got %d", len(page1))
		}

		page2, total, err := store.ListTechniquesPaginated(ctx, 2, 2)
		if err != nil {
			t.Fatalf("Failed to list techniques page 2: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected total 3, got %d", total)
		}
		if len(page2) != 1 {
			t.Errorf("Expected page size 1, got %d", len(page2))
		}

		// Combine pages and verify all techniques are present
		allPages := append(page1, page2...)
		if len(allPages) != 3 {
			t.Errorf("Expected total 3 techniques across pages, got %d", len(allPages))
		}

		// Verify all techniques are present
		expectedIDs := map[string]bool{
			"T1001.001": true,
			"T1002":     true,
			"T1003":     true,
		}

		actualIDs := make(map[string]bool)
		for _, tech := range allPages {
			actualIDs[tech.ID] = true
		}

		for expectedID := range expectedIDs {
			if !actualIDs[expectedID] {
				t.Errorf("Missing expected technique ID: %s", expectedID)
			}
		}
	})

}

// TestErrorConditions tests various error conditions
func TestErrorConditions(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestErrorConditions", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_error_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Test retrieving non-existent technique
		_, err = store.GetTechniqueByID(ctx, "NONEXISTENT-T1234")
		if err == nil {
			t.Error("Expected error when retrieving non-existent technique")
		}

		// Test retrieving non-existent tactic
		_, err = store.GetTacticByID(ctx, "NONEXISTENT-TA1234")
		if err == nil {
			t.Error("Expected error when retrieving non-existent tactic")
		}

		// Test retrieving non-existent mitigation
		_, err = store.GetMitigationByID(ctx, "NONEXISTENT-M1234")
		if err == nil {
			t.Error("Expected error when retrieving non-existent mitigation")
		}

		// Test importing non-existent XLSX file
		err = store.ImportFromXLSX("/nonexistent/file.xlsx", false)
		if err == nil {
			t.Error("Expected error when importing non-existent file")
		} else {
			if !strings.Contains(err.Error(), "XLSX file does not exist") {
				t.Errorf("Expected file not exist error, got: %v", err)
			}
		}

		// Test pagination with negative values
		_, _, err = store.ListTechniquesPaginated(ctx, -1, 10)
		if err == nil {
			t.Log("Negative offset did not produce an error (may be handled by GORM)")
		}

		_, _, err = store.ListTechniquesPaginated(ctx, 0, -1)
		if err == nil {
			t.Log("Negative limit did not produce an error (may be handled by GORM)")
		}
	})

}

// TestPerformanceBenchmark tests performance with large datasets
func TestPerformanceBenchmark(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestPerformanceBenchmark", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_performance_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Generate a large number of techniques for performance testing
		const numTechniques = 1000
		techniques := make([]AttackTechnique, numTechniques)

		for i := 0; i < numTechniques; i++ {
			techniques[i] = AttackTechnique{
				ID:          "T" + string(rune(1000+i)),
				Name:        "Performance Test Technique " + string(rune(1000+i)),
				Description: "Description for performance test technique " + string(rune(1000+i)),
				Domain:      "enterprise-attack",
				Platform:    "Windows",
				Created:     "2023-01-01",
				Modified:    "2023-06-01",
				Revoked:     false,
				Deprecated:  false,
			}
		}

		// Measure insertion performance
		startTime := time.Now()
		dbTx := store.db.WithContext(ctx).Begin()
		for _, tech := range techniques {
			if err := dbTx.Create(&tech).Error; err != nil {
				t.Fatalf("Failed to create technique: %v", err)
			}
		}
		dbTx.Commit()
		insertDuration := time.Since(startTime)

		t.Logf("Inserted %d techniques in %v", numTechniques, insertDuration)

		// Measure retrieval performance
		retrieveStartTime := time.Now()
		for i := 0; i < 10; i++ { // Test retrieval of 10 random techniques
			_, err := store.GetTechniqueByID(ctx, techniques[i].ID)
			if err != nil {
				t.Errorf("Failed to retrieve technique %s: %v", techniques[i].ID, err)
			}
		}
		retrieveDuration := time.Since(retrieveStartTime)

		t.Logf("Retrieved 10 techniques in %v", retrieveDuration)

		// Measure pagination performance
		paginateStartTime := time.Now()
		page1, total, err := store.ListTechniquesPaginated(ctx, 0, 100)
		if err != nil {
			t.Fatalf("Failed to paginate: %v", err)
		}
		paginateDuration := time.Since(paginateStartTime)

		t.Logf("Paginated 100 techniques (total %d) in %v", total, paginateDuration)

		if total != int64(numTechniques) {
			t.Errorf("Expected total %d, got %d", numTechniques, total)
		}
		if len(page1) != 100 {
			t.Errorf("Expected page size 100, got %d", len(page1))
		}

		// Performance thresholds (adjust based on expected performance)
		maxInsertDuration := time.Second * 5          // Allow up to 5 seconds for insertion
		maxRetrieveDuration := time.Millisecond * 50  // Allow up to 50ms for retrieval
		maxPaginateDuration := time.Millisecond * 100 // Allow up to 100ms for pagination

		if insertDuration > maxInsertDuration {
			t.Logf("Insertion took longer than expected: %v (threshold: %v)", insertDuration, maxInsertDuration)
		}
		if retrieveDuration > maxRetrieveDuration {
			t.Logf("Retrieval took longer than expected: %v (threshold: %v)", retrieveDuration, maxRetrieveDuration)
		}
		if paginateDuration > maxPaginateDuration {
			t.Logf("Pagination took longer than expected: %v (threshold: %v)", paginateDuration, maxPaginateDuration)
		}
	})

}

// TestConcurrentAccess tests concurrent access to the store
func TestConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_concurrent_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Pre-populate with some test data
		testTechniques := []AttackTechnique{
			{ID: "T1001", Name: "Test Technique 1", Domain: "enterprise-attack"},
			{ID: "T1002", Name: "Test Technique 2", Domain: "enterprise-attack"},
			{ID: "T1003", Name: "Test Technique 3", Domain: "enterprise-attack"},
		}

		dbTx := store.db.WithContext(ctx).Begin()
		for _, tech := range testTechniques {
			if err := dbTx.Create(&tech).Error; err != nil {
				t.Fatalf("Failed to create technique: %v", err)
			}
		}
		dbTx.Commit()

		var wg sync.WaitGroup
		const numGoroutines = 10

		// Test concurrent reads
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// Perform multiple operations in each goroutine
				for j := 0; j < 5; j++ {
					// Test GetTechniqueByID
					_, err := store.GetTechniqueByID(ctx, "T1001")
					if err != nil && !strings.Contains(err.Error(), "record not found") {
						t.Errorf("Goroutine %d: GetTechniqueByID failed: %v", goroutineID, err)
					}

					// Test ListTechniquesPaginated
					_, _, err = store.ListTechniquesPaginated(ctx, 0, 10)
					if err != nil {
						t.Errorf("Goroutine %d: ListTechniquesPaginated failed: %v", goroutineID, err)
					}

					// Small delay to allow other goroutines to interleave
					time.Sleep(time.Millisecond * 1)
				}
			}(i)
		}

		wg.Wait()
	})

}

// TestRelationshipFunctionality tests relationship functionality
func TestRelationshipFunctionality(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestRelationshipFunctionality", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_relationship_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Create test techniques and tactics
		technique := AttackTechnique{
			ID:          "T1055",
			Name:        "Process Injection",
			Description: "Adversaries may inject code into processes",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-06-01",
		}
		tactic := AttackTactic{
			ID:          "TA0005",
			Name:        "Defense Evasion",
			Description: "The adversary is trying to avoid being detected",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-06-01",
		}

		dbTx := store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&technique).Error; err != nil {
			t.Fatalf("Failed to create technique: %v", err)
		}
		if err := dbTx.Create(&tactic).Error; err != nil {
			t.Fatalf("Failed to create tactic: %v", err)
		}
		dbTx.Commit()

		// Create a relationship between tactic and technique (tactic has sub-technique)
		relationship := AttackRelationship{
			ID:               "rel-1",
			SourceRef:        "TA0005", // Tactic is the source (has sub-technique)
			TargetRef:        "T1055",  // Technique is the target (is a sub-technique)
			RelationshipType: "has-subtechnique",
			SourceObjectType: "x-mitre-tactic",
			TargetObjectType: "attack-pattern",
			Description:      "Defense Evasion has sub-technique Process Injection",
			Domain:           "enterprise-attack",
			Created:          "2023-01-01",
			Modified:         "2023-06-01",
		}

		dbTx = store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&relationship).Error; err != nil {
			t.Fatalf("Failed to create relationship: %v", err)
		}
		dbTx.Commit()

		// Test getting related techniques by tactic
		relatedTechniques, err := store.GetRelatedTechniquesByTactic(ctx, "TA0005")
		if err != nil {
			t.Fatalf("Failed to get related techniques: %v", err)
		}

		if len(relatedTechniques) != 1 {
			t.Errorf("Expected 1 related technique, got %d", len(relatedTechniques))
		} else if relatedTechniques[0].ID != "T1055" {
			t.Errorf("Expected technique ID 'T1055', got '%s'", relatedTechniques[0].ID)
		}

		// Test with non-existent tactic
		emptyRelated, err := store.GetRelatedTechniquesByTactic(ctx, "NONEXISTENT-TA9999")
		if err != nil {
			t.Fatalf("Error with non-existent tactic: %v", err)
		}
		if len(emptyRelated) != 0 {
			t.Errorf("Expected 0 related techniques for non-existent tactic, got %d", len(emptyRelated))
		}
	})

}

// TestSoftwareAndGroupFunctionality tests software and group functionality
func TestSoftwareAndGroupFunctionality(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestSoftwareAndGroupFunctionality", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_software_group_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Test software functionality
		software := AttackSoftware{
			ID:          "S0001",
			Name:        "Compiled HTML File",
			Description: "Adversaries may abuse Compiled HTML files",
			Type:        "tool",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-06-01",
		}

		dbTx := store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&software).Error; err != nil {
			t.Fatalf("Failed to create software: %v", err)
		}
		dbTx.Commit()

		// Retrieve software
		retrievedSoftware, err := store.GetSoftwareByID(ctx, "S0001")
		if err != nil {
			t.Fatalf("Failed to retrieve software: %v", err)
		}

		if retrievedSoftware.ID != software.ID {
			t.Errorf("Expected software ID %s, got %s", software.ID, retrievedSoftware.ID)
		}
		if retrievedSoftware.Name != software.Name {
			t.Errorf("Expected software name %s, got %s", software.Name, retrievedSoftware.Name)
		}

		// Test group functionality
		group := AttackGroup{
			ID:          "G0001",
			Name:        "APT1",
			Description: "APT1 is a Chinese threat group",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-06-01",
		}

		dbTx = store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&group).Error; err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}
		dbTx.Commit()

		// Retrieve group
		retrievedGroup, err := store.GetGroupByID(ctx, "G0001")
		if err != nil {
			t.Fatalf("Failed to retrieve group: %v", err)
		}

		if retrievedGroup.ID != group.ID {
			t.Errorf("Expected group ID %s, got %s", group.ID, retrievedGroup.ID)
		}
		if retrievedGroup.Name != group.Name {
			t.Errorf("Expected group name %s, got %s", group.Name, retrievedGroup.Name)
		}

		// Test pagination for software and groups
		softwareList, total, err := store.ListSoftwarePaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list software: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1 software, got %d", total)
		}
		if len(softwareList) != 1 {
			t.Errorf("Expected 1 software item, got %d", len(softwareList))
		}

		groupList, total, err := store.ListGroupsPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list groups: %v", err)
		}
		if total != 1 {
			t.Errorf("Expected total 1 group, got %d", total)
		}
		if len(groupList) != 1 {
			t.Errorf("Expected 1 group item, got %d", len(groupList))
		}
	})

}

// TestImportMetadata tests import metadata functionality
func TestImportMetadata(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestImportMetadata", nil, func(t *testing.T, tx *gorm.DB) {
		tempDir := t.TempDir()
		dbPath := filepath.Join(tempDir, "attack_metadata_test.db")

		store, err := NewLocalAttackStore(dbPath)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer func() {
			sqlDB, _ := store.db.DB()
			sqlDB.Close()
		}()

		ctx := context.Background()

		// Create test metadata
		metadata := AttackMetadata{
			ImportedAt:    time.Now().Unix(),
			SourceFile:    "/path/to/test/file.xlsx",
			TotalRecords:  100,
			ImportVersion: "1.0",
		}

		dbTx := store.db.WithContext(ctx).Begin()
		if err := dbTx.Create(&metadata).Error; err != nil {
			t.Fatalf("Failed to create metadata: %v", err)
		}
		dbTx.Commit()

		// Retrieve metadata
		retrievedMetadata, err := store.GetImportMetadata(ctx)
		if err != nil {
			t.Fatalf("Failed to retrieve metadata: %v", err)
		}

		if retrievedMetadata.SourceFile != metadata.SourceFile {
			t.Errorf("Expected source file %s, got %s", metadata.SourceFile, retrievedMetadata.SourceFile)
		}
		if retrievedMetadata.TotalRecords != metadata.TotalRecords {
			t.Errorf("Expected total records %d, got %d", metadata.TotalRecords, retrievedMetadata.TotalRecords)
		}
		if retrievedMetadata.ImportVersion != metadata.ImportVersion {
			t.Errorf("Expected import version %s, got %s", metadata.ImportVersion, retrievedMetadata.ImportVersion)
		}

		// Test with empty database (should return error)
		emptyTempDir := t.TempDir()
		emptyDbPath := filepath.Join(emptyTempDir, "attack_empty_test.db")

		emptyStore, err := NewLocalAttackStore(emptyDbPath)
		if err != nil {
			t.Fatalf("Failed to create empty store: %v", err)
		}
		defer func() {
			sqlDB, _ := emptyStore.db.DB()
			sqlDB.Close()
		}()

		_, err = emptyStore.GetImportMetadata(ctx)
		if err == nil {
			t.Error("Expected error when retrieving metadata from empty database")
		}
	})

}

// TestHelperFunctionsWithNilAndEmptySlices tests that helper functions
// handle nil and empty slices without panicking (TODO-007).
func TestHelperFunctionsWithNilAndEmptySlices(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestHelperFunctionsWithNilAndEmptySlices", nil, func(t *testing.T, tx *gorm.DB) {
		// Test getStringValue with nil headers
		t.Run("getStringValue with nil headers", func(t *testing.T) {
			// This should not panic
			result := getStringValue([]string{"value1", "value2"}, 0, nil, "ID")
			if result != "value1" {
				t.Errorf("Expected 'value1', got '%s'", result)
			}
		})

		// Test getStringValue with empty headers
		t.Run("getStringValue with empty headers", func(t *testing.T) {
			result := getStringValue([]string{"value1", "value2"}, 0, []string{}, "ID")
			if result != "value1" {
				t.Errorf("Expected 'value1', got '%s'", result)
			}
		})

		// Test getStringValue with nil row
		t.Run("getStringValue with nil row", func(t *testing.T) {
			result := getStringValue(nil, 0, []string{"ID"}, "ID")
			if result != "" {
				t.Errorf("Expected empty string, got '%s'", result)
			}
		})

		// Test getStringValue with nil row and nil headers
		t.Run("getStringValue with nil row and nil headers", func(t *testing.T) {
			result := getStringValue(nil, 0, nil, "ID")
			if result != "" {
				t.Errorf("Expected empty string, got '%s'", result)
			}
		})

		// Test getStringValue with out-of-bounds index
		t.Run("getStringValue with out-of-bounds index", func(t *testing.T) {
			result := getStringValue([]string{"value1"}, 10, nil, "ID")
			if result != "" {
				t.Errorf("Expected empty string, got '%s'", result)
			}
		})

		// Test getStringValue with nil entries in headers
		t.Run("getStringValue with nil entries in headers", func(t *testing.T) {
			headers := []string{"", "ID", ""}
			row := []string{"value1", "value2", "value3"}
			result := getStringValue(row, 0, headers, "ID")
			if result != "value2" {
				t.Errorf("Expected 'value2', got '%s'", result)
			}
		})

		// Test getStringIndex with nil headers
		t.Run("getStringIndex with nil headers", func(t *testing.T) {
			result := getStringIndex(nil, []string{"ID"})
			if result != -1 {
				t.Errorf("Expected -1, got %d", result)
			}
		})

		// Test getStringIndex with empty headers
		t.Run("getStringIndex with empty headers", func(t *testing.T) {
			result := getStringIndex([]string{}, []string{"ID"})
			if result != -1 {
				t.Errorf("Expected -1, got %d", result)
			}
		})

		// Test getStringIndex with nil possibleHeaders
		t.Run("getStringIndex with nil possibleHeaders", func(t *testing.T) {
			result := getStringIndex([]string{"ID", "Name"}, nil)
			if result != -1 {
				t.Errorf("Expected -1, got %d", result)
			}
		})

		// Test getStringIndex with empty possibleHeaders
		t.Run("getStringIndex with empty possibleHeaders", func(t *testing.T) {
			result := getStringIndex([]string{"ID", "Name"}, []string{})
			if result != -1 {
				t.Errorf("Expected -1, got %d", result)
			}
		})

		// Test getStringIndex with nil entries in headers
		t.Run("getStringIndex with nil entries in headers", func(t *testing.T) {
			headers := []string{"", "ID", ""}
			result := getStringIndex(headers, []string{"ID"})
			if result != 1 {
				t.Errorf("Expected 1, got %d", result)
			}
		})

		// Test getBoolValue with nil row
		t.Run("getBoolValue with nil row", func(t *testing.T) {
			result := getBoolValue(nil, 0)
			if result != false {
				t.Errorf("Expected false, got %v", result)
			}
		})

		// Test getBoolValue with negative index
		t.Run("getBoolValue with negative index", func(t *testing.T) {
			result := getBoolValue([]string{"true"}, -1)
			if result != false {
				t.Errorf("Expected false, got %v", result)
			}
		})

		// Test getBoolValue with out-of-bounds index
		t.Run("getBoolValue with out-of-bounds index", func(t *testing.T) {
			result := getBoolValue([]string{"true"}, 10)
			if result != false {
				t.Errorf("Expected false, got %v", result)
			}
		})

		// Test getBoolValue with empty row
		t.Run("getBoolValue with empty row", func(t *testing.T) {
			result := getBoolValue([]string{}, 0)
			if result != false {
				t.Errorf("Expected false, got %v", result)
			}
		})
	})
}

// TestImportFromXLSX_WithUnexpectedDataTypes tests that ImportFromXLSX
// handles Excel files with unexpected data types without panicking (TODO-007).
func TestImportFromXLSX_WithUnexpectedDataTypes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestImportFromXLSX_WithUnexpectedDataTypes", nil, func(t *testing.T, tx *gorm.DB) {
		store, err := NewLocalAttackStore(":memory:")
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		tmpDir := t.TempDir()
		xlsxPath := filepath.Join(tmpDir, "attack_unexpected_types.xlsx")

		f := excelize.NewFile()
		f.DeleteSheet("Sheet1")

		// Create a Techniques sheet with various edge cases
		sheetName := "Techniques"
		f.NewSheet(sheetName)

		// Set header row
		headers := []interface{}{"ID", "Name", "Description", "Domain", "Platform", "Created", "Modified", "Revoked", "Deprecated"}
		if err := f.SetSheetRow(sheetName, "A1", &headers); err != nil {
			t.Fatalf("failed to set header row: %v", err)
		}

		// Test cases for various unexpected data scenarios
		testCases := [][]interface{}{
			// Normal row
			{"T1001", "Normal Technique", "Description", "enterprise", "Windows", "2020-01-01", "2021-01-01", false, false},
			// Empty string ID (should be skipped)
			{"", "Empty ID Technique", "Description", "enterprise", "Windows", "2020-01-01", "2021-01-01", false, false},
			// Row with fewer columns than expected
			{"T1002", "Short Row", "Description"},
			// Boolean values as strings
			{"T1003", "Bool Technique", "Description", "enterprise", "Linux", "2020-01-01", "2021-01-01", "true", "false"},
			// Integer values (Excel may return numbers)
			{1004, "Numeric ID", "Description", "enterprise", "macOS", "2020-01-01", "2021-01-01", 0, 1},
		}

		for i, testCase := range testCases {
			rowNum := i + 2
			if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowNum), &testCase); err != nil {
				t.Fatalf("failed to set data row %d: %v", rowNum, err)
			}
		}

		if err := f.SaveAs(xlsxPath); err != nil {
			t.Fatalf("failed to save xlsx: %v", err)
		}

		// This should not panic even with unexpected data types
		if err := store.ImportFromXLSX(xlsxPath, true); err != nil {
			t.Fatalf("ImportFromXLSX returned error: %v", err)
		}

		// Verify that valid records were imported
		ctx := context.Background()

		// T1001 should be imported
		tech1, err := store.GetTechniqueByID(ctx, "T1001")
		if err != nil {
			t.Errorf("Failed to get T1001: %v", err)
		} else if tech1.Name != "Normal Technique" {
			t.Errorf("T1001 has unexpected name: %s", tech1.Name)
		}

		// T1002 with short row should be imported (partial data)
		tech2, err := store.GetTechniqueByID(ctx, "T1002")
		if err != nil {
			t.Errorf("Failed to get T1002: %v", err)
		} else if tech2.Name != "Short Row" {
			t.Errorf("T1002 has unexpected name: %s", tech2.Name)
		}

		// T1003 with bool strings should be imported
		tech3, err := store.GetTechniqueByID(ctx, "T1003")
		if err != nil {
			t.Errorf("Failed to get T1003: %v", err)
		} else if tech3.Name != "Bool Technique" {
			t.Errorf("T1003 has unexpected name: %s", tech3.Name)
		}

		// Numeric ID should be converted to string (if Excel returns it as number)
		// The GetRows function should convert everything to strings
		tech4, err := store.GetTechniqueByID(ctx, "1004")
		if err == nil {
			// If numeric ID was converted, verify the data
			if tech4.Name != "Numeric ID" {
				t.Errorf("1004 has unexpected name: %s", tech4.Name)
			}
		}
	})
}

// TestImportFromXLSX_WithEmptySheet tests that ImportFromXLSX
// handles Excel files with empty sheets without panicking (TODO-007).
func TestImportFromXLSX_WithEmptySheet(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestImportFromXLSX_WithEmptySheet", nil, func(t *testing.T, tx *gorm.DB) {
		store, err := NewLocalAttackStore(":memory:")
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		tmpDir := t.TempDir()
		xlsxPath := filepath.Join(tmpDir, "attack_empty_sheet.xlsx")

		f := excelize.NewFile()
		f.DeleteSheet("Sheet1")

		// Create an empty Techniques sheet (just headers, no data rows)
		sheetName := "Techniques"
		f.NewSheet(sheetName)
		headers := []interface{}{"ID", "Name", "Description", "Domain", "Platform", "Created", "Modified", "Revoked", "Deprecated"}
		if err := f.SetSheetRow(sheetName, "A1", &headers); err != nil {
			t.Fatalf("failed to set header row: %v", err)
		}

		// Create another empty sheet
		sheetName2 := "Tactics"
		f.NewSheet(sheetName2)
		// Don't even add headers to this one

		if err := f.SaveAs(xlsxPath); err != nil {
			t.Fatalf("failed to save xlsx: %v", err)
		}

		// This should not panic with empty sheets
		if err := store.ImportFromXLSX(xlsxPath, true); err != nil {
			t.Fatalf("ImportFromXLSX returned error: %v", err)
		}

		// Verify metadata was created even with no data
		ctx := context.Background()
		meta, err := store.GetImportMetadata(ctx)
		if err != nil {
			t.Fatalf("GetImportMetadata error: %v", err)
		}
		if meta.TotalRecords != 0 {
			t.Errorf("Expected 0 total records, got %d", meta.TotalRecords)
		}
	})
}

// TestImportFromXLSX_WithMalformedHeaders tests that ImportFromXLSX
// handles Excel files with malformed headers without panicking (TODO-007).
func TestImportFromXLSX_WithMalformedHeaders(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestImportFromXLSX_WithMalformedHeaders", nil, func(t *testing.T, tx *gorm.DB) {
		store, err := NewLocalAttackStore(":memory:")
		if err != nil {
			t.Fatalf("failed to create store: %v", err)
		}

		tmpDir := t.TempDir()
		xlsxPath := filepath.Join(tmpDir, "attack_malformed_headers.xlsx")

		f := excelize.NewFile()
		f.DeleteSheet("Sheet1")

		// Create a Techniques sheet with malformed headers
		sheetName := "Techniques"
		f.NewSheet(sheetName)

		// Set header row with empty strings and special characters
		headers := []interface{}{"", " ", "ID", "\tName\t", "Description", "", "Domain", "Platform", "Created"}
		if err := f.SetSheetRow(sheetName, "A1", &headers); err != nil {
			t.Fatalf("failed to set header row: %v", err)
		}

		// Set data row
		dataRow := []interface{}{"ignored", "ignored", "T9999", "Test Technique", "Desc", "ignored", "enterprise", "Windows", "2020-01-01"}
		if err := f.SetSheetRow(sheetName, "A2", &dataRow); err != nil {
			t.Fatalf("failed to set data row: %v", err)
		}

		if err := f.SaveAs(xlsxPath); err != nil {
			t.Fatalf("failed to save xlsx: %v", err)
		}

		// This should not panic with malformed headers
		if err := store.ImportFromXLSX(xlsxPath, true); err != nil {
			t.Fatalf("ImportFromXLSX returned error: %v", err)
		}

		// Verify that data was still imported correctly by header matching
		ctx := context.Background()
		tech, err := store.GetTechniqueByID(ctx, "T9999")
		if err != nil {
			t.Errorf("Failed to get T9999: %v", err)
		} else {
			// The function should have matched by header name "ID" at index 2
			if tech.Name != "Test Technique" {
				t.Errorf("T9999 has unexpected name: %s (expected 'Test Technique')", tech.Name)
			}
		}
	})
}
