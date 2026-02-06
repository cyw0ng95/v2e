package local

import (
	"os"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestDataVolume_SmallBatch(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataVolume_SmallBatch", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_volume_small.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		numCVEs := 10
		cves := make([]cve.CVEItem, numCVEs)
		for i := 0; i < numCVEs; i++ {
			cves[i] = cve.CVEItem{
				ID:           "CVE-SMALL-" + string(rune('A'+i)),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Small batch test CVE"}},
			}
		}

		start := time.Now()
		err = db.SaveCVEs(cves)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		t.Logf("Saved %d CVEs in %v", numCVEs, elapsed)

		count, err := db.Count()
		if err != nil {
			t.Fatalf("Failed to count CVEs: %v", err)
		}
		if int(count) != numCVEs {
			t.Errorf("Expected %d CVEs, got %d", numCVEs, count)
		}
	})
}

func TestDataVolume_MediumBatch(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataVolume_MediumBatch", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_volume_medium.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		numCVEs := 100
		cves := make([]cve.CVEItem, numCVEs)
		for i := 0; i < numCVEs; i++ {
			cves[i] = cve.CVEItem{
				ID:           "CVE-MEDIUM-" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Medium batch test CVE with more data"}},
				Metrics: &cve.Metrics{
					CvssMetricV31: []cve.CVSSMetricV3{
						{
							CvssData: cve.CVSSDataV3{
								BaseScore: float64(5 + i%5),
							},
						},
					},
				},
			}
		}

		start := time.Now()
		err = db.SaveCVEs(cves)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		t.Logf("Saved %d CVEs in %v", numCVEs, elapsed)

		count, err := db.Count()
		if err != nil {
			t.Fatalf("Failed to count CVEs: %v", err)
		}
		if int(count) != numCVEs {
			t.Errorf("Expected %d CVEs, got %d", numCVEs, count)
		}
	})
}

func TestDataVolume_LargeBatch(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataVolume_LargeBatch", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_volume_large.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		numCVEs := 500
		cves := make([]cve.CVEItem, numCVEs)
		for i := 0; i < numCVEs; i++ {
			cves[i] = cve.CVEItem{
				ID:           "CVE-LARGE-" + string(rune('A'+i%26)) + string(rune('0'+i/26%26)) + string(rune('0'+i/676)),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{
					{Lang: "en", Value: "Large batch test CVE with comprehensive data for testing performance"},
					{Lang: "es", Value: "Vulnerabilidad de prueba de lote grande"},
				},
				References: []cve.Reference{
					{URL: "https://example.com/ref1", Source: "vendor"},
					{URL: "https://example.com/ref2", Source: "vendor"},
				},
			}
		}

		start := time.Now()
		err = db.SaveCVEs(cves)
		elapsed := time.Since(start)

		if err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		t.Logf("Saved %d CVEs in %v (%.2f CVEs/sec)",
			numCVEs, elapsed, float64(numCVEs)/elapsed.Seconds())

		count, err := db.Count()
		if err != nil {
			t.Fatalf("Failed to count CVEs: %v", err)
		}
		if int(count) != numCVEs {
			t.Errorf("Expected %d CVEs, got %d", numCVEs, count)
		}
	})
}

func TestDataVolume_VariousCVETypes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataVolume_VariousCVETypes", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_volume_types.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		cves := []cve.CVEItem{
			{
				ID:           "CVE-WITH-METRICS-V31",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "CVE with CVSS v3.1 metrics"}},
				Metrics: &cve.Metrics{
					CvssMetricV31: []cve.CVSSMetricV3{
						{
							Source:              "nvd@nist.gov",
							Type:                "Primary",
							CvssData:            cve.CVSSDataV3{BaseScore: 9.8},
							ExploitabilityScore: 3.9,
							ImpactScore:         6.0,
						},
					},
				},
			},
			{
				ID:           "CVE-WITH-METRICS-V2",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "CVE with CVSS v2 metrics"}},
				Metrics: &cve.Metrics{
					CvssMetricV2: []cve.CVSSMetricV2{
						{
							CvssData: cve.CVSSDataV2{BaseScore: 6.5},
						},
					},
				},
			},
			{
				ID:           "CVE-WITHOUT-METRICS",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "CVE without metrics"}},
			},
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		retrieved, err := db.GetCVE("CVE-WITH-METRICS-V31")
		if err != nil {
			t.Fatalf("Failed to retrieve CVE: %v", err)
		}
		if retrieved.Metrics == nil {
			t.Error("Expected metrics to be preserved")
		}

		retrieved2, err := db.GetCVE("CVE-WITHOUT-METRICS")
		if err != nil {
			t.Fatalf("Failed to retrieve CVE without metrics: %v", err)
		}
		if retrieved2.Metrics != nil {
			t.Error("Expected metrics to be nil")
		}
	})
}

func TestDataVolume_BatchInsertPerformance(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataVolume_BatchInsertPerformance", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_volume_perf.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		batchSizes := []int{1, 10, 50, 100}
		for _, batchSize := range batchSizes {
			cves := make([]cve.CVEItem, batchSize)
			for i := 0; i < batchSize; i++ {
				cves[i] = cve.CVEItem{
					ID:           "CVE-PERF-" + string(rune('A'+i%26)) + string(rune('0'+i/26)),
					SourceID:     "nvd@nist.gov",
					Published:    cve.NewNVDTime(time.Now()),
					LastModified: cve.NewNVDTime(time.Now()),
					VulnStatus:   "Analyzed",
					Descriptions: []cve.Description{{Lang: "en", Value: "Performance test CVE"}},
				}
			}

			start := time.Now()
			err = db.SaveCVEs(cves)
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("Failed to save batch of %d CVEs: %v", batchSize, err)
			}

			t.Logf("Batch size %d: saved in %v (%.2f ops/sec)",
				batchSize, elapsed, float64(batchSize)/elapsed.Seconds())
		}
	})
}

func TestDataVolume_MixedDataTypes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestDataVolume_MixedDataTypes", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_volume_mixed.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		cves := []cve.CVEItem{
			{
				ID:           "CVE-MINIMAL",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Minimal CVE"}},
			},
			{
				ID:           "CVE-WITH-REFERENCES",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "CVE with references"}},
				References: []cve.Reference{
					{URL: "https://nvd.nist.gov/vuln/detail/CVE-2021-44228", Source: "nist"},
					{URL: "https://logging.apache.org/log4j/2.x/security.html", Source: "vendor"},
					{URL: "https://www.cisa.gov/known-exploited-vulnerabilities-catalog", Source: "cisa"},
				},
			},
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save mixed CVEs: %v", err)
		}

		count, err := db.Count()
		if err != nil {
			t.Fatalf("Failed to count CVEs: %v", err)
		}
		if count != 2 {
			t.Errorf("Expected 2 CVEs, got %d", count)
		}

		t.Log("Mixed data types test completed successfully")
	})
}
