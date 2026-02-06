package cve

import (
	"encoding/json"
	"fmt"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
	"strings"
	"testing"
)

// TestCVEItem_JSONMarshalUnmarshal covers CVE JSON serialization edge cases.
func TestCVEItem_JSONMarshalUnmarshal(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCVEItem_JSONMarshalUnmarshal", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			item CVEItem
		}{
			{
				name: "minimal-item",
				item: CVEItem{ID: "CVE-2021-0001"},
			},
			{
				name: "item-with-status",
				item: CVEItem{ID: "CVE-2021-0002", VulnStatus: "Analyzed"},
			},
			{
				name: "item-with-source",
				item: CVEItem{ID: "CVE-2021-0003", SourceID: "cve@mitre.org"},
			},
			{
				name: "unicode-id",
				item: CVEItem{ID: "CVE-2021-0004"},
			},
			{
				name: "item-with-descriptions",
				item: CVEItem{
					ID: "CVE-2021-0005",
					Descriptions: []Description{
						{Lang: "en", Value: "Test description"},
					},
				},
			},
			{
				name: "unicode-description",
				item: CVEItem{
					ID: "CVE-2021-0006",
					Descriptions: []Description{
						{Lang: "zh", Value: "漏洞描述 - уязвимость - 脆弱性"},
					},
				},
			},
			{
				name: "long-description",
				item: CVEItem{
					ID: "CVE-2021-0007",
					Descriptions: []Description{
						{Lang: "en", Value: strings.Repeat("x", 10000)},
					},
				},
			},
			{
				name: "html-in-description",
				item: CVEItem{
					ID: "CVE-2021-0008",
					Descriptions: []Description{
						{Lang: "en", Value: "<script>alert('xss')</script>"},
					},
				},
			},
			{
				name: "multiple-descriptions",
				item: CVEItem{
					ID: "CVE-2021-0009",
					Descriptions: []Description{
						{Lang: "en", Value: "English description"},
						{Lang: "es", Value: "Descripción en español"},
						{Lang: "zh", Value: "中文描述"},
					},
				},
			},
			{
				name: "item-with-references",
				item: CVEItem{
					ID: "CVE-2021-0010",
					References: []Reference{
						{URL: "https://example.com/1"},
						{URL: "https://example.com/2"},
					},
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CVEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.item.ID {
					t.Fatalf("ID mismatch: want %s got %s", tc.item.ID, decoded.ID)
				}
			})
		}
	})

}

// TestCVEItem_IDFormats validates various CVE ID formats.
func TestCVEItem_IDFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCVEItem_IDFormats", nil, func(t *testing.T, tx *gorm.DB) {
		validIDs := []string{
			"CVE-1999-0001",
			"CVE-2000-0001",
			"CVE-2021-12345",
			"CVE-2024-99999",
			"CVE-2025-100000",
			"CVE-2030-1234567",
		}

		for _, id := range validIDs {
			t.Run(id, func(t *testing.T) {
				item := CVEItem{ID: id}
				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CVEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != id {
					t.Fatalf("ID mismatch: want %s got %s", id, decoded.ID)
				}
			})
		}
	})

}

// TestDescription_Formats validates description edge cases.
func TestDescription_Formats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestDescription_Formats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			desc Description
		}{
			{name: "english", desc: Description{Lang: "en", Value: "English"}},
			{name: "spanish", desc: Description{Lang: "es", Value: "Español"}},
			{name: "chinese", desc: Description{Lang: "zh", Value: "中文"}},
			{name: "russian", desc: Description{Lang: "ru", Value: "Русский"}},
			{name: "japanese", desc: Description{Lang: "ja", Value: "日本語"}},
			{name: "empty-value", desc: Description{Lang: "en", Value: ""}},
			{name: "multiline", desc: Description{Lang: "en", Value: "Line1\nLine2\nLine3"}},
			{name: "special-chars", desc: Description{Lang: "en", Value: "<>&\"'"}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.desc)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded Description
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.Lang != tc.desc.Lang {
					t.Fatalf("Lang mismatch: want %s got %s", tc.desc.Lang, decoded.Lang)
				}
			})
		}
	})

}

// TestReference_URLFormats validates reference URL edge cases.
func TestReference_URLFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestReference_URLFormats", nil, func(t *testing.T, tx *gorm.DB) {
		urls := []string{
			"http://example.com",
			"https://example.com",
			"https://example.com:8080/path",
			"https://example.com?param=value",
			"https://example.com#section",
			"https://example.com/path%20with%20spaces",
			"https://example.com/" + strings.Repeat("path/", 100),
		}

		for _, url := range urls {
			t.Run(fmt.Sprintf("url-len-%d", len(url)), func(t *testing.T) {
				ref := Reference{URL: url}
				data, err := json.Marshal(&ref)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded Reference
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.URL != url {
					t.Fatalf("URL mismatch")
				}
			})
		}
	})

}

// TestCVEItem_StatusValues validates vulnerability status values.
func TestCVEItem_StatusValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCVEItem_StatusValues", nil, func(t *testing.T, tx *gorm.DB) {
		statuses := []string{
			"Analyzed",
			"Modified",
			"Awaiting Analysis",
			"Undergoing Analysis",
			"Deferred",
			"Rejected",
			"",
		}

		for _, status := range statuses {
			t.Run(fmt.Sprintf("status-%s", status), func(t *testing.T) {
				item := CVEItem{ID: "CVE-2021-0001", VulnStatus: status}
				data, err := json.Marshal(&item)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CVEItem
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.VulnStatus != status {
					t.Fatalf("Status mismatch: want %s got %s", status, decoded.VulnStatus)
				}
			})
		}
	})

}

// TestCVEResponse_JSONFormats validates top-level response structure.
func TestCVEResponse_JSONFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestCVEResponse_JSONFormats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name string
			resp CVEResponse
		}{
			{
				name: "empty-response",
				resp: CVEResponse{},
			},
			{
				name: "with-pagination",
				resp: CVEResponse{
					ResultsPerPage: 100,
					StartIndex:     0,
					TotalResults:   1000,
				},
			},
			{
				name: "with-single-vulnerability",
				resp: CVEResponse{
					ResultsPerPage: 1,
					Vulnerabilities: []struct {
						CVE CVEItem `json:"cve"`
					}{{CVE: CVEItem{ID: "CVE-2021-0001"}}},
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.resp)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded CVEResponse
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ResultsPerPage != tc.resp.ResultsPerPage {
					t.Fatalf("ResultsPerPage mismatch")
				}
			})
		}
	})

}
