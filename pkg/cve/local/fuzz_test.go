package local

import (
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
)

// FuzzSaveCVE tests saving CVE with arbitrary fields
func FuzzSaveCVE(f *testing.F) {
	// Seed corpus with typical CVE IDs
	f.Add("CVE-2021-44228", "Log4j vulnerability", 9.8)
	f.Add("CVE-2021-45046", "Log4j bypass", 9.0)
	f.Add("", "", 0.0)
	f.Add("CVE-2024-99999", "Test CVE", 10.0)

	// Fuzz test
	f.Fuzz(func(t *testing.T, cveID, description string, score float64) {
		// Create in-memory database
		db, err := NewDB(":memory:")
		if err != nil {
			return
		}
		defer db.Close()

		// Create CVE with fuzzed data
		testCVE := &cve.CVEItem{
			ID:           cveID,
			SourceID:     "fuzz-test",
			Published:    cve.NewNVDTime(time.Now()),
			LastModified: cve.NewNVDTime(time.Now()),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: description,
				},
			},
			Metrics: &cve.Metrics{
				CvssMetricV31: []cve.CVSSMetricV3{
					{
						Type:   "Primary",
						Source: "nvd@nist.gov",
						CvssData: cve.CVSSDataV3{
							BaseScore: score,
						},
					},
				},
			},
		}

		// Save CVE - should not panic
		_ = db.SaveCVE(testCVE)

		// Try to retrieve it - should not panic
		if cveID != "" {
			_, _ = db.GetCVE(cveID)
		}
	})
}

// FuzzListCVEs tests pagination with arbitrary limits and offsets
func FuzzListCVEs(f *testing.F) {
	// Seed corpus
	f.Add(0, 10)
	f.Add(10, 20)
	f.Add(-1, 100)
	f.Add(0, -5)
	f.Add(1000000, 999999)

	// Fuzz test
	f.Fuzz(func(t *testing.T, limit, offset int) {
		// Create in-memory database
		db, err := NewDB(":memory:")
		if err != nil {
			return
		}
		defer db.Close()

		// List CVEs with fuzzed pagination - should not panic
		_, _ = db.ListCVEs(limit, offset)
	})
}
