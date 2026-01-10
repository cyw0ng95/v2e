package repo

import (
	"os"
	"testing"
	"time"
)

func TestNewDB(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_cve.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("NewDB should not return nil")
	}
}

func TestSaveCVE(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_save_cve.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create a test CVE
	cve := &CVEItem{
		ID:           "CVE-2021-44228",
		SourceID:     "nvd@nist.gov",
		Published:    time.Now(),
		LastModified: time.Now(),
		VulnStatus:   "Analyzed",
		Descriptions: []Description{
			{
				Lang:  "en",
				Value: "Apache Log4j2 vulnerability",
			},
		},
	}

	// Save CVE
	err = db.SaveCVE(cve)
	if err != nil {
		t.Errorf("Failed to save CVE: %v", err)
	}

	// Verify it was saved
	count, err := db.Count()
	if err != nil {
		t.Errorf("Failed to count CVEs: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 CVE, got %d", count)
	}
}

func TestSaveCVE_Update(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_update_cve.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create a test CVE
	cve := &CVEItem{
		ID:           "CVE-2021-44228",
		SourceID:     "nvd@nist.gov",
		Published:    time.Now(),
		LastModified: time.Now(),
		VulnStatus:   "Analyzed",
		Descriptions: []Description{
			{
				Lang:  "en",
				Value: "Original description",
			},
		},
	}

	// Save CVE
	err = db.SaveCVE(cve)
	if err != nil {
		t.Errorf("Failed to save CVE: %v", err)
	}

	// Update CVE
	cve.Descriptions[0].Value = "Updated description"
	err = db.SaveCVE(cve)
	if err != nil {
		t.Errorf("Failed to update CVE: %v", err)
	}

	// Verify count is still 1 (not duplicated)
	count, err := db.Count()
	if err != nil {
		t.Errorf("Failed to count CVEs: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 CVE after update, got %d", count)
	}

	// Verify the description was updated
	retrieved, err := db.GetCVE("CVE-2021-44228")
	if err != nil {
		t.Errorf("Failed to retrieve CVE: %v", err)
	}
	if retrieved.Descriptions[0].Value != "Updated description" {
		t.Errorf("Expected updated description, got %s", retrieved.Descriptions[0].Value)
	}
}

func TestSaveCVEs(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_save_cves.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create multiple test CVEs
	cves := []CVEItem{
		{
			ID:           "CVE-2021-44228",
			SourceID:     "nvd@nist.gov",
			Published:    time.Now(),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "Log4j"}},
		},
		{
			ID:           "CVE-2021-45046",
			SourceID:     "nvd@nist.gov",
			Published:    time.Now(),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "Log4j 2"}},
		},
	}

	// Save CVEs
	err = db.SaveCVEs(cves)
	if err != nil {
		t.Errorf("Failed to save CVEs: %v", err)
	}

	// Verify count
	count, err := db.Count()
	if err != nil {
		t.Errorf("Failed to count CVEs: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 CVEs, got %d", count)
	}
}

func TestGetCVE(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_get_cve.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create and save a test CVE
	cve := &CVEItem{
		ID:           "CVE-2021-44228",
		SourceID:     "nvd@nist.gov",
		Published:    time.Date(2021, 12, 10, 0, 0, 0, 0, time.UTC),
		LastModified: time.Date(2021, 12, 15, 0, 0, 0, 0, time.UTC),
		VulnStatus:   "Analyzed",
		Descriptions: []Description{
			{
				Lang:  "en",
				Value: "Apache Log4j2 vulnerability",
			},
		},
	}

	err = db.SaveCVE(cve)
	if err != nil {
		t.Errorf("Failed to save CVE: %v", err)
	}

	// Retrieve CVE
	retrieved, err := db.GetCVE("CVE-2021-44228")
	if err != nil {
		t.Errorf("Failed to retrieve CVE: %v", err)
	}

	if retrieved.ID != "CVE-2021-44228" {
		t.Errorf("Expected ID CVE-2021-44228, got %s", retrieved.ID)
	}
	if retrieved.SourceID != "nvd@nist.gov" {
		t.Errorf("Expected SourceID nvd@nist.gov, got %s", retrieved.SourceID)
	}
	if retrieved.VulnStatus != "Analyzed" {
		t.Errorf("Expected VulnStatus Analyzed, got %s", retrieved.VulnStatus)
	}
	if len(retrieved.Descriptions) != 1 {
		t.Fatalf("Expected 1 description, got %d", len(retrieved.Descriptions))
	}
	if retrieved.Descriptions[0].Value != "Apache Log4j2 vulnerability" {
		t.Errorf("Expected description 'Apache Log4j2 vulnerability', got %s", retrieved.Descriptions[0].Value)
	}
}

func TestGetCVE_NotFound(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_get_cve_notfound.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Try to retrieve non-existent CVE
	_, err = db.GetCVE("CVE-9999-99999")
	if err == nil {
		t.Error("Expected error when retrieving non-existent CVE")
	}
}

func TestListCVEs(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_list_cves.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create and save multiple CVEs with different publish dates
	cves := []CVEItem{
		{
			ID:           "CVE-2021-44228",
			SourceID:     "nvd@nist.gov",
			Published:    time.Date(2021, 12, 10, 0, 0, 0, 0, time.UTC),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "CVE 1"}},
		},
		{
			ID:           "CVE-2021-45046",
			SourceID:     "nvd@nist.gov",
			Published:    time.Date(2021, 12, 14, 0, 0, 0, 0, time.UTC),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "CVE 2"}},
		},
		{
			ID:           "CVE-2022-12345",
			SourceID:     "nvd@nist.gov",
			Published:    time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "CVE 3"}},
		},
	}

	err = db.SaveCVEs(cves)
	if err != nil {
		t.Errorf("Failed to save CVEs: %v", err)
	}

	// List all CVEs
	retrieved, err := db.ListCVEs(0, 10)
	if err != nil {
		t.Errorf("Failed to list CVEs: %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("Expected 3 CVEs, got %d", len(retrieved))
	}

	// Verify order (should be descending by publish date)
	if retrieved[0].ID != "CVE-2022-12345" {
		t.Errorf("Expected first CVE to be CVE-2022-12345, got %s", retrieved[0].ID)
	}
}

func TestListCVEs_Pagination(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_list_cves_pagination.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create and save 5 CVEs
	cves := []CVEItem{}
	for i := 1; i <= 5; i++ {
		cves = append(cves, CVEItem{
			ID:           "CVE-2021-" + string(rune(10000+i)),
			SourceID:     "nvd@nist.gov",
			Published:    time.Date(2021, 12, i, 0, 0, 0, 0, time.UTC),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "CVE"}},
		})
	}

	err = db.SaveCVEs(cves)
	if err != nil {
		t.Errorf("Failed to save CVEs: %v", err)
	}

	// Get first page (limit 2)
	page1, err := db.ListCVEs(0, 2)
	if err != nil {
		t.Errorf("Failed to list CVEs page 1: %v", err)
	}
	if len(page1) != 2 {
		t.Errorf("Expected 2 CVEs on page 1, got %d", len(page1))
	}

	// Get second page (offset 2, limit 2)
	page2, err := db.ListCVEs(2, 2)
	if err != nil {
		t.Errorf("Failed to list CVEs page 2: %v", err)
	}
	if len(page2) != 2 {
		t.Errorf("Expected 2 CVEs on page 2, got %d", len(page2))
	}

	// Verify pages don't overlap
	if page1[0].ID == page2[0].ID {
		t.Error("Pages should not contain the same CVE")
	}
}

func TestCount(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_count.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initially should be 0
	count, err := db.Count()
	if err != nil {
		t.Errorf("Failed to count CVEs: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 CVEs initially, got %d", count)
	}

	// Add CVEs
	cves := []CVEItem{
		{
			ID:           "CVE-2021-44228",
			SourceID:     "nvd@nist.gov",
			Published:    time.Now(),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "CVE 1"}},
		},
		{
			ID:           "CVE-2021-45046",
			SourceID:     "nvd@nist.gov",
			Published:    time.Now(),
			LastModified: time.Now(),
			VulnStatus:   "Analyzed",
			Descriptions: []Description{{Lang: "en", Value: "CVE 2"}},
		},
	}

	err = db.SaveCVEs(cves)
	if err != nil {
		t.Errorf("Failed to save CVEs: %v", err)
	}

	// Count should be 2
	count, err = db.Count()
	if err != nil {
		t.Errorf("Failed to count CVEs: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 CVEs, got %d", count)
	}
}

func TestSaveCVE_WithFullData(t *testing.T) {
	// Create a temporary database
	dbPath := "/tmp/test_save_full_cve.db"
	defer os.Remove(dbPath)

	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create a CVE with full data including metrics
	cve := &CVEItem{
		ID:           "CVE-2021-44228",
		SourceID:     "nvd@nist.gov",
		Published:    time.Date(2021, 12, 10, 0, 0, 0, 0, time.UTC),
		LastModified: time.Date(2021, 12, 15, 0, 0, 0, 0, time.UTC),
		VulnStatus:   "Analyzed",
		Descriptions: []Description{
			{
				Lang:  "en",
				Value: "Apache Log4j2 vulnerability",
			},
		},
		Metrics: &Metrics{
			CvssMetricV31: []CVSSMetricV3{
				{
					Source: "nvd@nist.gov",
					Type:   "Primary",
					CvssData: CVSSDataV3{
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
		References: []Reference{
			{
				URL:    "https://logging.apache.org/log4j/2.x/security.html",
				Source: "nvd@nist.gov",
				Tags:   []string{"Vendor Advisory"},
			},
		},
	}

	// Save CVE
	err = db.SaveCVE(cve)
	if err != nil {
		t.Errorf("Failed to save CVE with full data: %v", err)
	}

	// Retrieve and verify
	retrieved, err := db.GetCVE("CVE-2021-44228")
	if err != nil {
		t.Errorf("Failed to retrieve CVE: %v", err)
	}

	// Verify metrics were preserved
	if retrieved.Metrics == nil {
		t.Fatal("Metrics should not be nil")
	}
	if len(retrieved.Metrics.CvssMetricV31) != 1 {
		t.Fatalf("Expected 1 CVSS v3.1 metric, got %d", len(retrieved.Metrics.CvssMetricV31))
	}
	if retrieved.Metrics.CvssMetricV31[0].CvssData.BaseScore != 10.0 {
		t.Errorf("Expected base score 10.0, got %f", retrieved.Metrics.CvssMetricV31[0].CvssData.BaseScore)
	}

	// Verify references were preserved
	if len(retrieved.References) != 1 {
		t.Fatalf("Expected 1 reference, got %d", len(retrieved.References))
	}
	if retrieved.References[0].URL != "https://logging.apache.org/log4j/2.x/security.html" {
		t.Errorf("Expected reference URL to be preserved, got %s", retrieved.References[0].URL)
	}
}
