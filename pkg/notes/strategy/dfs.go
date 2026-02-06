package strategy

import (
	"context"
	"fmt"
	"sync"
)

// ItemGraph maintains bidirectional relationships for DFS navigation
type ItemGraph struct {
	links map[string][]string // URN -> linked URNs
	mu    sync.RWMutex
}

// NewItemGraph creates a new item graph
func NewItemGraph() *ItemGraph {
	return &ItemGraph{
		links: make(map[string][]string),
	}
}

// AddLink adds a bidirectional link between two items
func (g *ItemGraph) AddLink(fromURN, toURN string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.links[fromURN] == nil {
		g.links[fromURN] = make([]string, 0)
	}
	g.links[fromURN] = append(g.links[fromURN], toURN)
}

// GetLinks returns all URNs linked to the given URN
func (g *ItemGraph) GetLinks(urn string) []string {
	g.mu.RLock()
	defer g.mu.RUnlock()

	if links, ok := g.links[urn]; ok {
		result := make([]string, len(links))
		copy(result, links)
		return result
	}
	return nil
}

// DFSStrategy follows link relationships (depth-first)
type DFSStrategy struct {
	mu        sync.RWMutex
	pathStack []string // Navigation stack for backtracking
	viewed    map[string]bool
	itemGraph *ItemGraph // Pre-built link graph
}

// NewDFSStrategy creates a new DFS strategy
func NewDFSStrategy(graph *ItemGraph) *DFSStrategy {
	return &DFSStrategy{
		pathStack: make([]string, 0),
		viewed:    make(map[string]bool),
		itemGraph: graph,
	}
}

// Name returns the strategy name
func (s *DFSStrategy) Name() string {
	return "dfs"
}

// NextItem returns the next item from the DFS path stack
func (s *DFSStrategy) NextItem(ctx context.Context, current *LearningContext) (*LearningItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pathStack) == 0 {
		// No more items in stack, signal to switch to BFS
		return nil, ErrSwitchStrategy
	}

	// Pop next item from stack
	urn := s.pathStack[len(s.pathStack)-1]
	s.pathStack = s.pathStack[:len(s.pathStack)-1]

	// Mark as viewed
	s.viewed[urn] = true

	// Find item in available items
	for _, item := range current.AvailableItems {
		if item.URN == urn {
			return &LearningItem{
				URN:     item.URN,
				Type:    item.Type,
				Title:   item.Title,
				Context: "deep_dive",
			}, nil
		}
	}

	return nil, fmt.Errorf("item not found: %s", urn)
}

// OnView is called when an item is viewed
func (s *DFSStrategy) OnView(urn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.viewed[urn] = true
	return nil
}

// OnFollowLink handles user clicking a related link
func (s *DFSStrategy) OnFollowLink(fromURN, toURN string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Push current item to stack for backtracking
	if fromURN != "" {
		s.pathStack = append(s.pathStack, fromURN)
	}

	// Mark the target as viewed
	s.viewed[toURN] = true

	return nil
}

// OnGoBack navigates to the previous item
func (s *DFSStrategy) OnGoBack(ctx context.Context) (*LearningItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.pathStack) == 0 {
		// No more items in stack, signal to switch to BFS
		return nil, ErrSwitchStrategy
	}

	// Pop previous item from stack
	urn := s.pathStack[len(s.pathStack)-1]
	s.pathStack = s.pathStack[:len(s.pathStack)-1]

	return &LearningItem{
		URN:     urn,
		Context: "deep_dive",
	}, nil
}

// Reset resets the strategy state
func (s *DFSStrategy) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pathStack = make([]string, 0)
	s.viewed = make(map[string]bool)
	return nil
}

// GetLinkedItems returns all items linked to the given URN
func (s *DFSStrategy) GetLinkedItems(urn string) []string {
	if s.itemGraph == nil {
		return nil
	}
	return s.itemGraph.GetLinks(urn)
}

// PushStack adds an item to the navigation stack
func (s *DFSStrategy) PushStack(urn string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pathStack = append(s.pathStack, urn)
}

// GetStackDepth returns the current depth of the navigation stack
func (s *DFSStrategy) GetStackDepth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pathStack)
}

// GetViewedCount returns the number of viewed items
func (s *DFSStrategy) GetViewedCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	count := 0
	for _, viewed := range s.viewed {
		if viewed {
			count++
		}
	}
	return count
}
