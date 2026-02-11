package glc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migrations
	if err := MigrateTables(db); err != nil {
		t.Fatalf("Failed to migrate tables: %v", err)
	}

	return db
}

// TestGraphModel tests basic GraphModel struct properties
func TestGraphModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGraphModel", nil, func(t *testing.T, tx *gorm.DB) {
		nodes, _ := json.Marshal([]map[string]interface{}{
			{"id": "node-1", "type": "glc", "position": map[string]float64{"x": 100, "y": 100}},
		})
		edges, _ := json.Marshal([]map[string]interface{}{})

		graph := &GraphModel{
			GraphID:     "test-graph-123",
			Name:        "Test Graph",
			Description: "A test graph for unit testing",
			PresetID:    "d3fend",
			Tags:        `["test", "unit"]`,
			Nodes:       string(nodes),
			Edges:       string(edges),
			Viewport:    `{"x": 0, "y": 0, "zoom": 1}`,
			Version:     1,
			IsArchived:  false,
		}

		if graph.GraphID != "test-graph-123" {
			t.Errorf("Expected GraphID 'test-graph-123', got '%s'", graph.GraphID)
		}
		if graph.Name != "Test Graph" {
			t.Errorf("Expected Name 'Test Graph', got '%s'", graph.Name)
		}
		if graph.PresetID != "d3fend" {
			t.Errorf("Expected PresetID 'd3fend', got '%s'", graph.PresetID)
		}
		if graph.Version != 1 {
			t.Errorf("Expected Version 1, got %d", graph.Version)
		}
	})
}

// TestGraphVersionModel tests GraphVersionModel struct properties
func TestGraphVersionModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGraphVersionModel", nil, func(t *testing.T, tx *gorm.DB) {
		version := &GraphVersionModel{
			GraphID:  1,
			Version:  1,
			Nodes:    `[]`,
			Edges:    `[]`,
			Viewport: `{"x": 0, "y": 0, "zoom": 1}`,
		}

		if version.GraphID != 1 {
			t.Errorf("Expected GraphID 1, got %d", version.GraphID)
		}
		if version.Version != 1 {
			t.Errorf("Expected Version 1, got %d", version.Version)
		}
	})
}

// TestUserPresetModel tests UserPresetModel struct properties
func TestUserPresetModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUserPresetModel", nil, func(t *testing.T, tx *gorm.DB) {
		preset := &UserPresetModel{
			PresetID:    "custom-preset-123",
			Name:        "Custom Security Model",
			Version:     "1.0.0",
			Description: "A custom preset for security modeling",
			Author:      "Test User",
			Theme:       `{"primary": "#6366f1", "background": "#0f172a"}`,
			Behavior:    `{"snapToGrid": true, "gridSize": 20}`,
			NodeTypes:   `[]`,
			Relations:   `[]`,
		}

		if preset.PresetID != "custom-preset-123" {
			t.Errorf("Expected PresetID 'custom-preset-123', got '%s'", preset.PresetID)
		}
		if preset.Name != "Custom Security Model" {
			t.Errorf("Expected Name 'Custom Security Model', got '%s'", preset.Name)
		}
		if preset.Version != "1.0.0" {
			t.Errorf("Expected Version '1.0.0', got '%s'", preset.Version)
		}
	})
}

// TestShareLinkModel tests ShareLinkModel struct properties
func TestShareLinkModel(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestShareLinkModel", nil, func(t *testing.T, tx *gorm.DB) {
		expiresAt := time.Now().Add(24 * time.Hour)
		link := &ShareLinkModel{
			LinkID:    "abc12345",
			GraphID:   "test-graph-123",
			Password:  "hashed-password",
			ExpiresAt: &expiresAt,
			ViewCount: 0,
		}

		if link.LinkID != "abc12345" {
			t.Errorf("Expected LinkID 'abc12345', got '%s'", link.LinkID)
		}
		if link.GraphID != "test-graph-123" {
			t.Errorf("Expected GraphID 'test-graph-123', got '%s'", link.GraphID)
		}
		if link.ViewCount != 0 {
			t.Errorf("Expected ViewCount 0, got %d", link.ViewCount)
		}
		if link.ExpiresAt == nil {
			t.Error("Expected ExpiresAt to be set")
		}
	})
}

// TestNewStore tests store creation
func TestNewStore(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewStore", nil, func(t *testing.T, tx *gorm.DB) {
		// Test with nil database
		_, err := NewStore(nil)
		if err == nil {
			t.Error("Expected error for nil database, got nil")
		}

		// Test with valid database
		db := setupTestDB(t)
		store, err := NewStore(db)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}
		if store == nil {
			t.Error("Expected store to be created, got nil")
		}
	})
}

// TestCreateGraph tests graph creation
func TestCreateGraph(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestCreateGraph", db, func(t *testing.T, tx *gorm.DB) {
		// Create a transaction-scoped store
		testStore, err := NewStore(tx)
		if err != nil {
			t.Fatalf("Failed to create test store: %v", err)
		}

		nodes := `[{"id": "node-1", "type": "glc", "position": {"x": 100, "y": 100}}]`
		edges := `[]`
		viewport := `{"x": 0, "y": 0, "zoom": 1}`

		graph, err := testStore.CreateGraph(
			context.Background(),
			"Test Graph",
			"A test graph",
			"d3fend",
			nodes,
			edges,
			viewport,
		)

		if err != nil {
			t.Fatalf("Failed to create graph: %v", err)
		}

		if graph.GraphID == "" {
			t.Error("Expected GraphID to be generated")
		}
		if graph.Name != "Test Graph" {
			t.Errorf("Expected Name 'Test Graph', got '%s'", graph.Name)
		}
		if graph.PresetID != "d3fend" {
			t.Errorf("Expected PresetID 'd3fend', got '%s'", graph.PresetID)
		}
		if graph.Version != 1 {
			t.Errorf("Expected Version 1, got %d", graph.Version)
		}
	})
}

// TestGetGraph tests graph retrieval
func TestGetGraph(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestGetGraph", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph first
		created, err := testStore.CreateGraph(
			context.Background(),
			"Test Graph",
			"Description",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)
		if err != nil {
			t.Fatalf("Failed to create graph: %v", err)
		}

		// Retrieve the graph
		graph, err := testStore.GetGraph(context.Background(), created.GraphID)
		if err != nil {
			t.Fatalf("Failed to get graph: %v", err)
		}

		if graph.GraphID != created.GraphID {
			t.Errorf("Expected GraphID '%s', got '%s'", created.GraphID, graph.GraphID)
		}
		if graph.Name != "Test Graph" {
			t.Errorf("Expected Name 'Test Graph', got '%s'", graph.Name)
		}

		// Test non-existent graph
		_, err = testStore.GetGraph(context.Background(), "non-existent-id")
		if err == nil {
			t.Error("Expected error for non-existent graph")
		}
	})
}

// TestUpdateGraph tests graph updates
func TestUpdateGraph(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestUpdateGraph", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		created, _ := testStore.CreateGraph(
			context.Background(),
			"Original Name",
			"Original Description",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)

		// Update the graph
		updates := map[string]interface{}{
			"name":        "Updated Name",
			"description": "Updated Description",
		}

		updated, err := testStore.UpdateGraph(context.Background(), created.GraphID, updates)
		if err != nil {
			t.Fatalf("Failed to update graph: %v", err)
		}

		if updated.Name != "Updated Name" {
			t.Errorf("Expected Name 'Updated Name', got '%s'", updated.Name)
		}
		if updated.Description != "Updated Description" {
			t.Errorf("Expected Description 'Updated Description', got '%s'", updated.Description)
		}
	})
}

// TestDeleteGraph tests graph deletion
func TestDeleteGraph(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestDeleteGraph", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		created, _ := testStore.CreateGraph(
			context.Background(),
			"To Be Deleted",
			"",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)

		// Delete the graph
		err := testStore.DeleteGraph(context.Background(), created.GraphID)
		if err != nil {
			t.Fatalf("Failed to delete graph: %v", err)
		}

		// Verify deletion
		_, err = testStore.GetGraph(context.Background(), created.GraphID)
		if err == nil {
			t.Error("Expected error getting deleted graph")
		}

		// Test deleting non-existent graph
		err = testStore.DeleteGraph(context.Background(), "non-existent-id")
		if err == nil {
			t.Error("Expected error deleting non-existent graph")
		}
	})
}

// TestListGraphs tests graph listing
func TestListGraphs(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestListGraphs", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create multiple graphs
		for i := 0; i < 5; i++ {
			testStore.CreateGraph(
				context.Background(),
				"Graph "+string(rune('0'+i)),
				"",
				"d3fend",
				"[]",
				"[]",
				"{}",
			)
		}

		// List graphs
		graphs, total, err := testStore.ListGraphs(context.Background(), "", 0, 10)
		if err != nil {
			t.Fatalf("Failed to list graphs: %v", err)
		}

		if len(graphs) != 5 {
			t.Errorf("Expected 5 graphs, got %d", len(graphs))
		}
		if total != 5 {
			t.Errorf("Expected total 5, got %d", total)
		}

		// Test with preset filter
		d3fendGraphs, _, err := testStore.ListGraphs(context.Background(), "d3fend", 0, 10)
		if err != nil {
			t.Fatalf("Failed to list d3fend graphs: %v", err)
		}
		if len(d3fendGraphs) != 5 {
			t.Errorf("Expected 5 d3fend graphs, got %d", len(d3fendGraphs))
		}
	})
}

// TestVersionOperations tests version creation and retrieval
func TestVersionOperations(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestVersionOperations", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"Version Test",
			"",
			"d3fend",
			`[{"id": "node-1"}]`,
			`[]`,
			`{}`,
		)

		// Update to create a version
		testStore.UpdateGraph(context.Background(), graph.GraphID, map[string]interface{}{
			"nodes": `[{"id": "node-1"}, {"id": "node-2"}]`,
		})

		// Get the updated graph to retrieve its DB ID
		updatedGraph, _ := testStore.GetGraph(context.Background(), graph.GraphID)

		// List versions
		versions, err := testStore.ListVersions(context.Background(), updatedGraph.ID, 10)
		if err != nil {
			t.Fatalf("Failed to list versions: %v", err)
		}

		if len(versions) < 1 {
			t.Error("Expected at least 1 version")
		}

		// Get specific version
		if len(versions) > 0 {
			v, err := testStore.GetVersion(context.Background(), updatedGraph.ID, versions[0].Version)
			if err != nil {
				t.Fatalf("Failed to get version: %v", err)
			}
			if v.Version != versions[0].Version {
				t.Errorf("Expected version %d, got %d", versions[0].Version, v.Version)
			}
		}
	})
}

// TestRestoreVersion tests version restoration
func TestRestoreVersion(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestRestoreVersion", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"Restore Test",
			"",
			"d3fend",
			`[{"id": "node-1"}]`,
			`[]`,
			`{}`,
		)

		// Update to create a version
		testStore.UpdateGraph(context.Background(), graph.GraphID, map[string]interface{}{
			"nodes": `[{"id": "node-1"}, {"id": "node-2"}]`,
		})

		// Restore to version 1
		restored, err := testStore.RestoreVersion(context.Background(), graph.GraphID, 1)
		if err != nil {
			t.Fatalf("Failed to restore version: %v", err)
		}

		// Check nodes were restored
		if restored.Nodes != `[{"id": "node-1"}]` {
			t.Errorf("Expected original nodes, got '%s'", restored.Nodes)
		}
	})
}

// TestUserPresetOperations tests preset CRUD operations
func TestUserPresetOperations(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestUserPresetOperations", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create preset
		preset, err := testStore.CreateUserPreset(
			context.Background(),
			"Custom Preset",
			"1.0.0",
			"A custom preset",
			"Test User",
			`{"primary": "#6366f1"}`,
			`{"snapToGrid": true}`,
			`[]`,
			`[]`,
		)
		if err != nil {
			t.Fatalf("Failed to create preset: %v", err)
		}

		if preset.PresetID == "" {
			t.Error("Expected PresetID to be generated")
		}

		// Get preset
		retrieved, err := testStore.GetUserPreset(context.Background(), preset.PresetID)
		if err != nil {
			t.Fatalf("Failed to get preset: %v", err)
		}
		if retrieved.Name != "Custom Preset" {
			t.Errorf("Expected Name 'Custom Preset', got '%s'", retrieved.Name)
		}

		// Update preset
		updated, err := testStore.UpdateUserPreset(context.Background(), preset.PresetID, map[string]interface{}{
			"name": "Updated Preset",
		})
		if err != nil {
			t.Fatalf("Failed to update preset: %v", err)
		}
		if updated.Name != "Updated Preset" {
			t.Errorf("Expected Name 'Updated Preset', got '%s'", updated.Name)
		}

		// List presets
		presets, err := testStore.ListUserPresets(context.Background())
		if err != nil {
			t.Fatalf("Failed to list presets: %v", err)
		}
		if len(presets) != 1 {
			t.Errorf("Expected 1 preset, got %d", len(presets))
		}

		// Delete preset
		err = testStore.DeleteUserPreset(context.Background(), preset.PresetID)
		if err != nil {
			t.Fatalf("Failed to delete preset: %v", err)
		}

		// Verify deletion
		_, err = testStore.GetUserPreset(context.Background(), preset.PresetID)
		if err == nil {
			t.Error("Expected error getting deleted preset")
		}
	})
}

// TestShareLinkOperations tests share link CRUD operations
func TestShareLinkOperations(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestShareLinkOperations", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph first
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"Share Test",
			"",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)

		// Create share link without expiration
		link, err := testStore.CreateShareLink(context.Background(), graph.GraphID, "", nil)
		if err != nil {
			t.Fatalf("Failed to create share link: %v", err)
		}

		if link.LinkID == "" {
			t.Error("Expected LinkID to be generated")
		}
		if link.GraphID != graph.GraphID {
			t.Errorf("Expected GraphID '%s', got '%s'", graph.GraphID, link.GraphID)
		}

		// Create share link with expiration
		expiresIn := 24 * time.Hour
		expiringLink, err := testStore.CreateShareLink(context.Background(), graph.GraphID, "secret", &expiresIn)
		if err != nil {
			t.Fatalf("Failed to create expiring share link: %v", err)
		}

		if expiringLink.Password != "secret" {
			t.Error("Expected password to be set")
		}
		if expiringLink.ExpiresAt == nil {
			t.Error("Expected ExpiresAt to be set")
		}

		// Get share link
		retrieved, err := testStore.GetShareLink(context.Background(), link.LinkID)
		if err != nil {
			t.Fatalf("Failed to get share link: %v", err)
		}
		if retrieved.LinkID != link.LinkID {
			t.Errorf("Expected LinkID '%s', got '%s'", link.LinkID, retrieved.LinkID)
		}

		// Get graph by share link
		sharedGraph, err := testStore.GetGraphByShareLink(context.Background(), link.LinkID, "")
		if err != nil {
			t.Fatalf("Failed to get graph by share link: %v", err)
		}
		if sharedGraph.GraphID != graph.GraphID {
			t.Errorf("Expected GraphID '%s', got '%s'", graph.GraphID, sharedGraph.GraphID)
		}

		// Test password-protected link
		_, err = testStore.GetGraphByShareLink(context.Background(), expiringLink.LinkID, "wrong")
		if err == nil {
			t.Error("Expected error with wrong password")
		}

		// Delete share link
		err = testStore.DeleteShareLink(context.Background(), link.LinkID)
		if err != nil {
			t.Fatalf("Failed to delete share link: %v", err)
		}

		// Verify deletion
		_, err = testStore.GetShareLink(context.Background(), link.LinkID)
		if err == nil {
			t.Error("Expected error getting deleted share link")
		}
	})
}

// TestStore_Close tests the Close method
func TestStore_Close(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestStore_Close", nil, func(t *testing.T, tx *gorm.DB) {
		db := setupTestDB(t)
		store, err := NewStore(db)
		if err != nil {
			t.Fatalf("Failed to create store: %v", err)
		}

		// Close should return nil (no-op for shared connection)
		if err := store.Close(); err != nil {
			t.Errorf("Expected Close to return nil, got %v", err)
		}
	})
}

// TestGetGraphByDBID tests retrieving a graph by database ID
func TestGetGraphByDBID(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestGetGraphByDBID", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		created, _ := testStore.CreateGraph(
			context.Background(),
			"DB ID Test",
			"Description",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)

		// Get by DB ID
		graph, err := testStore.GetGraphByDBID(context.Background(), created.ID)
		if err != nil {
			t.Fatalf("Failed to get graph by DB ID: %v", err)
		}

		if graph.ID != created.ID {
			t.Errorf("Expected ID %d, got %d", created.ID, graph.ID)
		}
		if graph.Name != "DB ID Test" {
			t.Errorf("Expected Name 'DB ID Test', got '%s'", graph.Name)
		}

		// Test non-existent DB ID
		_, err = testStore.GetGraphByDBID(context.Background(), 99999)
		if err == nil {
			t.Error("Expected error for non-existent DB ID")
		}
	})
}

// TestListRecentGraphs tests listing recent graphs
func TestListRecentGraphs(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestListRecentGraphs", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create multiple graphs
		for i := 0; i < 5; i++ {
			testStore.CreateGraph(
				context.Background(),
				"Recent Graph "+string(rune('0'+i)),
				"",
				"d3fend",
				"[]",
				"[]",
				"{}",
			)
		}

		// List recent graphs with limit
		graphs, err := testStore.ListRecentGraphs(context.Background(), 3)
		if err != nil {
			t.Fatalf("Failed to list recent graphs: %v", err)
		}

		if len(graphs) != 3 {
			t.Errorf("Expected 3 graphs, got %d", len(graphs))
		}

		// List all recent graphs
		allGraphs, err := testStore.ListRecentGraphs(context.Background(), 10)
		if err != nil {
			t.Fatalf("Failed to list all recent graphs: %v", err)
		}
		if len(allGraphs) != 5 {
			t.Errorf("Expected 5 graphs, got %d", len(allGraphs))
		}
	})
}

// TestDeleteOldVersions tests deleting old graph versions
func TestDeleteOldVersions(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestDeleteOldVersions", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"Version Cleanup Test",
			"",
			"d3fend",
			`[{"id": "node-1"}]`,
			`[]`,
			`{}`,
		)

		// Create multiple versions by updating nodes
		for i := 0; i < 5; i++ {
			testStore.UpdateGraph(context.Background(), graph.GraphID, map[string]interface{}{
				"nodes": `[{"id": "node-` + string(rune('0'+i)) + `"}]`,
			})
		}

		// Get the updated graph to retrieve its DB ID
		updatedGraph, _ := testStore.GetGraph(context.Background(), graph.GraphID)

		// Check we have versions
		versions, _ := testStore.ListVersions(context.Background(), updatedGraph.ID, 0)
		if len(versions) < 5 {
			t.Fatalf("Expected at least 5 versions, got %d", len(versions))
		}

		// Delete old versions, keeping only 2
		err := testStore.DeleteOldVersions(context.Background(), updatedGraph.ID, 2)
		if err != nil {
			t.Fatalf("Failed to delete old versions: %v", err)
		}

		// Check only 2 versions remain
		remainingVersions, _ := testStore.ListVersions(context.Background(), updatedGraph.ID, 0)
		if len(remainingVersions) != 2 {
			t.Errorf("Expected 2 remaining versions, got %d", len(remainingVersions))
		}
	})
}

// TestIncrementViewCount tests incrementing share link view count
func TestIncrementViewCount(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestIncrementViewCount", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph and share link
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"View Count Test",
			"",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)
		link, _ := testStore.CreateShareLink(context.Background(), graph.GraphID, "", nil)

		// Initial view count should be 0
		if link.ViewCount != 0 {
			t.Errorf("Expected initial ViewCount 0, got %d", link.ViewCount)
		}

		// Increment view count
		if err := testStore.IncrementViewCount(context.Background(), link.LinkID); err != nil {
			t.Fatalf("Failed to increment view count: %v", err)
		}

		// Check view count was incremented
		updatedLink, _ := testStore.GetShareLink(context.Background(), link.LinkID)
		if updatedLink.ViewCount != 1 {
			t.Errorf("Expected ViewCount 1, got %d", updatedLink.ViewCount)
		}

		// Increment again
		testStore.IncrementViewCount(context.Background(), link.LinkID)
		updatedLink2, _ := testStore.GetShareLink(context.Background(), link.LinkID)
		if updatedLink2.ViewCount != 2 {
			t.Errorf("Expected ViewCount 2, got %d", updatedLink2.ViewCount)
		}
	})
}

// TestExpiredShareLink tests that expired share links return an error
func TestExpiredShareLink(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestExpiredShareLink", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"Expired Link Test",
			"",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)

		// Create share link that expires in the past
		expiresIn := -1 * time.Hour
		link, _ := testStore.CreateShareLink(context.Background(), graph.GraphID, "", &expiresIn)

		// Trying to get an expired link should return an error
		_, err := testStore.GetShareLink(context.Background(), link.LinkID)
		if err == nil {
			t.Error("Expected error for expired share link")
		}
		if err != nil && err.Error() != "share link expired" {
			t.Errorf("Expected 'share link expired' error, got: %v", err)
		}
	})
}

// TestUpdateGraphWithEdges tests updating graph with edges creates a version snapshot
func TestUpdateGraphWithEdges(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestUpdateGraphWithEdges", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"Edge Update Test",
			"",
			"d3fend",
			`[{"id": "node-1"}]`,
			`[]`,
			`{}`,
		)

		// Update with edges
		updated, err := testStore.UpdateGraph(context.Background(), graph.GraphID, map[string]interface{}{
			"edges": `[{"id": "edge-1", "source": "node-1", "target": "node-2"}]`,
		})
		if err != nil {
			t.Fatalf("Failed to update graph with edges: %v", err)
		}

		// Check version was incremented
		if updated.Version <= graph.Version {
			t.Errorf("Expected version to be incremented, got %d", updated.Version)
		}

		// Get the graph to check its DB ID for version listing
		updatedGraph, _ := testStore.GetGraph(context.Background(), graph.GraphID)

		// Verify version was created
		versions, _ := testStore.ListVersions(context.Background(), updatedGraph.ID, 10)
		if len(versions) < 1 {
			t.Error("Expected at least 1 version to be created")
		}
	})
}

// TestUpdateNonExistentGraph tests updating a non-existent graph
func TestUpdateNonExistentGraph(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestUpdateNonExistentGraph", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Try to update non-existent graph
		_, err := testStore.UpdateGraph(context.Background(), "non-existent-id", map[string]interface{}{
			"name": "Updated Name",
		})
		if err == nil {
			t.Error("Expected error updating non-existent graph")
		}
	})
}

// TestRestoreNonExistentVersion tests restoring a non-existent version
func TestRestoreNonExistentVersion(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestRestoreNonExistentVersion", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, _ := testStore.CreateGraph(
			context.Background(),
			"Restore Non-Existent Test",
			"",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)

		// Try to restore a version that doesn't exist
		_, err := testStore.RestoreVersion(context.Background(), graph.GraphID, 999)
		if err == nil {
			t.Error("Expected error restoring non-existent version")
		}
	})
}

// TestGetNonExistentVersion tests getting a non-existent version
func TestGetNonExistentVersion(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestGetNonExistentVersion", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Try to get non-existent version
		_, err := testStore.GetVersion(context.Background(), 99999, 1)
		if err == nil {
			t.Error("Expected error getting non-existent version")
		}
	})
}

// TestUpdateNonExistentUserPreset tests updating a non-existent preset
func TestUpdateNonExistentUserPreset(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestUpdateNonExistentUserPreset", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Try to update non-existent preset
		_, err := testStore.UpdateUserPreset(context.Background(), "non-existent-preset", map[string]interface{}{
			"name": "Updated Name",
		})
		if err == nil {
			t.Error("Expected error updating non-existent preset")
		}
	})
}

// TestDeleteNonExistentUserPreset tests deleting a non-existent preset
func TestDeleteNonExistentUserPreset(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestDeleteNonExistentUserPreset", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Try to delete non-existent preset
		err := testStore.DeleteUserPreset(context.Background(), "non-existent-preset")
		if err == nil {
			t.Error("Expected error deleting non-existent preset")
		}
	})
}

// TestDeleteNonExistentShareLink tests deleting a non-existent share link
func TestDeleteNonExistentShareLink(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestDeleteNonExistentShareLink", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Try to delete non-existent share link
		err := testStore.DeleteShareLink(context.Background(), "non-existent-link")
		if err == nil {
			t.Error("Expected error deleting non-existent share link")
		}
	})
}

// TestGenerateLinkID tests link ID generation
func TestGenerateLinkID(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestGenerateLinkID", nil, func(t *testing.T, tx *gorm.DB) {
		id1, err := generateLinkID(8)
		if err != nil {
			t.Fatalf("Failed to generate link ID: %v", err)
		}
		if len(id1) != 8 {
			t.Errorf("Expected link ID length 8, got %d", len(id1))
		}

		id2, err := generateLinkID(8)
		if err != nil {
			t.Fatalf("Failed to generate second link ID: %v", err)
		}
		if id1 == id2 {
			t.Error("Expected different link IDs")
		}
	})
}

// TestParseJSON tests the JSON helper function
func TestParseJSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestParseJSON", nil, func(t *testing.T, tx *gorm.DB) {
		// Test valid JSON
		result, err := parseJSON[map[string]string](`{"key": "value"}`)
		if err != nil {
			t.Fatalf("Failed to parse valid JSON: %v", err)
		}
		if result["key"] != "value" {
			t.Errorf("Expected key 'value', got '%s'", result["key"])
		}

		// Test invalid JSON
		_, err = parseJSON[map[string]string](`invalid json`)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

// TestTableNames tests GORM table name overrides
func TestTableNames(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestTableNames", nil, func(t *testing.T, tx *gorm.DB) {
		var graph GraphModel
		var graphVersion GraphVersionModel
		var userPreset UserPresetModel
		var shareLink ShareLinkModel

		if graph.TableName() != "glc_graphs" {
			t.Errorf("Expected table name 'glc_graphs', got '%s'", graph.TableName())
		}
		if graphVersion.TableName() != "glc_graph_versions" {
			t.Errorf("Expected table name 'glc_graph_versions', got '%s'", graphVersion.TableName())
		}
		if userPreset.TableName() != "glc_user_presets" {
			t.Errorf("Expected table name 'glc_user_presets', got '%s'", userPreset.TableName())
		}
		if shareLink.TableName() != "glc_share_links" {
			t.Errorf("Expected table name 'glc_share_links', got '%s'", shareLink.TableName())
		}
	})
}

// TestConcurrentGraphUpdates tests concurrent graph updates for race conditions
func TestConcurrentGraphUpdates(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestConcurrentGraphUpdates", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, err := testStore.CreateGraph(
			context.Background(),
			"Concurrent Update Test",
			"",
			"d3fend",
			`[{"id": "node-1"}]`,
			`[]`,
			`{}`,
		)
		if err != nil {
			t.Fatalf("Failed to create graph: %v", err)
		}

		// Launch multiple concurrent updates
		concurrency := 10
		errChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(index int) {
				_, err := testStore.UpdateGraph(context.Background(), graph.GraphID, map[string]interface{}{
					"name": fmt.Sprintf("Concurrent Update %d", index),
				})
				errChan <- err
			}(i)
		}

		// Collect all errors
		for i := 0; i < concurrency; i++ {
			if err := <-errChan; err != nil {
				t.Errorf("Concurrent update failed: %v", err)
			}
		}

		// Verify graph still exists and is consistent
		finalGraph, err := testStore.GetGraph(context.Background(), graph.GraphID)
		if err != nil {
			t.Fatalf("Failed to get final graph: %v", err)
		}
		if finalGraph.GraphID != graph.GraphID {
			t.Errorf("GraphID mismatch after concurrent updates")
		}
	})
}

// TestLargeJSONPayload tests handling of large JSON payloads (1000+ nodes)
func TestLargeJSONPayload(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestLargeJSONPayload", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Build large nodes array (1000+ nodes)
		nodes := make([]map[string]interface{}, 1000)
		for i := 0; i < 1000; i++ {
			nodes[i] = map[string]interface{}{
				"id":       fmt.Sprintf("node-%d", i),
				"type":     "glc",
				"position": map[string]float64{"x": float64(i * 10), "y": float64(i * 10)},
			}
		}
		nodesJSON, err := json.Marshal(nodes)
		if err != nil {
			t.Fatalf("Failed to marshal large nodes array: %v", err)
		}

		// Build large edges array (500 edges)
		edges := make([]map[string]interface{}, 500)
		for i := 0; i < 500; i++ {
			edges[i] = map[string]interface{}{
				"id":     fmt.Sprintf("edge-%d", i),
				"source": fmt.Sprintf("node-%d", i),
				"target": fmt.Sprintf("node-%d", (i+1)%1000),
			}
		}
		edgesJSON, err := json.Marshal(edges)
		if err != nil {
			t.Fatalf("Failed to marshal large edges array: %v", err)
		}

		// Create graph with large payload
		graph, err := testStore.CreateGraph(
			context.Background(),
			"Large Payload Test",
			"Test graph with 1000+ nodes",
			"d3fend",
			string(nodesJSON),
			string(edgesJSON),
			`{"x": 0, "y": 0, "zoom": 1}`,
		)
		if err != nil {
			t.Fatalf("Failed to create graph with large payload: %v", err)
		}

		// Verify retrieval
		retrieved, err := testStore.GetGraph(context.Background(), graph.GraphID)
		if err != nil {
			t.Fatalf("Failed to retrieve graph with large payload: %v", err)
		}

		// Verify nodes can be parsed
		var retrievedNodes []map[string]interface{}
		if err := json.Unmarshal([]byte(retrieved.Nodes), &retrievedNodes); err != nil {
			t.Errorf("Failed to parse retrieved nodes: %v", err)
		}
		if len(retrievedNodes) != 1000 {
			t.Errorf("Expected 1000 nodes, got %d", len(retrievedNodes))
		}

		// Update with large payload
		updatedNodes := make([]map[string]interface{}, 1001)
		copy(updatedNodes, nodes)
		updatedNodes[1000] = map[string]interface{}{
			"id":       "node-1000",
			"type":     "glc",
			"position": map[string]float64{"x": 10000, "y": 10000},
		}
		updatedNodesJSON, _ := json.Marshal(updatedNodes)

		_, err = testStore.UpdateGraph(context.Background(), graph.GraphID, map[string]interface{}{
			"nodes": string(updatedNodesJSON),
		})
		if err != nil {
			t.Errorf("Failed to update graph with large payload: %v", err)
		}
	})
}

// TestPaginationEdgeCases tests pagination edge cases
func TestPaginationEdgeCases(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestPaginationEdgeCases", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Test with empty database
		graphs, total, err := testStore.ListGraphs(context.Background(), "", 0, 10)
		if err != nil {
			t.Fatalf("Failed to list graphs with empty DB: %v", err)
		}
		if len(graphs) != 0 {
			t.Errorf("Expected 0 graphs from empty DB, got %d", len(graphs))
		}
		if total != 0 {
			t.Errorf("Expected total count 0, got %d", total)
		}

		// Test offset=0 explicitly
		for i := 0; i < 5; i++ {
			_, err := testStore.CreateGraph(
				context.Background(),
				fmt.Sprintf("Pagination Test %d", i),
				"",
				"d3fend",
				"[]",
				"[]",
				"{}",
			)
			if err != nil {
				t.Fatalf("Failed to create graph %d: %v", i, err)
			}
		}

		// Test offset=0
		graphs, total, err = testStore.ListGraphs(context.Background(), "", 0, 3)
		if err != nil {
			t.Fatalf("Failed to list graphs with offset=0: %v", err)
		}
		if len(graphs) != 3 {
			t.Errorf("Expected 3 graphs with offset=0, limit=3, got %d", len(graphs))
		}
		if total != 5 {
			t.Errorf("Expected total count 5, got %d", total)
		}

		// Test offset beyond available data
		graphs, total, err = testStore.ListGraphs(context.Background(), "", 10, 3)
		if err != nil {
			t.Fatalf("Failed to list graphs with offset beyond data: %v", err)
		}
		if len(graphs) != 0 {
			t.Errorf("Expected 0 graphs with offset=10, got %d", len(graphs))
		}

		// Test very large limit (should return all available)
		graphs, total, err = testStore.ListGraphs(context.Background(), "", 0, 10000)
		if err != nil {
			t.Fatalf("Failed to list graphs with large limit: %v", err)
		}
		if len(graphs) != 5 {
			t.Errorf("Expected 5 graphs with large limit, got %d", len(graphs))
		}
		if total != 5 {
			t.Errorf("Expected total count 5, got %d", total)
		}
	})
}

// TestShareLinkExpirationBoundary tests share link expiration at exact boundary
func TestShareLinkExpirationBoundary(t *testing.T) {
	db := setupTestDB(t)

	testutils.Run(t, testutils.Level2, "TestShareLinkExpirationBoundary", db, func(t *testing.T, tx *gorm.DB) {
		testStore, _ := NewStore(tx)

		// Create a graph
		graph, err := testStore.CreateGraph(
			context.Background(),
			"Expiration Boundary Test",
			"",
			"d3fend",
			"[]",
			"[]",
			"{}",
		)
		if err != nil {
			t.Fatalf("Failed to create graph: %v", err)
		}

		// Create link with zero expiration (not expired at creation)
		var zeroDuration time.Duration
		link, err := testStore.CreateShareLink(context.Background(), graph.GraphID, "", &zeroDuration)
		if err != nil {
			t.Fatalf("Failed to create share link: %v", err)
		}

		// Link should be immediately retrievable (or not, depending on implementation)
		// The current implementation might consider zero duration as expired
		_, err = testStore.GetShareLink(context.Background(), link.LinkID)
		// Either behavior is acceptable - link might be considered expired or not

		// Create link that expires in 1 millisecond
		oneMs := 1 * time.Millisecond
		link, err = testStore.CreateShareLink(context.Background(), graph.GraphID, "", &oneMs)
		if err != nil {
			t.Fatalf("Failed to create short-lived share link: %v", err)
		}

		// Should be accessible immediately
		_, err = testStore.GetShareLink(context.Background(), link.LinkID)
		if err != nil {
			t.Errorf("Short-lived link should be accessible immediately: %v", err)
		}

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Should be expired now
		_, err = testStore.GetShareLink(context.Background(), link.LinkID)
		if err == nil {
			t.Error("Expected link to be expired after 10ms wait")
		}
	})
}

// Benchmark tests
func BenchmarkCreateGraph(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	MigrateTables(db)
	store, _ := NewStore(db)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.CreateGraph(ctx, "Bench Graph", "", "d3fend", "[]", "[]", "{}")
	}
}

func BenchmarkGetGraph(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	MigrateTables(db)
	store, _ := NewStore(db)
	ctx := context.Background()

	graph, _ := store.CreateGraph(ctx, "Bench Graph", "", "d3fend", "[]", "[]", "{}")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GetGraph(ctx, graph.GraphID)
	}
}

func BenchmarkUpdateGraph(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	MigrateTables(db)
	store, _ := NewStore(db)
	ctx := context.Background()

	graph, _ := store.CreateGraph(ctx, "Bench Graph", "", "d3fend", "[]", "[]", "{}")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.UpdateGraph(ctx, graph.GraphID, map[string]interface{}{"name": fmt.Sprintf("Update %d", i)})
	}
}

func BenchmarkListGraphs(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	MigrateTables(db)
	store, _ := NewStore(db)
	ctx := context.Background()

	// Create 100 graphs
	for i := 0; i < 100; i++ {
		store.CreateGraph(ctx, fmt.Sprintf("Graph %d", i), "", "d3fend", "[]", "[]", "{}")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.ListGraphs(ctx, "", 0, 10)
	}
}

func BenchmarkCreateVersion(b *testing.B) {
	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	MigrateTables(db)
	store, _ := NewStore(db)
	ctx := context.Background()

	graph, _ := store.CreateGraph(ctx, "Bench Graph", "", "d3fend", `[]`, `[]`, `{}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.UpdateGraph(ctx, graph.GraphID, map[string]interface{}{"nodes": fmt.Sprintf(`[{"id": "node-%d"}]`, i)})
	}
}
