package attack

// ATT&CK Technique represents a technique in the ATT&CK framework
type AttackTechnique struct {
	ID          string `json:"id" gorm:"primaryKey"` // e.g. "T1001"
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`     // e.g. "enterprise-attack", "mobile-attack", "ics-attack"
	Platform    string `json:"platform"`   // e.g. "Windows", "Linux", "macOS"
	Created     string `json:"created"`    // Creation date
	Modified    string `json:"modified"`   // Last modified date
	Revoked     bool   `json:"revoked"`    // Whether the technique is revoked
	Deprecated  bool   `json:"deprecated"` // Whether the technique is deprecated
}

// ATT&CK Tactic represents a tactic in the ATT&CK framework
type AttackTactic struct {
	ID          string `json:"id" gorm:"primaryKey"` // e.g. "TA0001"
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`   // e.g. "enterprise-attack", "mobile-attack", "ics-attack"
	Created     string `json:"created"`  // Creation date
	Modified    string `json:"modified"` // Last modified date
}

// ATT&CK Mitigation represents a mitigation in the ATT&CK framework
type AttackMitigation struct {
	ID          string `json:"id"` // e.g. "M1001"
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`   // e.g. "enterprise-attack", "mobile-attack", "ics-attack"
	Created     string `json:"created"`  // Creation date
	Modified    string `json:"modified"` // Last modified date
}

// ATT&CK Software represents software in the ATT&CK framework
type AttackSoftware struct {
	ID          string `json:"id"` // e.g. "S0001"
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`     // e.g. "malware", "tool"
	Domain      string `json:"domain"`   // e.g. "enterprise-attack", "mobile-attack", "ics-attack"
	Created     string `json:"created"`  // Creation date
	Modified    string `json:"modified"` // Last modified date
}

// ATT&CK Group represents an adversary group in the ATT&CK framework
type AttackGroup struct {
	ID          string `json:"id"` // e.g. "G0001"
	Name        string `json:"name"`
	Description string `json:"description"`
	Domain      string `json:"domain"`   // e.g. "enterprise-attack", "mobile-attack", "ics-attack"
	Created     string `json:"created"`  // Creation date
	Modified    string `json:"modified"` // Last modified date
}

// ATT&CK Relationship represents relationships between ATT&CK objects
type AttackRelationship struct {
	ID               string `json:"id" gorm:"primaryKey"`
	SourceRef        string `json:"source_ref"`         // ID of source object (e.g. "attack-pattern--...")
	TargetRef        string `json:"target_ref"`         // ID of target object (e.g. "course-of-action--...")
	RelationshipType string `json:"relationship_type"`  // e.g. "mitigates", "uses"
	SourceObjectType string `json:"source_object_type"` // e.g. "attack-pattern", "malware", "tool", "intrusion-set"
	TargetObjectType string `json:"target_object_type"` // e.g. "course-of-action", "attack-pattern", "malware", "tool"
	Description      string `json:"description"`
	Domain           string `json:"domain"`
	Created          string `json:"created"`
	Modified         string `json:"modified"`
}

// ATT&CK Metadata stores import metadata
type AttackMetadata struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	ImportedAt    int64  `json:"imported_at"`    // Unix timestamp
	SourceFile    string `json:"source_file"`    // Path to the imported XLSX file
	TotalRecords  int    `json:"total_records"`  // Total number of records imported
	ImportVersion string `json:"import_version"` // Version of the ATT&CK dataset
}
