package capec

// Shared GORM model definitions for CAPEC
type CAPECItemModel struct {
	CAPECID         int `gorm:"primaryKey"`
	Name            string
	Summary         string `gorm:"type:text"`
	Description     string `gorm:"type:text"`
	Status          string
	Abstraction     string
	Likelihood      string
	TypicalSeverity string
}

type CAPECRelatedWeaknessModel struct {
	ID      uint   `gorm:"primaryKey"`
	CAPECID int    `gorm:"index"`
	CWEID   string `gorm:"index" gorm:"uniqueIndex:ux_capec_cwe,priority:2"`
	// Composite unique index to avoid duplicate (capec_id, cweid)
	// CAPECID field participates in the composite index as priority 1
}

type CAPECExampleModel struct {
	ID          uint   `gorm:"primaryKey"`
	CAPECID     int    `gorm:"index" gorm:"uniqueIndex:ux_capec_example,priority:1"`
	ExampleText string `gorm:"type:text" gorm:"uniqueIndex:ux_capec_example,priority:2"`
}

type CAPECMitigationModel struct {
	ID             uint   `gorm:"primaryKey"`
	CAPECID        int    `gorm:"index" gorm:"uniqueIndex:ux_capec_mitigation,priority:1"`
	MitigationText string `gorm:"type:text" gorm:"uniqueIndex:ux_capec_mitigation,priority:2"`
}

type CAPECReferenceModel struct {
	ID                uint   `gorm:"primaryKey"`
	CAPECID           int    `gorm:"index" gorm:"uniqueIndex:ux_capec_reference,priority:1"`
	ExternalReference string `gorm:"index" gorm:"uniqueIndex:ux_capec_reference,priority:2"`
	URL               string
}

// CAPECCatalogMeta stores metadata about the imported CAPEC catalog
type CAPECCatalogMeta struct {
	ID            uint   `gorm:"primaryKey"`
	Version       string `gorm:"index"`
	Source        string
	ImportedAtUTC int64
}
