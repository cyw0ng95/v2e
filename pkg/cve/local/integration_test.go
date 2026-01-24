package local

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/cve"
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
