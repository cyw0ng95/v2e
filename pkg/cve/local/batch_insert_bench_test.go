package local

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
)

// BenchmarkBatchInsertThroughput_100 benchmarks inserting 100 CVEs in one batch
func BenchmarkBatchInsertThroughput_100(b *testing.B) {
	dbPath := "/tmp/bench_batch_100.db"
	defer os.Remove(dbPath)

	cves := make([]cve.CVEItem, 100)
	for i := 0; i < 100; i++ {
		cves[i] = cve.CVEItem{
			ID:           fmt.Sprintf("CVE-2021-%05d", i),
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Now()),
			LastModified: cve.NewNVDTime(time.Now()),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Test CVE description",
				},
			},
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.Remove(dbPath)
		db, err := NewDB(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		if err := db.SaveCVEs(cves); err != nil {
			b.Fatal(err)
		}
		db.Close()
	}
}

// BenchmarkBatchInsertThroughput_500 benchmarks inserting 500 CVEs in one batch
func BenchmarkBatchInsertThroughput_500(b *testing.B) {
	dbPath := "/tmp/bench_batch_500.db"
	defer os.Remove(dbPath)

	cves := make([]cve.CVEItem, 500)
	for i := 0; i < 500; i++ {
		cves[i] = cve.CVEItem{
			ID:           fmt.Sprintf("CVE-2021-%05d", i),
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Now()),
			LastModified: cve.NewNVDTime(time.Now()),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Test CVE description",
				},
			},
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.Remove(dbPath)
		db, err := NewDB(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		if err := db.SaveCVEs(cves); err != nil {
			b.Fatal(err)
		}
		db.Close()
	}
}

// BenchmarkBatchInsertThroughput_1000 benchmarks inserting 1000 CVEs in one batch
func BenchmarkBatchInsertThroughput_1000(b *testing.B) {
	dbPath := "/tmp/bench_batch_1000.db"
	defer os.Remove(dbPath)

	cves := make([]cve.CVEItem, 1000)
	for i := 0; i < 1000; i++ {
		cves[i] = cve.CVEItem{
			ID:           fmt.Sprintf("CVE-2021-%05d", i),
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Now()),
			LastModified: cve.NewNVDTime(time.Now()),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Test CVE description",
				},
			},
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.Remove(dbPath)
		db, err := NewDB(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		if err := db.SaveCVEs(cves); err != nil {
			b.Fatal(err)
		}
		db.Close()
	}
}

// BenchmarkBatchInsertThroughput_5000 benchmarks inserting 5000 CVEs in one batch
func BenchmarkBatchInsertThroughput_5000(b *testing.B) {
	dbPath := "/tmp/bench_batch_5000.db"
	defer os.Remove(dbPath)

	cves := make([]cve.CVEItem, 5000)
	for i := 0; i < 5000; i++ {
		cves[i] = cve.CVEItem{
			ID:           fmt.Sprintf("CVE-2021-%05d", i),
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Now()),
			LastModified: cve.NewNVDTime(time.Now()),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Test CVE description",
				},
			},
		}
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.Remove(dbPath)
		db, err := NewDB(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		if err := db.SaveCVEs(cves); err != nil {
			b.Fatal(err)
		}
		db.Close()
	}
}

// BenchmarkBatchInsertSizeComparison benchmarks different batch sizes
func BenchmarkBatchInsertSizeComparison(b *testing.B) {
	batchSizes := []int{10, 50, 100, 200, 500}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("BatchSize_%d", batchSize), func(b *testing.B) {
			dbPath := fmt.Sprintf("/tmp/bench_batch_%d.db", batchSize)
			defer os.Remove(dbPath)

			cves := make([]cve.CVEItem, batchSize)
			for i := 0; i < batchSize; i++ {
				cves[i] = cve.CVEItem{
					ID:           fmt.Sprintf("CVE-2021-%05d", i),
					SourceID:     "nvd@nist.gov",
					Published:    cve.NewNVDTime(time.Now()),
					LastModified: cve.NewNVDTime(time.Now()),
					VulnStatus:   "Analyzed",
					Descriptions: []cve.Description{
						{
							Lang:  "en",
							Value: "Test CVE description",
						},
					},
				}
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				os.Remove(dbPath)
				db, err := NewDB(dbPath)
				if err != nil {
					b.Fatal(err)
				}
				if err := db.SaveCVEs(cves); err != nil {
					b.Fatal(err)
				}
				db.Close()
			}
		})
	}
}

// BenchmarkBatchInsertVsSingle compares batch insert vs single inserts
func BenchmarkBatchInsertVsSingle(b *testing.B) {
	numCVEs := 100

	b.Run("BatchInsert", func(b *testing.B) {
		dbPath := "/tmp/bench_batch_vs_single_batch.db"
		defer os.Remove(dbPath)

		cves := make([]cve.CVEItem, numCVEs)
		for i := 0; i < numCVEs; i++ {
			cves[i] = cve.CVEItem{
				ID:           fmt.Sprintf("CVE-2021-%05d", i),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{
					{
						Lang:  "en",
						Value: "Test CVE description",
					},
				},
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			os.Remove(dbPath)
			db, err := NewDB(dbPath)
			if err != nil {
				b.Fatal(err)
			}
			if err := db.SaveCVEs(cves); err != nil {
				b.Fatal(err)
			}
			db.Close()
		}
	})

	b.Run("SingleInserts", func(b *testing.B) {
		dbPath := "/tmp/bench_batch_vs_single_single.db"
		defer os.Remove(dbPath)

		cves := make([]cve.CVEItem, numCVEs)
		for i := 0; i < numCVEs; i++ {
			cves[i] = cve.CVEItem{
				ID:           fmt.Sprintf("CVE-2021-%05d", i),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{
					{
						Lang:  "en",
						Value: "Test CVE description",
					},
				},
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			os.Remove(dbPath)
			db, err := NewDB(dbPath)
			if err != nil {
				b.Fatal(err)
			}
			for j := range cves {
				if err := db.SaveCVE(&cves[j]); err != nil {
					b.Fatal(err)
				}
			}
			db.Close()
		}
	})
}

// BenchmarkBatchInsertThroughput_ComplexCVEs benchmarks inserting complex CVEs with metrics
func BenchmarkBatchInsertThroughput_ComplexCVEs(b *testing.B) {
	dbPath := "/tmp/bench_batch_complex.db"
	defer os.Remove(dbPath)

	cves := make([]cve.CVEItem, 100)
	for i := 0; i < 100; i++ {
		cves[i] = cve.CVEItem{
			ID:           fmt.Sprintf("CVE-2021-%05d", i),
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Now()),
			LastModified: cve.NewNVDTime(time.Now()),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Test CVE description with more details",
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

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		os.Remove(dbPath)
		db, err := NewDB(dbPath)
		if err != nil {
			b.Fatal(err)
		}
		if err := db.SaveCVEs(cves); err != nil {
			b.Fatal(err)
		}
		db.Close()
	}
}
