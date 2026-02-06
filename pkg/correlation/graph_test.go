package correlation

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCorrelationGraph(t *testing.T) {
	g := NewCorrelationGraph()
	assert.NotNil(t, g)
	assert.Empty(t, g.nodes)
	assert.Empty(t, g.reverseEdges)
}

func TestAddNode(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Add CVE node
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	
	node, exists := g.GetNode(cveURN)
	require.True(t, exists)
	assert.Equal(t, cveURN, node.URN)
	assert.Equal(t, URNTypeCVE, node.Type)
	assert.Equal(t, "CVE-2021-44228", node.ID)
	
	// Check statistics
	stats := g.GetStats()
	assert.Equal(t, 1, stats.NodeCount)
	assert.Equal(t, 1, stats.CVECount)
	assert.Equal(t, 0, stats.EdgeCount)
}

func TestAddEdge(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Add nodes
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cweURN := BuildURN(URNTypeCWE, "CWE-502")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cweURN, URNTypeCWE, "CWE-502")
	
	// Add edge
	err := g.AddEdge(cveURN, cweURN)
	require.NoError(t, err)
	
	// Verify edge exists
	edges := g.GetOutgoingEdges(cveURN)
	require.Len(t, edges, 1)
	assert.Equal(t, cweURN, edges[0])
	
	// Verify reverse edge
	reverseEdges := g.GetIncomingEdges(cweURN)
	require.Len(t, reverseEdges, 1)
	assert.Equal(t, cveURN, reverseEdges[0])
	
	// Check statistics
	stats := g.GetStats()
	assert.Equal(t, 1, stats.EdgeCount)
}

func TestAddEdgeErrors(t *testing.T) {
	g := NewCorrelationGraph()
	
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cweURN := BuildURN(URNTypeCWE, "CWE-502")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	
	// Try to add edge to non-existent target
	err := g.AddEdge(cveURN, cweURN)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	
	// Try to add edge from non-existent source
	g.AddNode(cweURN, URNTypeCWE, "CWE-502")
	err = g.AddEdge("nonexistent", cweURN)
	assert.Error(t, err)
}

func TestDuplicateEdge(t *testing.T) {
	g := NewCorrelationGraph()
	
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cweURN := BuildURN(URNTypeCWE, "CWE-502")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cweURN, URNTypeCWE, "CWE-502")
	
	// Add edge twice
	err := g.AddEdge(cveURN, cweURN)
	require.NoError(t, err)
	
	err = g.AddEdge(cveURN, cweURN)
	require.NoError(t, err)
	
	// Should still have only one edge
	edges := g.GetOutgoingEdges(cveURN)
	assert.Len(t, edges, 1)
}

func TestTraverseCVEToCWE(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Build graph: CVE-2021-44228 → CWE-502, CWE-20
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cwe1URN := BuildURN(URNTypeCWE, "CWE-502")
	cwe2URN := BuildURN(URNTypeCWE, "CWE-20")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cwe1URN, URNTypeCWE, "CWE-502")
	g.AddNode(cwe2URN, URNTypeCWE, "CWE-20")
	
	g.AddEdge(cveURN, cwe1URN)
	g.AddEdge(cveURN, cwe2URN)
	
	// Traverse
	cwes := g.TraverseCVEToCWE(cveURN)
	require.Len(t, cwes, 2)
	assert.Contains(t, cwes, cwe1URN)
	assert.Contains(t, cwes, cwe2URN)
}

func TestTraverseCWEToCAPEC(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Build graph: CWE-502 ← CAPEC-586, CAPEC-588
	cweURN := BuildURN(URNTypeCWE, "CWE-502")
	capec1URN := BuildURN(URNTypeCAPEC, "CAPEC-586")
	capec2URN := BuildURN(URNTypeCAPEC, "CAPEC-588")
	
	g.AddNode(cweURN, URNTypeCWE, "CWE-502")
	g.AddNode(capec1URN, URNTypeCAPEC, "CAPEC-586")
	g.AddNode(capec2URN, URNTypeCAPEC, "CAPEC-588")
	
	// CAPEC → CWE edges
	g.AddEdge(capec1URN, cweURN)
	g.AddEdge(capec2URN, cweURN)
	
	// Traverse
	capecs := g.TraverseCWEToCAPEC(cweURN)
	require.Len(t, capecs, 2)
	assert.Contains(t, capecs, capec1URN)
	assert.Contains(t, capecs, capec2URN)
}

func TestTraverseCVEToCAPEC(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Build complete graph:
	// CVE-2021-44228 → CWE-502, CWE-20
	// CAPEC-586, CAPEC-588 → CWE-502
	// CAPEC-10 → CWE-20
	
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cwe1URN := BuildURN(URNTypeCWE, "CWE-502")
	cwe2URN := BuildURN(URNTypeCWE, "CWE-20")
	capec1URN := BuildURN(URNTypeCAPEC, "CAPEC-586")
	capec2URN := BuildURN(URNTypeCAPEC, "CAPEC-588")
	capec3URN := BuildURN(URNTypeCAPEC, "CAPEC-10")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cwe1URN, URNTypeCWE, "CWE-502")
	g.AddNode(cwe2URN, URNTypeCWE, "CWE-20")
	g.AddNode(capec1URN, URNTypeCAPEC, "CAPEC-586")
	g.AddNode(capec2URN, URNTypeCAPEC, "CAPEC-588")
	g.AddNode(capec3URN, URNTypeCAPEC, "CAPEC-10")
	
	// CVE → CWE
	g.AddEdge(cveURN, cwe1URN)
	g.AddEdge(cveURN, cwe2URN)
	
	// CAPEC → CWE
	g.AddEdge(capec1URN, cwe1URN)
	g.AddEdge(capec2URN, cwe1URN)
	g.AddEdge(capec3URN, cwe2URN)
	
	// Traverse CVE → CAPEC
	capecs := g.TraverseCVEToCAPEC(cveURN)
	require.Len(t, capecs, 3)
	assert.Contains(t, capecs, capec1URN)
	assert.Contains(t, capecs, capec2URN)
	assert.Contains(t, capecs, capec3URN)
}

func TestGetFullCorrelation(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Build complete graph
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cwe1URN := BuildURN(URNTypeCWE, "CWE-502")
	cwe2URN := BuildURN(URNTypeCWE, "CWE-20")
	capec1URN := BuildURN(URNTypeCAPEC, "CAPEC-586")
	capec2URN := BuildURN(URNTypeCAPEC, "CAPEC-588")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cwe1URN, URNTypeCWE, "CWE-502")
	g.AddNode(cwe2URN, URNTypeCWE, "CWE-20")
	g.AddNode(capec1URN, URNTypeCAPEC, "CAPEC-586")
	g.AddNode(capec2URN, URNTypeCAPEC, "CAPEC-588")
	
	g.AddEdge(cveURN, cwe1URN)
	g.AddEdge(cveURN, cwe2URN)
	g.AddEdge(capec1URN, cwe1URN)
	g.AddEdge(capec2URN, cwe2URN)
	
	// Get full correlation
	path := g.GetFullCorrelation(cveURN)
	require.NotNil(t, path)
	assert.Equal(t, cveURN, path.CVE)
	assert.Len(t, path.CWEs, 2)
	assert.Len(t, path.CAPECs, 2)
	assert.Contains(t, path.CWEs, cwe1URN)
	assert.Contains(t, path.CWEs, cwe2URN)
	assert.Contains(t, path.CAPECs, capec1URN)
	assert.Contains(t, path.CAPECs, capec2URN)
}

func TestGetFullCorrelationNonExistent(t *testing.T) {
	g := NewCorrelationGraph()
	
	path := g.GetFullCorrelation("nonexistent")
	require.NotNil(t, path)
	assert.Empty(t, path.CWEs)
	assert.Empty(t, path.CAPECs)
}

func TestGraphStatistics(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Add nodes and edges
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cweURN := BuildURN(URNTypeCWE, "CWE-502")
	capecURN := BuildURN(URNTypeCAPEC, "CAPEC-586")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cweURN, URNTypeCWE, "CWE-502")
	g.AddNode(capecURN, URNTypeCAPEC, "CAPEC-586")
	
	g.AddEdge(cveURN, cweURN)
	g.AddEdge(capecURN, cweURN)
	
	stats := g.GetStats()
	assert.Equal(t, 3, stats.NodeCount)
	assert.Equal(t, 1, stats.CVECount)
	assert.Equal(t, 1, stats.CWECount)
	assert.Equal(t, 1, stats.CAPECCount)
	assert.Equal(t, 2, stats.EdgeCount)
}

func TestClear(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Add some data
	cveURN := BuildURN(URNTypeCVE, "CVE-2021-44228")
	cweURN := BuildURN(URNTypeCWE, "CWE-502")
	
	g.AddNode(cveURN, URNTypeCVE, "CVE-2021-44228")
	g.AddNode(cweURN, URNTypeCWE, "CWE-502")
	g.AddEdge(cveURN, cweURN)
	
	// Verify data exists
	stats := g.GetStats()
	assert.Equal(t, 2, stats.NodeCount)
	
	// Clear
	g.Clear()
	
	// Verify empty
	stats = g.GetStats()
	assert.Equal(t, 0, stats.NodeCount)
	assert.Equal(t, 0, stats.EdgeCount)
	
	_, exists := g.GetNode(cveURN)
	assert.False(t, exists)
}

func TestBuildURN(t *testing.T) {
	tests := []struct {
		name     string
		urnType  URNType
		id       string
		expected string
	}{
		{"CVE", URNTypeCVE, "CVE-2021-44228", "v2e::nvd::cve::CVE-2021-44228"},
		{"CWE", URNTypeCWE, "CWE-502", "v2e::mitre::cwe::CWE-502"},
		{"CAPEC", URNTypeCAPEC, "CAPEC-586", "v2e::mitre::capec::CAPEC-586"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urn := BuildURN(tt.urnType, tt.id)
			assert.Equal(t, tt.expected, urn)
		})
	}
}

// Benchmark to ensure <5ms lookup performance
func BenchmarkGetFullCorrelation(b *testing.B) {
	g := NewCorrelationGraph()
	
	// Build a realistic graph with 1000 CVEs, 500 CWEs, 1000 CAPECs
	for i := 0; i < 1000; i++ {
		cveURN := BuildURN(URNTypeCVE, fmt.Sprintf("CVE-2021-%05d", i))
		g.AddNode(cveURN, URNTypeCVE, fmt.Sprintf("CVE-2021-%05d", i))
	}
	
	for i := 0; i < 500; i++ {
		cweURN := BuildURN(URNTypeCWE, fmt.Sprintf("CWE-%d", i))
		g.AddNode(cweURN, URNTypeCWE, fmt.Sprintf("CWE-%d", i))
	}
	
	for i := 0; i < 1000; i++ {
		capecURN := BuildURN(URNTypeCAPEC, fmt.Sprintf("CAPEC-%d", i))
		g.AddNode(capecURN, URNTypeCAPEC, fmt.Sprintf("CAPEC-%d", i))
	}
	
	// Create edges: each CVE → 2-3 CWEs, each CAPEC → 1-2 CWEs
	for i := 0; i < 1000; i++ {
		cveURN := BuildURN(URNTypeCVE, fmt.Sprintf("CVE-2021-%05d", i))
		cwe1URN := BuildURN(URNTypeCWE, fmt.Sprintf("CWE-%d", i%500))
		cwe2URN := BuildURN(URNTypeCWE, fmt.Sprintf("CWE-%d", (i+1)%500))
		g.AddEdge(cveURN, cwe1URN)
		g.AddEdge(cveURN, cwe2URN)
	}
	
	for i := 0; i < 1000; i++ {
		capecURN := BuildURN(URNTypeCAPEC, fmt.Sprintf("CAPEC-%d", i))
		cweURN := BuildURN(URNTypeCWE, fmt.Sprintf("CWE-%d", i%500))
		g.AddEdge(capecURN, cweURN)
	}
	
	testCVE := BuildURN(URNTypeCVE, "CVE-2021-00500")
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		_ = g.GetFullCorrelation(testCVE)
	}
}

// Test concurrent access
func TestConcurrentAccess(t *testing.T) {
	g := NewCorrelationGraph()
	
	// Pre-populate graph
	for i := 0; i < 100; i++ {
		cveURN := BuildURN(URNTypeCVE, fmt.Sprintf("CVE-2021-%05d", i))
		cweURN := BuildURN(URNTypeCWE, fmt.Sprintf("CWE-%d", i))
		g.AddNode(cveURN, URNTypeCVE, fmt.Sprintf("CVE-2021-%05d", i))
		g.AddNode(cweURN, URNTypeCWE, fmt.Sprintf("CWE-%d", i))
		g.AddEdge(cveURN, cweURN)
	}
	
	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				cveURN := BuildURN(URNTypeCVE, fmt.Sprintf("CVE-2021-%05d", j))
				_ = g.GetFullCorrelation(cveURN)
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("Concurrent access test timed out")
		}
	}
}
