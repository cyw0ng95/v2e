package attack

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// TestAttackTechnique_JSONMarshalUnmarshal covers ATT&CK technique JSON serialization.
func TestAttackTechnique_JSONMarshalUnmarshal(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackTechnique_JSONMarshalUnmarshal", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name      string
			technique AttackTechnique
		}{
			{
				name:      "minimal-technique",
				technique: AttackTechnique{ID: "T1001"},
			},
			{
				name:      "technique-with-name",
				technique: AttackTechnique{ID: "T1001", Name: "Data Obfuscation"},
			},
			{
				name:      "unicode-name",
				technique: AttackTechnique{ID: "T1002", Name: "数据混淆 - обфускация данных"},
			},
			{
				name:      "long-name",
				technique: AttackTechnique{ID: "T1003", Name: strings.Repeat("x", 500)},
			},
			{
				name:      "special-chars-name",
				technique: AttackTechnique{ID: "T1004", Name: "Technique & Method <Test>"},
			},
			{
				name:      "technique-with-description",
				technique: AttackTechnique{ID: "T1005", Description: "Full description"},
			},
			{
				name:      "long-description",
				technique: AttackTechnique{ID: "T1006", Description: strings.Repeat("desc ", 1000)},
			},
			{
				name:      "unicode-description",
				technique: AttackTechnique{ID: "T1007", Description: "技术描述 🔒"},
			},
			{
				name:      "technique-with-domain",
				technique: AttackTechnique{ID: "T1008", Domain: "enterprise-attack"},
			},
			{
				name:      "technique-with-platform",
				technique: AttackTechnique{ID: "T1009", Platform: "Windows"},
			},
			{
				name:      "revoked-technique",
				technique: AttackTechnique{ID: "T1010", Revoked: true},
			},
			{
				name:      "deprecated-technique",
				technique: AttackTechnique{ID: "T1011", Deprecated: true},
			},
			{
				name: "all-fields",
				technique: AttackTechnique{
					ID:          "T1099",
					Name:        "Full Technique",
					Description: "Description",
					Domain:      "enterprise-attack",
					Platform:    "Windows",
					Created:     "2021-01-01T00:00:00.000Z",
					Modified:    "2021-12-31T23:59:59.000Z",
					Revoked:     false,
					Deprecated:  false,
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.technique)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackTechnique
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.technique.ID {
					t.Fatalf("ID mismatch: want %s got %s", tc.technique.ID, decoded.ID)
				}
			})
		}
	})

}

// TestAttackTactic_JSONFormats covers ATT&CK tactic JSON serialization.
func TestAttackTactic_JSONFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackTactic_JSONFormats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name   string
			tactic AttackTactic
		}{
			{
				name:   "minimal-tactic",
				tactic: AttackTactic{ID: "TA0001"},
			},
			{
				name:   "tactic-with-name",
				tactic: AttackTactic{ID: "TA0001", Name: "Initial Access"},
			},
			{
				name:   "unicode-name",
				tactic: AttackTactic{ID: "TA0002", Name: "初始访问 - начальный доступ"},
			},
			{
				name:   "tactic-with-description",
				tactic: AttackTactic{ID: "TA0003", Description: "Description text"},
			},
			{
				name:   "long-description",
				tactic: AttackTactic{ID: "TA0004", Description: strings.Repeat("tactic ", 100)},
			},
			{
				name: "all-fields",
				tactic: AttackTactic{
					ID:          "TA0099",
					Name:        "Full Tactic",
					Description: "Description",
					Domain:      "enterprise-attack",
					Created:     "2021-01-01T00:00:00.000Z",
					Modified:    "2021-12-31T23:59:59.000Z",
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.tactic)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackTactic
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.tactic.ID {
					t.Fatalf("ID mismatch: want %s got %s", tc.tactic.ID, decoded.ID)
				}
			})
		}
	})

}

// TestAttackTechnique_IDFormats validates various technique ID formats.
func TestAttackTechnique_IDFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackTechnique_IDFormats", nil, func(t *testing.T, tx *gorm.DB) {
		validIDs := []string{
			"T1001",
			"T1002",
			"T1003.001", // sub-technique
			"T1003.002",
			"T1055.001",
			"T1055.012",
			"T1548",
			"T9999",
		}

		for _, id := range validIDs {
			t.Run(id, func(t *testing.T) {
				technique := AttackTechnique{ID: id}
				data, err := json.Marshal(&technique)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackTechnique
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

// TestAttackTactic_IDFormats validates various tactic ID formats.
func TestAttackTactic_IDFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackTactic_IDFormats", nil, func(t *testing.T, tx *gorm.DB) {
		validIDs := []string{
			"TA0001",
			"TA0002",
			"TA0010",
			"TA0040",
			"TA9999",
		}

		for _, id := range validIDs {
			t.Run(id, func(t *testing.T) {
				tactic := AttackTactic{ID: id}
				data, err := json.Marshal(&tactic)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackTactic
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

// TestAttackTechnique_DomainValues validates domain field values.
func TestAttackTechnique_DomainValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackTechnique_DomainValues", nil, func(t *testing.T, tx *gorm.DB) {
		domains := []string{
			"enterprise-attack",
			"mobile-attack",
			"ics-attack",
			"",
		}

		for _, domain := range domains {
			t.Run(fmt.Sprintf("domain-%s", domain), func(t *testing.T) {
				technique := AttackTechnique{ID: "T1001", Domain: domain}
				data, err := json.Marshal(&technique)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackTechnique
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.Domain != domain {
					t.Fatalf("Domain mismatch: want %s got %s", domain, decoded.Domain)
				}
			})
		}
	})

}

// TestAttackTechnique_PlatformValues validates platform field values.
func TestAttackTechnique_PlatformValues(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackTechnique_PlatformValues", nil, func(t *testing.T, tx *gorm.DB) {
		platforms := []string{
			"Windows",
			"Linux",
			"macOS",
			"AWS",
			"Azure",
			"GCP",
			"Android",
			"iOS",
			"",
		}

		for _, platform := range platforms {
			t.Run(fmt.Sprintf("platform-%s", platform), func(t *testing.T) {
				technique := AttackTechnique{ID: "T1001", Platform: platform}
				data, err := json.Marshal(&technique)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackTechnique
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.Platform != platform {
					t.Fatalf("Platform mismatch: want %s got %s", platform, decoded.Platform)
				}
			})
		}
	})

}

// TestAttackMitigation_JSONFormats validates mitigation JSON serialization.
func TestAttackMitigation_JSONFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackMitigation_JSONFormats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name       string
			mitigation AttackMitigation
		}{
			{name: "minimal", mitigation: AttackMitigation{ID: "M1001"}},
			{name: "with-name", mitigation: AttackMitigation{ID: "M1001", Name: "Patch"}},
			{name: "with-description", mitigation: AttackMitigation{ID: "M1002", Description: "Description"}},
			{name: "all-fields", mitigation: AttackMitigation{
				ID:          "M1099",
				Name:        "Full Mitigation",
				Description: "Description",
				Domain:      "enterprise-attack",
				Created:     "2021-01-01T00:00:00.000Z",
				Modified:    "2021-12-31T23:59:59.000Z",
			}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.mitigation)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackMitigation
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.mitigation.ID {
					t.Fatalf("ID mismatch")
				}
			})
		}
	})

}

// TestAttackSoftware_JSONFormats validates software JSON serialization.
func TestAttackSoftware_JSONFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackSoftware_JSONFormats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name     string
			software AttackSoftware
		}{
			{name: "minimal", software: AttackSoftware{ID: "S0001"}},
			{name: "malware", software: AttackSoftware{ID: "S0001", Type: "malware"}},
			{name: "tool", software: AttackSoftware{ID: "S0002", Type: "tool"}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.software)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackSoftware
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.software.ID {
					t.Fatalf("ID mismatch")
				}
			})
		}
	})

}

// TestAttackGroup_JSONFormats validates group JSON serialization.
func TestAttackGroup_JSONFormats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAttackGroup_JSONFormats", nil, func(t *testing.T, tx *gorm.DB) {
		cases := []struct {
			name  string
			group AttackGroup
		}{
			{name: "minimal", group: AttackGroup{ID: "G0001"}},
			{name: "with-name", group: AttackGroup{ID: "G0001", Name: "APT28"}},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				data, err := json.Marshal(&tc.group)
				if err != nil {
					t.Fatalf("json.Marshal failed: %v", err)
				}

				var decoded AttackGroup
				if err := json.Unmarshal(data, &decoded); err != nil {
					t.Fatalf("json.Unmarshal failed: %v", err)
				}

				if decoded.ID != tc.group.ID {
					t.Fatalf("ID mismatch")
				}
			})
		}
	})

}
