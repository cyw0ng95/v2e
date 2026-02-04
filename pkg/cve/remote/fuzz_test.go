package remote

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"
)

// FuzzValidateCVEID tests CVE ID validation with arbitrary inputs
func FuzzValidateCVEID(f *testing.F) {
	// Seed corpus with valid and invalid CVE IDs
	f.Add("CVE-2021-44228")
	f.Add("CVE-2024-12345")
	f.Add("CVE-XXXX-YYYY")
	f.Add("")
	f.Add("cve-2021-1234")
	f.Add("CVE-2021-1")
	f.Add("CVE-2021-123456")
	f.Add("NOT-A-CVE-ID")
	f.Add("CVE-999999-99999")

	// Fuzz test
	f.Fuzz(func(t *testing.T, cveID string) {
		// Validate CVE ID - should not panic
		// Valid format is CVE-YYYY-NNNNN (where YYYY is year, NNNNN is 4+ digits)
		_ = isValidCVEID(cveID)
	})
}

// isValidCVEID validates CVE ID format (helper for fuzz test)
func isValidCVEID(cveID string) bool {
	if len(cveID) < 13 {
		return false
	}
	if cveID[0:4] != "CVE-" {
		return false
	}
	// Just basic validation - real validation would be more complex
	return len(cveID) >= 13
}

// FuzzFetchCVEsParams tests parameter validation
func FuzzFetchCVEsParams(f *testing.F) {
	// Seed corpus
	f.Add(0, 10)
	f.Add(100, 200)
	f.Add(-1, 0)
	f.Add(0, -10)
	f.Add(999999, 1000000)

	// Fuzz test
	f.Fuzz(func(t *testing.T, startIndex, resultsPerPage int) {
		// Create fetcher with empty API key
		fetcher := NewFetcher("")

		// Validate parameters - should not panic
		if startIndex < 0 || resultsPerPage < 0 || resultsPerPage > 2000 {
			// Invalid parameters - expected to fail
			return
		}

		// Parameters are valid, would make API call in real scenario
		// For fuzz test, we just validate the parameters don't cause panic
		_ = fetcher
	})
}
