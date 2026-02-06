package fsm

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// LearningState tracks user's overall learning session state
type LearningState string

const (
	// LearningStateIdle is the initial state
	LearningStateIdle LearningState = "idle"
	// LearningStateBrowsing indicates BFS mode (list order)
	LearningStateBrowsing LearningState = "browsing"
	// LearningStateDeepDive indicates DFS mode (following links)
	LearningStateDeepDive LearningState = "deep_dive"
	// LearningStateReviewing indicates card review mode
	LearningStateReviewing LearningState = "reviewing"
	// LearningStatePaused indicates session is paused
	LearningStatePaused LearningState = "paused"
)

// LearningItem represents an item being learned
type LearningItem struct {
	URN     string                 `json:"urn"`
	Type    string                 `json:"type"` // "cve", "cwe", "capec", "attack"
	Title   string                 `json:"title"`
	Content string                 `json:"content"`
	Context string                 `json:"context"` // "browsing", "deep_dive", "reviewing"
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// SecurityItem represents a security item from UEE data
type SecurityItem struct {
	URN    string `json:"urn"`
	Type   string `json:"type"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	Source string `json:"source"`
}

// LearningContext provides context for strategy decisions
type LearningContext struct {
	ViewedItems    []string       `json:"viewed_items"`
	CompletedItems []string       `json:"completed_items"`
	AvailableItems []SecurityItem `json:"available_items"`
	PathStack      []string       `json:"path_stack"` // For DFS navigation
}

// LearningFSM manages user learning progress
type LearningFSM struct {
	mu              sync.RWMutex
	state           LearningState
	currentStrategy string // "bfs" or "dfs"
	currentItemURN  string
	viewedItems     []string // URNs viewed in session
	completedItems  []string // URNs marked learned
	pathStack       []string // DFS navigation stack
	sessionStart    time.Time
	lastActivity    time.Time
	storage         Storage
	availableItems  []SecurityItem // Items available for learning
	itemGraph       *ItemGraph     // Pre-built link graph for DFS
}

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

// NewLearningFSM creates a new learning FSM
func NewLearningFSM(storage Storage, items []SecurityItem) (*LearningFSM, error) {
	now := time.Now()
	fsm := &LearningFSM{
		state:           LearningStateIdle,
		currentStrategy: "bfs",
		viewedItems:     make([]string, 0),
		completedItems:  make([]string, 0),
		pathStack:       make([]string, 0),
		sessionStart:    now,
		lastActivity:    now,
		storage:         storage,
		availableItems:  items,
		itemGraph:       NewItemGraph(),
	}

	// Build item graph from cross-references
	fsm.buildItemGraph()

	// Try to load existing state
	if storage != nil {
		if err := fsm.LoadState(); err != nil {
			// If load fails, start fresh
			fsm.state = LearningStateIdle
		}
	}

	return fsm, nil
}

// buildItemGraph constructs the link graph from available items
func (l *LearningFSM) buildItemGraph() {
	if l.itemGraph == nil {
		l.itemGraph = NewItemGraph()
	}

	// Group items by type
	itemsByType := make(map[string][]SecurityItem)
	for _, item := range l.availableItems {
		itemsByType[item.Type] = append(itemsByType[item.Type], item)
	}

	// Build links based on type relationships
	// CVE -> CWE -> CAPEC -> ATT&CK is a common security hierarchy
	l.buildTypeLinks(itemsByType, "cve", "cwe")
	l.buildTypeLinks(itemsByType, "cwe", "capec")
	l.buildTypeLinks(itemsByType, "capec", "attack")

	// Also add intra-type links (within same type)
	l.buildIntraTypeLinks(itemsByType)
}

// buildTypeLinks creates links from items of one type to items of another type
func (l *LearningFSM) buildTypeLinks(itemsByType map[string][]SecurityItem, fromType, toType string) {
	fromItems, fromExists := itemsByType[fromType]
	toItems, toExists := itemsByType[toType]

	if !fromExists || !toExists || len(toItems) == 0 {
		return
	}

	// Link each fromType item to up to 3 toType items
	for _, fromItem := range fromItems {
		for i := 0; i < min(3, len(toItems)); i++ {
			l.itemGraph.AddLink(fromItem.URN, toItems[i].URN)
		}
	}
}

// buildIntraTypeLinks creates links between items of the same type
func (l *LearningFSM) buildIntraTypeLinks(itemsByType map[string][]SecurityItem) {
	for _, items := range itemsByType {
		if len(items) < 2 {
			continue
		}

		// Create a simple chain: item1 -> item2 -> item3 -> ...
		for i := 0; i < len(items)-1; i++ {
			l.itemGraph.AddLink(items[i].URN, items[i+1].URN)
		}
	}
}

// LoadItem presents the next item based on current strategy
func (l *LearningFSM) LoadItem(ctx context.Context) (*LearningItem, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Update activity time
	l.lastActivity = time.Now()

	// If we have a current item, return it
	if l.currentItemURN != "" {
		return l.getItemByURN(l.currentItemURN)
	}

	// Get next item based on strategy
	switch l.currentStrategy {
	case "bfs":
		return l.getNextBFSItem(ctx)
	case "dfs":
		return l.getNextDFSItem(ctx)
	default:
		return nil, fmt.Errorf("unknown strategy: %s", l.currentStrategy)
	}
}

// getNextBFSItem returns the next unviewed item in list order
func (l *LearningFSM) getNextBFSItem(ctx context.Context) (*LearningItem, error) {
	l.state = LearningStateBrowsing

	viewedSet := make(map[string]bool)
	for _, urn := range l.viewedItems {
		viewedSet[urn] = true
	}

	for _, item := range l.availableItems {
		if !viewedSet[item.URN] {
			l.currentItemURN = item.URN
			return &LearningItem{
				URN:     item.URN,
				Type:    item.Type,
				Title:   item.Title,
				Context: "browsing",
			}, nil
		}
	}

	return nil, fmt.Errorf("no more items to review")
}

// getNextDFSItem returns the next item from the DFS path stack
func (l *LearningFSM) getNextDFSItem(ctx context.Context) (*LearningItem, error) {
	l.state = LearningStateDeepDive

	if len(l.pathStack) == 0 {
		// No more items in DFS path, switch to BFS
		l.currentStrategy = "bfs"
		return l.getNextBFSItem(ctx)
	}

	// Pop next item from stack
	urn := l.pathStack[len(l.pathStack)-1]
	l.pathStack = l.pathStack[:len(l.pathStack)-1]
	l.currentItemURN = urn

	return l.getItemByURN(urn)
}

// getItemByURN retrieves an item by its URN
func (l *LearningFSM) getItemByURN(urn string) (*LearningItem, error) {
	for _, item := range l.availableItems {
		if item.URN == urn {
			return &LearningItem{
				URN:     item.URN,
				Type:    item.Type,
				Title:   item.Title,
				Context: string(l.state),
			}, nil
		}
	}
	return nil, fmt.Errorf("item not found: %s", urn)
}

// MarkViewed records an item as viewed
func (l *LearningFSM) MarkViewed(urn string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if already viewed
	for _, viewedURN := range l.viewedItems {
		if viewedURN == urn {
			return nil // Already viewed
		}
	}

	l.viewedItems = append(l.viewedItems, urn)
	l.lastActivity = time.Now()

	// Save state with timeout
	if l.storage != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := l.SaveStateWithContext(ctx); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	return nil
}

// MarkLearned records an item as completed/learned
func (l *LearningFSM) MarkLearned(urn string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Check if already learned
	for _, learnedURN := range l.completedItems {
		if learnedURN == urn {
			return nil // Already learned
		}
	}

	l.completedItems = append(l.completedItems, urn)
	l.lastActivity = time.Now()

	// Clear current item if it's the one being marked learned
	if l.currentItemURN == urn {
		l.currentItemURN = ""
	}

	// Save state with timeout
	if l.storage != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := l.SaveStateWithContext(ctx); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	return nil
}

// FollowLink handles user clicking a related link (DFS transition)
func (l *LearningFSM) FollowLink(fromURN, toURN string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Push current item to stack for backtracking
	if fromURN != "" && l.currentItemURN != "" {
		l.pathStack = append(l.pathStack, fromURN)
	}

	// Switch to DFS strategy
	l.currentStrategy = "dfs"
	l.state = LearningStateDeepDive
	l.currentItemURN = toURN
	l.lastActivity = time.Now()

	// Mark the new item as viewed
	l.viewedItems = append(l.viewedItems, toURN)

	// Save state
	if l.storage != nil {
		if err := l.SaveState(); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	return nil
}

// GoBack navigates to the previous item (DFS backtracking)
func (l *LearningFSM) GoBack(ctx context.Context) (*LearningItem, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.pathStack) == 0 {
		// No more items in stack, switch to BFS
		l.currentStrategy = "bfs"
		l.state = LearningStateBrowsing
		l.currentItemURN = ""

		// Get next BFS item
		l.mu.Unlock()
		return l.LoadItem(ctx)
	}

	// Pop previous item from stack
	urn := l.pathStack[len(l.pathStack)-1]
	l.pathStack = l.pathStack[:len(l.pathStack)-1]
	l.currentItemURN = urn
	l.lastActivity = time.Now()

	return l.getItemByURN(urn)
}

// GetState returns the current learning FSM state
func (l *LearningFSM) GetState() LearningState {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state
}

// GetContext returns the current learning context
func (l *LearningFSM) GetContext() *LearningContext {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return &LearningContext{
		ViewedItems:    append([]string{}, l.viewedItems...),
		CompletedItems: append([]string{}, l.completedItems...),
		AvailableItems: l.availableItems,
		PathStack:      append([]string{}, l.pathStack...),
	}
}

// SaveStateWithContext persists the learning FSM state with context timeout
func (l *LearningFSM) SaveStateWithContext(ctx context.Context) error {
	if l.storage == nil {
		return nil
	}

	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Create a channel to handle timeout
	type result struct {
		err error
	}
	resultCh := make(chan result, 1)

	go func() {
		state := &LearningFSMState{
			State:           l.state,
			CurrentStrategy: l.currentStrategy,
			CurrentItemURN:  l.currentItemURN,
			ViewedItems:     l.viewedItems,
			CompletedItems:  l.completedItems,
			PathStack:       l.pathStack,
			SessionStart:    l.sessionStart,
			LastActivity:    l.lastActivity,
			UpdatedAt:       time.Now(),
		}
		resultCh <- result{l.storage.SaveLearningFSMState(state)}
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("save state timeout: %w", ctx.Err())
	case res := <-resultCh:
		return res.err
	}
}

// SaveState persists the learning FSM state
func (l *LearningFSM) SaveState() error {
	return l.SaveStateWithContext(context.Background())
}

// LoadState restores the learning FSM state
func (l *LearningFSM) LoadState() error {
	if l.storage == nil {
		return nil // No storage configured
	}

	state, err := l.storage.LoadLearningFSMState()
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.state = state.State
	l.currentStrategy = state.CurrentStrategy
	l.currentItemURN = state.CurrentItemURN
	l.viewedItems = state.ViewedItems
	l.completedItems = state.CompletedItems
	l.pathStack = state.PathStack
	l.sessionStart = state.SessionStart
	l.lastActivity = state.LastActivity
	l.mu.Unlock()

	return nil
}

// Pause pauses the learning session
func (l *LearningFSM) Pause() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.state = LearningStatePaused
	l.lastActivity = time.Now()

	if l.storage != nil {
		return l.SaveState()
	}

	return nil
}

// Resume resumes the learning session
func (l *LearningFSM) Resume() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.state != LearningStatePaused {
		return fmt.Errorf("cannot resume from state: %s", l.state)
	}

	l.state = LearningStateBrowsing
	l.lastActivity = time.Now()

	if l.storage != nil {
		return l.SaveState()
	}

	return nil
}

// LearningFSMState represents the persisted learning FSM state
type LearningFSMState struct {
	State           LearningState `json:"state"`
	CurrentStrategy string        `json:"current_strategy"`
	CurrentItemURN  string        `json:"current_item_urn"`
	ViewedItems     []string      `json:"viewed_items"`
	CompletedItems  []string      `json:"completed_items"`
	PathStack       []string      `json:"path_stack"`
	SessionStart    time.Time     `json:"session_start"`
	LastActivity    time.Time     `json:"last_activity"`
	UpdatedAt       time.Time     `json:"updated_at"`
}
