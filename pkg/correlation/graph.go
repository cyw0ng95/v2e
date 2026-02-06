package correlation

import (
	"fmt"
	"sync"
	"time"
)

// URNType represents the type of URN in the correlation graph
type URNType string

const (
	// URNTypeCVE represents v2e::nvd::cve URN
	URNTypeCVE URNType = "v2e::nvd::cve"
	// URNTypeCWE represents v2e::mitre::cwe URN
	URNTypeCWE URNType = "v2e::mitre::cwe"
	// URNTypeCAPEC represents v2e::mitre::capec URN
	URNTypeCAPEC URNType = "v2e::mitre::capec"
)

// Node represents a node in the correlation graph
type Node struct {
	URN         string    // Full URN (e.g., "v2e::nvd::cve::CVE-2021-44228")
	Type        URNType   // Type of the URN
	ID          string    // ID portion (e.g., "CVE-2021-44228")
	Edges       []string  // Outgoing edges (URNs this node links to)
	LastUpdated time.Time // Last update timestamp
}

// CorrelationGraph is a high-performance graph structure for CVE→CWE→CAPEC correlations
type CorrelationGraph struct {
	// nodes stores all nodes indexed by URN
	nodes map[string]*Node
	
	// reverseEdges stores reverse index: URN -> list of URNs that point to it
	reverseEdges map[string][]string
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Statistics
	stats GraphStats
}

// GraphStats contains statistics about the correlation graph
type GraphStats struct {
	NodeCount      int           // Total number of nodes
	EdgeCount      int           // Total number of edges
	CVECount       int           // Number of CVE nodes
	CWECount       int           // Number of CWE nodes
	CAPECCount     int           // Number of CAPEC nodes
	LastBuildTime  time.Time     // Last time the graph was built
	BuildDuration  time.Duration // Time taken to build the graph
}

// NewCorrelationGraph creates a new correlation graph
func NewCorrelationGraph() *CorrelationGraph {
	return &CorrelationGraph{
		nodes:        make(map[string]*Node),
		reverseEdges: make(map[string][]string),
	}
}

// AddNode adds a node to the graph
func (g *CorrelationGraph) AddNode(urn string, urnType URNType, id string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	if _, exists := g.nodes[urn]; !exists {
		g.nodes[urn] = &Node{
			URN:         urn,
			Type:        urnType,
			ID:          id,
			Edges:       make([]string, 0),
			LastUpdated: time.Now(),
		}
		
		// Update statistics
		switch urnType {
		case URNTypeCVE:
			g.stats.CVECount++
		case URNTypeCWE:
			g.stats.CWECount++
		case URNTypeCAPEC:
			g.stats.CAPECCount++
		}
		g.stats.NodeCount++
	}
}

// AddEdge adds a directed edge from sourceURN to targetURN
func (g *CorrelationGraph) AddEdge(sourceURN, targetURN string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	sourceNode, sourceExists := g.nodes[sourceURN]
	if !sourceExists {
		return fmt.Errorf("source node %s not found", sourceURN)
	}
	
	// Check if target exists
	if _, targetExists := g.nodes[targetURN]; !targetExists {
		return fmt.Errorf("target node %s not found", targetURN)
	}
	
	// Add edge if it doesn't already exist
	for _, edge := range sourceNode.Edges {
		if edge == targetURN {
			return nil // Edge already exists
		}
	}
	
	sourceNode.Edges = append(sourceNode.Edges, targetURN)
	g.stats.EdgeCount++
	
	// Update reverse index
	g.reverseEdges[targetURN] = append(g.reverseEdges[targetURN], sourceURN)
	
	return nil
}

// GetNode retrieves a node by URN
func (g *CorrelationGraph) GetNode(urn string) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	node, exists := g.nodes[urn]
	return node, exists
}

// GetOutgoingEdges returns all outgoing edges from a node
func (g *CorrelationGraph) GetOutgoingEdges(urn string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	node, exists := g.nodes[urn]
	if !exists {
		return nil
	}
	
	// Return a copy to prevent external modification
	edges := make([]string, len(node.Edges))
	copy(edges, node.Edges)
	return edges
}

// GetIncomingEdges returns all incoming edges to a node
func (g *CorrelationGraph) GetIncomingEdges(urn string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	edges, exists := g.reverseEdges[urn]
	if !exists {
		return nil
	}
	
	// Return a copy
	result := make([]string, len(edges))
	copy(result, edges)
	return result
}

// TraverseCVEToCWE finds all CWE nodes reachable from a CVE node
func (g *CorrelationGraph) TraverseCVEToCWE(cveURN string) []string {
	return g.GetOutgoingEdges(cveURN)
}

// TraverseCWEToCAPEC finds all CAPEC nodes reachable from a CWE node
func (g *CorrelationGraph) TraverseCWEToCAPEC(cweURN string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	// Get all CAPEC nodes that have this CWE as an incoming edge
	incoming := g.reverseEdges[cweURN]
	
	capecs := make([]string, 0, len(incoming))
	for _, urn := range incoming {
		if node, exists := g.nodes[urn]; exists && node.Type == URNTypeCAPEC {
			capecs = append(capecs, urn)
		}
	}
	
	return capecs
}

// TraverseCVEToCAPEC finds all CAPEC nodes reachable from a CVE via CWE
func (g *CorrelationGraph) TraverseCVEToCAPEC(cveURN string) []string {
	// Get all CWEs linked to this CVE
	cwes := g.TraverseCVEToCWE(cveURN)
	
	// Collect all unique CAPECs from all CWEs
	capecMap := make(map[string]bool)
	for _, cweURN := range cwes {
		capecs := g.TraverseCWEToCAPEC(cweURN)
		for _, capecURN := range capecs {
			capecMap[capecURN] = true
		}
	}
	
	// Convert map to slice
	result := make([]string, 0, len(capecMap))
	for capecURN := range capecMap {
		result = append(result, capecURN)
	}
	
	return result
}

// GetFullCorrelation returns the complete correlation path: CVE → CWEs → CAPECs
func (g *CorrelationGraph) GetFullCorrelation(cveURN string) *CorrelationPath {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	path := &CorrelationPath{
		CVE:    cveURN,
		CWEs:   make([]string, 0),
		CAPECs: make([]string, 0),
	}
	
	// Get CVE node
	cveNode, exists := g.nodes[cveURN]
	if !exists {
		return path
	}
	
	// Get CWEs
	path.CWEs = make([]string, len(cveNode.Edges))
	copy(path.CWEs, cveNode.Edges)
	
	// Get unique CAPECs from all CWEs
	capecMap := make(map[string]bool)
	for _, cweURN := range path.CWEs {
		// Get incoming edges to CWE (these are CAPECs)
		incoming := g.reverseEdges[cweURN]
		for _, urn := range incoming {
			if node, ok := g.nodes[urn]; ok && node.Type == URNTypeCAPEC {
				capecMap[urn] = true
			}
		}
	}
	
	for capecURN := range capecMap {
		path.CAPECs = append(path.CAPECs, capecURN)
	}
	
	return path
}

// CorrelationPath represents a full correlation path
type CorrelationPath struct {
	CVE    string   // CVE URN
	CWEs   []string // CWE URNs
	CAPECs []string // CAPEC URNs
}

// GetStats returns current graph statistics
func (g *CorrelationGraph) GetStats() GraphStats {
	g.mu.RLock()
	defer g.mu.RUnlock()
	
	return g.stats
}

// Clear removes all nodes and edges from the graph
func (g *CorrelationGraph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()
	
	g.nodes = make(map[string]*Node)
	g.reverseEdges = make(map[string][]string)
	g.stats = GraphStats{}
}

// BuildURN constructs a URN from type and ID
func BuildURN(urnType URNType, id string) string {
	return string(urnType) + "::" + id
}
