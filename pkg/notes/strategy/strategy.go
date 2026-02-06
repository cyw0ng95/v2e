package strategy

import (
	"context"
	"fmt"
)

// LearningStrategy defines the navigation strategy interface
type LearningStrategy interface {
	// Name returns the strategy name
	Name() string

	// NextItem returns the next item to learn based on the strategy
	NextItem(ctx context.Context, current *LearningContext) (*LearningItem, error)

	// OnView is called when an item is viewed
	OnView(urn string) error

	// OnFollowLink is called when user follows a related link (DFS)
	OnFollowLink(fromURN, toURN string) error

	// OnGoBack is called when user navigates back (DFS)
	OnGoBack(ctx context.Context) (*LearningItem, error)

	// Reset resets the strategy state
	Reset() error
}

// LearningItem represents an item being learned
type LearningItem struct {
	URN     string                 `json:"urn"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	Content string                 `json:"content"`
	Context string                 `json:"context"`
	Meta    map[string]interface{} `json:"meta,omitempty"`
}

// LearningContext provides strategy context
type LearningContext struct {
	ViewedItems    []string `json:"viewed_items"`
	CompletedItems []string `json:"completed_items"`
	AvailableItems []SecurityItem
	PathStack      []string // For DFS
}

// SecurityItem represents a security item from UEE data
type SecurityItem struct {
	URN    string `json:"urn"`
	Type   string `json:"type"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	Source string `json:"source"`
}

// ErrNoMoreItems is returned when there are no more items to review
var ErrNoMoreItems = fmt.Errorf("no more items to review")

// ErrSwitchStrategy is returned when a strategy wants to switch to another
var ErrSwitchStrategy = fmt.Errorf("switch strategy")
