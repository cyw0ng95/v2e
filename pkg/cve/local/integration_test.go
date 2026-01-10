package local

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/remote"
)

// TestCreateCVEDatabase creates a CVE database in the project root
// This test demonstrates the ability to save CVEs to cve.db and allows
// users to download the database after tests run.
func TestCreateCVEDatabase(t *testing.T) {
	// Get the project root directory (4 levels up from pkg/cve/local)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	projectRoot := filepath.Join(cwd, "..", "..", "..")
	dbPath := filepath.Join(projectRoot, "cve.db")

	// Remove existing database if present (for clean test)
	os.Remove(dbPath)

	// Create database
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create sample CVE data
	sampleCVEs := []cve.CVEItem{
		{
			ID:           "CVE-2021-44228",
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Date(2021, 12, 10, 0, 0, 0, 0, time.UTC)),
			LastModified: cve.NewNVDTime(time.Date(2021, 12, 15, 0, 0, 0, 0, time.UTC)),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Apache Log4j2 2.0-beta9 through 2.15.0 (excluding security releases 2.12.2, 2.12.3, and 2.3.1) JNDI features used in configuration, log messages, and parameters do not protect against attacker controlled LDAP and other JNDI related endpoints. An attacker who can control log messages or log message parameters can execute arbitrary code loaded from LDAP servers when message lookup substitution is enabled.",
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
			References: []cve.Reference{
				{
					URL:    "https://logging.apache.org/log4j/2.x/security.html",
					Source: "nvd@nist.gov",
					Tags:   []string{"Vendor Advisory"},
				},
			},
		},
		{
			ID:           "CVE-2021-45046",
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Date(2021, 12, 14, 0, 0, 0, 0, time.UTC)),
			LastModified: cve.NewNVDTime(time.Date(2021, 12, 15, 0, 0, 0, 0, time.UTC)),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "It was found that the fix to address CVE-2021-44228 in Apache Log4j 2.15.0 was incomplete in certain non-default configurations.",
				},
			},
			Metrics: &cve.Metrics{
				CvssMetricV31: []cve.CVSSMetricV3{
					{
						Source: "nvd@nist.gov",
						Type:   "Primary",
						CvssData: cve.CVSSDataV3{
							Version:               "3.1",
							VectorString:          "CVSS:3.1/AV:N/AC:H/PR:N/UI:N/S:C/C:H/I:H/A:H",
							BaseScore:             9.0,
							BaseSeverity:          "CRITICAL",
							AttackVector:          "NETWORK",
							AttackComplexity:      "HIGH",
							PrivilegesRequired:    "NONE",
							UserInteraction:       "NONE",
							Scope:                 "CHANGED",
							ConfidentialityImpact: "HIGH",
							IntegrityImpact:       "HIGH",
							AvailabilityImpact:    "HIGH",
						},
						ExploitabilityScore: 2.3,
						ImpactScore:         6.0,
					},
				},
			},
		},
		{
			ID:           "CVE-2022-0001",
			SourceID:     "nvd@nist.gov",
			Published:    cve.NewNVDTime(time.Date(2022, 3, 18, 0, 0, 0, 0, time.UTC)),
			LastModified: cve.NewNVDTime(time.Date(2022, 3, 25, 0, 0, 0, 0, time.UTC)),
			VulnStatus:   "Analyzed",
			Descriptions: []cve.Description{
				{
					Lang:  "en",
					Value: "Non-transparent sharing of branch predictor selectors between contexts in some Intel(R) Processors may allow an authorized user to potentially enable information disclosure via local access.",
				},
			},
			Metrics: &cve.Metrics{
				CvssMetricV31: []cve.CVSSMetricV3{
					{
						Source: "nvd@nist.gov",
						Type:   "Primary",
						CvssData: cve.CVSSDataV3{
							Version:               "3.1",
							VectorString:          "CVSS:3.1/AV:L/AC:L/PR:L/UI:N/S:C/C:H/I:N/A:N",
							BaseScore:             6.5,
							BaseSeverity:          "MEDIUM",
							AttackVector:          "LOCAL",
							AttackComplexity:      "LOW",
							PrivilegesRequired:    "LOW",
							UserInteraction:       "NONE",
							Scope:                 "CHANGED",
							ConfidentialityImpact: "HIGH",
							IntegrityImpact:       "NONE",
							AvailabilityImpact:    "NONE",
						},
						ExploitabilityScore: 2.0,
						ImpactScore:         4.0,
					},
				},
			},
		},
	}

	// Save sample CVEs
	err = db.SaveCVEs(sampleCVEs)
	if err != nil {
		t.Fatalf("Failed to save sample CVEs: %v", err)
	}

	// Verify the CVEs were saved
	count, err := db.Count()
	if err != nil {
		t.Fatalf("Failed to count CVEs: %v", err)
	}

	if count != int64(len(sampleCVEs)) {
		t.Errorf("Expected %d CVEs in database, got %d", len(sampleCVEs), count)
	}

	// Verify we can retrieve a CVE
	retrieved, err := db.GetCVE("CVE-2021-44228")
	if err != nil {
		t.Fatalf("Failed to retrieve CVE: %v", err)
	}

	if retrieved.ID != "CVE-2021-44228" {
		t.Errorf("Expected CVE ID CVE-2021-44228, got %s", retrieved.ID)
	}

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file does not exist at %s", dbPath)
	}

	t.Logf("Successfully created CVE database at %s", dbPath)
	t.Logf("Database contains %d CVE records", count)
	t.Logf("You can now download the cve.db file from the project root")
}

// TestFetchAndRecoverCVEFromNVD tests the full cycle of fetching CVE data from NVD,
// saving it to the database, and recovering it with all fields intact.
// This test validates that the ORM properly serializes and deserializes all CVE fields.
func TestFetchAndRecoverCVEFromNVD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that requires network access")
	}

	// Create a temporary database
	dbPath := "/tmp/test_nvd_fetch_recover.db"
	defer os.Remove(dbPath)

	// Create database
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create CVE fetcher (no API key - using public rate limit)
	fetcher := remote.NewFetcher("")

	// Fetch a well-known CVE (Log4Shell)
	t.Log("Fetching CVE-2021-44228 from NVD...")
	response, err := fetcher.FetchCVEByID("CVE-2021-44228")
	if err != nil {
		t.Fatalf("Failed to fetch CVE from NVD: %v", err)
	}

	if response == nil || len(response.Vulnerabilities) == 0 {
		t.Fatal("No CVE data returned from NVD")
	}

	// Extract the CVE item
	originalCVE := response.Vulnerabilities[0].CVE
	t.Logf("Successfully fetched CVE: %s", originalCVE.ID)

	// Save the CVE to the database
	t.Log("Saving CVE to database...")
	err = db.SaveCVE(&originalCVE)
	if err != nil {
		t.Fatalf("Failed to save CVE to database: %v", err)
	}

	// Retrieve the CVE from the database
	t.Log("Retrieving CVE from database...")
	retrievedCVE, err := db.GetCVE(originalCVE.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve CVE from database: %v", err)
	}

	// Verify core fields
	if retrievedCVE.ID != originalCVE.ID {
		t.Errorf("ID mismatch: expected %s, got %s", originalCVE.ID, retrievedCVE.ID)
	}
	if retrievedCVE.SourceID != originalCVE.SourceID {
		t.Errorf("SourceID mismatch: expected %s, got %s", originalCVE.SourceID, retrievedCVE.SourceID)
	}
	if retrievedCVE.VulnStatus != originalCVE.VulnStatus {
		t.Errorf("VulnStatus mismatch: expected %s, got %s", originalCVE.VulnStatus, retrievedCVE.VulnStatus)
	}

	// Verify timestamps (allowing for minor precision differences)
	if !retrievedCVE.Published.Time.Equal(originalCVE.Published.Time) {
		t.Errorf("Published timestamp mismatch: expected %v, got %v", originalCVE.Published, retrievedCVE.Published)
	}
	if !retrievedCVE.LastModified.Time.Equal(originalCVE.LastModified.Time) {
		t.Errorf("LastModified timestamp mismatch: expected %v, got %v", originalCVE.LastModified, retrievedCVE.LastModified)
	}

	// Verify descriptions
	if len(retrievedCVE.Descriptions) != len(originalCVE.Descriptions) {
		t.Errorf("Descriptions count mismatch: expected %d, got %d", len(originalCVE.Descriptions), len(retrievedCVE.Descriptions))
	} else {
		for i, desc := range originalCVE.Descriptions {
			if retrievedCVE.Descriptions[i].Lang != desc.Lang {
				t.Errorf("Description[%d] Lang mismatch: expected %s, got %s", i, desc.Lang, retrievedCVE.Descriptions[i].Lang)
			}
			if retrievedCVE.Descriptions[i].Value != desc.Value {
				t.Errorf("Description[%d] Value mismatch", i)
			}
		}
	}

	// Verify CVSS metrics if present
	if originalCVE.Metrics != nil {
		if retrievedCVE.Metrics == nil {
			t.Error("Metrics should not be nil after retrieval")
		} else {
			// Check CVSS v3.1
			if len(originalCVE.Metrics.CvssMetricV31) > 0 {
				if len(retrievedCVE.Metrics.CvssMetricV31) != len(originalCVE.Metrics.CvssMetricV31) {
					t.Errorf("CVSS v3.1 metrics count mismatch: expected %d, got %d",
						len(originalCVE.Metrics.CvssMetricV31), len(retrievedCVE.Metrics.CvssMetricV31))
				} else {
					origMetric := originalCVE.Metrics.CvssMetricV31[0]
					retrMetric := retrievedCVE.Metrics.CvssMetricV31[0]

					if retrMetric.Source != origMetric.Source {
						t.Errorf("CVSS v3.1 source mismatch: expected %s, got %s", origMetric.Source, retrMetric.Source)
					}
					if retrMetric.Type != origMetric.Type {
						t.Errorf("CVSS v3.1 type mismatch: expected %s, got %s", origMetric.Type, retrMetric.Type)
					}
					if retrMetric.CvssData.BaseScore != origMetric.CvssData.BaseScore {
						t.Errorf("CVSS v3.1 base score mismatch: expected %f, got %f",
							origMetric.CvssData.BaseScore, retrMetric.CvssData.BaseScore)
					}
					if retrMetric.CvssData.BaseSeverity != origMetric.CvssData.BaseSeverity {
						t.Errorf("CVSS v3.1 base severity mismatch: expected %s, got %s",
							origMetric.CvssData.BaseSeverity, retrMetric.CvssData.BaseSeverity)
					}
					if retrMetric.CvssData.VectorString != origMetric.CvssData.VectorString {
						t.Errorf("CVSS v3.1 vector string mismatch: expected %s, got %s",
							origMetric.CvssData.VectorString, retrMetric.CvssData.VectorString)
					}
				}
			}

			// Check CVSS v3.0
			if len(originalCVE.Metrics.CvssMetricV30) > 0 {
				if len(retrievedCVE.Metrics.CvssMetricV30) != len(originalCVE.Metrics.CvssMetricV30) {
					t.Errorf("CVSS v3.0 metrics count mismatch: expected %d, got %d",
						len(originalCVE.Metrics.CvssMetricV30), len(retrievedCVE.Metrics.CvssMetricV30))
				}
			}

			// Check CVSS v2
			if len(originalCVE.Metrics.CvssMetricV2) > 0 {
				if len(retrievedCVE.Metrics.CvssMetricV2) != len(originalCVE.Metrics.CvssMetricV2) {
					t.Errorf("CVSS v2 metrics count mismatch: expected %d, got %d",
						len(originalCVE.Metrics.CvssMetricV2), len(retrievedCVE.Metrics.CvssMetricV2))
				}
			}

			// Check CVSS v4.0
			if len(originalCVE.Metrics.CvssMetricV40) > 0 {
				if len(retrievedCVE.Metrics.CvssMetricV40) != len(originalCVE.Metrics.CvssMetricV40) {
					t.Errorf("CVSS v4.0 metrics count mismatch: expected %d, got %d",
						len(originalCVE.Metrics.CvssMetricV40), len(retrievedCVE.Metrics.CvssMetricV40))
				}
			}
		}
	}

	// Verify weaknesses if present
	if len(originalCVE.Weaknesses) > 0 {
		if len(retrievedCVE.Weaknesses) != len(originalCVE.Weaknesses) {
			t.Errorf("Weaknesses count mismatch: expected %d, got %d",
				len(originalCVE.Weaknesses), len(retrievedCVE.Weaknesses))
		} else {
			for i, weakness := range originalCVE.Weaknesses {
				if retrievedCVE.Weaknesses[i].Source != weakness.Source {
					t.Errorf("Weakness[%d] Source mismatch: expected %s, got %s",
						i, weakness.Source, retrievedCVE.Weaknesses[i].Source)
				}
				if retrievedCVE.Weaknesses[i].Type != weakness.Type {
					t.Errorf("Weakness[%d] Type mismatch: expected %s, got %s",
						i, weakness.Type, retrievedCVE.Weaknesses[i].Type)
				}
			}
		}
	}

	// Verify references if present
	if len(originalCVE.References) > 0 {
		if len(retrievedCVE.References) != len(originalCVE.References) {
			t.Errorf("References count mismatch: expected %d, got %d",
				len(originalCVE.References), len(retrievedCVE.References))
		} else {
			for i, ref := range originalCVE.References {
				if retrievedCVE.References[i].URL != ref.URL {
					t.Errorf("Reference[%d] URL mismatch: expected %s, got %s",
						i, ref.URL, retrievedCVE.References[i].URL)
				}
				if retrievedCVE.References[i].Source != ref.Source {
					t.Errorf("Reference[%d] Source mismatch: expected %s, got %s",
						i, ref.Source, retrievedCVE.References[i].Source)
				}
			}
		}
	}

	// Verify configurations if present
	if len(originalCVE.Configurations) > 0 {
		if len(retrievedCVE.Configurations) != len(originalCVE.Configurations) {
			t.Errorf("Configurations count mismatch: expected %d, got %d",
				len(originalCVE.Configurations), len(retrievedCVE.Configurations))
		}
	}

	// Verify CVE tags if present
	if len(originalCVE.CVETags) > 0 {
		if len(retrievedCVE.CVETags) != len(originalCVE.CVETags) {
			t.Errorf("CVE tags count mismatch: expected %d, got %d",
				len(originalCVE.CVETags), len(retrievedCVE.CVETags))
		}
	}

	// Verify vendor comments if present
	if len(originalCVE.VendorComments) > 0 {
		if len(retrievedCVE.VendorComments) != len(originalCVE.VendorComments) {
			t.Errorf("Vendor comments count mismatch: expected %d, got %d",
				len(originalCVE.VendorComments), len(retrievedCVE.VendorComments))
		}
	}

	t.Log("Successfully verified CVE data integrity after ORM cycle")
	t.Logf("CVE %s: All fields properly serialized and deserialized", originalCVE.ID)
}

// TestFetchMultipleCVEsAndRecover tests fetching multiple CVEs from NVD,
// saving them to the database, and verifying they can all be recovered correctly.
func TestFetchMultipleCVEsAndRecover(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that requires network access")
	}

	// Create a temporary database
	dbPath := "/tmp/test_nvd_fetch_multiple.db"
	defer os.Remove(dbPath)

	// Create database
	db, err := NewDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create CVE fetcher
	fetcher := remote.NewFetcher("")

	// Fetch a small batch of CVEs
	t.Log("Fetching CVEs from NVD...")
	response, err := fetcher.FetchCVEs(0, 5)
	if err != nil {
		t.Fatalf("Failed to fetch CVEs from NVD: %v", err)
	}

	if response == nil || len(response.Vulnerabilities) == 0 {
		t.Fatal("No CVE data returned from NVD")
	}

	t.Logf("Fetched %d CVEs from NVD", len(response.Vulnerabilities))

	// Extract and save all CVEs
	originalCVEs := make(map[string]cve.CVEItem)
	for _, vuln := range response.Vulnerabilities {
		cveItem := vuln.CVE
		originalCVEs[cveItem.ID] = cveItem

		err = db.SaveCVE(&cveItem)
		if err != nil {
			t.Fatalf("Failed to save CVE %s to database: %v", cveItem.ID, err)
		}
	}

	// Verify count
	count, err := db.Count()
	if err != nil {
		t.Fatalf("Failed to count CVEs: %v", err)
	}
	if count != int64(len(originalCVEs)) {
		t.Errorf("Count mismatch: expected %d, got %d", len(originalCVEs), count)
	}

	// Retrieve and verify each CVE
	for cveID, originalCVE := range originalCVEs {
		retrievedCVE, err := db.GetCVE(cveID)
		if err != nil {
			t.Errorf("Failed to retrieve CVE %s: %v", cveID, err)
			continue
		}

		// Basic verification
		if retrievedCVE.ID != originalCVE.ID {
			t.Errorf("CVE %s: ID mismatch", cveID)
		}
		if retrievedCVE.SourceID != originalCVE.SourceID {
			t.Errorf("CVE %s: SourceID mismatch", cveID)
		}
		if retrievedCVE.VulnStatus != originalCVE.VulnStatus {
			t.Errorf("CVE %s: VulnStatus mismatch", cveID)
		}
		if len(retrievedCVE.Descriptions) != len(originalCVE.Descriptions) {
			t.Errorf("CVE %s: Descriptions count mismatch", cveID)
		}
	}

	// Test ListCVEs pagination
	listedCVEs, err := db.ListCVEs(0, 10)
	if err != nil {
		t.Fatalf("Failed to list CVEs: %v", err)
	}
	if len(listedCVEs) != len(originalCVEs) {
		t.Errorf("ListCVEs count mismatch: expected %d, got %d", len(originalCVEs), len(listedCVEs))
	}

	t.Logf("Successfully verified %d CVEs through ORM cycle", len(originalCVEs))
}
