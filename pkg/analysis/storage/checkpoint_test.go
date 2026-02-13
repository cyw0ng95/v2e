package storage

import (
	"fmt"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/graph"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

func TestGraphStore_SaveFullCheckpoint(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_SaveFullCheckpoint", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_full.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		// Create test graph
		g := graph.New()
		cveURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		g.AddNode(cveURN, map[string]interface{}{"id": "CVE-2024-1234"})

		// Save full checkpoint
		checkID, err := store.SaveFullCheckpoint(g)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}

		if checkID != 1 {
			t.Errorf("Expected checkpoint ID 1, got %d", checkID)
		}

		// Verify checkpoint exists
		latest, err := store.GetLatestCheckpointID()
		if err != nil {
			t.Fatalf("Failed to get latest checkpoint: %v", err)
		}

		if latest != checkID {
			t.Errorf("Expected latest checkpoint %d, got %d", checkID, latest)
		}
	})
}

func TestGraphStore_SaveIncrementalCheckpoint(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_SaveIncrementalCheckpoint", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_incremental.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		// Create test graph
		g := graph.New()
		cveURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		g.AddNode(cveURN, map[string]interface{}{"id": "CVE-2024-1234"})

		// Save base checkpoint
		baseID, err := store.SaveFullCheckpoint(g)
		if err != nil {
			t.Fatalf("Failed to save base checkpoint: %v", err)
		}

		// Add more nodes
		for i := 0; i < 5; i++ {
			urn, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
			g.AddNode(urn, map[string]interface{}{"id": fmt.Sprintf("CVE-2024-%d", i)})
		}

		// Create incremental changes
		changes := []IncrementalChange{}
		for i := 0; i < 5; i++ {
			changes = append(changes, IncrementalChange{
				Type:    "add_node",
				NodeURN: fmt.Sprintf("urn:nvd:cve:CVE-2024-%d", i),
				Data:    map[string]interface{}{"id": fmt.Sprintf("CVE-2024-%d", i)},
			})
		}

		// Save incremental checkpoint
		incID, err := store.SaveIncrementalCheckpoint(g, changes, baseID)
		if err != nil {
			t.Fatalf("Failed to save incremental checkpoint: %v", err)
		}

		if incID != baseID+1 {
			t.Errorf("Expected checkpoint ID %d, got %d", baseID+1, incID)
		}
	})
}

func TestGraphStore_LoadFromCheckpoint_Full(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_LoadFromCheckpoint_Full", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_load_full.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		// Create and save graph
		g := graph.New()
		cveURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		g.AddNode(cveURN, map[string]interface{}{"id": "CVE-2024-1234", "severity": "HIGH"})

		checkID, err := store.SaveFullCheckpoint(g)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}

		store.Close()

		// Load checkpoint in new store
		store2, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store2.Close()

		loaded, err := store2.LoadFromCheckpoint(checkID)
		if err != nil {
			t.Fatalf("Failed to load checkpoint: %v", err)
		}

		if loaded.NodeCount() != 1 {
			t.Errorf("Expected 1 node, got %d", loaded.NodeCount())
		}

		node, exists := loaded.GetNode(cveURN)
		if !exists {
			t.Error("CVE node not found")
		}

		if node.Properties["id"] != "CVE-2024-1234" {
			t.Errorf("Node data mismatch")
		}
	})
}

func TestGraphStore_LoadFromCheckpoint_Incremental(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_LoadFromCheckpoint_Incremental", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_load_inc.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		// Create base graph
		g := graph.New()
		cveURN1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-0001")
		g.AddNode(cveURN1, map[string]interface{}{"id": "CVE-2024-0001"})

		baseID, err := store.SaveFullCheckpoint(g)
		if err != nil {
			t.Fatalf("Failed to save base checkpoint: %v", err)
		}

		// Add incremental changes
		cveURN2, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-0002")
		g.AddNode(cveURN2, map[string]interface{}{"id": "CVE-2024-0002"})

		changes := []IncrementalChange{
			{
				Type:    "add_node",
				NodeURN: cveURN2.String(),
				Data:    map[string]interface{}{"id": "CVE-2024-0002"},
			},
		}

		incID, err := store.SaveIncrementalCheckpoint(g, changes, baseID)
		if err != nil {
			t.Fatalf("Failed to save incremental checkpoint: %v", err)
		}

		store.Close()

		// Load incremental checkpoint
		store2, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store2.Close()

		loaded, err := store2.LoadFromCheckpoint(incID)
		if err != nil {
			t.Fatalf("Failed to load incremental checkpoint: %v", err)
		}

		if loaded.NodeCount() != 2 {
			t.Errorf("Expected 2 nodes after incremental load, got %d", loaded.NodeCount())
		}
	})
}

func TestGraphStore_ListCheckpoints(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_ListCheckpoints", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_list.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		// Save multiple checkpoints
		g := graph.New()
		for i := 0; i < 3; i++ {
			store.SaveFullCheckpoint(g)
		}

		checkpoints, err := store.ListCheckpoints()
		if err != nil {
			t.Fatalf("Failed to list checkpoints: %v", err)
		}

		if len(checkpoints) != 3 {
			t.Errorf("Expected 3 checkpoints, got %d", len(checkpoints))
		}
	})
}

func TestGraphStore_DeleteCheckpoint(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_DeleteCheckpoint", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_delete.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		// Save checkpoint
		g := graph.New()
		checkID, err := store.SaveFullCheckpoint(g)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}

		// Delete it
		if err := store.DeleteCheckpoint(checkID); err != nil {
			t.Fatalf("Failed to delete checkpoint: %v", err)
		}

		// Verify it's gone
		_, err = store.LoadFromCheckpoint(checkID)
		if err == nil {
			t.Error("Expected error when loading deleted checkpoint")
		}
	})
}

func TestGraphStore_CompactCheckpoints(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_CompactCheckpoints", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_compact.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		// Save 5 checkpoints
		g := graph.New()
		for i := 0; i < 5; i++ {
			store.SaveFullCheckpoint(g)
		}

		// Compact to keep only 3
		if err := store.CompactCheckpoints(3); err != nil {
			t.Fatalf("Failed to compact checkpoints: %v", err)
		}

		// Verify only 3 remain
		checkpoints, err := store.ListCheckpoints()
		if err != nil {
			t.Fatalf("Failed to list checkpoints: %v", err)
		}

		if len(checkpoints) != 3 {
			t.Errorf("Expected 3 checkpoints after compaction, got %d", len(checkpoints))
		}
	})
}

func TestGraphStore_RecoverFromLatestCheckpoint(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_RecoverFromLatestCheckpoint", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_recover.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		// Save checkpoint
		g := graph.New()
		cveURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-9999")
		g.AddNode(cveURN, map[string]interface{}{"id": "CVE-2024-9999"})

		expectedID, err := store.SaveFullCheckpoint(g)
		if err != nil {
			t.Fatalf("Failed to save checkpoint: %v", err)
		}

		store.Close()

		// Recover in new store
		store2, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store2.Close()

		recovered, checkID, err := store2.RecoverFromLatestCheckpoint()
		if err != nil {
			t.Fatalf("Failed to recover: %v", err)
		}

		if checkID != expectedID {
			t.Errorf("Expected checkpoint ID %d, got %d", expectedID, checkID)
		}

		if recovered.NodeCount() != 1 {
			t.Errorf("Expected 1 node, got %d", recovered.NodeCount())
		}
	})
}

func TestGraphStore_RecoverFromLatestCheckpoint_NoCheckpoints(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphStore_RecoverFromLatestCheckpoint_NoCheckpoints", nil, func(t *testing.T, _ *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
		tmpFile := "/tmp/test_checkpoint_no_data.db"
		defer os.Remove(tmpFile)

		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		defer store.Close()

		// Try to recover with no checkpoints
		_, _, err = store.RecoverFromLatestCheckpoint()
		if err == nil {
			t.Error("Expected error when recovering with no checkpoints")
		}
	})
}

func TestApplyChanges_AddNode(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ApplyChanges_AddNode", nil, func(t *testing.T, _ *gorm.DB) {
		g := graph.New()
		store := &GraphStore{logger: common.NewLogger(os.Stdout, "test", common.WarnLevel)}

		changes := []IncrementalChange{
			{
				Type:    "add_node",
				NodeURN: "urn:nvd:cve:CVE-2024-1234",
				Data:    map[string]interface{}{"id": "CVE-2024-1234"},
			},
		}

		err := store.applyChanges(g, changes)
		if err != nil {
			t.Fatalf("Failed to apply changes: %v", err)
		}

		if g.NodeCount() != 1 {
			t.Errorf("Expected 1 node, got %d", g.NodeCount())
		}
	})
}

func TestApplyChanges_AddEdge(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ApplyChanges_AddEdge", nil, func(t *testing.T, _ *gorm.DB) {
		g := graph.New()
		store := &GraphStore{logger: common.NewLogger(os.Stdout, "test", common.WarnLevel)}

		// First add nodes
		cveURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cweURN, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")
		g.AddNode(cveURN, nil)
		g.AddNode(cweURN, nil)

		// Then apply edge change
		changes := []IncrementalChange{
			{
				Type:     "add_edge",
				From:     cveURN.String(),
				To:       cweURN.String(),
				EdgeType: string(graph.EdgeTypeReferences),
			},
		}

		err := store.applyChanges(g, changes)
		if err != nil {
			t.Fatalf("Failed to apply changes: %v", err)
		}

		if g.EdgeCount() != 1 {
			t.Errorf("Expected 1 edge, got %d", g.EdgeCount())
		}
	})
}
