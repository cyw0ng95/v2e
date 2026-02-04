package asvs

// ASVSRequirementModel is the GORM model for ASVS requirements
type ASVSRequirementModel struct {
	RequirementID string `gorm:"primaryKey;column:requirement_id"`
	Chapter       string `gorm:"column:chapter;index"`
	Section       string `gorm:"column:section"`
	Description   string `gorm:"column:description"`
	Level1        bool   `gorm:"column:level1;index"`
	Level2        bool   `gorm:"column:level2;index"`
	Level3        bool   `gorm:"column:level3;index"`
	CWE           string `gorm:"column:cwe;index"`
}

// TableName specifies the table name for ASVSRequirementModel
func (ASVSRequirementModel) TableName() string {
	return "asvs_requirements"
}
