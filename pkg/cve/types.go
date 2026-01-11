package cve

import (
	"strings"
	"time"
)

const (
	// NVDAPIURL is the base URL for the NVD CVE API v2.0
	NVDAPIURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"
	// nvdTimeFormat is the NVD timestamp format: "2021-12-10T10:15:09.143"
	nvdTimeFormat = "2006-01-02T15:04:05.999"
)

// NVDTime is a custom time type that handles NVD API timestamp format
type NVDTime struct {
	time.Time
}

// UnmarshalJSON implements sonic.Unmarshaler for NVDTime
func (t *NVDTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	if s == "null" || s == "" {
		t.Time = time.Time{}
		return nil
	}
	
	// Try parsing with the NVD format first
	parsed, err := time.Parse(nvdTimeFormat, s)
	if err != nil {
		// Fallback to RFC3339 format for compatibility
		parsed, err = time.Parse(time.RFC3339, s)
		if err != nil {
			return err
		}
	}
	t.Time = parsed
	return nil
}

// MarshalJSON implements sonic.Marshaler for NVDTime
func (t NVDTime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return []byte("\"" + t.Time.Format(nvdTimeFormat) + "\""), nil
}

// NewNVDTime creates a new NVDTime from a time.Time
func NewNVDTime(t time.Time) NVDTime {
	return NVDTime{Time: t}
}

// CVEResponse represents the top-level response from the NVD API
type CVEResponse struct {
	ResultsPerPage  int       `json:"resultsPerPage"`
	StartIndex      int       `json:"startIndex"`
	TotalResults    int       `json:"totalResults"`
	Format          string    `json:"format"`
	Version         string    `json:"version"`
	Timestamp       NVDTime   `json:"timestamp"`
	Vulnerabilities []struct {
		CVE CVEItem `json:"cve"`
	} `json:"vulnerabilities"`
}

// CVEItem represents a single CVE item from the NVD API
type CVEItem struct {
	ID                    string          `json:"id"`
	SourceID              string          `json:"sourceIdentifier"`
	Published             NVDTime         `json:"published"`
	LastModified          NVDTime         `json:"lastModified"`
	VulnStatus            string          `json:"vulnStatus"`
	EvaluatorComment      string          `json:"evaluatorComment,omitempty"`
	EvaluatorSolution     string          `json:"evaluatorSolution,omitempty"`
	EvaluatorImpact       string          `json:"evaluatorImpact,omitempty"`
	CisaExploitAdd        string          `json:"cisaExploitAdd,omitempty"`
	CisaActionDue         string          `json:"cisaActionDue,omitempty"`
	CisaRequiredAction    string          `json:"cisaRequiredAction,omitempty"`
	CisaVulnerabilityName string          `json:"cisaVulnerabilityName,omitempty"`
	CVETags               []CVETag        `json:"cveTags,omitempty"`
	Descriptions          []Description   `json:"descriptions"`
	Metrics               *Metrics        `json:"metrics,omitempty"`
	Weaknesses            []Weakness      `json:"weaknesses,omitempty"`
	Configurations        []Config        `json:"configurations,omitempty"`
	References            []Reference     `json:"references,omitempty"`
	VendorComments        []VendorComment `json:"vendorComments,omitempty"`
}

// Description represents a CVE description
type Description struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// CVETag represents tags associated with a CVE
type CVETag struct {
	SourceIdentifier string   `json:"sourceIdentifier"`
	Tags             []string `json:"tags,omitempty"`
}

// Weakness represents CWE (Common Weakness Enumeration) data
type Weakness struct {
	Source      string        `json:"source"`
	Type        string        `json:"type"`
	Description []Description `json:"description"`
}

// Config represents a configuration node with logical operators
type Config struct {
	Operator string `json:"operator,omitempty"`
	Negate   bool   `json:"negate,omitempty"`
	Nodes    []Node `json:"nodes"`
}

// Node represents a configuration node in an NVD applicability statement
type Node struct {
	Operator string     `json:"operator"`
	Negate   bool       `json:"negate,omitempty"`
	CPEMatch []CPEMatch `json:"cpeMatch"`
}

// CPEMatch represents CPE match string or range
type CPEMatch struct {
	Vulnerable            bool   `json:"vulnerable"`
	Criteria              string `json:"criteria"`
	MatchCriteriaID       string `json:"matchCriteriaId"`
	VersionStartExcluding string `json:"versionStartExcluding,omitempty"`
	VersionStartIncluding string `json:"versionStartIncluding,omitempty"`
	VersionEndExcluding   string `json:"versionEndExcluding,omitempty"`
	VersionEndIncluding   string `json:"versionEndIncluding,omitempty"`
}

// VendorComment represents a comment from a vendor
type VendorComment struct {
	Organization string  `json:"organization"`
	Comment      string  `json:"comment"`
	LastModified NVDTime `json:"lastModified"`
}

// Metrics contains CVSS metrics for a CVE
type Metrics struct {
	CvssMetricV40 []CVSSMetricV40 `json:"cvssMetricV40,omitempty"`
	CvssMetricV31 []CVSSMetricV3  `json:"cvssMetricV31,omitempty"`
	CvssMetricV30 []CVSSMetricV3  `json:"cvssMetricV30,omitempty"`
	CvssMetricV2  []CVSSMetricV2  `json:"cvssMetricV2,omitempty"`
}

// CVSSMetricV3 represents CVSS v3.x scoring data
type CVSSMetricV3 struct {
	Source              string     `json:"source"`
	Type                string     `json:"type"`
	CvssData            CVSSDataV3 `json:"cvssData"`
	ExploitabilityScore float64    `json:"exploitabilityScore,omitempty"`
	ImpactScore         float64    `json:"impactScore,omitempty"`
}

// CVSSDataV3 contains the actual CVSS v3.x score information
type CVSSDataV3 struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
	BaseSeverity string  `json:"baseSeverity"`

	// Base Metric Group
	AttackVector          string `json:"attackVector,omitempty"`
	AttackComplexity      string `json:"attackComplexity,omitempty"`
	PrivilegesRequired    string `json:"privilegesRequired,omitempty"`
	UserInteraction       string `json:"userInteraction,omitempty"`
	Scope                 string `json:"scope,omitempty"`
	ConfidentialityImpact string `json:"confidentialityImpact,omitempty"`
	IntegrityImpact       string `json:"integrityImpact,omitempty"`
	AvailabilityImpact    string `json:"availabilityImpact,omitempty"`

	// Temporal Metric Group
	TemporalScore       float64 `json:"temporalScore,omitempty"`
	TemporalSeverity    string  `json:"temporalSeverity,omitempty"`
	ExploitCodeMaturity string  `json:"exploitCodeMaturity,omitempty"`
	RemediationLevel    string  `json:"remediationLevel,omitempty"`
	ReportConfidence    string  `json:"reportConfidence,omitempty"`

	// Environmental Metric Group
	EnvironmentalScore            float64 `json:"environmentalScore,omitempty"`
	EnvironmentalSeverity         string  `json:"environmentalSeverity,omitempty"`
	ConfidentialityRequirement    string  `json:"confidentialityRequirement,omitempty"`
	IntegrityRequirement          string  `json:"integrityRequirement,omitempty"`
	AvailabilityRequirement       string  `json:"availabilityRequirement,omitempty"`
	ModifiedAttackVector          string  `json:"modifiedAttackVector,omitempty"`
	ModifiedAttackComplexity      string  `json:"modifiedAttackComplexity,omitempty"`
	ModifiedPrivilegesRequired    string  `json:"modifiedPrivilegesRequired,omitempty"`
	ModifiedUserInteraction       string  `json:"modifiedUserInteraction,omitempty"`
	ModifiedScope                 string  `json:"modifiedScope,omitempty"`
	ModifiedConfidentialityImpact string  `json:"modifiedConfidentialityImpact,omitempty"`
	ModifiedIntegrityImpact       string  `json:"modifiedIntegrityImpact,omitempty"`
	ModifiedAvailabilityImpact    string  `json:"modifiedAvailabilityImpact,omitempty"`
}

// CVSSMetricV2 represents CVSS v2.0 scoring data
type CVSSMetricV2 struct {
	Source                  string     `json:"source"`
	Type                    string     `json:"type"`
	CvssData                CVSSDataV2 `json:"cvssData"`
	BaseSeverity            string     `json:"baseSeverity,omitempty"`
	ExploitabilityScore     float64    `json:"exploitabilityScore,omitempty"`
	ImpactScore             float64    `json:"impactScore,omitempty"`
	AcInsufInfo             bool       `json:"acInsufInfo,omitempty"`
	ObtainAllPrivilege      bool       `json:"obtainAllPrivilege,omitempty"`
	ObtainUserPrivilege     bool       `json:"obtainUserPrivilege,omitempty"`
	ObtainOtherPrivilege    bool       `json:"obtainOtherPrivilege,omitempty"`
	UserInteractionRequired bool       `json:"userInteractionRequired,omitempty"`
}

// CVSSMetricV40 represents CVSS v4.0 scoring data
type CVSSMetricV40 struct {
	Source   string      `json:"source"`
	Type     string      `json:"type"`
	CvssData CVSSDataV40 `json:"cvssData"`
}

// CVSSDataV40 contains the actual CVSS v4.0 score information
type CVSSDataV40 struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`
	BaseSeverity string  `json:"baseSeverity"`

	// Attack Vector (AV)
	AttackVector string `json:"attackVector,omitempty"`

	// Attack Complexity (AC)
	AttackComplexity string `json:"attackComplexity,omitempty"`

	// Attack Requirements (AT)
	AttackRequirements string `json:"attackRequirements,omitempty"`

	// Privileges Required (PR)
	PrivilegesRequired string `json:"privilegesRequired,omitempty"`

	// User Interaction (UI)
	UserInteraction string `json:"userInteraction,omitempty"`

	// Vulnerable System Impact Metrics
	VulnConfidentialityImpact string `json:"vulnConfidentialityImpact,omitempty"`
	VulnIntegrityImpact       string `json:"vulnIntegrityImpact,omitempty"`
	VulnAvailabilityImpact    string `json:"vulnAvailabilityImpact,omitempty"`

	// Subsequent System Impact Metrics
	SubConfidentialityImpact string `json:"subConfidentialityImpact,omitempty"`
	SubIntegrityImpact       string `json:"subIntegrityImpact,omitempty"`
	SubAvailabilityImpact    string `json:"subAvailabilityImpact,omitempty"`

	// Exploit Maturity (E)
	ExploitMaturity string `json:"exploitMaturity,omitempty"`

	// Confidentiality Requirement (CR)
	ConfidentialityRequirement string `json:"confidentialityRequirement,omitempty"`

	// Integrity Requirement (IR)
	IntegrityRequirement string `json:"integrityRequirement,omitempty"`

	// Availability Requirement (AR)
	AvailabilityRequirement string `json:"availabilityRequirement,omitempty"`

	// Modified Attack Vector (MAV)
	ModifiedAttackVector string `json:"modifiedAttackVector,omitempty"`

	// Modified Attack Complexity (MAC)
	ModifiedAttackComplexity string `json:"modifiedAttackComplexity,omitempty"`

	// Modified Attack Requirements (MAT)
	ModifiedAttackRequirements string `json:"modifiedAttackRequirements,omitempty"`

	// Modified Privileges Required (MPR)
	ModifiedPrivilegesRequired string `json:"modifiedPrivilegesRequired,omitempty"`

	// Modified User Interaction (MUI)
	ModifiedUserInteraction string `json:"modifiedUserInteraction,omitempty"`

	// Modified Vulnerable System Impact Metrics
	ModifiedVulnConfidentialityImpact string `json:"modifiedVulnConfidentialityImpact,omitempty"`
	ModifiedVulnIntegrityImpact       string `json:"modifiedVulnIntegrityImpact,omitempty"`
	ModifiedVulnAvailabilityImpact    string `json:"modifiedVulnAvailabilityImpact,omitempty"`

	// Modified Subsequent System Impact Metrics
	ModifiedSubConfidentialityImpact string `json:"modifiedSubConfidentialityImpact,omitempty"`
	ModifiedSubIntegrityImpact       string `json:"modifiedSubIntegrityImpact,omitempty"`
	ModifiedSubAvailabilityImpact    string `json:"modifiedSubAvailabilityImpact,omitempty"`

	// Safety (S)
	Safety string `json:"safety,omitempty"`

	// Automatable (AU)
	Automatable string `json:"automatable,omitempty"`

	// Recovery (R)
	Recovery string `json:"recovery,omitempty"`

	// Value Density (V)
	ValueDensity string `json:"valueDensity,omitempty"`

	// Vulnerability Response Effort (RE)
	VulnerabilityResponseEffort string `json:"vulnerabilityResponseEffort,omitempty"`

	// Provider Urgency (U)
	ProviderUrgency string `json:"providerUrgency,omitempty"`

	// Modified Safety (MS)
	ModifiedSafety string `json:"modifiedSafety,omitempty"`

	// Modified Automatable (MAU)
	ModifiedAutomatable string `json:"modifiedAutomatable,omitempty"`

	// Modified Recovery (MR)
	ModifiedRecovery string `json:"modifiedRecovery,omitempty"`

	// Modified Value Density (MV)
	ModifiedValueDensity string `json:"modifiedValueDensity,omitempty"`

	// Modified Vulnerability Response Effort (MRE)
	ModifiedVulnerabilityResponseEffort string `json:"modifiedVulnerabilityResponseEffort,omitempty"`

	// Modified Provider Urgency (MU)
	ModifiedProviderUrgency string `json:"modifiedProviderUrgency,omitempty"`
}

// CVSSDataV2 contains the actual CVSS v2.0 score information
type CVSSDataV2 struct {
	Version      string  `json:"version"`
	VectorString string  `json:"vectorString"`
	BaseScore    float64 `json:"baseScore"`

	// Base Metric Group
	AccessVector          string `json:"accessVector,omitempty"`
	AccessComplexity      string `json:"accessComplexity,omitempty"`
	Authentication        string `json:"authentication,omitempty"`
	ConfidentialityImpact string `json:"confidentialityImpact,omitempty"`
	IntegrityImpact       string `json:"integrityImpact,omitempty"`
	AvailabilityImpact    string `json:"availabilityImpact,omitempty"`

	// Temporal Metric Group
	TemporalScore    float64 `json:"temporalScore,omitempty"`
	Exploitability   string  `json:"exploitability,omitempty"`
	RemediationLevel string  `json:"remediationLevel,omitempty"`
	ReportConfidence string  `json:"reportConfidence,omitempty"`

	// Environmental Metric Group
	EnvironmentalScore         float64 `json:"environmentalScore,omitempty"`
	CollateralDamagePotential  string  `json:"collateralDamagePotential,omitempty"`
	TargetDistribution         string  `json:"targetDistribution,omitempty"`
	ConfidentialityRequirement string  `json:"confidentialityRequirement,omitempty"`
	IntegrityRequirement       string  `json:"integrityRequirement,omitempty"`
	AvailabilityRequirement    string  `json:"availabilityRequirement,omitempty"`
}

// Reference represents a reference link for a CVE
type Reference struct {
	URL    string   `json:"url"`
	Source string   `json:"source,omitempty"`
	Tags   []string `json:"tags,omitempty"`
}
