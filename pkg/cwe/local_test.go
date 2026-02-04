package cwe

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalCWEStore_ImportAndQuery(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestLocalCWEStore_ImportAndQuery", nil, func(t *testing.T, tx *gorm.DB) {
		tmp := t.TempDir()
		dbPath := filepath.Join(tmp, "cwe_test.db")

		store, err := NewLocalCWEStore(dbPath)
		if err != nil {
			t.Fatalf("NewLocalCWEStore failed: %v", err)
		}

		// Prepare a JSON file with one CWE item that includes nested fields
		jsonPath := filepath.Join(tmp, "cwe.json")
		jsonData := `[
		{
			"ID": "CWE-1",
			"Name": "Test CWE",
			"Description": "A test CWE",
			"RelatedWeaknesses": [
				{"Nature": "example", "CweID": "CWE-2", "ViewID": "v1", "Ordinal": "1"}
			],
			"DemonstrativeExamples": [
				{"ID": "de1", "Entries": [{"IntroText": "intro", "BodyText": "body", "Nature": "nat", "Language": "go", "ExampleCode": "fmt.Println(1)", "Reference": "ref"}]}
			],
			"ObservedExamples": [{"Reference": "oref", "Description": "od", "Link": "link"}],
			"TaxonomyMappings": [{"TaxonomyName": "tax", "EntryName": "entry", "EntryID": "e1", "MappingFit": "fit"}],
			"Notes": [{"Type": "noteType", "Note": "note content"}],
			"ContentHistory": [{"Type": "submission", "SubmissionName": "sub", "Date": "2020-01-01", "Version": "v1"}]
		}
	]
	`
		if err := os.WriteFile(jsonPath, []byte(jsonData), 0644); err != nil {
			t.Fatalf("failed to write json file: %v", err)
		}

		// Import
		if err := store.ImportFromJSON(jsonPath); err != nil {
			t.Fatalf("ImportFromJSON failed: %v", err)
		}

		ctx := context.Background()
		// Query by ID
		item, err := store.GetByID(ctx, "CWE-1")
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if item.ID != "CWE-1" {
			t.Fatalf("unexpected ID: %s", item.ID)
		}
		if item.Name != "Test CWE" {
			t.Fatalf("unexpected Name: %s", item.Name)
		}
		if len(item.RelatedWeaknesses) != 1 {
			t.Fatalf("expected RelatedWeaknesses, got %d", len(item.RelatedWeaknesses))
		}
		if len(item.DemonstrativeExamples) != 1 {
			t.Fatalf("expected DemonstrativeExamples, got %d", len(item.DemonstrativeExamples))
		}
		if len(item.ObservedExamples) != 1 {
			t.Fatalf("expected ObservedExamples, got %d", len(item.ObservedExamples))
		}
		if len(item.TaxonomyMappings) != 1 {
			t.Fatalf("expected TaxonomyMappings, got %d", len(item.TaxonomyMappings))
		}
		if len(item.Notes) != 1 {
			t.Fatalf("expected Notes, got %d", len(item.Notes))
		}
		if len(item.ContentHistory) != 1 {
			t.Fatalf("expected ContentHistory, got %d", len(item.ContentHistory))
		}

		// List paginated
		items, total, err := store.ListCWEsPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("ListCWEsPaginated failed: %v", err)
		}
		if total != 1 {
			t.Fatalf("expected total 1, got %d", total)
		}
		if len(items) != 1 {
			t.Fatalf("expected 1 item, got %d", len(items))
		}

		// Import again - should be a no-op (skip) and not error
		if err := store.ImportFromJSON(jsonPath); err != nil {
			t.Fatalf("second ImportFromJSON failed: %v", err)
		}

		// Missing ID should return error
		if _, err := store.GetByID(ctx, "MISSING"); err == nil {
			t.Fatalf("expected error for missing ID")
		}
	})

}
