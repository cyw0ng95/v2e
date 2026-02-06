package capec

// Domain types for CAPEC XML parsing
// These structs are minimal and map the commonly-needed fields.
type CAPECAttackPattern struct {
	ID              int      `xml:"ID,attr" json:"id"`
	Name            string   `xml:"Name,attr" json:"name"`
	Abstraction     string   `xml:"Abstraction,attr" json:"abstraction"`
	Status          string   `xml:"Status,attr" json:"status"`
	Summary         string   `xml:"Summary" json:"summary"`
	Description     InnerXML `xml:"Description" json:"description"`
	Likelihood      string   `xml:"Likelihood_Of_Attack" json:"likelihood"`
	TypicalSeverity string   `xml:"Typical_Severity" json:"typical_severity"`
	// Related weaknesses are parsed from attributes
	RelatedWeaknesses []RelatedWeakness `xml:"Related_Weaknesses>Related_Weakness" json:"related_weaknesses"`
	Examples          []InnerXML        `xml:"Example_Instances>Example" json:"examples"`
	Mitigations       []InnerXML        `xml:"Mitigations>Mitigation" json:"mitigations"`
	References        []RelatedRef      `xml:"References>Reference" json:"references"`
}

type RelatedWeakness struct {
	CWEID string `xml:"CWE_ID,attr" json:"cwe_id"`
}

type InnerXML struct {
	XML string `xml:",innerxml"`
}

type RelatedRef struct {
	ExternalRef string `xml:"External_Reference_ID,attr" json:"external_reference_id"`
}
