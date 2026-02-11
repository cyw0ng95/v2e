package attack

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"

	"github.com/stretchr/testify/assert"
)

func TestAttackTypes_Structs(t *testing.T) {
	// Test AttackTechnique struct
	testutils.Run(t, testutils.Level1, "AttackTechnique", nil, func(t *testing.T, tx *gorm.DB) {
		tech := AttackTechnique{
			ID:          "T1001",
			Name:        "Test Technique",
			Description: "Test Description",
			Domain:      "enterprise-attack",
			Platform:    "Windows",
			Created:     "2023-01-01",
			Modified:    "2023-01-02",
			Revoked:     false,
			Deprecated:  true,
		}

		assert.Equal(t, "T1001", tech.ID)
		assert.Equal(t, "Test Technique", tech.Name)
		assert.Equal(t, "Test Description", tech.Description)
		assert.Equal(t, "enterprise-attack", tech.Domain)
		assert.Equal(t, "Windows", tech.Platform)
		assert.Equal(t, "2023-01-01", tech.Created)
		assert.Equal(t, "2023-01-02", tech.Modified)
		assert.False(t, tech.Revoked)
		assert.True(t, tech.Deprecated)
	})

	// Test AttackTactic struct
	testutils.Run(t, testutils.Level1, "AttackTactic", nil, func(t *testing.T, tx *gorm.DB) {
		tactic := AttackTactic{
			ID:          "TA0001",
			Name:        "Test Tactic",
			Description: "Test Description",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-01-02",
		}

		assert.Equal(t, "TA0001", tactic.ID)
		assert.Equal(t, "Test Tactic", tactic.Name)
		assert.Equal(t, "Test Description", tactic.Description)
		assert.Equal(t, "enterprise-attack", tactic.Domain)
		assert.Equal(t, "2023-01-01", tactic.Created)
		assert.Equal(t, "2023-01-02", tactic.Modified)
	})

	// Test AttackMitigation struct
	testutils.Run(t, testutils.Level1, "AttackMitigation", nil, func(t *testing.T, tx *gorm.DB) {
		mitigation := AttackMitigation{
			ID:          "M1001",
			Name:        "Test Mitigation",
			Description: "Test Description",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-01-02",
		}

		assert.Equal(t, "M1001", mitigation.ID)
		assert.Equal(t, "Test Mitigation", mitigation.Name)
		assert.Equal(t, "Test Description", mitigation.Description)
		assert.Equal(t, "enterprise-attack", mitigation.Domain)
		assert.Equal(t, "2023-01-01", mitigation.Created)
		assert.Equal(t, "2023-01-02", mitigation.Modified)
	})

	// Test AttackSoftware struct
	testutils.Run(t, testutils.Level1, "AttackSoftware", nil, func(t *testing.T, tx *gorm.DB) {
		software := AttackSoftware{
			ID:          "S0001",
			Name:        "Test Software",
			Description: "Test Description",
			Type:        "malware",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-01-02",
		}

		assert.Equal(t, "S0001", software.ID)
		assert.Equal(t, "Test Software", software.Name)
		assert.Equal(t, "Test Description", software.Description)
		assert.Equal(t, "malware", software.Type)
		assert.Equal(t, "enterprise-attack", software.Domain)
		assert.Equal(t, "2023-01-01", software.Created)
		assert.Equal(t, "2023-01-02", software.Modified)
	})

	// Test AttackGroup struct
	testutils.Run(t, testutils.Level1, "AttackGroup", nil, func(t *testing.T, tx *gorm.DB) {
		group := AttackGroup{
			ID:          "G0001",
			Name:        "Test Group",
			Description: "Test Description",
			Domain:      "enterprise-attack",
			Created:     "2023-01-01",
			Modified:    "2023-01-02",
		}

		assert.Equal(t, "G0001", group.ID)
		assert.Equal(t, "Test Group", group.Name)
		assert.Equal(t, "Test Description", group.Description)
		assert.Equal(t, "enterprise-attack", group.Domain)
		assert.Equal(t, "2023-01-01", group.Created)
		assert.Equal(t, "2023-01-02", group.Modified)
	})

	// Test AttackRelationship struct
	testutils.Run(t, testutils.Level1, "AttackRelationship", nil, func(t *testing.T, tx *gorm.DB) {
		relationship := AttackRelationship{
			ID:               "rel-1",
			SourceRef:        "T1001",
			TargetRef:        "TA0001",
			RelationshipType: "mitigates",
			SourceObjectType: "attack-pattern",
			TargetObjectType: "tactic",
			Description:      "Test relationship",
			Domain:           "enterprise-attack",
			Created:          "2023-01-01",
			Modified:         "2023-01-02",
		}

		assert.Equal(t, "rel-1", relationship.ID)
		assert.Equal(t, "T1001", relationship.SourceRef)
		assert.Equal(t, "TA0001", relationship.TargetRef)
		assert.Equal(t, "mitigates", relationship.RelationshipType)
		assert.Equal(t, "attack-pattern", relationship.SourceObjectType)
		assert.Equal(t, "tactic", relationship.TargetObjectType)
		assert.Equal(t, "Test relationship", relationship.Description)
		assert.Equal(t, "enterprise-attack", relationship.Domain)
		assert.Equal(t, "2023-01-01", relationship.Created)
		assert.Equal(t, "2023-01-02", relationship.Modified)
	})

	// Test AttackMetadata struct
	testutils.Run(t, testutils.Level1, "AttackMetadata", nil, func(t *testing.T, tx *gorm.DB) {
		metadata := AttackMetadata{
			ID:            1,
			ImportedAt:    1234567890,
			SourceFile:    "/path/to/file.xlsx",
			TotalRecords:  100,
			ImportVersion: "1.0",
		}

		assert.Equal(t, uint(1), metadata.ID)
		assert.Equal(t, int64(1234567890), metadata.ImportedAt)
		assert.Equal(t, "/path/to/file.xlsx", metadata.SourceFile)
		assert.Equal(t, 100, metadata.TotalRecords)
		assert.Equal(t, "1.0", metadata.ImportVersion)
	})
}
