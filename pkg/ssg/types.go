package ssg

import "encoding/xml"

// SSGDataStreamCollection represents the top-level SCAP DataStream collection
type SSGDataStreamCollection struct {
	XMLName xml.Name         `xml:"data-stream-collection"`
	Streams []SSGDataStream  `xml:"data-stream"`
}

// SSGDataStream represents a single SCAP DataStream
type SSGDataStream struct {
	ID        string           `xml:"id,attr"`
	ScapVersion string         `xml:"scap-version,attr"`
	Timestamp string           `xml:"timestamp,attr"`
	Checklists SSGChecklists  `xml:"checklists"`
}

// SSGChecklists contains references to XCCDF checklists
type SSGChecklists struct {
	ComponentRefs []SSGComponentRef `xml:"component-ref"`
}

// SSGComponentRef references embedded XCCDF components
type SSGComponentRef struct {
	ID   string `xml:"id,attr"`
	Href string `xml:"href,attr"`
}

// SSGBenchmark represents an XCCDF Benchmark
type SSGBenchmark struct {
	XMLName     xml.Name     `xml:"Benchmark"`
	ID          string       `xml:"id,attr"`
	Style       string       `xml:"style,attr"`
	Resolved    string       `xml:"resolved,attr"`
	Status      SSGStatus    `xml:"status"`
	Title       string       `xml:"title"`
	Description string       `xml:"description"`
	Version     string       `xml:"version"`
	Profiles    []SSGProfile `xml:"Profile"`
	Groups      []SSGGroup   `xml:"Group"`
	Rules       []SSGRule    `xml:"Rule"`
	Values      []SSGValue   `xml:"Value"`
}

// SSGStatus represents the status of a benchmark
type SSGStatus struct {
	Date string `xml:"date,attr"`
	Text string `xml:",chardata"`
}

// SSGProfile represents an XCCDF Profile (e.g., NIST 800-53, PCI-DSS)
type SSGProfile struct {
	ID          string            `xml:"id,attr"`
	Title       string            `xml:"title"`
	Description string            `xml:"description"`
	Selects     []SSGProfileSelect `xml:"select"`
	SetValues   []SSGSetValue     `xml:"set-value"`
	RefineValues []SSGRefineValue `xml:"refine-value"`
}

// SSGProfileSelect represents a rule selection in a profile
type SSGProfileSelect struct {
	IDRef    string `xml:"idref,attr"`
	Selected string `xml:"selected,attr"`
}

// SSGSetValue represents a value setting in a profile
type SSGSetValue struct {
	IDRef string `xml:"idref,attr"`
	Value string `xml:",chardata"`
}

// SSGRefineValue represents a value refinement in a profile
type SSGRefineValue struct {
	IDRef    string `xml:"idref,attr"`
	Selector string `xml:"selector,attr"`
}

// SSGGroup represents a group of rules in a benchmark
type SSGGroup struct {
	ID          string      `xml:"id,attr"`
	Title       string      `xml:"title"`
	Description string      `xml:"description"`
	Groups      []SSGGroup  `xml:"Group"`
	Rules       []SSGRule   `xml:"Rule"`
}

// SSGRule represents an individual security rule
type SSGRule struct {
	ID          string            `xml:"id,attr"`
	Severity    string            `xml:"severity,attr"`
	Selected    string            `xml:"selected,attr"`
	Title       string            `xml:"title"`
	Description string            `xml:"description"`
	Rationale   string            `xml:"rationale"`
	Warning     string            `xml:"warning"`
	References  []SSGReference    `xml:"reference"`
	Idents      []SSGIdent        `xml:"ident"`
	Checks      []SSGCheck        `xml:"check"`
	Fixes       []SSGFix          `xml:"fix"`
}

// SSGReference represents an external reference
type SSGReference struct {
	Href string `xml:"href,attr"`
	Text string `xml:",chardata"`
}

// SSGIdent represents an identifier (e.g., CCE)
type SSGIdent struct {
	System string `xml:"system,attr"`
	Text   string `xml:",chardata"`
}

// SSGCheck represents a compliance check
type SSGCheck struct {
	System      string `xml:"system,attr"`
	CheckContentRef SSGCheckContentRef `xml:"check-content-ref"`
}

// SSGCheckContentRef references check content
type SSGCheckContentRef struct {
	Href string `xml:"href,attr"`
	Name string `xml:"name,attr"`
}

// SSGFix represents a remediation fix
type SSGFix struct {
	System     string `xml:"system,attr"`
	Complexity string `xml:"complexity,attr"`
	Disruption string `xml:"disruption,attr"`
	ID         string `xml:"id,attr"`
	Content    string `xml:",chardata"`
}

// SSGValue represents a configurable value in a benchmark
type SSGValue struct {
	ID          string          `xml:"id,attr"`
	Type        string          `xml:"type,attr"`
	Operator    string          `xml:"operator,attr"`
	Title       string          `xml:"title"`
	Description string          `xml:"description"`
	Value       []SSGValueItem  `xml:"value"`
}

// SSGValueItem represents a value option
type SSGValueItem struct {
	Selector string `xml:"selector,attr"`
	Text     string `xml:",chardata"`
}

// SSGMetadata stores metadata about an SSG package
type SSGMetadata struct {
	Version      string `json:"version"`
	BenchmarkID  string `json:"benchmark_id"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ProfileCount int    `json:"profile_count"`
	RuleCount    int    `json:"rule_count"`
	LastUpdated  string `json:"last_updated"`
}

// SSGProfileSummary is a lightweight profile representation for listing
type SSGProfileSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	RuleCount   int    `json:"rule_count"`
}

// SSGRuleSummary is a lightweight rule representation for listing
type SSGRuleSummary struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
}
