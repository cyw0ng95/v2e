package attack

import (
	"context"
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestGetTechniqueByID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetTechniqueByID_Existing", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_technique.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		technique := AttackTechnique{ID: "T1001", Name: "Test Technique"}
		if err := store.db.Create(&technique).Error; err != nil {
			t.Fatalf("Failed to create technique: %v", err)
		}

		result, err := store.GetTechniqueByID(ctx, "T1001")
		if err != nil {
			t.Fatalf("Failed to get technique: %v", err)
		}

		if result.ID != "T1001" {
			t.Errorf("Expected ID T1001, got %s", result.ID)
		}
	})

	testutils.Run(t, testutils.Level2, "GetTechniqueByID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_technique_notfound.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		_, err = store.GetTechniqueByID(ctx, "T9999")
		if err == nil {
			t.Error("Expected error for non-existent technique")
		}
	})
}

func TestGetTacticByID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetTacticByID_Existing", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_tactic.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		tactic := AttackTactic{ID: "TA0001", Name: "Initial Access"}
		if err := store.db.Create(&tactic).Error; err != nil {
			t.Fatalf("Failed to create tactic: %v", err)
		}

		result, err := store.GetTacticByID(ctx, "TA0001")
		if err != nil {
			t.Fatalf("Failed to get tactic: %v", err)
		}

		if result.ID != "TA0001" {
			t.Errorf("Expected ID TA0001, got %s", result.ID)
		}
	})

	testutils.Run(t, testutils.Level2, "GetTacticByID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_tactic_notfound.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		_, err = store.GetTacticByID(ctx, "TA9999")
		if err == nil {
			t.Error("Expected error for non-existent tactic")
		}
	})
}

func TestGetMitigationByID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetMitigationByID_Existing", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_mitigation.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		mitigation := AttackMitigation{ID: "M1001", Name: "Test Mitigation"}
		if err := store.db.Create(&mitigation).Error; err != nil {
			t.Fatalf("Failed to create mitigation: %v", err)
		}

		result, err := store.GetMitigationByID(ctx, "M1001")
		if err != nil {
			t.Fatalf("Failed to get mitigation: %v", err)
		}

		if result.ID != "M1001" {
			t.Errorf("Expected ID M1001, got %s", result.ID)
		}
	})

	testutils.Run(t, testutils.Level2, "GetMitigationByID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_mitigation_notfound.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		_, err = store.GetMitigationByID(ctx, "M9999")
		if err == nil {
			t.Error("Expected error for non-existent mitigation")
		}
	})
}

func TestGetSoftwareByID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetSoftwareByID_Existing", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_software.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		software := AttackSoftware{ID: "S0001", Name: "Test Software"}
		if err := store.db.Create(&software).Error; err != nil {
			t.Fatalf("Failed to create software: %v", err)
		}

		result, err := store.GetSoftwareByID(ctx, "S0001")
		if err != nil {
			t.Fatalf("Failed to get software: %v", err)
		}

		if result.ID != "S0001" {
			t.Errorf("Expected ID S0001, got %s", result.ID)
		}
	})

	testutils.Run(t, testutils.Level2, "GetSoftwareByID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_software_notfound.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		_, err = store.GetSoftwareByID(ctx, "S9999")
		if err == nil {
			t.Error("Expected error for non-existent software")
		}
	})
}

func TestGetGroupByID(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetGroupByID_Existing", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_group.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		group := AttackGroup{ID: "G0001", Name: "Test Group"}
		if err := store.db.Create(&group).Error; err != nil {
			t.Fatalf("Failed to create group: %v", err)
		}

		result, err := store.GetGroupByID(ctx, "G0001")
		if err != nil {
			t.Fatalf("Failed to get group: %v", err)
		}

		if result.ID != "G0001" {
			t.Errorf("Expected ID G0001, got %s", result.ID)
		}
	})

	testutils.Run(t, testutils.Level2, "GetGroupByID_NotFound", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_get_group_notfound.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		_, err = store.GetGroupByID(ctx, "G9999")
		if err == nil {
			t.Error("Expected error for non-existent group")
		}
	})
}

func TestListTechniquesPaginated(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ListTechniques_Empty", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_techniques_empty.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		techniques, total, err := store.ListTechniquesPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list techniques: %v", err)
		}

		if total != 0 {
			t.Errorf("Expected 0 techniques, got %d", total)
		}

		if len(techniques) != 0 {
			t.Errorf("Expected empty list, got %d items", len(techniques))
		}
	})

	testutils.Run(t, testutils.Level2, "ListTechniques_WithPagination", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_techniques_pagination.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()

		techniques := []AttackTechnique{
			{ID: "T1001", Name: "Technique 1"},
			{ID: "T1002", Name: "Technique 2"},
			{ID: "T1003", Name: "Technique 3"},
		}

		for _, technique := range techniques {
			if err := store.db.Create(&technique).Error; err != nil {
				t.Fatalf("Failed to create technique: %v", err)
			}
		}

		result, total, err := store.ListTechniquesPaginated(ctx, 0, 2)
		if err != nil {
			t.Fatalf("Failed to list techniques: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected 3 total techniques, got %d", total)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 techniques in page, got %d", len(result))
		}
	})
}

func TestListSoftwarePaginated(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ListSoftware_Empty", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_software_empty.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		software, total, err := store.ListSoftwarePaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list software: %v", err)
		}

		if total != 0 {
			t.Errorf("Expected 0 software, got %d", total)
		}

		if len(software) != 0 {
			t.Errorf("Expected empty list, got %d items", len(software))
		}
	})

	testutils.Run(t, testutils.Level2, "ListSoftware_WithPagination", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_software_pagination.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()

		software := []AttackSoftware{
			{ID: "S0001", Name: "Software 1"},
			{ID: "S0002", Name: "Software 2"},
			{ID: "S0003", Name: "Software 3"},
		}

		for _, s := range software {
			if err := store.db.Create(&s).Error; err != nil {
				t.Fatalf("Failed to create software: %v", err)
			}
		}

		result, total, err := store.ListSoftwarePaginated(ctx, 0, 2)
		if err != nil {
			t.Fatalf("Failed to list software: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected 3 total software, got %d", total)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 software in page, got %d", len(result))
		}
	})
}

func TestListGroupsPaginated(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ListGroups_Empty", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_groups_empty.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		groups, total, err := store.ListGroupsPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list groups: %v", err)
		}

		if total != 0 {
			t.Errorf("Expected 0 groups, got %d", total)
		}

		if len(groups) != 0 {
			t.Errorf("Expected empty list, got %d items", len(groups))
		}
	})

	testutils.Run(t, testutils.Level2, "ListGroups_WithPagination", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_groups_pagination.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()

		groups := []AttackGroup{
			{ID: "G0001", Name: "Group 1"},
			{ID: "G0002", Name: "Group 2"},
			{ID: "G0003", Name: "Group 3"},
		}

		for _, g := range groups {
			if err := store.db.Create(&g).Error; err != nil {
				t.Fatalf("Failed to create group: %v", err)
			}
		}

		result, total, err := store.ListGroupsPaginated(ctx, 0, 2)
		if err != nil {
			t.Fatalf("Failed to list groups: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected 3 total groups, got %d", total)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 groups in page, got %d", len(result))
		}
	})
}

func TestGetRelatedTechniquesByTactic(t *testing.T) {
	testutils.Run(t, testutils.Level2, "GetRelatedTechniquesByTactic_NoMatches", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_related_techniques_none.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		techniques, err := store.GetRelatedTechniquesByTactic(ctx, "TA0001")
		if err != nil {
			t.Fatalf("Failed to get related techniques: %v", err)
		}

		if len(techniques) != 0 {
			t.Errorf("Expected 0 techniques, got %d", len(techniques))
		}
	})

	testutils.Run(t, testutils.Level2, "GetRelatedTechniquesByTactic_WithMatches", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_related_techniques_matches.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()

		techniques := []AttackTechnique{
			{ID: "T1001", Name: "Technique 1"},
			{ID: "T1002", Name: "Technique 2"},
			{ID: "T1003", Name: "Technique 3"},
		}

		for _, technique := range techniques {
			if err := store.db.Create(&technique).Error; err != nil {
				t.Fatalf("Failed to create technique: %v", err)
			}
		}

		relationships := []AttackRelationship{
			{ID: "rel1", SourceRef: "T1001", TargetRef: "TA0001", RelationshipType: "mitigates"},
			{ID: "rel2", SourceRef: "T1002", TargetRef: "TA0001", RelationshipType: "mitigates"},
		}

		for _, rel := range relationships {
			if err := store.db.Create(&rel).Error; err != nil {
				t.Fatalf("Failed to create relationship: %v", err)
			}
		}

		relatedTechniques, err := store.GetRelatedTechniquesByTactic(ctx, "TA0001")
		if err != nil {
			t.Fatalf("Failed to get related techniques: %v", err)
		}

		if len(relatedTechniques) != 2 {
			t.Errorf("Expected 2 techniques, got %d", len(relatedTechniques))
		}
	})
}

func TestListTacticsPaginated(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ListTactics_Empty", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_tactics_empty.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		tactics, total, err := store.ListTacticsPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list tactics: %v", err)
		}

		if total != 0 {
			t.Errorf("Expected 0 tactics, got %d", total)
		}

		if len(tactics) != 0 {
			t.Errorf("Expected empty list, got %d items", len(tactics))
		}
	})

	testutils.Run(t, testutils.Level2, "ListTactics_WithPagination", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_tactics_pagination.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()

		tactics := []AttackTactic{
			{ID: "TA0001", Name: "Initial Access"},
			{ID: "TA0002", Name: "Execution"},
			{ID: "TA0003", Name: "Persistence"},
		}

		for _, tactic := range tactics {
			if err := store.db.Create(&tactic).Error; err != nil {
				t.Fatalf("Failed to create tactic: %v", err)
			}
		}

		result, total, err := store.ListTacticsPaginated(ctx, 0, 2)
		if err != nil {
			t.Fatalf("Failed to list tactics: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected 3 total tactics, got %d", total)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 tactics in page, got %d", len(result))
		}
	})
}

func TestListMitigationsPaginated(t *testing.T) {
	testutils.Run(t, testutils.Level2, "ListMitigations_Empty", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_mitigations_empty.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()
		mitigations, total, err := store.ListMitigationsPaginated(ctx, 0, 10)
		if err != nil {
			t.Fatalf("Failed to list mitigations: %v", err)
		}

		if total != 0 {
			t.Errorf("Expected 0 mitigations, got %d", total)
		}

		if len(mitigations) != 0 {
			t.Errorf("Expected empty list, got %d items", len(mitigations))
		}
	})

	testutils.Run(t, testutils.Level2, "ListMitigations_WithPagination", nil, func(t *testing.T, tx *gorm.DB) {
		tempDB := "/tmp/test_attack_mitigations_pagination.db"
		defer os.Remove(tempDB)

		store, err := NewLocalAttackStore(tempDB)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		ctx := context.Background()

		mitigations := []AttackMitigation{
			{ID: "M1001", Name: "Mitigation 1", Description: "Test mitigation"},
			{ID: "M1002", Name: "Mitigation 2", Description: "Test mitigation"},
			{ID: "M1003", Name: "Mitigation 3", Description: "Test mitigation"},
		}

		for _, mitigation := range mitigations {
			if err := store.db.Create(&mitigation).Error; err != nil {
				t.Fatalf("Failed to create mitigation: %v", err)
			}
		}

		result, total, err := store.ListMitigationsPaginated(ctx, 0, 2)
		if err != nil {
			t.Fatalf("Failed to list mitigations: %v", err)
		}

		if total != 3 {
			t.Errorf("Expected 3 total mitigations, got %d", total)
		}

		if len(result) != 2 {
			t.Errorf("Expected 2 mitigations in page, got %d", len(result))
		}
	})
}
