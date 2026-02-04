package asvs

import (
	"context"
	"os"
	"testing"
)

func TestNewLocalASVSStore(t *testing.T) {
	// Create a temporary database file
	tmpDB := "/tmp/test_asvs.db"
	defer os.Remove(tmpDB)

	store, err := NewLocalASVSStore(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create ASVS store: %v", err)
	}

	if store == nil {
		t.Fatal("Store is nil")
	}

	// Verify database is accessible
	count, err := store.Count(context.Background())
	if err != nil {
		t.Fatalf("Failed to count records: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 records, got %d", count)
	}
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
		result := parseBoolColumn(tt.input)
		if result != tt.expected {
			t.Errorf("parseBoolColumn(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestListASVSPaginated(t *testing.T) {
	// Create a temporary database file
	tmpDB := "/tmp/test_asvs_list.db"
	defer os.Remove(tmpDB)

	store, err := NewLocalASVSStore(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create ASVS store: %v", err)
	}

	ctx := context.Background()

	// Insert test data
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

	// Test listing all
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

	// Test filtering by chapter
	requirements, total, err = store.ListASVSPaginated(ctx, 0, 10, "V1", 0)
	if err != nil {
		t.Fatalf("Failed to list requirements by chapter: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected total 2 for V1, got %d", total)
	}

	// Test filtering by level
	requirements, total, err = store.ListASVSPaginated(ctx, 0, 10, "", 1)
	if err != nil {
		t.Fatalf("Failed to list requirements by level: %v", err)
	}

	if total != 2 {
		t.Errorf("Expected total 2 for Level 1, got %d", total)
	}
}

func TestGetByID(t *testing.T) {
	// Create a temporary database file
	tmpDB := "/tmp/test_asvs_getbyid.db"
	defer os.Remove(tmpDB)

	store, err := NewLocalASVSStore(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create ASVS store: %v", err)
	}

	ctx := context.Background()

	// Insert test data
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

	// Test GetByID
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

	// Test GetByID with non-existent ID
	_, err = store.GetByID(ctx, "999.999.999")
	if err == nil {
		t.Error("Expected error for non-existent ID, got nil")
	}
}

func TestGetByCWE(t *testing.T) {
	// Create a temporary database file
	tmpDB := "/tmp/test_asvs_getbycwe.db"
	defer os.Remove(tmpDB)

	store, err := NewLocalASVSStore(tmpDB)
	if err != nil {
		t.Fatalf("Failed to create ASVS store: %v", err)
	}

	ctx := context.Background()

	// Insert test data
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

	// Test GetByCWE
	requirements, err := store.GetByCWE(ctx, "CWE-79")
	if err != nil {
		t.Fatalf("Failed to get requirements by CWE: %v", err)
	}

	if len(requirements) != 2 {
		t.Errorf("Expected 2 requirements for CWE-79, got %d", len(requirements))
	}

	// Test GetByCWE with CWE-89
	requirements, err = store.GetByCWE(ctx, "CWE-89")
	if err != nil {
		t.Fatalf("Failed to get requirements by CWE: %v", err)
	}

	if len(requirements) != 2 {
		t.Errorf("Expected 2 requirements for CWE-89, got %d", len(requirements))
	}
}
