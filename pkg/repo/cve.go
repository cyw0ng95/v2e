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
	ID          string       `json:"id"`
	SourceID    string       `json:"sourceIdentifier"`
	Published   time.Time    `json:"published"`
	LastModified time.Time   `json:"lastModified"`
	VulnStatus  string       `json:"vulnStatus"`
	Descriptions []Description `json:"descriptions"`
	Metrics     *Metrics     `json:"metrics,omitempty"`
	References  []Reference  `json:"references,omitempty"`
}

// Description represents a CVE description
type Description struct {
	Lang  string `json:"lang"`
	Value string `json:"value"`
}

// Metrics contains CVSS metrics for a CVE
type Metrics struct {
	CvssMetricV31 []CVSSMetricV3 `json:"cvssMetricV31,omitempty"`
	CvssMetricV30 []CVSSMetricV3 `json:"cvssMetricV30,omitempty"`
	CvssMetricV2  []CVSSMetricV2 `json:"cvssMetricV2,omitempty"`
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
