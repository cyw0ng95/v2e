package local

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
)

// Helper function to create a test CVE item
func createTestCVEItem(id string) *cve.CVEItem {
	return &cve.CVEItem{
		ID:           id,
		SourceID:     "nvd@nist.gov",
		Published:    cve.NewNVDTime(time.Now()),
		LastModified: cve.NewNVDTime(time.Now()),
		VulnStatus:   "Analyzed",
		Descriptions: []cve.Description{
			{
				Lang:  "en",
				Value: "This is a test CVE description for benchmarking purposes",
			},
		},
		Metrics: &cve.Metrics{
			CvssMetricV31: []cve.CVSSMetricV3{
				{
					Source: "nvd@nist.gov",
					Type:   "Primary",
					CvssData: cve.CVSSDataV3{
						Version:               "3.1",
						VectorString:          "CVSS:3.1/AV:N/AC:L/PR:N/UI:N/S:C/C:H/I:H/A:H",
						BaseScore:             10.0,
						BaseSeverity:          "CRITICAL",
						AttackVector:          "NETWORK",
						AttackComplexity:      "LOW",
						PrivilegesRequired:    "NONE",
						UserInteraction:       "NONE",
						Scope:                 "CHANGED",
						ConfidentialityImpact: "HIGH",
						IntegrityImpact:       "HIGH",
						AvailabilityImpact:    "HIGH",
					},
					ExploitabilityScore: 3.9,
					ImpactScore:         6.0,
				},
			},
		},
	}
}

// BenchmarkNewDB benchmarks creating a new database connection
func BenchmarkNewDB(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		dbPath := "/tmp/bench_newdb_" + string(rune('a'+i%26)) + ".db"
		db, err := NewDB(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		db.Close()
		os.Remove(dbPath)
	}
}

// BenchmarkSaveCVE benchmarks saving a single CVE to the database
func BenchmarkSaveCVE(b *testing.B) {
	dbPath := "/tmp/bench_save.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	cveItem := createTestCVEItem("CVE-2021-44228")

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Use different IDs to avoid uniqueness constraint
		cveItem.ID = fmt.Sprintf("CVE-2021-%05d", i)
		b.StartTimer()

		if err := db.SaveCVE(cveItem); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGetCVE benchmarks retrieving a CVE from the database
func BenchmarkGetCVE(b *testing.B) {
	dbPath := "/tmp/bench_get.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Pre-populate database
	cveItem := createTestCVEItem("CVE-2021-44228")
	if err := db.SaveCVE(cveItem); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.GetCVE("CVE-2021-44228")
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkListCVEs benchmarks listing CVEs with pagination
func BenchmarkListCVEs(b *testing.B) {
	dbPath := "/tmp/bench_list.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Pre-populate database with 100 CVEs
	for i := 0; i < 100; i++ {
		cveID := fmt.Sprintf("CVE-2021-%05d", i)
		cveItem := createTestCVEItem(cveID)
		if err := db.SaveCVE(cveItem); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.ListCVEs(0, 10)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDeleteCVE benchmarks deleting a CVE from the database
func BenchmarkDeleteCVE(b *testing.B) {
	dbPath := "/tmp/bench_delete.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Create CVE to delete
		cveID := fmt.Sprintf("CVE-2021-%05d", i)
		cveItem := createTestCVEItem(cveID)
		if err := db.SaveCVE(cveItem); err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		if err := db.DeleteCVE(cveID); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkUpdateCVE benchmarks updating an existing CVE
func BenchmarkUpdateCVE(b *testing.B) {
	dbPath := "/tmp/bench_update.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Pre-populate database
	cveItem := createTestCVEItem("CVE-2021-44228")
	if err := db.SaveCVE(cveItem); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Modify the CVE
		cveItem.VulnStatus = "Modified"
		cveItem.LastModified = cve.NewNVDTime(time.Now())

		if err := db.SaveCVE(cveItem); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkBulkSaveCVEs benchmarks saving multiple CVEs in bulk
func BenchmarkBulkSaveCVEs(b *testing.B) {
	dbPath := "/tmp/bench_bulk_save.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Create 10 CVEs to save
	cves := make([]*cve.CVEItem, 10)
	for i := 0; i < 10; i++ {
		cveID := "CVE-2021-0000" + string(rune('0'+i))
		cves[i] = createTestCVEItem(cveID)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// Clear database between iterations
		os.Remove(dbPath)
		db, _ = NewDB(dbPath)
		b.StartTimer()

		for _, cveItem := range cves {
			if err := db.SaveCVE(cveItem); err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkCount benchmarks counting CVEs in the database
func BenchmarkCount(b *testing.B) {
	dbPath := "/tmp/bench_count.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		b.Fatal(err)
	}
	defer db.Close()

	// Pre-populate database with 100 CVEs
	for i := 0; i < 100; i++ {
		cveID := fmt.Sprintf("CVE-2021-%05d", i)
		cveItem := createTestCVEItem(cveID)
		if err := db.SaveCVE(cveItem); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := db.Count()
		if err != nil {
			b.Fatal(err)
		}
	}
}
