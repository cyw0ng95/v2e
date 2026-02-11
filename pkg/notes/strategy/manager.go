package strategy

import (
	"context"
	"log"
	"sync"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/notes"
)

// Manager handles automatic strategy switching based on user behavior
type Manager struct {
	mu              sync.RWMutex
	currentStrategy LearningStrategy
	bfsStrategy     *BFSStrategy
	dfsStrategy     *DFSStrategy
	items           []SecurityItem
	itemGraph       *ItemGraph
	db              *gorm.DB

	// FSM state (simplified, without direct FSM dependency)
	viewedItems    []string
	completedItems []string
	pathStack      []string
}

// NewManager creates a new strategy manager with cross-reference graph
func NewManager(items []SecurityItem, db *gorm.DB) *Manager {
	graph := NewItemGraph()

	// Build item graph from cross-references in database
	if db != nil {
		ctx := context.Background()
		crossRefService := notes.NewCrossReferenceService(db)

		for _, item := range items {
			// Get all cross-references where this item is the source
			crossRefs, err := crossRefService.GetCrossReferencesBySource(ctx, item.URN)
			if err != nil {
				log.Printf("Warning: failed to get cross-references for %s: %v", item.URN, err)
				continue
			}

			// Add links from cross-references
			for _, ref := range crossRefs {
				graph.AddLink(ref.SourceItemID, ref.TargetItemID)
			}
		}
	} else {
		// Fallback: create simple links based on item types if no DB
		buildSimpleGraph(graph, items)
	}

	bfs := NewBFSStrategy(items)
	dfs := NewDFSStrategy(graph)

	return &Manager{
		currentStrategy: bfs,
		bfsStrategy:     bfs,
		dfsStrategy:     dfs,
		items:           items,
		itemGraph:       graph,
		db:              db,
		viewedItems:     make([]string, 0),
		completedItems:  make([]string, 0),
		pathStack:       make([]string, 0),
	}
}

// buildSimpleGraph creates a simple type-based graph as fallback
func buildSimpleGraph(graph *ItemGraph, items []SecurityItem) {
	itemsByType := make(map[string][]SecurityItem)
	for _, item := range items {
		itemsByType[item.Type] = append(itemsByType[item.Type], item)
	}

	// Create links based on type relationships
	buildTypeLinks(graph, itemsByType, "cve", "cwe")
	buildTypeLinks(graph, itemsByType, "cwe", "capec")
	buildTypeLinks(graph, itemsByType, "capec", "attack")

	// Add intra-type links
	for _, typeItems := range itemsByType {
		if len(typeItems) < 2 {
			continue
		}
		for i := 0; i < len(typeItems)-1; i++ {
			graph.AddLink(typeItems[i].URN, typeItems[i+1].URN)
		}
	}
}

// buildTypeLinks creates links between items of different types
func buildTypeLinks(graph *ItemGraph, itemsByType map[string][]SecurityItem, fromType, toType string) {
	fromItems, fromExists := itemsByType[fromType]
	toItems, toExists := itemsByType[toType]

	if !fromExists || !toExists || len(toItems) == 0 {
		return
	}

	// Link each fromType item to up to 3 toType items
	for _, fromItem := range fromItems {
		for i := 0; i < min(3, len(toItems)); i++ {
			graph.AddLink(fromItem.URN, toItems[i].URN)
		}
	}
}

// GetCurrentStrategy returns the current strategy name
func (m *Manager) GetCurrentStrategy() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentStrategy.Name()
}

// GetNextItem returns the next item based on current strategy
func (m *Manager) GetNextItem(ctx context.Context) (*LearningItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Build context
	context := &LearningContext{
		ViewedItems:    m.viewedItems,
		CompletedItems: m.completedItems,
		AvailableItems: m.items,
		PathStack:      m.pathStack,
	}

	// Try current strategy
	item, err := m.currentStrategy.NextItem(ctx, context)
	if err == ErrSwitchStrategy {
		// Auto-switch to BFS when DFS path is exhausted
		m.currentStrategy = m.bfsStrategy
		item, err = m.bfsStrategy.NextItem(ctx, context)
	}

	return item, err
}

// MarkViewed marks an item as viewed
func (m *Manager) MarkViewed(urn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already viewed
	for _, viewedURN := range m.viewedItems {
		if viewedURN == urn {
			return nil
		}
	}

	m.viewedItems = append(m.viewedItems, urn)

	// Mark in strategy
	return m.currentStrategy.OnView(urn)
}

// MarkLearned marks an item as completed
func (m *Manager) MarkLearned(urn string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already learned
	for _, learnedURN := range m.completedItems {
		if learnedURN == urn {
			return nil
		}
	}

	m.completedItems = append(m.completedItems, urn)
	return nil
}

// FollowLink handles user clicking a related link (switches to DFS)
func (m *Manager) FollowLink(fromURN, toURN string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Switch to DFS strategy
	m.currentStrategy = m.dfsStrategy

	// Push to path stack
	m.pathStack = append(m.pathStack, fromURN)

	// Mark the new item as viewed
	m.viewedItems = append(m.viewedItems, toURN)

	// Handle the link follow
	return m.dfsStrategy.OnFollowLink(fromURN, toURN)
}

// GoBack navigates to the previous item
func (m *Manager) GoBack(ctx context.Context) (*LearningItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Try DFS first
	if m.currentStrategy.Name() == "dfs" {
		item, err := m.dfsStrategy.OnGoBack(ctx)
		if err == nil {
			// Pop from path stack
			if len(m.pathStack) > 0 {
				m.pathStack = m.pathStack[:len(m.pathStack)-1]
			}
			return item, nil
		}
		if err == ErrSwitchStrategy {
			// Switch to BFS
			m.currentStrategy = m.bfsStrategy
		}
	}

	// If DFS has no more items, pop from path stack
	if len(m.pathStack) > 0 {
		urn := m.pathStack[len(m.pathStack)-1]
		m.pathStack = m.pathStack[:len(m.pathStack)-1]

		// Find item in available items
		for _, item := range m.items {
			if item.URN == urn {
				return &LearningItem{
					URN:     item.URN,
					Type:    item.Type,
					Title:   item.Title,
					Context: "browsing",
				}, nil
			}
		}
	}

	return nil, ErrNoMoreItems
}

// GetProgress returns overall learning progress
func (m *Manager) GetProgress() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"strategy":        m.currentStrategy.Name(),
		"viewed_count":    len(m.viewedItems),
		"completed_count": len(m.completedItems),
		"total_items":     len(m.items),
		"bfs_progress":    m.bfsStrategy.GetProgress(),
		"dfs_stack_depth": m.dfsStrategy.GetStackDepth(),
	}
}

// GetContext returns the current learning context
func (m *Manager) GetContext() *LearningContext {
	m.mu.RLock()
	defer m.mu.RUnlock()

	viewedCopy := make([]string, len(m.viewedItems))
	copy(viewedCopy, m.viewedItems)

	completedCopy := make([]string, len(m.completedItems))
	copy(completedCopy, m.completedItems)

	pathStackCopy := make([]string, len(m.pathStack))
	copy(pathStackCopy, m.pathStack)

	return &LearningContext{
		ViewedItems:    viewedCopy,
		CompletedItems: completedCopy,
		AvailableItems: m.items,
		PathStack:      pathStackCopy,
	}
}

// UpdateItems updates the available items
func (m *Manager) UpdateItems(items []SecurityItem) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = items
	m.bfsStrategy.SetItems(items)
	// Rebuild item graph
	m.itemGraph = NewItemGraph()
	for _, item := range items {
		m.itemGraph.AddLink(item.URN, item.URN)
	}
	m.dfsStrategy = NewDFSStrategy(m.itemGraph)
}

// Reset resets all strategies
func (m *Manager) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.bfsStrategy.Reset(); err != nil {
		return err
	}

	if err := m.dfsStrategy.Reset(); err != nil {
		return err
	}

	m.viewedItems = make([]string, 0)
	m.completedItems = make([]string, 0)
	m.pathStack = make([]string, 0)

	return nil
}
