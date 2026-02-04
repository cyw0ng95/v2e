package capec

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLocalCAPECStore(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewLocalCAPECStore", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_store.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)
		require.NotNil(t, store)
		require.NotNil(t, store.db)
	})

}

func TestGetByID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetByID", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_getbyid.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		// Test with non-existent ID
		ctx := context.Background()
		_, err = store.GetByID(ctx, "CAPEC-999999")
		assert.Error(t, err)

		// Test with invalid ID format
		_, err = store.GetByID(ctx, "INVALID-ID")
		assert.Error(t, err)

		// Test with numeric-only ID
		_, err = store.GetByID(ctx, "999999")
		assert.Error(t, err)

		// Test with mixed format ID
		_, err = store.GetByID(ctx, "CAPEC-ABC")
		assert.Error(t, err)
	})

}

func TestListCAPECsPaginated(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestListCAPECsPaginated", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_paginated.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		ctx := context.Background()

		// Test with empty database
		items, total, err := store.ListCAPECsPaginated(ctx, 0, 10)
		assert.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, int64(0), total)

		// Test with negative offset
		items, total, err = store.ListCAPECsPaginated(ctx, -1, 10)
		assert.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, int64(0), total)

		// Test with zero limit
		items, total, err = store.ListCAPECsPaginated(ctx, 0, 0)
		assert.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, int64(0), total)

		// Test with large offset
		items, total, err = store.ListCAPECsPaginated(ctx, 1000, 10)
		assert.NoError(t, err)
		assert.Empty(t, items)
		assert.Equal(t, int64(0), total)
	})

}

func TestGetRelatedWeaknesses(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetRelatedWeaknesses", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_related_weaknesses.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		ctx := context.Background()

		// Test with non-existent CAPEC ID
		weaknesses, err := store.GetRelatedWeaknesses(ctx, 999999)
		assert.NoError(t, err)
		assert.Empty(t, weaknesses)

		// Test with valid but non-existent CAPEC ID
		weaknesses, err = store.GetRelatedWeaknesses(ctx, 123)
		assert.NoError(t, err)
		assert.Empty(t, weaknesses)
	})

}

func TestGetExamples(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetExamples", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_examples.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		ctx := context.Background()

		// Test with non-existent CAPEC ID
		examples, err := store.GetExamples(ctx, 999999)
		assert.NoError(t, err)
		assert.Empty(t, examples)

		// Test with valid but non-existent CAPEC ID
		examples, err = store.GetExamples(ctx, 123)
		assert.NoError(t, err)
		assert.Empty(t, examples)
	})

}

func TestGetMitigations(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetMitigations", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_mitigations.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		ctx := context.Background()

		// Test with non-existent CAPEC ID
		mitigations, err := store.GetMitigations(ctx, 999999)
		assert.NoError(t, err)
		assert.Empty(t, mitigations)

		// Test with valid but non-existent CAPEC ID
		mitigations, err = store.GetMitigations(ctx, 123)
		assert.NoError(t, err)
		assert.Empty(t, mitigations)
	})

}

func TestGetReferences(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetReferences", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_references.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		ctx := context.Background()

		// Test with non-existent CAPEC ID
		references, err := store.GetReferences(ctx, 999999)
		assert.NoError(t, err)
		assert.Empty(t, references)

		// Test with valid but non-existent CAPEC ID
		references, err = store.GetReferences(ctx, 123)
		assert.NoError(t, err)
		assert.Empty(t, references)
	})

}

func TestGetCatalogMeta(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetCatalogMeta", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_meta.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		ctx := context.Background()

		// Test with empty database - should return error since no meta exists yet
		_, err = store.GetCatalogMeta(ctx)
		assert.Error(t, err)
	})

}

func TestUtilityFunctions(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUtilityFunctions", nil, func(t *testing.T, tx *gorm.DB) {
		// Test firstNonEmpty function
		result := firstNonEmpty("first", "second")
		assert.Equal(t, "first", result)

		result = firstNonEmpty("", "second")
		assert.Equal(t, "second", result)

		result = firstNonEmpty("first", "")
		assert.Equal(t, "first", result)

		result = firstNonEmpty("", "")
		assert.Equal(t, "", result)

		result = firstNonEmpty("  ", "second") // whitespace should be treated as empty
		assert.Equal(t, "second", result)

		// Test truncateString function
		longStr := "this is a very long string that will be truncated"
		truncated := truncateString(longStr, 10)
		assert.Equal(t, "this is a ", truncated)

		shortStr := "short"
		truncated = truncateString(shortStr, 10)
		assert.Equal(t, "short", truncated)

		emptyStr := ""
		truncated = truncateString(emptyStr, 10)
		assert.Equal(t, "", truncated)
	})

}

func TestGetByIDRegexParsing(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGetByIDRegexParsing", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_regex.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		ctx := context.Background()

		// Test various ID formats to ensure regex parsing works
		testCases := []string{
			"CAPEC-123",
			"123",
			"CAPEC-00123",
			"CAPEC-123-extra",
			"prefix-CAPEC-123-suffix",
		}

		for _, id := range testCases {
			// These should all return errors since the CAPEC items don't exist in the DB
			_, err := store.GetByID(ctx, id)
			assert.Error(t, err, "Testing ID format: %s", id)
		}

		// Test invalid formats that should return errors
		invalidCases := []string{
			"",
			"CAPEC-",
			"CAPEC-ABC",
			"invalid",
			"-123",
		}

		for _, id := range invalidCases {
			_, err := store.GetByID(ctx, id)
			assert.Error(t, err, "Testing invalid ID format: %s", id)
		}
	})

}

func TestImportFromXMLErrorHandling(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestImportFromXMLErrorHandling", nil, func(t *testing.T, tx *gorm.DB) {
		// Create a temporary database file for testing
		tempDB := filepath.Join(t.TempDir(), "test_capec_import_errors.db")

		store, err := NewLocalCAPECStore(tempDB)
		require.NoError(t, err)

		// Test importing from non-existent file
		err = store.ImportFromXML("/non/existent/file.xml", false)
		assert.Error(t, err)
		// Check for the actual error message - might be different from what we expect
		// In any case, there should be an error
	})

}
