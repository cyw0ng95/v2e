package graph

import (
	"fmt"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/urn"
)

// EdgeType represents the type of relationship between nodes
type EdgeType string

const (
	// EdgeTypeReferences indicates a reference relationship (e.g., CVE references CWE)
	EdgeTypeReferences EdgeType = "references"
	// EdgeTypeRelatedTo indicates a general relationship
	EdgeTypeRelatedTo EdgeType = "related_to"
	// EdgeTypeMitigates indicates a mitigation relationship
	EdgeTypeMitigates EdgeType = "mitigates"
	// EdgeTypeExploits indicates an exploitation relationship
	EdgeTypeExploits EdgeType = "exploits"
	// EdgeTypeContains indicates a containment relationship
	EdgeTypeContains EdgeType = "contains"
)

// Node represents a graph node identified by a URN
type Node struct {
	URN        *urn.URN
	Properties map[string]interface{}
}

// Edge represents a directed edge between two nodes
type Edge struct {
	From       *urn.URN
	To         *urn.URN
	Type       EdgeType
	Properties map[string]interface{}
}

// Graph is an in-memory graph database for URN-based relationships
type Graph struct {
	mu           sync.RWMutex
	nodes        map[string]*Node   // key is URN.Key()
	edges        map[string][]*Edge // key is from URN.Key()
	reverseEdges map[string][]*Edge // key is to URN.Key(), for reverse lookups
}

// New creates a new empty graph
func New() *Graph {
	return &Graph{
		nodes:        make(map[string]*Node),
		edges:        make(map[string][]*Edge),
		reverseEdges: make(map[string][]*Edge),
	}
}

// AddNode adds or updates a node in the graph
func (g *Graph) AddNode(u *urn.URN, properties map[string]interface{}) {
	g.mu.Lock()
	defer g.mu.Unlock()

	key := u.Key()
	if properties == nil {
		properties = make(map[string]interface{})
	}
	g.nodes[key] = &Node{
		URN:        u,
		Properties: properties,
	}
}

// GetNode retrieves a node by URN
func (g *Graph) GetNode(u *urn.URN) (*Node, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	node, exists := g.nodes[u.Key()]
	return node, exists
}

// AddEdge adds a directed edge from one URN to another
func (g *Graph) AddEdge(from, to *urn.URN, edgeType EdgeType, properties map[string]interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	fromKey := from.Key()
	toKey := to.Key()

	// Ensure both nodes exist
	if _, exists := g.nodes[fromKey]; !exists {
		return fmt.Errorf("from node %s does not exist", fromKey)
	}
	if _, exists := g.nodes[toKey]; !exists {
		return fmt.Errorf("to node %s does not exist", toKey)
	}

	if properties == nil {
		properties = make(map[string]interface{})
	}

	edge := &Edge{
		From:       from,
		To:         to,
		Type:       edgeType,
		Properties: properties,
	}

	g.edges[fromKey] = append(g.edges[fromKey], edge)
	g.reverseEdges[toKey] = append(g.reverseEdges[toKey], edge)

	return nil
}

// GetOutgoingEdges returns all edges originating from a URN
func (g *Graph) GetOutgoingEdges(u *urn.URN) []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := g.edges[u.Key()]
	// Return a copy to prevent concurrent modification
	result := make([]*Edge, len(edges))
	copy(result, edges)
	return result
}

// GetIncomingEdges returns all edges pointing to a URN
func (g *Graph) GetIncomingEdges(u *urn.URN) []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	edges := g.reverseEdges[u.Key()]
	// Return a copy to prevent concurrent modification
	result := make([]*Edge, len(edges))
	copy(result, edges)
	return result
}

// GetNeighbors returns all URNs connected to the given URN (both incoming and outgoing)
func (g *Graph) GetNeighbors(u *urn.URN) []*urn.URN {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Pre-allocate with capacity hint
	outEdges := g.edges[u.Key()]
	inEdges := g.reverseEdges[u.Key()]
	neighborsMap := make(map[string]*urn.URN, len(outEdges)+len(inEdges))

	// Add outgoing neighbors
	for _, edge := range outEdges {
		neighborsMap[edge.To.Key()] = edge.To
	}

	// Add incoming neighbors
	for _, edge := range inEdges {
		neighborsMap[edge.From.Key()] = edge.From
	}

	// Convert map to slice
	neighbors := make([]*urn.URN, 0, len(neighborsMap))
	for _, neighbor := range neighborsMap {
		neighbors = append(neighbors, neighbor)
	}

	return neighbors
}

// NodeCount returns the total number of nodes in the graph
func (g *Graph) NodeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.nodes)
}

// EdgeCount returns the total number of edges in the graph
func (g *Graph) EdgeCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	count := 0
	for _, edges := range g.edges {
		count += len(edges)
	}
	return count
}

// Clear removes all nodes and edges from the graph
func (g *Graph) Clear() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.nodes = make(map[string]*Node)
	g.edges = make(map[string][]*Edge)
	g.reverseEdges = make(map[string][]*Edge)
}

// GetNodesByType returns all nodes of a specific resource type
func (g *Graph) GetNodesByType(resourceType urn.ResourceType) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []*Node
	for _, node := range g.nodes {
		if node.URN.Type == resourceType {
			result = append(result, node)
		}
	}
	return result
}

// GetNodesByProvider returns all nodes from a specific provider
func (g *Graph) GetNodesByProvider(provider urn.Provider) []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []*Node
	for _, node := range g.nodes {
		if node.URN.Provider == provider {
			result = append(result, node)
		}
	}
	return result
}

// FindPath finds a path between two URNs using breadth-first search
func (g *Graph) FindPath(from, to *urn.URN) ([]*urn.URN, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	fromKey := from.Key()
	toKey := to.Key()

	// Check if both nodes exist
	if _, exists := g.nodes[fromKey]; !exists {
		return nil, false
	}
	if _, exists := g.nodes[toKey]; !exists {
		return nil, false
	}

	// BFS to find path
	queue := [][]*urn.URN{{from}}
	visited := make(map[string]bool)
	visited[fromKey] = true

	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]

		current := path[len(path)-1]
		currentKey := current.Key()

		if currentKey == toKey {
			return path, true
		}

		// Explore neighbors (only outgoing edges for directed path)
		for _, edge := range g.edges[currentKey] {
			neighborKey := edge.To.Key()
			if !visited[neighborKey] {
				visited[neighborKey] = true
				newPath := make([]*urn.URN, len(path)+1)
				copy(newPath, path)
				newPath[len(path)] = edge.To
				queue = append(queue, newPath)
			}
		}
	}

	return nil, false
}

// GetAllNodes returns all nodes in the graph
func (g *Graph) GetAllNodes() []*Node {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*Node, 0, len(g.nodes))
	for _, node := range g.nodes {
		result = append(result, node)
	}
	return result
}

// GetAllEdges returns all edges in the graph
func (g *Graph) GetAllEdges() []*Edge {
	g.mu.RLock()
	defer g.mu.RUnlock()

	result := make([]*Edge, 0)
	for _, edges := range g.edges {
		result = append(result, edges...)
	}
	return result
}
