package local

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/capec"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cwe"
)

// BenchmarkGetCWEOpt benchmarks retrieving a CWE from the database
func BenchmarkGetCWEOpt(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "cwe_bench.db")

	store, err := cwe.NewLocalCWEStore(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		os.Remove(dbPath)
	}()

	// Pre-populate database with a CWE item
	jsonData := `[
	{
		"ID": "CWE-1",
		"Name": "Test CWE",
		"Description": "A test CWE for benchmarking",
		"RelatedWeaknesses": [
			{"Nature": "example", "CweID": "CWE-2", "ViewID": "v1", "Ordinal": "1"}
		]
	}
]`
	jsonPath := filepath.Join(b.TempDir(), "cwe_bench.json")
	if err := os.WriteFile(jsonPath, []byte(jsonData), 0644); err != nil {
		b.Fatal(err)
	}

	if err := store.ImportFromJSON(jsonPath); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.GetByID(ctx, "CWE-1")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListCWEsOpt benchmarks listing CWEs from the database
func BenchmarkListCWEsOpt(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "cwe_list_bench.db")

	store, err := cwe.NewLocalCWEStore(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		os.Remove(dbPath)
	}()

	// Pre-populate database with 100 CWE items
	var jsonData string
	jsonData = "[\n"
	for i := 0; i < 100; i++ {
		cweID := fmt.Sprintf("CWE-%d", i+1)
		if i > 0 {
			jsonData += ",\n"
		}
		jsonData += fmt.Sprintf(`{
			"ID": "%s",
			"Name": "Test CWE %d",
			"Description": "A test CWE for benchmarking %d"
		}`, cweID, i+1, i+1)
	}
	jsonData += "\n]"

	jsonPath := filepath.Join(b.TempDir(), "cwe_list_bench.json")
	if err := os.WriteFile(jsonPath, []byte(jsonData), 0644); err != nil {
		b.Fatal(err)
	}

	if err := store.ImportFromJSON(jsonPath); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := store.ListCWEsPaginated(ctx, 0, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetCAPECOpt benchmarks retrieving a CAPEC from the database
func BenchmarkGetCAPECOpt(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "capec_bench.db")

	store, err := capec.NewLocalCAPECStore(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		os.Remove(dbPath)
	}()

	// Create a minimal CAPEC XML file
	xmlContent := `<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v-test" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>
    <Attack_Pattern ID="1" Name="Test Pattern" Abstraction="Detailed" Status="Stable">
      <Description>A test CAPEC pattern for benchmarking</Description>
      <Related_Weaknesses>
        <Related_Weakness CWE_ID="123"/>
      </Related_Weaknesses>
    </Attack_Pattern>
  </Attack_Patterns>
</Attack_Pattern_Catalog>`

	xmlPath := filepath.Join(b.TempDir(), "capec_bench.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		b.Fatal(err)
	}

	if err := store.ImportFromXML(xmlPath, false); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := store.GetByID(ctx, "CAPEC-1") // Changed to string ID
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListCAPECsOpt benchmarks listing CAPECs from the database
func BenchmarkListCAPECsOpt(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "capec_list_bench.db")

	store, err := capec.NewLocalCAPECStore(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer func() {
		os.Remove(dbPath)
	}()

	// Create XML with multiple CAPEC items for benchmarking
	var xmlContent string
	xmlContent = `<?xml version="1.0"?>
<Attack_Pattern_Catalog Version="v-test" xmlns="http://capec.mitre.org/capec-3">
  <Attack_Patterns>`

	for i := 1; i <= 100; i++ {
		xmlContent += fmt.Sprintf(`
    <Attack_Pattern ID="%d" Name="Test Pattern %d" Abstraction="Detailed" Status="Stable">
      <Description>A test CAPEC pattern for benchmarking %d</Description>
    </Attack_Pattern>`, i, i, i)
	}

	xmlContent += `
  </Attack_Patterns>
</Attack_Pattern_Catalog>`

	xmlPath := filepath.Join(b.TempDir(), "capec_list_bench.xml")
	if err := os.WriteFile(xmlPath, []byte(xmlContent), 0644); err != nil {
		b.Fatal(err)
	}

	if err := store.ImportFromXML(xmlPath, false); err != nil {
		b.Fatal(err)
	}

	ctx := context.Background()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := store.ListCAPECsPaginated(ctx, 0, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkConcurrentReadsOpt benchmarks concurrent reads of CVEs
func BenchmarkConcurrentReadsOpt(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "concurrent_bench.db")

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Pre-populate database with 1000 CVEs
	for i := 0; i < 1000; i++ {
		cveID := fmt.Sprintf("CVE-2021-%05d", i)
		cveItem := createTestCVEItem(cveID)
		if err := db.SaveCVE(cveItem); err != nil {
			b.Fatal(err)
		}
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			id := fmt.Sprintf("CVE-2021-%05d", i%1000)
			_, err := db.GetCVE(id)
			if err != nil {
				b.Fatal(err)
			}
			i++
		}
	})
}

// BenchmarkCVERelationshipsOpt benchmarks retrieving CVEs with relationships (simulating real-world usage)
func BenchmarkCVERelationshipsOpt(b *testing.B) {
	dbPath := filepath.Join(b.TempDir(), "relationship_bench.db")

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Pre-populate database with CVEs that have complex relationships
	cveItem := createTestCVEItem("CVE-2021-44228")
	cveItem.Weaknesses = []cve.Weakness{
		{
			Source: "nvd@nist.gov",
			Type:   "Primary",
			Description: []cve.Description{
				{
					Lang:  "en",
					Value: "CWE-79: Improper Neutralization of Input During Web Page Generation ('Cross-site Scripting')",
				},
			},
		},
	}

	if err := db.SaveCVE(cveItem); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		item, err := db.GetCVE("CVE-2021-44228")
		if err != nil {
			b.Fatal(err)
		}

		// Simulate processing the retrieved CVE (e.g., extracting related CWEs)
		if item.Weaknesses != nil && len(item.Weaknesses) > 0 {
			_ = item.Weaknesses[0].Description[0].Value
		}
	}
}
