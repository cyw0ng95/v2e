package strategy

import (
	"context"
	"fmt"
	"sync"
)

// BFSStrategy presents items in list order (breadth-first)
type BFSStrategy struct {
	mu           sync.RWMutex
	currentIndex int
	items        []SecurityItem
	viewed       map[string]bool
}

// NewBFSStrategy creates a new BFS strategy
func NewBFSStrategy(items []SecurityItem) *BFSStrategy {
	viewed := make(map[string]bool)
	for _, item := range items {
		viewed[item.URN] = false
	}

	return &BFSStrategy{
		items:  items,
		viewed: viewed,
	}
}

// Name returns the strategy name
func (s *BFSStrategy) Name() string {
	return "bfs"
}

// NextItem returns the next unviewed item in list order
func (s *BFSStrategy) NextItem(ctx context.Context, current *LearningContext) (*LearningItem, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find next unviewed item
	for i := s.currentIndex; i < len(s.items); i++ {
		item := s.items[i]
		if !s.viewed[item.URN] {
			s.currentIndex = i
			return &LearningItem{
				URN:     item.URN,
				Type:    item.Type,
				Title:   item.Title,
				Context: "browsing",
			}, nil
		}
	}

	return nil, ErrNoMoreItems
}

// OnView marks an item as viewed
func (s *BFSStrategy) OnView(urn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.viewed[urn] = true
	return nil
}

// OnFollowLink switches to DFS strategy
func (s *BFSStrategy) OnFollowLink(fromURN, toURN string) error {
	// BFS doesn't handle link following - signal to switch strategy
	return ErrSwitchStrategy
}

// OnGoBack is not applicable for BFS
func (s *BFSStrategy) OnGoBack(ctx context.Context) (*LearningItem, error) {
	return nil, fmt.Errorf("go back not supported in BFS mode")
}

// Reset resets the strategy state
func (s *BFSStrategy) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.currentIndex = 0
	for urn := range s.viewed {
		s.viewed[urn] = false
	}

	return nil
}

// SetItems updates the available items
func (s *BFSStrategy) SetItems(items []SecurityItem) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = items
	s.currentIndex = 0
	s.viewed = make(map[string]bool)
	for _, item := range items {
		s.viewed[item.URN] = false
	}
}

// GetViewedCount returns the number of viewed items
func (s *BFSStrategy) GetViewedCount() int {
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

// GetTotalCount returns the total number of items
func (s *BFSStrategy) GetTotalCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// GetProgress returns the progress as a percentage
func (s *BFSStrategy) GetProgress() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.items) == 0 {
		return 0
	}

	viewedCount := 0
	for _, viewed := range s.viewed {
		if viewed {
			viewedCount++
		}
	}

	return float64(viewedCount) / float64(len(s.items)) * 100
}
