package main

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

func TestNewAnalysisService(t *testing.T) {
	testutils.Run(t, testutils.Level1, "NewAnalysisService", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		dbPath := "/tmp/test_analysis_service.db"
		defer os.Remove(dbPath)

		service, err := NewAnalysisService(nil, logger, dbPath)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
		defer service.Close()

		if service == nil {
			t.Fatal("Expected service to be created")
		}

		if service.graph == nil {
			t.Error("Expected graph to be initialized")
		}

		if service.logger == nil {
			t.Error("Expected logger to be set")
		}
	})
}

func TestAnalysisServiceGraphOperations(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GraphOperations", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		dbPath := "/tmp/test_graph_ops.db"
		defer os.Remove(dbPath)

		service, err := NewAnalysisService(nil, logger, dbPath)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
		defer service.Close()

		// Test adding nodes
		cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")

		service.graph.AddNode(cve, map[string]interface{}{"severity": "HIGH"})
		service.graph.AddNode(cwe, map[string]interface{}{"name": "XSS"})

		if service.graph.NodeCount() != 2 {
			t.Errorf("Expected 2 nodes, got %d", service.graph.NodeCount())
		}

		// Test adding edges
		err = service.graph.AddEdge(cve, cwe, graph.EdgeTypeReferences, nil)
		if err != nil {
			t.Errorf("Failed to add edge: %v", err)
		}

		if service.graph.EdgeCount() != 1 {
			t.Errorf("Expected 1 edge, got %d", service.graph.EdgeCount())
		}

		// Test getting neighbors
		neighbors := service.graph.GetNeighbors(cve)
		if len(neighbors) != 1 {
			t.Errorf("Expected 1 neighbor, got %d", len(neighbors))
		}

		// Test clearing graph
		service.graph.Clear()
		if service.graph.NodeCount() != 0 {
			t.Errorf("Expected 0 nodes after clear, got %d", service.graph.NodeCount())
		}
	})
}

func TestAnalysisServiceMultipleEdgeTypes(t *testing.T) {
	testutils.Run(t, testutils.Level1, "MultipleEdgeTypes", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		dbPath := "/tmp/test_edge_types.db"
		defer os.Remove(dbPath)

		service, err := NewAnalysisService(nil, logger, dbPath)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
		defer service.Close()

		cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")
		capec, _ := urn.New(urn.ProviderMITRE, urn.TypeCAPEC, "CAPEC-66")

		service.graph.AddNode(cve, nil)
		service.graph.AddNode(cwe, nil)
		service.graph.AddNode(capec, nil)

		// Add different edge types
		service.graph.AddEdge(cve, cwe, graph.EdgeTypeReferences, nil)
		service.graph.AddEdge(cwe, capec, graph.EdgeTypeRelatedTo, nil)

		// Verify edges
		outgoing := service.graph.GetOutgoingEdges(cve)
		if len(outgoing) != 1 || outgoing[0].Type != graph.EdgeTypeReferences {
			t.Error("Expected one 'references' edge from CVE")
		}

		outgoing = service.graph.GetOutgoingEdges(cwe)
		if len(outgoing) != 1 || outgoing[0].Type != graph.EdgeTypeRelatedTo {
			t.Error("Expected one 'related_to' edge from CWE")
		}
	})
}

func TestAnalysisServicePathFinding(t *testing.T) {
	testutils.Run(t, testutils.Level1, "PathFinding", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		dbPath := "/tmp/test_path_finding.db"
		defer os.Remove(dbPath)

		service, err := NewAnalysisService(nil, logger, dbPath)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
		defer service.Close()

		// Build a chain: CVE -> CWE -> CAPEC -> ATT&CK
		cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")
		capec, _ := urn.New(urn.ProviderMITRE, urn.TypeCAPEC, "CAPEC-66")
		attack, _ := urn.New(urn.ProviderMITRE, urn.TypeATTACK, "T1566")

		service.graph.AddNode(cve, nil)
		service.graph.AddNode(cwe, nil)
		service.graph.AddNode(capec, nil)
		service.graph.AddNode(attack, nil)

		service.graph.AddEdge(cve, cwe, graph.EdgeTypeReferences, nil)
		service.graph.AddEdge(cwe, capec, graph.EdgeTypeRelatedTo, nil)
		service.graph.AddEdge(capec, attack, graph.EdgeTypeExploits, nil)

		// Find path from CVE to ATT&CK
		path, found := service.graph.FindPath(cve, attack)
		if !found {
			t.Error("Expected to find path from CVE to ATT&CK")
		}
		if len(path) != 4 {
			t.Errorf("Expected path length of 4, got %d", len(path))
		}

		// Verify path order
		if !path[0].Equal(cve) {
			t.Error("Path should start with CVE")
		}
		if !path[3].Equal(attack) {
			t.Error("Path should end with ATT&CK")
		}
	})
}

func TestAnalysisServiceGetNodesByType(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetNodesByType", nil, func(t *testing.T, tx *gorm.DB) {
		logger := common.NewLogger(os.Stdout, "test", common.InfoLevel)
		dbPath := "/tmp/test_nodes_by_type.db"
		defer os.Remove(dbPath)

		service, err := NewAnalysisService(nil, logger, dbPath)
		if err != nil {
			t.Fatalf("Failed to create service: %v", err)
		}
		defer service.Close()

		// Add multiple nodes of different types
		for i := 0; i < 5; i++ {
			cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%04d", i))
			service.graph.AddNode(cve, nil)
		}

		for i := 0; i < 3; i++ {
			cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, fmt.Sprintf("CWE-%d", 79+i))
			service.graph.AddNode(cwe, nil)
		}

		cveNodes := service.graph.GetNodesByType(urn.TypeCVE)
		if len(cveNodes) != 5 {
			t.Errorf("Expected 5 CVE nodes, got %d", len(cveNodes))
		}

		cweNodes := service.graph.GetNodesByType(urn.TypeCWE)
		if len(cweNodes) != 3 {
			t.Errorf("Expected 3 CWE nodes, got %d", len(cweNodes))
		}
	})
}
