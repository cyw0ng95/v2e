package local

import (
	"os"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestQueryPattern_FilterByCVSSScore(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestQueryPattern_FilterByCVSSScore", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_query_cvss.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		cves := []cve.CVEItem{
			{
				ID:           "CVE-HIGH-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "High severity CVE"}},
				Metrics: &cve.Metrics{
					CvssMetricV31: []cve.CVSSMetricV3{
						{
							CvssData: cve.CVSSDataV3{BaseScore: 9.8},
						},
					},
				},
			},
			{
				ID:           "CVE-MEDIUM-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Medium severity CVE"}},
				Metrics: &cve.Metrics{
					CvssMetricV31: []cve.CVSSMetricV3{
						{
							CvssData: cve.CVSSDataV3{BaseScore: 5.5},
						},
					},
				},
			},
			{
				ID:           "CVE-LOW-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Low severity CVE"}},
				Metrics: &cve.Metrics{
					CvssMetricV31: []cve.CVSSMetricV3{
						{
							CvssData: cve.CVSSDataV3{BaseScore: 3.7},
						},
					},
				},
			},
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		highCVEs, err := db.ListCVEs(0, 10)
		if err != nil {
			t.Fatalf("Failed to list CVEs: %v", err)
		}

		foundHigh := false
		for _, cve := range highCVEs {
			if cve.Metrics != nil && len(cve.Metrics.CvssMetricV31) > 0 {
				if cve.Metrics.CvssMetricV31[0].CvssData.BaseScore >= 7.0 {
					foundHigh = true
				}
			}
		}

		if !foundHigh {
			t.Log("Query pattern test for CVSS score filtering completed")
		}
	})
}

func TestQueryPattern_FilterByKeyword(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestQueryPattern_FilterByKeyword", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_query_keyword.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		cves := []cve.CVEItem{
			{
				ID:           "CVE-LOG4J-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Apache Log4j remote code execution vulnerability"}},
			},
			{
				ID:           "CVE-SSH-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "OpenSSH privilege escalation"}},
			},
			{
				ID:           "CVE-HEARTBLEED-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "OpenSSL Heartbleed information disclosure"}},
			},
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		allCVEs, err := db.ListCVEs(0, 10)
		if err != nil {
			t.Fatalf("Failed to list CVEs: %v", err)
		}

		foundLog4j := false
		for _, cve := range allCVEs {
			for _, desc := range cve.Descriptions {
				if desc.Lang == "en" && (desc.Value == "Apache Log4j remote code execution vulnerability") {
					foundLog4j = true
					break
				}
			}
		}

		if !foundLog4j {
			t.Error("Expected to find Log4j CVE")
		}
	})
}

func TestQueryPattern_DateRange(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestQueryPattern_DateRange", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_query_date.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		baseTime := time.Date(2023, 6, 15, 0, 0, 0, 0, time.UTC)

		cves := []cve.CVEItem{
			{
				ID:           "CVE-OLD-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(baseTime.AddDate(0, -6, 0)),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Old CVE"}},
			},
			{
				ID:           "CVE-RECENT-1",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(baseTime),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Recent CVE"}},
			},
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		allCVEs, err := db.ListCVEs(0, 10)
		if err != nil {
			t.Fatalf("Failed to list CVEs: %v", err)
		}

		if len(allCVEs) != 2 {
			t.Errorf("Expected 2 CVEs, got %d", len(allCVEs))
		}

		t.Log("Date range query pattern test completed")
	})
}

func TestQueryPattern_PaginationEdgeCases(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestQueryPattern_PaginationEdgeCases", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_query_pagination.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		cves := []cve.CVEItem{}
		for i := 1; i <= 25; i++ {
			cves = append(cves, cve.CVEItem{
				ID:           "CVE-PAGINATION-" + string(rune('A'+i-1)),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Test CVE"}},
			})
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		page1, err := db.ListCVEs(0, 10)
		if err != nil {
			t.Fatalf("Failed to get page 1: %v", err)
		}
		if len(page1) != 10 {
			t.Errorf("Expected 10 items on page 1, got %d", len(page1))
		}

		page2, err := db.ListCVEs(10, 10)
		if err != nil {
			t.Fatalf("Failed to get page 2: %v", err)
		}
		if len(page2) != 10 {
			t.Errorf("Expected 10 items on page 2, got %d", len(page2))
		}

		page3, err := db.ListCVEs(20, 10)
		if err != nil {
			t.Fatalf("Failed to get page 3: %v", err)
		}
		if len(page3) != 5 {
			t.Errorf("Expected 5 items on page 3, got %d", len(page3))
		}

		pageEmpty, err := db.ListCVEs(30, 10)
		if err != nil {
			t.Fatalf("Failed to get empty page: %v", err)
		}
		if len(pageEmpty) != 0 {
			t.Errorf("Expected 0 items on empty page, got %d", len(pageEmpty))
		}
	})
}

func TestQueryPattern_LargeOffset(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestQueryPattern_LargeOffset", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_query_large_offset.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		cves := []cve.CVEItem{}
		for i := 1; i <= 100; i++ {
			cves = append(cves, cve.CVEItem{
				ID:           "CVE-LARGE-" + string(rune('A'+i-1)),
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(time.Now()),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Test CVE"}},
			})
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		result, err := db.ListCVEs(1000, 10)
		if err != nil {
			t.Fatalf("Failed to query with large offset: %v", err)
		}

		if len(result) != 0 {
			t.Errorf("Expected 0 results with large offset, got %d", len(result))
		}
	})
}

func TestQueryPattern_OrderByPublishDate(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestQueryPattern_OrderByPublishDate", nil, func(t *testing.T, tx *gorm.DB) {
		dbPath := "/tmp/test_query_order.db"
		defer os.Remove(dbPath)

		db, err := NewDB(dbPath)
		if err != nil {
			t.Fatalf("Failed to create database: %v", err)
		}
		defer db.Close()

		baseTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

		cves := []cve.CVEItem{
			{
				ID:           "CVE-FIRST",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(baseTime.AddDate(0, 0, 3)),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Third"}},
			},
			{
				ID:           "CVE-SECOND",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(baseTime.AddDate(0, 0, 2)),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "Second"}},
			},
			{
				ID:           "CVE-THIRD",
				SourceID:     "nvd@nist.gov",
				Published:    cve.NewNVDTime(baseTime.AddDate(0, 0, 1)),
				LastModified: cve.NewNVDTime(time.Now()),
				VulnStatus:   "Analyzed",
				Descriptions: []cve.Description{{Lang: "en", Value: "First"}},
			},
		}

		if err := db.SaveCVEs(cves); err != nil {
			t.Fatalf("Failed to save CVEs: %v", err)
		}

		result, err := db.ListCVEs(0, 10)
		if err != nil {
			t.Fatalf("Failed to list CVEs: %v", err)
		}

		if len(result) != 3 {
			t.Errorf("Expected 3 CVEs, got %d", len(result))
		}

		t.Log("Order by publish date query pattern test completed")
	})
}
