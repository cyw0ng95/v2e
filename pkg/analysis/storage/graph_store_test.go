package storage

import (
	"fmt"
	"os"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/graph"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

func BenchmarkSaveGraph(b *testing.B) {
	logger := common.NewLogger(os.Stdout, "bench", common.WarnLevel)
	
	// Create a test graph with nodes and edges
	g := graph.New()
	
	// Add 1000 nodes
	for i := 0; i < 1000; i++ {
		u, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		g.AddNode(u, map[string]interface{}{
			"id":          fmt.Sprintf("CVE-2024-%d", i),
			"severity":    "HIGH",
			"description": "Test vulnerability description",
		})
	}
	
	// Add 2000 edges
	for i := 0; i < 1000; i++ {
		fromURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		
		// Add 2 edges per CVE
		for j := 0; j < 2; j++ {
			cweID := fmt.Sprintf("CWE-%d", (i+j)%100)
			toURN, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, cweID)
			
			// Add CWE node if needed
			if _, exists := g.GetNode(toURN); !exists {
				g.AddNode(toURN, map[string]interface{}{"id": cweID})
			}
			
			g.AddEdge(fromURN, toURN, graph.EdgeTypeReferences, nil)
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		tmpFile := fmt.Sprintf("/tmp/bench_graph_%d.db", i)
		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			b.Fatalf("Failed to create store: %v", err)
		}
		
		if err := store.SaveGraph(g); err != nil {
			b.Fatalf("Failed to save graph: %v", err)
		}
		
		store.Close()
		os.Remove(tmpFile)
	}
}

func BenchmarkLoadGraph(b *testing.B) {
	logger := common.NewLogger(os.Stdout, "bench", common.WarnLevel)
	
	// Create and save a test graph once
	g := graph.New()
	
	for i := 0; i < 1000; i++ {
		u, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		g.AddNode(u, map[string]interface{}{
			"id":       fmt.Sprintf("CVE-2024-%d", i),
			"severity": "HIGH",
		})
	}
	
	for i := 0; i < 1000; i++ {
		fromURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		for j := 0; j < 2; j++ {
			cweID := fmt.Sprintf("CWE-%d", (i+j)%100)
			toURN, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, cweID)
			if _, exists := g.GetNode(toURN); !exists {
				g.AddNode(toURN, map[string]interface{}{"id": cweID})
			}
			g.AddEdge(fromURN, toURN, graph.EdgeTypeReferences, nil)
		}
	}
	
	tmpFile := "/tmp/bench_load_graph.db"
	store, err := NewGraphStore(tmpFile, logger)
	if err != nil {
		b.Fatalf("Failed to create store: %v", err)
	}
	
	if err := store.SaveGraph(g); err != nil {
		b.Fatalf("Failed to save graph: %v", err)
	}
	store.Close()
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		store, err := NewGraphStore(tmpFile, logger)
		if err != nil {
			b.Fatalf("Failed to create store: %v", err)
		}
		
		if _, err := store.LoadGraph(); err != nil {
			b.Fatalf("Failed to load graph: %v", err)
		}
		
		store.Close()
	}
	
	os.Remove(tmpFile)
}

func TestGraphStore_SaveAndLoad(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
	tmpFile := "/tmp/test_graph_store.db"
	defer os.Remove(tmpFile)
	
	// Create a test graph
	g := graph.New()
	
	// Add some nodes
	cveURN, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
	g.AddNode(cveURN, map[string]interface{}{
		"id":       "CVE-2024-1234",
		"severity": "HIGH",
	})
	
	cweURN, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")
	g.AddNode(cweURN, map[string]interface{}{
		"id":   "CWE-79",
		"name": "Cross-site Scripting",
	})
	
	// Add an edge
	if err := g.AddEdge(cveURN, cweURN, graph.EdgeTypeReferences, nil); err != nil {
		t.Fatalf("Failed to add edge: %v", err)
	}
	
	// Save the graph
	store, err := NewGraphStore(tmpFile, logger)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	
	if err := store.SaveGraph(g); err != nil {
		t.Fatalf("Failed to save graph: %v", err)
	}
	
	// Check metadata
	metadata, err := store.GetMetadata()
	if err != nil {
		t.Fatalf("Failed to get metadata: %v", err)
	}
	
	if metadata.NodeCount != 2 {
		t.Errorf("Expected 2 nodes, got %d", metadata.NodeCount)
	}
	
	if metadata.EdgeCount != 1 {
		t.Errorf("Expected 1 edge, got %d", metadata.EdgeCount)
	}
	
	store.Close()
	
	// Load the graph
	store2, err := NewGraphStore(tmpFile, logger)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store2.Close()
	
	loadedGraph, err := store2.LoadGraph()
	if err != nil {
		t.Fatalf("Failed to load graph: %v", err)
	}
	
	// Verify loaded graph
	if loadedGraph.NodeCount() != 2 {
		t.Errorf("Expected 2 nodes, got %d", loadedGraph.NodeCount())
	}
	
	if loadedGraph.EdgeCount() != 1 {
		t.Errorf("Expected 1 edge, got %d", loadedGraph.EdgeCount())
	}
	
	// Verify node data
	node, exists := loadedGraph.GetNode(cveURN)
	if !exists {
		t.Error("CVE node not found in loaded graph")
	}
	
	if severity, ok := node.Properties["severity"].(string); !ok || severity != "HIGH" {
		t.Errorf("Expected severity HIGH, got %v", node.Properties["severity"])
	}
}

func TestGraphStore_Checkpoint(t *testing.T) {
	logger := common.NewLogger(os.Stdout, "test", common.WarnLevel)
	tmpFile := "/tmp/test_checkpoint.db"
	defer os.Remove(tmpFile)
	
	store, err := NewGraphStore(tmpFile, logger)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()
	
	// Save a checkpoint
	checkpointData := map[string]interface{}{
		"last_processed": "CVE-2024-1234",
		"count":          100,
	}
	
	if err := store.SaveCheckpoint("test_checkpoint", checkpointData); err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}
	
	// Load the checkpoint
	loaded, err := store.GetCheckpoint("test_checkpoint")
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}
	
	data := loaded["data"].(map[string]interface{})
	if data["last_processed"] != "CVE-2024-1234" {
		t.Errorf("Checkpoint data mismatch")
	}
}
