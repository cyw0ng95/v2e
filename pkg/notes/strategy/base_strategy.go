package strategy

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// BaseStrategy provides common functionality for navigation strategies
type BaseStrategy struct {
	mu          sync.RWMutex
	items       []SecurityItem
	viewed      map[string]bool
	viewedCount atomic.Int32
}

// NewBaseStrategy creates a new base strategy
func NewBaseStrategy(items []SecurityItem) *BaseStrategy {
	viewed := make(map[string]bool)
	for _, item := range items {
		viewed[item.URN] = false
	}

	return &BaseStrategy{
		items:       items,
		viewed:      viewed,
		viewedCount: atomic.Int32{},
	}
}

// SetItems updates the available items
func (s *BaseStrategy) SetItems(items []SecurityItem) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = items
	s.viewed = make(map[string]bool)
	for _, item := range items {
		s.viewed[item.URN] = false
	}
	s.viewedCount.Store(0)
}

// GetItems returns the current items
func (s *BaseStrategy) GetItems() []SecurityItem {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]SecurityItem, len(s.items))
	copy(items, s.items)
	return items
}

// MarkViewed marks an item as viewed
func (s *BaseStrategy) MarkViewed(urn string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.viewed[urn] {
		s.viewed[urn] = true
		s.viewedCount.Add(1)
	}

	return nil
}

// IsViewed checks if an item has been viewed
func (s *BaseStrategy) IsViewed(urn string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.viewed[urn]
}

// GetViewedCount returns the number of viewed items
func (s *BaseStrategy) GetViewedCount() int {
	return int(s.viewedCount.Load())
}

// GetTotalCount returns the total number of items
func (s *BaseStrategy) GetTotalCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.items)
}

// GetProgress returns the progress as a percentage
func (s *BaseStrategy) GetProgress() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.items) == 0 {
		return 0
	}

	viewedCount := int(s.viewedCount.Load())
	return float64(viewedCount) / float64(len(s.items)) * 100
}

// Reset clears all viewed state
func (s *BaseStrategy) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for urn := range s.viewed {
		s.viewed[urn] = false
	}
	s.viewedCount.Store(0)

	return nil
}

// FindItemByURN finds an item by its URN
func (s *BaseStrategy) FindItemByURN(urn string) (*SecurityItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.items {
		if s.items[i].URN == urn {
			return &s.items[i], nil
		}
	}

	return nil, fmt.Errorf("item not found: %s", urn)
}

// StrategyNotFoundError is returned when an item is not found
type StrategyNotFoundError struct {
	URN string
}

func (e *StrategyNotFoundError) Error() string {
	return fmt.Sprintf("item not found: %s", e.URN)
}

// StrategyError is a generic strategy error
type StrategyError struct {
	Message string
}

func (e *StrategyError) Error() string {
	return e.Message
}

// IsStrategyError checks if an error is a strategy error
func IsStrategyError(err error) bool {
	_, ok := err.(*StrategyError)
	return ok
}

// NewStrategyError creates a new strategy error
func NewStrategyError(message string) *StrategyError {
	return &StrategyError{Message: message}
}
