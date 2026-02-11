package graph

import (
	"fmt"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/urn"
)

func BenchmarkGraphAddNode(b *testing.B) {
	g := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		u, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		g.AddNode(u, map[string]interface{}{"id": i})
	}
}

func BenchmarkGraphGetNode(b *testing.B) {
	g := New()
	// Pre-populate graph
	nodes := make([]*urn.URN, 1000)
	for i := 0; i < 1000; i++ {
		u, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		nodes[i] = u
		g.AddNode(u, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetNode(nodes[i%1000])
	}
}

func BenchmarkGraphAddEdge(b *testing.B) {
	g := New()
	// Pre-populate nodes
	cves := make([]*urn.URN, 100)
	cwes := make([]*urn.URN, 100)
	for i := 0; i < 100; i++ {
		cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, fmt.Sprintf("CWE-%d", i))
		cves[i] = cve
		cwes[i] = cwe
		g.AddNode(cve, nil)
		g.AddNode(cwe, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		from := cves[i%100]
		to := cwes[i%100]
		g.AddEdge(from, to, EdgeTypeReferences, nil)
	}
}

func BenchmarkGraphGetOutgoingEdges(b *testing.B) {
	g := New()
	// Create a node with multiple outgoing edges
	cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
	g.AddNode(cve, nil)

	for i := 0; i < 50; i++ {
		cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, fmt.Sprintf("CWE-%d", i))
		g.AddNode(cwe, nil)
		g.AddEdge(cve, cwe, EdgeTypeReferences, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetOutgoingEdges(cve)
	}
}

func BenchmarkGraphGetNeighbors(b *testing.B) {
	g := New()
	cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
	g.AddNode(cve, nil)

	// Add outgoing neighbors
	for i := 0; i < 25; i++ {
		cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, fmt.Sprintf("CWE-%d", i))
		g.AddNode(cwe, nil)
		g.AddEdge(cve, cwe, EdgeTypeReferences, nil)
	}

	// Add incoming neighbors
	for i := 0; i < 25; i++ {
		other, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2023-%d", i))
		g.AddNode(other, nil)
		g.AddEdge(other, cve, EdgeTypeRelatedTo, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetNeighbors(cve)
	}
}

func BenchmarkGraphGetNodesByType(b *testing.B) {
	g := New()
	// Add mixed node types
	for i := 0; i < 500; i++ {
		cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		g.AddNode(cve, nil)
	}
	for i := 0; i < 500; i++ {
		cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, fmt.Sprintf("CWE-%d", i))
		g.AddNode(cwe, nil)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.GetNodesByType(urn.TypeCVE)
	}
}

func BenchmarkGraphFindPath(b *testing.B) {
	g := New()
	// Create a chain: CVE -> CWE -> CAPEC -> ATT&CK
	cve, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, "CVE-2024-1234")
	cwe, _ := urn.New(urn.ProviderMITRE, urn.TypeCWE, "CWE-79")
	capec, _ := urn.New(urn.ProviderMITRE, urn.TypeCAPEC, "CAPEC-66")
	attack, _ := urn.New(urn.ProviderMITRE, urn.TypeATTACK, "T1566")

	g.AddNode(cve, nil)
	g.AddNode(cwe, nil)
	g.AddNode(capec, nil)
	g.AddNode(attack, nil)

	g.AddEdge(cve, cwe, EdgeTypeReferences, nil)
	g.AddEdge(cwe, capec, EdgeTypeRelatedTo, nil)
	g.AddEdge(capec, attack, EdgeTypeRelatedTo, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.FindPath(cve, attack)
	}
}

func BenchmarkGraphConcurrentReads(b *testing.B) {
	g := New()
	// Populate graph
	nodes := make([]*urn.URN, 100)
	for i := 0; i < 100; i++ {
		u, _ := urn.New(urn.ProviderNVD, urn.TypeCVE, fmt.Sprintf("CVE-2024-%d", i))
		nodes[i] = u
		g.AddNode(u, nil)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			g.GetNode(nodes[i%100])
			i++
		}
	})
}
