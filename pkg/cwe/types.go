package cwe

// CWEItem represents a CWE Weakness object as defined in the CWE-CAPEC OpenAPI spec.
type CWEItem struct {
	ID                    string                 `json:"ID"`
	Name                  string                 `json:"Name"`
	Diagram               string                 `json:"Diagram,omitempty"`
	Abstraction           string                 `json:"Abstraction"`
	Structure             string                 `json:"Structure"`
	Status                string                 `json:"Status"`
	Description           string                 `json:"Description"`
	ExtendedDescription   string                 `json:"ExtendedDescription,omitempty"`
	LikelihoodOfExploit   string                 `json:"LikelihoodOfExploit,omitempty"`
	RelatedWeaknesses     []RelatedWeakness      `json:"RelatedWeaknesses,omitempty"`
	WeaknessOrdinalities  []WeaknessOrdinality   `json:"WeaknessOrdinalities,omitempty"`
	ApplicablePlatforms   []ApplicablePlatform   `json:"ApplicablePlatforms,omitempty"`
	BackgroundDetails     []string               `json:"BackgroundDetails,omitempty"`
	AlternateTerms        []AlternateTerm        `json:"AlternateTerms,omitempty"`
	ModesOfIntroduction   []ModeOfIntroduction   `json:"ModesOfIntroduction,omitempty"`
	CommonConsequences    []Consequence          `json:"CommonConsequences,omitempty"`
	DetectionMethods      []DetectionMethod      `json:"DetectionMethods,omitempty"`
	PotentialMitigations  []Mitigation           `json:"PotentialMitigations,omitempty"`
	DemonstrativeExamples []DemonstrativeExample `json:"DemonstrativeExamples,omitempty"`
	ObservedExamples      []ObservedExample      `json:"ObservedExamples,omitempty"`
	FunctionalAreas       []string               `json:"FunctionalAreas,omitempty"`
	AffectedResources     []string               `json:"AffectedResources,omitempty"`
	TaxonomyMappings      []TaxonomyMapping      `json:"TaxonomyMappings,omitempty"`
	RelatedAttackPatterns []string               `json:"RelatedAttackPatterns,omitempty"`
	References            []Reference            `json:"References,omitempty"`
	MappingNotes          *MappingNotes          `json:"MappingNotes,omitempty"`
	Notes                 []Note                 `json:"Notes,omitempty"`
	ContentHistory        []ContentHistory       `json:"ContentHistory"`
}

// Subtypes for nested fields (minimal, expand as needed)
type RelatedWeakness struct {
	Nature  string `json:"Nature"`
	CweID   string `json:"CweID"`
	ViewID  string `json:"ViewID"`
	Ordinal string `json:"Ordinal,omitempty"`
}

type WeaknessOrdinality struct {
	Ordinality  string `json:"Ordinality"`
	Description string `json:"Description,omitempty"`
}

type ApplicablePlatform struct {
	Type       string `json:"Type"`
	Name       string `json:"Name,omitempty"`
	Class      string `json:"Class,omitempty"`
	Prevalence string `json:"Prevalence"`
}

type AlternateTerm struct {
	Term        string `json:"Term"`
	Description string `json:"Description,omitempty"`
}

type ModeOfIntroduction struct {
	Phase string `json:"Phase"`
	Note  string `json:"Note,omitempty"`
}

type Consequence struct {
	Scope      []string `json:"Scope"`
	Impact     []string `json:"Impact,omitempty"`
	Likelihood []string `json:"Likelihood,omitempty"`
	Note       string   `json:"Note,omitempty"`
}

type DetectionMethod struct {
	DetectionMethodID  string `json:"DetectionMethodID,omitempty"`
	Method             string `json:"Method"`
	Description        string `json:"Description"`
	Effectiveness      string `json:"Effectiveness,omitempty"`
	EffectivenessNotes string `json:"EffectivenessNotes,omitempty"`
}

type Mitigation struct {
	MitigationID       string   `json:"MitigationID,omitempty"`
	Phase              []string `json:"Phase,omitempty"`
	Strategy           string   `json:"Strategy,omitempty"`
	Description        string   `json:"Description"`
	Effectiveness      string   `json:"Effectiveness,omitempty"`
	EffectivenessNotes string   `json:"EffectivenessNotes,omitempty"`
}

type DemonstrativeExample struct {
	ID      string               `json:"ID,omitempty"`
	Entries []DemonstrativeEntry `json:"Entries"`
}

type DemonstrativeEntry struct {
	IntroText   string `json:"IntroText,omitempty"`
	BodyText    string `json:"BodyText,omitempty"`
	Nature      string `json:"Nature,omitempty"`
	Language    string `json:"Language,omitempty"`
	ExampleCode string `json:"ExampleCode,omitempty"`
	Reference   string `json:"Reference,omitempty"`
}

type ObservedExample struct {
	Reference   string `json:"Reference"`
	Description string `json:"Description"`
	Link        string `json:"Link"`
}

type TaxonomyMapping struct {
	TaxonomyName string `json:"TaxonomyName"`
	EntryName    string `json:"EntryName,omitempty"`
	EntryID      string `json:"EntryID,omitempty"`
	MappingFit   string `json:"MappingFit,omitempty"`
}

type Reference struct {
	ExternalReferenceID string   `json:"ExternalReferenceID"`
	Section             string   `json:"Section,omitempty"`
	Authors             []string `json:"Authors,omitempty"`
	Title               string   `json:"Title,omitempty"`
	PublicationYear     string   `json:"PublicationYear,omitempty"`
	PublicationMonth    string   `json:"PublicationMonth,omitempty"`
	PublicationDay      string   `json:"PublicationDay,omitempty"`
	Publisher           string   `json:"Publisher,omitempty"`
	URL                 string   `json:"URL,omitempty"`
	URLDate             string   `json:"URLDate,omitempty"`
	Edition             string   `json:"Edition,omitempty"`
	Publication         string   `json:"Publication,omitempty"`
}

type MappingNotes struct {
	Usage       string              `json:"Usage"`
	Rationale   string              `json:"Rationale"`
	Comments    string              `json:"Comments"`
	Reasons     []string            `json:"Reasons"`
	Suggestions []SuggestionComment `json:"Suggestions,omitempty"`
}

type SuggestionComment struct {
	Comment string `json:"Comment"`
	CweID   string `json:"CweID"`
}

type Note struct {
	Type string `json:"Type"`
	Note string `json:"Note"`
}

type ContentHistory struct {
	Type                     string `json:"Type"`
	SubmissionName           string `json:"SubmissionName,omitempty"`
	SubmissionOrganization   string `json:"SubmissionOrganization,omitempty"`
	SubmissionDate           string `json:"SubmissionDate,omitempty"`
	SubmissionVersion        string `json:"SubmissionVersion,omitempty"`
	SubmissionReleaseDate    string `json:"SubmissionReleaseDate,omitempty"`
	SubmissionComment        string `json:"SubmissionComment,omitempty"`
	ModificationName         string `json:"ModificationName,omitempty"`
	ModificationOrganization string `json:"ModificationOrganization,omitempty"`
	ModificationDate         string `json:"ModificationDate,omitempty"`
	ModificationVersion      string `json:"ModificationVersion,omitempty"`
	ModificationReleaseDate  string `json:"ModificationReleaseDate,omitempty"`
	ModificationComment      string `json:"ModificationComment,omitempty"`
	ContributionName         string `json:"ContributionName,omitempty"`
	ContributionOrganization string `json:"ContributionOrganization,omitempty"`
	ContributionDate         string `json:"ContributionDate,omitempty"`
	ContributionVersion      string `json:"ContributionVersion,omitempty"`
	ContributionReleaseDate  string `json:"ContributionReleaseDate,omitempty"`
	ContributionComment      string `json:"ContributionComment,omitempty"`
	ContributionType         string `json:"ContributionType,omitempty"`
	PreviousEntryName        string `json:"PreviousEntryName,omitempty"`
	Date                     string `json:"Date,omitempty"`
	Version                  string `json:"Version,omitempty"`
}
