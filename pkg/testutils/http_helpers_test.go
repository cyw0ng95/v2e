package testutils

import (
	"encoding/json"
	"testing"
)

type cveResponse struct {
	ResultsPerPage  int `json:"resultsPerPage"`
	StartIndex      int `json:"startIndex"`
	TotalResults    int `json:"totalResults"`
	Vulnerabilities []struct {
		CVE struct {
			ID           string `json:"id"`
			Descriptions []struct {
				Lang  string `json:"lang"`
				Value string `json:"value"`
			} `json:"descriptions"`
		} `json:"cve"`
	} `json:"vulnerabilities"`
}

func TestMakeCVEResponseJSON_WithID(t *testing.T) {
	data := MakeCVEResponseJSON("CVE-1234", 5)

	var resp cveResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.TotalResults != 5 || resp.ResultsPerPage != 1 || resp.StartIndex != 0 {
		t.Fatalf("Unexpected metadata: %+v", resp)
	}

	if len(resp.Vulnerabilities) != 1 {
		t.Fatalf("Expected a single vulnerability entry")
	}

	got := resp.Vulnerabilities[0].CVE
	if got.ID != "CVE-1234" {
		t.Fatalf("Unexpected CVE ID: %s", got.ID)
	}
	if len(got.Descriptions) != 1 || got.Descriptions[0].Value == "" {
		t.Fatalf("Description not populated: %+v", got.Descriptions)
	}
}

func TestMakeCVEListResponseJSON_EmptyList(t *testing.T) {
	data := MakeCVEListResponseJSON(3)

	var resp cveResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if resp.TotalResults != 3 {
		t.Fatalf("Unexpected total results: %d", resp.TotalResults)
	}
	if len(resp.Vulnerabilities) != 0 {
		t.Fatalf("Expected no vulnerabilities, got %d", len(resp.Vulnerabilities))
	}
}
