package graph

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"github.com/cyw0ng95/v2e/pkg/urn"
)

func TestGraphBasicOperations(t *testing.T) {
	testutils.Run(t, testutils.Level1, "BasicOperations", nil, func(t *testing.T, tx *gorm.DB) {
		g := New()

		cve1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe1, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")

		g.AddNode(cve1, map[string]interface{}{"severity": "HIGH"})
		g.AddNode(cwe1, map[string]interface{}{"name": "XSS"})

		if g.NodeCount() != 2 {
			t.Errorf("Expected 2 nodes, got %d", g.NodeCount())
		}

		node, exists := g.GetNode(cve1)
		if !exists {
			t.Error("Expected CVE node to exist")
		}
		if node.Properties["severity"] != "HIGH" {
			t.Error("Expected severity to be HIGH")
		}

		err := g.AddEdge(cve1, cwe1, EdgeTypeReferences, nil)
		if err != nil {
			t.Errorf("Failed to add edge: %v", err)
		}

		if g.EdgeCount() != 1 {
			t.Errorf("Expected 1 edge, got %d", g.EdgeCount())
		}

		outgoing := g.GetOutgoingEdges(cve1)
		if len(outgoing) != 1 {
			t.Errorf("Expected 1 outgoing edge, got %d", len(outgoing))
		}
		if outgoing[0].Type != EdgeTypeReferences {
			t.Errorf("Expected edge type to be references, got %s", outgoing[0].Type)
		}

		incoming := g.GetIncomingEdges(cwe1)
		if len(incoming) != 1 {
			t.Errorf("Expected 1 incoming edge, got %d", len(incoming))
		}

		neighbors := g.GetNeighbors(cve1)
		if len(neighbors) != 1 {
			t.Errorf("Expected 1 neighbor, got %d", len(neighbors))
		}
		if !neighbors[0].Equal(cwe1) {
			t.Error("Expected neighbor to be CWE-79")
		}
	})
}

func TestGraphEdgeRequiresNodes(t *testing.T) {
	testutils.Run(t, testutils.Level1, "EdgeRequiresNodes", nil, func(t *testing.T, tx *gorm.DB) {
		g := New()

		cve1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe1, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")

		err := g.AddEdge(cve1, cwe1, EdgeTypeReferences, nil)
		if err == nil {
			t.Error("Expected error when adding edge without nodes")
		}

		g.AddNode(cve1, nil)
		g.AddNode(cwe1, nil)
		err = g.AddEdge(cve1, cwe1, EdgeTypeReferences, nil)
		if err != nil {
			t.Errorf("Expected no error after adding nodes: %v", err)
		}
	})
}

func TestGraphGetNodesByType(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetNodesByType", nil, func(t *testing.T, tx *gorm.DB) {
		g := New()

		cve1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cve2, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-5678")
		cwe1, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")

		g.AddNode(cve1, nil)
		g.AddNode(cve2, nil)
		g.AddNode(cwe1, nil)

		cveNodes := g.GetNodesByType(urn.TypeCVE)
		if len(cveNodes) != 2 {
			t.Errorf("Expected 2 CVE nodes, got %d", len(cveNodes))
		}

		cweNodes := g.GetNodesByType(urn.TypeCWE)
		if len(cweNodes) != 1 {
			t.Errorf("Expected 1 CWE node, got %d", len(cweNodes))
		}
	})
}

func TestGraphGetNodesByProvider(t *testing.T) {
	testutils.Run(t, testutils.Level1, "GetNodesByProvider", nil, func(t *testing.T, tx *gorm.DB) {
		g := New()

		cve1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe1, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")
		capec1, _ := urn.New(urn.ProviderMITRE, urn.TypeCAPEC, "CAPEC-66")

		g.AddNode(cve1, nil)
		g.AddNode(cwe1, nil)
		g.AddNode(capec1, nil)

		nvdNodes := g.GetNodesByProvider(urn.ProviderNVD)
		if len(nvdNodes) != 1 {
			t.Errorf("Expected 1 NVD node, got %d", len(nvdNodes))
		}

		mitreNodes := g.GetNodesByProvider(urn.ProviderMITRE)
		if len(mitreNodes) != 2 {
			t.Errorf("Expected 2 MITRE nodes, got %d", len(mitreNodes))
		}
	})
}

func TestGraphFindPath(t *testing.T) {
	testutils.Run(t, testutils.Level1, "FindPath", nil, func(t *testing.T, tx *gorm.DB) {
		g := New()

		cve1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe1, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")
		capec1, _ := urn.New(urn.ProviderMITRE, urn.TypeCAPEC, "CAPEC-66")

		g.AddNode(cve1, nil)
		g.AddNode(cwe1, nil)
		g.AddNode(capec1, nil)

		g.AddEdge(cve1, cwe1, EdgeTypeReferences, nil)
		g.AddEdge(cwe1, capec1, EdgeTypeRelatedTo, nil)

		path, found := g.FindPath(cve1, capec1)
		if !found {
			t.Error("Expected to find path from CVE to CAPEC")
		}
		if len(path) != 3 {
			t.Errorf("Expected path length of 3, got %d", len(path))
		}
		if !path[0].Equal(cve1) || !path[1].Equal(cwe1) || !path[2].Equal(capec1) {
			t.Error("Path nodes don't match expected sequence")
		}

		_, found = g.FindPath(capec1, cve1)
		if found {
			t.Error("Should not find path in reverse direction for directed graph")
		}
	})
}

func TestGraphClear(t *testing.T) {
	testutils.Run(t, testutils.Level1, "Clear", nil, func(t *testing.T, tx *gorm.DB) {
		g := New()

		cve1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe1, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")

		g.AddNode(cve1, nil)
		g.AddNode(cwe1, nil)
		g.AddEdge(cve1, cwe1, EdgeTypeReferences, nil)

		g.Clear()

		if g.NodeCount() != 0 {
			t.Errorf("Expected 0 nodes after clear, got %d", g.NodeCount())
		}
		if g.EdgeCount() != 0 {
			t.Errorf("Expected 0 edges after clear, got %d", g.EdgeCount())
		}
	})
}

func TestGraphConcurrentAccess(t *testing.T) {
	testutils.Run(t, testutils.Level1, "ConcurrentAccess", nil, func(t *testing.T, tx *gorm.DB) {
		g := New()

		cve1, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
		cwe1, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")

		g.AddNode(cve1, nil)
		g.AddNode(cwe1, nil)
		g.AddEdge(cve1, cwe1, EdgeTypeReferences, nil)

		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_ = g.GetOutgoingEdges(cve1)
				_ = g.GetIncomingEdges(cwe1)
				_ = g.NodeCount()
				_ = g.EdgeCount()
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
