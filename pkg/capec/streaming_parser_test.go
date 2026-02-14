//go:build CONFIG_USE_LIBXML2

package capec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// TestDefaultStreamingBatchConfig tests default configuration
func TestDefaultStreamingBatchConfig(t *testing.T) {
	config := DefaultStreamingBatchConfig()
	assert.Equal(t, 100, config.BatchSize, "Default batch size should be 100")
}

// TestStreamingBatchConfig_DefaultHandling tests zero batch size handling
func TestStreamingBatchConfig_DefaultHandling(t *testing.T) {
	config := StreamingBatchConfig{BatchSize: 0}
	parser := NewStreamingCAPECParser(nil, config)
	assert.Equal(t, 100, parser.config.BatchSize, "Zero batch size should use default")
}

func TestStreamingCAPECParser_BasicImport(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStreamingCAPECParser_BasicImport", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_streaming_test.db")

		store, err := NewLocalCAPECStore(dbPath)
		require.NoError(t, err)

		// Create test XML with multiple Attack_Patterns
		xmlContent := `<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v-test-streaming" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>
    <Attack_Pattern ID="1" Name="Test Pattern 1" Abstraction="Detailed" Status="Stable">
      <Description>Test description 1</Description>
      <Summary>Test summary 1</Summary>
      <Likelihood_Of_Attack>High</Likelihood_Of_Attack>
      <Typical_Severity>High</Typical_Severity>
      <Related_Weaknesses>
        <Related_Weakness CWE_ID="CWE-79"/>
        <Related_Weakness CWE_ID="CWE-89"/>
      </Related_Weaknesses>
      <Example_Instances>
        <Example>Example 1 text</Example>
      </Example_Instances>
      <Mitigations>
        <Mitigation>Mitigation 1</Mitigation>
      </Mitigations>
      <References>
        <Reference External_Reference_ID="REF-1"/>
      </References>
    </Attack_Pattern>
    <Attack_Pattern ID="2" Name="Test Pattern 2" Abstraction="Detailed" Status="Stable">
      <Description>Test description 2</Description>
      <Summary>Test summary 2</Summary>
      <Likelihood_Of_Attack>Medium</Likelihood_Of_Attack>
      <Typical_Severity>Medium</Typical_Severity>
      <Related_Weaknesses>
        <Related_Weakness CWE_ID="CWE-20"/>
      </Related_Weaknesses>
      <Example_Instances>
        <Example>Example 2 text</Example>
        <Example>Example 2 text 2</Example>
      </Example_Instances>
      <Mitigations>
        <Mitigation>Mitigation 2</Mitigation>
      </Mitigations>
      <References>
        <Reference External_Reference_ID="REF-2"/>
      </References>
    </Attack_Pattern>
  </Attack_Patterns>
</Attack_Pattern_Catalog>`

		xmlPath := filepath.Join(dir, "test_capec_streaming.xml")
		require.NoError(t, os.WriteFile(xmlPath, []byte(xmlContent), 0644))

		// Import using streaming parser
		imported, err := ImportWithStreamingParser(store, xmlPath, false)
		require.NoError(t, err)
		assert.True(t, imported)

		// Verify items were imported
		item1, err := store.GetByID(nil, "CAPEC-1")
		require.NoError(t, err)
		assert.Equal(t, 1, item1.CAPECID)
		assert.Equal(t, "Test Pattern 1", item1.Name)
		assert.Equal(t, "Test summary 1", item1.Summary)
		assert.Equal(t, "High", item1.Likelihood)
		assert.Equal(t, "High", item1.TypicalSeverity)

		item2, err := store.GetByID(nil, "CAPEC-2")
		require.NoError(t, err)
		assert.Equal(t, 2, item2.CAPECID)
		assert.Equal(t, "Test Pattern 2", item2.Name)

		// Verify related weaknesses
		weaknesses1, err := store.GetRelatedWeaknesses(nil, 1)
		require.NoError(t, err)
		assert.Len(t, weaknesses1, 2)

		// Verify examples
		examples2, err := store.GetExamples(nil, 2)
		require.NoError(t, err)
		assert.Len(t, examples2, 2)

		// Verify mitigations
		mitigations1, err := store.GetMitigations(nil, 1)
		require.NoError(t, err)
		assert.Len(t, mitigations1, 1)

		// Verify references
		refs1, err := store.GetReferences(nil, 1)
		require.NoError(t, err)
		assert.Len(t, refs1, 1)
		assert.Equal(t, "REF-1", refs1[0].ExternalReference)

		// Verify catalog meta
		meta, err := store.GetCatalogMeta(nil)
		require.NoError(t, err)
		assert.Equal(t, "v-test-streaming", meta.Version)
	})
}

func TestStreamingCAPECParser_LargeFile(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStreamingCAPECParser_LargeFile", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_large_test.db")

		store, err := NewLocalCAPECStore(dbPath)
		require.NoError(t, err)

		// Generate a large XML file with many patterns
		var xmlBuilder strings.Builder
		xmlBuilder.WriteString(`<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v-test-large" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>`)

		// Generate 250 patterns (enough to test batching with default batch size 100)
		for i := 1; i <= 250; i++ {
			xmlBuilder.WriteString(fmt.Sprintf(`
    <Attack_Pattern ID="%d" Name="Test Pattern %d" Abstraction="Detailed" Status="Stable">
      <Description>Test description %d</Description>
      <Summary>Test summary %d</Summary>
      <Likelihood_Of_Attack>High</Likelihood_Of_Attack>
      <Typical_Severity>Medium</Typical_Severity>
      <Related_Weaknesses>
        <Related_Weakness CWE_ID="CWE-%d"/>
      </Related_Weaknesses>
      <Example_Instances>
        <Example>Example %d</Example>
      </Example_Instances>
      <Mitigations>
        <Mitigation>Mitigation %d</Mitigation>
      </Mitigations>
      <References>
        <Reference External_Reference_ID="REF-%d"/>
      </References>
    </Attack_Pattern>`, i, i, i, i, i, i, i, i))
		}

		xmlBuilder.WriteString(`
  </Attack_Patterns>
</Attack_Pattern_Catalog>`)

		xmlPath := filepath.Join(dir, "test_capec_large.xml")
		require.NoError(t, os.WriteFile(xmlPath, []byte(xmlBuilder.String()), 0644))

		// Import using streaming parser
		start := time.Now()
		imported, err := ImportWithStreamingParser(store, xmlPath, false)
		require.NoError(t, err)
		assert.True(t, imported)
		t.Logf("Large file import took: %v", time.Since(start))

		// Verify all items were imported
		items, total, err := store.ListCAPECsPaginated(nil, 0, 1000)
		require.NoError(t, err)
		assert.Equal(t, int64(250), total)
		assert.Len(t, items, 250)

		// Verify a sample item
		item100, err := store.GetByID(nil, "CAPEC-100")
		require.NoError(t, err)
		assert.Equal(t, 100, item100.CAPECID)
	})
}

func TestStreamingCAPECParser_BatchSize(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStreamingCAPECParser_BatchSize", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_batch_test.db")

		store, err := NewLocalCAPECStore(dbPath)
		require.NoError(t, err)

		// Create test XML with 15 patterns, use batch size of 5
		var xmlBuilder strings.Builder
		xmlBuilder.WriteString(`<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v-test-batch" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>`)

		for i := 1; i <= 15; i++ {
			xmlBuilder.WriteString(fmt.Sprintf(`
    <Attack_Pattern ID="%d" Name="Test Pattern %d">
      <Description>Test description %d</Description>
    </Attack_Pattern>`, i, i, i))
		}

		xmlBuilder.WriteString(`
  </Attack_Patterns>
</Attack_Pattern_Catalog>`)

		xmlPath := filepath.Join(dir, "test_capec_batch.xml")
		require.NoError(t, os.WriteFile(xmlPath, []byte(xmlBuilder.String()), 0644))

		// Import with custom batch size
		config := StreamingBatchConfig{BatchSize: 5}
		imported, err := ImportWithStreamingParserAndConfig(store, xmlPath, false, config)
		require.NoError(t, err)
		assert.True(t, imported)

		// Verify all items
		items, total, err := store.ListCAPECsPaginated(nil, 0, 100)
		require.NoError(t, err)
		assert.Equal(t, int64(15), total)
		assert.Len(t, items, 15)
	})
}

func TestStreamingCAPECParser_ForceImport(t *testing.T) {
	testutils.Run(t, testutils.Level2, "StreamingCAPECParser_ForceImport", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_force_test.db")
		xmlPath := filepath.Join(dir, "test_capec_force.xml")

		// Create initial XML
		xmlContent := `<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v1.0" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>
    <Attack_Pattern ID="1" Name="Original">
      <Description>Original description</Description>
    </Attack_Pattern>
  </Attack_Patterns>
</Attack_Pattern_Catalog>`
		require.NoError(t, os.WriteFile(xmlPath, []byte(xmlContent), 0644))

		store, err := NewLocalCAPECStore(dbPath)
		require.NoError(t, err)

		// First import
		imported, err := ImportWithStreamingParser(store, xmlPath, false)
		require.NoError(t, err)
		assert.True(t, imported)

		meta1, err := store.GetCatalogMeta(nil)
		require.NoError(t, err)
		firstImportTime := meta1.ImportedAtUTC

		// Second import without force - should skip
		time.Sleep(time.Second)
		imported, err = ImportWithStreamingParser(store, xmlPath, false)
		require.NoError(t, err)
		assert.False(t, imported, "Should skip import when version unchanged")

		meta2, err := store.GetCatalogMeta(nil)
		require.NoError(t, err)
		assert.Equal(t, firstImportTime, meta2.ImportedAtUTC, "Import time should not change")

		// Force import
		imported, err = ImportWithStreamingParser(store, xmlPath, true)
		require.NoError(t, err)
		assert.True(t, imported)

		meta3, err := store.GetCatalogMeta(nil)
		require.NoError(t, err)
		assert.Greater(t, meta3.ImportedAtUTC, firstImportTime, "Import time should update on force")
	})
}

func TestStreamingCAPECParser_SummaryFallback(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStreamingCAPECParser_SummaryFallback", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_summary_test.db")

		store, err := NewLocalCAPECStore(dbPath)
		require.NoError(t, err)

		// Test with empty summary (should use truncated description)
		xmlContent := `<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v-test-summary" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>
    <Attack_Pattern ID="1" Name="Test">
      <Description>This is a very long description that should be truncated to 200 characters when the summary field is empty. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco.</Description>
    </Attack_Pattern>
  </Attack_Patterns>
</Attack_Pattern_Catalog>`

		xmlPath := filepath.Join(dir, "test_capec_summary.xml")
		require.NoError(t, os.WriteFile(xmlPath, []byte(xmlContent), 0644))

		imported, err := ImportWithStreamingParser(store, xmlPath, false)
		require.NoError(t, err)
		assert.True(t, imported)

		item, err := store.GetByID(nil, "CAPEC-1")
		require.NoError(t, err)
		assert.LessOrEqual(t, len(item.Summary), 200, "Summary should be truncated to 200 chars")
		assert.Equal(t, item.Summary, truncateString(strings.TrimSpace(item.Description), 200))
	})
}

func TestStreamingCAPECParser_Deduplication(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStreamingCAPECParser_Deduplication", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_dedup_test.db")

		store, err := NewLocalCAPECStore(dbPath)
		require.NoError(t, err)

		// Test XML with duplicate weaknesses and references
		xmlContent := `<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v-test-dedup" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>
    <Attack_Pattern ID="1" Name="Test Dedup" Abstraction="Detailed" Status="Stable">
      <Description>Test</Description>
      <Related_Weaknesses>
        <Related_Weakness CWE_ID="CWE-79"/>
        <Related_Weakness CWE_ID="CWE-79"/>
        <Related_Weakness CWE_ID="CWE-89"/>
        <Related_Weakness CWE_ID="CWE-89"/>
      </Related_Weaknesses>
      <References>
        <Reference External_Reference_ID="REF-1"/>
        <Reference External_Reference_ID="REF-1"/>
        <Reference External_Reference_ID="REF-2"/>
      </References>
    </Attack_Pattern>
  </Attack_Patterns>
</Attack_Pattern_Catalog>`

		xmlPath := filepath.Join(dir, "test_capec_dedup.xml")
		require.NoError(t, os.WriteFile(xmlPath, []byte(xmlContent), 0644))

		imported, err := ImportWithStreamingParser(store, xmlPath, false)
		require.NoError(t, err)
		assert.True(t, imported)

		// Check deduplicated weaknesses
		weaknesses, err := store.GetRelatedWeaknesses(nil, 1)
		require.NoError(t, err)
		assert.Len(t, weaknesses, 2, "Should have 2 unique weaknesses after deduplication")

		// Check deduplicated references
		refs, err := store.GetReferences(nil, 1)
		require.NoError(t, err)
		assert.Len(t, refs, 2, "Should have 2 unique references after deduplication")
	})
}

func TestStreamingCAPECParser_TransactionRollback(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestStreamingCAPECParser_TransactionRollback", nil, func(t *testing.T, tx *gorm.DB) {
		dir := t.TempDir()
		dbPath := filepath.Join(dir, "capec_rollback_test.db")

		store, err := NewLocalCAPECStore(dbPath)
		require.NoError(t, err)

		// First, import valid data
		validXML := `<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v1.0" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>
    <Attack_Pattern ID="1" Name="Valid">
      <Description>Valid description</Description>
    </Attack_Pattern>
  </Attack_Patterns>
</Attack_Pattern_Catalog>`

		validPath := filepath.Join(dir, "valid_capec.xml")
		require.NoError(t, os.WriteFile(validPath, []byte(validXML), 0644))

		imported, err := ImportWithStreamingParser(store, validPath, false)
		require.NoError(t, err)
		assert.True(t, imported)

		// Verify item exists
		item, err := store.GetByID(nil, "CAPEC-1")
		require.NoError(t, err)
		assert.Equal(t, "Valid", item.Name)

		// Now try to import malformed XML (should not affect existing data)
		malformedPath := filepath.Join(dir, "malformed_capec.xml")
		require.NoError(t, os.WriteFile(malformedPath, []byte("not valid xml"), 0644))

		imported, err = ImportWithStreamingParser(store, malformedPath, false)
		assert.Error(t, err, "Malformed XML should return error")
		assert.False(t, imported)

		// Verify original data is intact
		item, err = store.GetByID(nil, "CAPEC-1")
		require.NoError(t, err)
		assert.Equal(t, "Valid", item.Name, "Original data should be intact after failed import")
	})
}
