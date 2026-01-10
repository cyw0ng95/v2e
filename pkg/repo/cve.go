package repo

import (
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	// NVDAPIURL is the base URL for the NVD CVE API v2.0
	NVDAPIURL = "https://services.nvd.nist.gov/rest/json/cves/2.0"
)

// CVEResponse represents the top-level response from the NVD API
type CVEResponse struct {
	ResultsPerPage  int       `json:"resultsPerPage"`
	StartIndex      int       `json:"startIndex"`
	TotalResults    int       `json:"totalResults"`
	Format          string    `json:"format"`
	Version         string    `json:"version"`
	Timestamp       time.Time `json:"timestamp"`
	Vulnerabilities []struct {
		CVE CVEItem `json:"cve"`
	} `json:"vulnerabilities"`
}

// CVEItem represents a single CVE item from the NVD API
type CVEItem struct {
	ID                   string          `json:"id"`
	SourceID             string          `json:"sourceIdentifier"`
	Published            time.Time       `json:"published"`
	LastModified         time.Time       `json:"lastModified"`
	VulnStatus           string          `json:"vulnStatus"`
	EvaluatorComment     string          `json:"evaluatorComment,omitempty"`
	EvaluatorSolution    string          `json:"evaluatorSolution,omitempty"`
	EvaluatorImpact      string          `json:"evaluatorImpact,omitempty"`
	CisaExploitAdd       string          `json:"cisaExploitAdd,omitempty"`
	CisaActionDue        string          `json:"cisaActionDue,omitempty"`
	CisaRequiredAction   string          `json:"cisaRequiredAction,omitempty"`
	CisaVulnerabilityName string         `json:"cisaVulnerabilityName,omitempty"`
	CVETags              []CVETag        `json:"cveTags,omitempty"`
	Descriptions         []Description   `json:"descriptions"`
	Metrics              *Metrics        `json:"metrics,omitempty"`
	Weaknesses           []Weakness      `json:"weaknesses,omitempty"`
	Configurations       []Config        `json:"configurations,omitempty"`
	References           []Reference     `json:"references,omitempty"`
	VendorComments       []VendorComment `json:"vendorComments,omitempty"`
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
	Organization string    `json:"organization"`
	Comment      string    `json:"comment"`
	LastModified time.Time `json:"lastModified"`
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
	TemporalScore        float64 `json:"temporalScore,omitempty"`
	TemporalSeverity     string  `json:"temporalSeverity,omitempty"`
	ExploitCodeMaturity  string  `json:"exploitCodeMaturity,omitempty"`
	RemediationLevel     string  `json:"remediationLevel,omitempty"`
	ReportConfidence     string  `json:"reportConfidence,omitempty"`
	
	// Environmental Metric Group
	EnvironmentalScore         float64 `json:"environmentalScore,omitempty"`
	EnvironmentalSeverity      string  `json:"environmentalSeverity,omitempty"`
	ConfidentialityRequirement string  `json:"confidentialityRequirement,omitempty"`
	IntegrityRequirement       string  `json:"integrityRequirement,omitempty"`
	AvailabilityRequirement    string  `json:"availabilityRequirement,omitempty"`
	ModifiedAttackVector       string  `json:"modifiedAttackVector,omitempty"`
	ModifiedAttackComplexity   string  `json:"modifiedAttackComplexity,omitempty"`
	ModifiedPrivilegesRequired string  `json:"modifiedPrivilegesRequired,omitempty"`
	ModifiedUserInteraction    string  `json:"modifiedUserInteraction,omitempty"`
	ModifiedScope              string  `json:"modifiedScope,omitempty"`
	ModifiedConfidentialityImpact string `json:"modifiedConfidentialityImpact,omitempty"`
	ModifiedIntegrityImpact       string `json:"modifiedIntegrityImpact,omitempty"`
	ModifiedAvailabilityImpact    string `json:"modifiedAvailabilityImpact,omitempty"`
}

// CVSSMetricV2 represents CVSS v2.0 scoring data
type CVSSMetricV2 struct {
	Source                   string     `json:"source"`
	Type                     string     `json:"type"`
	CvssData                 CVSSDataV2 `json:"cvssData"`
	BaseSeverity             string     `json:"baseSeverity,omitempty"`
	ExploitabilityScore      float64    `json:"exploitabilityScore,omitempty"`
	ImpactScore              float64    `json:"impactScore,omitempty"`
	AcInsufInfo              bool       `json:"acInsufInfo,omitempty"`
	ObtainAllPrivilege       bool       `json:"obtainAllPrivilege,omitempty"`
	ObtainUserPrivilege      bool       `json:"obtainUserPrivilege,omitempty"`
	ObtainOtherPrivilege     bool       `json:"obtainOtherPrivilege,omitempty"`
	UserInteractionRequired  bool       `json:"userInteractionRequired,omitempty"`
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

// CVEFetcher handles fetching CVE data from the NVD API
type CVEFetcher struct {
	client  *resty.Client
	baseURL string
	apiKey  string
}

// NewCVEFetcher creates a new CVE fetcher
func NewCVEFetcher(apiKey string) *CVEFetcher {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	
	return &CVEFetcher{
		client:  client,
		baseURL: NVDAPIURL,
		apiKey:  apiKey,
	}
}

// FetchCVEByID fetches a specific CVE by its ID
func (f *CVEFetcher) FetchCVEByID(cveID string) (*CVEResponse, error) {
	if cveID == "" {
		return nil, fmt.Errorf("CVE ID cannot be empty")
	}

	req := f.client.R().
		SetResult(&CVEResponse{}).
		SetError(&map[string]interface{}{})

	// Add API key if provided
	if f.apiKey != "" {
		req.SetHeader("apiKey", f.apiKey)
	}

	resp, err := req.Get(f.baseURL + "?cveId=" + cveID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CVE: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	result, ok := resp.Result().(*CVEResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse CVE response")
	}

	return result, nil
}

// FetchCVEs fetches CVEs with optional filters
func (f *CVEFetcher) FetchCVEs(startIndex, resultsPerPage int) (*CVEResponse, error) {
	if startIndex < 0 {
		return nil, fmt.Errorf("startIndex must be non-negative")
	}
	if resultsPerPage < 1 || resultsPerPage > 2000 {
		return nil, fmt.Errorf("resultsPerPage must be between 1 and 2000")
	}

	req := f.client.R().
		SetResult(&CVEResponse{}).
		SetError(&map[string]interface{}{}).
		SetQueryParam("startIndex", fmt.Sprintf("%d", startIndex)).
		SetQueryParam("resultsPerPage", fmt.Sprintf("%d", resultsPerPage))

	// Add API key if provided
	if f.apiKey != "" {
		req.SetHeader("apiKey", f.apiKey)
	}

	resp, err := req.Get(f.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch CVEs: %w", err)
	}

	if resp.IsError() {
		return nil, fmt.Errorf("API returned error status: %d", resp.StatusCode())
	}

	result, ok := resp.Result().(*CVEResponse)
	if !ok {
		return nil, fmt.Errorf("failed to parse CVE response")
	}

	return result, nil
}
