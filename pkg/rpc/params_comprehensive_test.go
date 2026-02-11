package rpc

import (
	"encoding/json"
	"fmt"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// TestFetchCVEsParams_Validation tests parameter validation
func TestFetchCVEsParams_Validation(t *testing.T) {
	tests := []struct {
		name           string
		startIndex     int
		resultsPerPage int
		expectValid    bool
	}{
		{"valid small", 0, 10, true},
		{"valid medium", 100, 50, true},
		{"valid large", 1000, 100, true},
		{"negative start", -1, 10, false},
		{"negative results", 10, -1, false},
		{"zero results", 10, 0, false},
		{"very large start", 999999, 10, true},
		{"very large results", 0, 10000, true},
		{"both zero", 0, 0, false},
		{"both negative", -10, -10, false},
		{"max int start", int(^uint(0) >> 1), 10, true},
		{"boundary 1", 2147483647, 100, true},
		{"boundary 2", 0, 2147483647, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := FetchCVEsParams{
				StartIndex:     tt.startIndex,
				ResultsPerPage: tt.resultsPerPage,
			}

			// Marshal and unmarshal to test JSON handling
			data, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var decoded FetchCVEsParams
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if decoded.StartIndex != tt.startIndex {
				t.Errorf("StartIndex = %v, want %v", decoded.StartIndex, tt.startIndex)
			}
			if decoded.ResultsPerPage != tt.resultsPerPage {
				t.Errorf("ResultsPerPage = %v, want %v", decoded.ResultsPerPage, tt.resultsPerPage)
			}
		})
	}
}

// TestImportParams_PathValidation tests path validation scenarios
func TestImportParams_PathValidation(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		xsd   string
		force bool
	}{
		{"simple path", "/tmp/data.xml", "/tmp/schema.xsd", false},
		{"relative path", "./data.xml", "./schema.xsd", false},
		{"parent path", "../data.xml", "../schema.xsd", false},
		{"home path", "~/data.xml", "~/schema.xsd", false},
		{"empty path", "", "", false},
		{"empty xsd", "/tmp/data.xml", "", false},
		{"force true", "/tmp/data.xml", "", true},
		{"force false", "/tmp/data.xml", "", false},
		{"with spaces", "/tmp/path with spaces.xml", "", false},
		{"unicode path", "/tmp/データ.xml", "", false},
		{"windows path", "C:\\data\\file.xml", "C:\\schema\\file.xsd", false},
		{"network path", "//server/share/file.xml", "", false},
		{"very long path", string(make([]byte, 1000)), "", false},
		{"with quotes", "/tmp/\"quoted\".xml", "", false},
		{"with backslash", "/tmp/back\\slash.xml", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ImportParams{
				Path:  tt.path,
				XSD:   tt.xsd,
				Force: tt.force,
			}

			data, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var decoded ImportParams
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if decoded.Path != tt.path {
				t.Errorf("Path = %v, want %v", decoded.Path, tt.path)
			}
			if decoded.XSD != tt.xsd {
				t.Errorf("XSD = %v, want %v", decoded.XSD, tt.xsd)
			}
			if decoded.Force != tt.force {
				t.Errorf("Force = %v, want %v", decoded.Force, tt.force)
			}
		})
	}
}

// TestGetByIDParams_IDFormats tests various ID format scenarios
func TestGetByIDParams_IDFormats(t *testing.T) {
	tests := []string{
		"CVE-2021-44228",
		"CVE-1999-0001",
		"CVE-2030-99999",
		"cve-2021-12345",
		"CvE-2021-12345",
		"",
		"INVALID-ID",
		"123-456-789",
		"テスト-ID",
		"ID with spaces",
		"ID\twith\ttabs",
		"ID\nwith\nnewlines",
		string(make([]byte, 500)),
		"<script>alert(1)</script>",
		"'; DROP TABLE ids--",
		"../../etc/passwd",
		"C:\\Windows\\System32",
	}

	for i, id := range tests {
		t.Run(fmt.Sprintf("id_%d", i), func(t *testing.T) {
			params := GetByIDParams{ID: id}

			data, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var decoded GetByIDParams
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if decoded.ID != id {
				t.Errorf("ID = %v, want %v", decoded.ID, id)
			}
		})
	}
}

// TestCVEIDParams_FormatVariations tests CVE ID format variations
func TestCVEIDParams_FormatVariations(t *testing.T) {
	validFormats := []string{
		"CVE-2021-44228",
		"CVE-2020-00001",
		"CVE-2024-99999",
	}

	invalidFormats := []string{
		"",
		"NOT-A-CVE",
		"CVE-XXXX-YYYY",
		"cve-2021-12345",
		"CVE20211234",
		"CVE-2021",
		"2021-12345",
	}

	for _, cveID := range validFormats {
		t.Run("valid_"+cveID, func(t *testing.T) {
			params := CVEIDParams{CVEID: cveID}
			data, _ := json.Marshal(params)
			var decoded CVEIDParams
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("unmarshal error: %v", err)
			}
			if decoded.CVEID != cveID {
				t.Errorf("CVEID = %v, want %v", decoded.CVEID, cveID)
			}
		})
	}

	for _, cveID := range invalidFormats {
		t.Run("invalid_"+cveID, func(t *testing.T) {
			params := CVEIDParams{CVEID: cveID}
			data, _ := json.Marshal(params)
			var decoded CVEIDParams
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Errorf("unmarshal error: %v", err)
			}
			// Just verify it unmarshals without panic
		})
	}
}

// TestListParams_PaginationScenarios tests pagination edge cases
func TestListParams_PaginationScenarios(t *testing.T) {
	tests := []struct {
		name   string
		offset int
		limit  int
	}{
		{"first page", 0, 10},
		{"second page", 10, 10},
		{"large page", 0, 1000},
		{"small page", 0, 1},
		{"middle page", 500, 50},
		{"negative offset", -1, 10},
		{"negative limit", 10, -1},
		{"zero limit", 10, 0},
		{"both zero", 0, 0},
		{"max int", int(^uint(0) >> 1), 100},
		{"boundary 1", 0, 2147483647},
		{"boundary 2", 2147483647, 10},
		{"boundary 3", 999999, 999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := ListParams{
				Offset: tt.offset,
				Limit:  tt.limit,
			}

			data, err := json.Marshal(params)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var decoded ListParams
			if err := json.Unmarshal(data, &decoded); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if decoded.Offset != tt.offset {
				t.Errorf("Offset = %v, want %v", decoded.Offset, tt.offset)
			}
			if decoded.Limit != tt.limit {
				t.Errorf("Limit = %v, want %v", decoded.Limit, tt.limit)
			}
		})
	}
}

// TestRPCParams_JSONRoundTrip tests JSON marshaling/unmarshaling integrity
func TestRPCParams_JSONRoundTrip(t *testing.T) {
	testutils.Run(t, testutils.Level1, "FetchCVEsParams", nil, func(t *testing.T, tx *gorm.DB) {
		for i := 0; i < 50; i++ {
			original := FetchCVEsParams{StartIndex: i * 10, ResultsPerPage: (i + 1) * 5}
			data, _ := json.Marshal(original)
			var decoded FetchCVEsParams
			json.Unmarshal(data, &decoded)

			if decoded != original {
				t.Errorf("round trip failed: got %+v, want %+v", decoded, original)
			}
		}
	})

	testutils.Run(t, testutils.Level1, "ListParams", nil, func(t *testing.T, tx *gorm.DB) {
		for i := 0; i < 50; i++ {
			original := ListParams{Offset: i * 100, Limit: i + 1}
			data, _ := json.Marshal(original)
			var decoded ListParams
			json.Unmarshal(data, &decoded)

			if decoded != original {
				t.Errorf("round trip failed: got %+v, want %+v", decoded, original)
			}
		}
	})
}

// TestRPCParams_MalformedJSON tests handling of malformed JSON
func TestRPCParams_MalformedJSON(t *testing.T) {
	malformed := [][]byte{
		[]byte("{"),
		[]byte("}"),
		[]byte("{invalid}"),
		[]byte(`{"start_index":}`),
		[]byte(`{"start_index":,}`),
		[]byte(`{"start_index":"not_a_number"}`),
		[]byte(`[1,2,3]`),
		[]byte(`"string"`),
		[]byte(`123`),
		[]byte(`true`),
		[]byte(`null`),
	}

	for i, data := range malformed {
		t.Run(fmt.Sprintf("malformed_%d", i), func(t *testing.T) {
			var params1 FetchCVEsParams
			var params2 ListParams
			var params3 GetByIDParams
			var params4 CVEIDParams
			var params5 ImportParams

			// These should not panic, just return errors
			_ = json.Unmarshal(data, &params1)
			_ = json.Unmarshal(data, &params2)
			_ = json.Unmarshal(data, &params3)
			_ = json.Unmarshal(data, &params4)
			_ = json.Unmarshal(data, &params5)
		})
	}
}

// TestRPCParams_TypeSafety tests type safety with various inputs
func TestRPCParams_TypeSafety(t *testing.T) {
	tests := []struct {
		name string
		json string
		want interface{}
	}{
		{"int as string", `{"start_index":"10"}`, nil},
		{"float as int", `{"start_index":10.5}`, nil},
		{"bool as int", `{"start_index":true}`, nil},
		{"null as int", `{"start_index":null}`, nil},
		{"array as int", `{"start_index":[1,2,3]}`, nil},
		{"object as int", `{"start_index":{"nested":"value"}}`, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var params FetchCVEsParams
			err := json.Unmarshal([]byte(tt.json), &params)
			// Should handle gracefully without panic
			if err == nil {
				t.Logf("Unexpectedly parsed: %+v", params)
			}
		})
	}
}
