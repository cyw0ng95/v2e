package asvs

// ASVSRequirement represents an OWASP ASVS requirement
type ASVSRequirement struct {
	RequirementID string `json:"RequirementID"`
	Chapter       string `json:"Chapter"`
	Section       string `json:"Section"`
	Description   string `json:"Description"`
	Level1        bool   `json:"Level1"`
	Level2        bool   `json:"Level2"`
	Level3        bool   `json:"Level3"`
	CWE           string `json:"CWE,omitempty"`
}
