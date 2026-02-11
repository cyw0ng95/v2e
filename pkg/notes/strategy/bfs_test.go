package strategy

import (
	"context"
	"fmt"
	"testing"
)

func TestNewBFSStrategy(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "CVE 3"},
	}

	strategy := NewBFSStrategy(items)

	if strategy.Name() != "bfs" {
		t.Errorf("expected name 'bfs', got %s", strategy.Name())
	}

	if strategy.GetTotalCount() != len(items) {
		t.Errorf("expected total count %d, got %d", len(items), strategy.GetTotalCount())
	}

	if strategy.GetViewedCount() != 0 {
		t.Errorf("expected viewed count 0, got %d", strategy.GetViewedCount())
	}
}

func TestBFSStrategy_NextItem(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "CVE 3"},
	}

	strategy := NewBFSStrategy(items)
	ctx := context.Background()
	current := &LearningContext{}

	// Get first item
	item1, err := strategy.NextItem(ctx, current)
	if err != nil {
		t.Fatal(err)
	}

	if item1.URN != "v2e::cve::1" {
		t.Errorf("expected first item v2e::cve::1, got %s", item1.URN)
	}

	if item1.Context != "browsing" {
		t.Errorf("expected context 'browsing', got %s", item1.Context)
	}

	// Mark as viewed and get next
	strategy.OnView(item1.URN)

	item2, err := strategy.NextItem(ctx, current)
	if err != nil {
		t.Fatal(err)
	}

	if item2.URN != "v2e::cve::2" {
		t.Errorf("expected second item v2e::cve::2, got %s", item2.URN)
	}
}

func TestBFSStrategy_NextItem_NoMoreItems(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)
	ctx := context.Background()
	current := &LearningContext{}

	// Get item
	item, err := strategy.NextItem(ctx, current)
	if err != nil {
		t.Fatal(err)
	}

	// Mark as viewed
	strategy.OnView(item.URN)

	// Try to get next item - should error
	_, err = strategy.NextItem(ctx, current)
	if err == nil {
		t.Fatal("expected error when no more items")
	}

	if err != ErrNoMoreItems {
		t.Errorf("expected ErrNoMoreItems, got %v", err)
	}
}

func TestBFSStrategy_OnView(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
	}

	strategy := NewBFSStrategy(items)

	// Mark first item as viewed
	err := strategy.OnView("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	// Check viewed count
	if strategy.GetViewedCount() != 1 {
		t.Errorf("expected viewed count 1, got %d", strategy.GetViewedCount())
	}

	// Mark second item as viewed
	err = strategy.OnView("v2e::cve::2")
	if err != nil {
		t.Fatal(err)
	}

	if strategy.GetViewedCount() != 2 {
		t.Errorf("expected viewed count 2, got %d", strategy.GetViewedCount())
	}
}

func TestBFSStrategy_OnView_Duplicate(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)

	// Mark as viewed twice
	err := strategy.OnView("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	err = strategy.OnView("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	// Should still have count of 1
	if strategy.GetViewedCount() != 1 {
		t.Errorf("expected viewed count 1, got %d", strategy.GetViewedCount())
	}
}

func TestBFSStrategy_OnFollowLink(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)

	// OnFollowLink should return switch strategy error
	err := strategy.OnFollowLink("v2e::cve::1", "v2e::cwe::1")
	if err != ErrSwitchStrategy {
		t.Errorf("expected ErrSwitchStrategy, got %v", err)
	}
}

func TestBFSStrategy_OnGoBack(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)

	// OnGoBack should return error
	_, err := strategy.OnGoBack(context.Background())
	if err == nil {
		t.Fatal("expected error for OnGoBack in BFS mode")
	}
}

func TestBFSStrategy_Reset(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
	}

	strategy := NewBFSStrategy(items)
	ctx := context.Background()
	current := &LearningContext{}

	// View some items
	item, _ := strategy.NextItem(ctx, current)
	strategy.OnView(item.URN)

	item, _ = strategy.NextItem(ctx, current)
	strategy.OnView(item.URN)

	// Check viewed count
	if strategy.GetViewedCount() != 2 {
		t.Errorf("expected viewed count 2 before reset, got %d", strategy.GetViewedCount())
	}

	// Reset
	err := strategy.Reset()
	if err != nil {
		t.Fatal(err)
	}

	// Check viewed count is reset
	if strategy.GetViewedCount() != 0 {
		t.Errorf("expected viewed count 0 after reset, got %d", strategy.GetViewedCount())
	}
}

func TestBFSStrategy_SetItems(t *testing.T) {
	items1 := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
	}

	strategy := NewBFSStrategy(items1)

	// View first item
	strategy.OnView("v2e::cve::1")

	// Set new items
	items2 := []SecurityItem{
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "CVE 3"},
		{URN: "v2e::cve::4", Type: "cve", ID: "4", Title: "CVE 4"},
	}

	strategy.SetItems(items2)

	// Check total count updated
	if strategy.GetTotalCount() != len(items2) {
		t.Errorf("expected total count %d, got %d", len(items2), strategy.GetTotalCount())
	}

	// Check viewed count reset
	if strategy.GetViewedCount() != 0 {
		t.Errorf("expected viewed count 0 after SetItems, got %d", strategy.GetViewedCount())
	}
}

func TestBFSStrategy_GetProgress(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "CVE 3"},
		{URN: "v2e::cve::4", Type: "cve", ID: "4", Title: "CVE 4"},
	}

	strategy := NewBFSStrategy(items)

	// Initial progress should be 0%
	progress := strategy.GetProgress()
	if progress != 0 {
		t.Errorf("expected progress 0%%, got %.2f%%", progress)
	}

	// View 1 item (25%)
	strategy.OnView("v2e::cve::1")
	progress = strategy.GetProgress()
	if progress != 25.0 {
		t.Errorf("expected progress 25%%, got %.2f%%", progress)
	}

	// View 2 items (50%)
	strategy.OnView("v2e::cve::2")
	progress = strategy.GetProgress()
	if progress != 50.0 {
		t.Errorf("expected progress 50%%, got %.2f%%", progress)
	}

	// View 3 items (75%)
	strategy.OnView("v2e::cve::3")
	progress = strategy.GetProgress()
	if progress != 75.0 {
		t.Errorf("expected progress 75%%, got %.2f%%", progress)
	}

	// View 4 items (100%)
	strategy.OnView("v2e::cve::4")
	progress = strategy.GetProgress()
	if progress != 100.0 {
		t.Errorf("expected progress 100%%, got %.2f%%", progress)
	}
}

func TestBFSStrategy_GetProgress_EmptyItems(t *testing.T) {
	items := []SecurityItem{}

	strategy := NewBFSStrategy(items)

	// Progress should be 0 for empty items
	progress := strategy.GetProgress()
	if progress != 0 {
		t.Errorf("expected progress 0 for empty items, got %.2f%%", progress)
	}
}

func TestBFSStrategy_SequentialAccess(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "CVE 3"},
	}

	strategy := NewBFSStrategy(items)
	ctx := context.Background()
	current := &LearningContext{}

	// Access all items sequentially
	expectedOrder := []string{"v2e::cve::1", "v2e::cve::2", "v2e::cve::3"}
	for i, expectedURN := range expectedOrder {
		item, err := strategy.NextItem(ctx, current)
		if err != nil {
			t.Fatalf("item %d: failed to get next item: %v", i, err)
		}

		if item.URN != expectedURN {
			t.Errorf("item %d: expected %s, got %s", i, expectedURN, item.URN)
		}

		strategy.OnView(item.URN)
	}

	// No more items
	_, err := strategy.NextItem(ctx, current)
	if err != ErrNoMoreItems {
		t.Errorf("expected ErrNoMoreItems, got %v", err)
	}
}

func TestBFSStrategy_NextItem_AfterViewed(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
	}

	strategy := NewBFSStrategy(items)
	ctx := context.Background()
	current := &LearningContext{}

	// Get first item
	item1, err := strategy.NextItem(ctx, current)
	if err != nil {
		t.Fatal(err)
	}

	// Mark it as viewed
	strategy.OnView(item1.URN)

	// Get first item again - should skip to next
	item2, err := strategy.NextItem(ctx, current)
	if err != nil {
		t.Fatal(err)
	}

	if item2.URN == item1.URN {
		t.Error("expected to skip viewed item")
	}

	if item2.URN != "v2e::cve::2" {
		t.Errorf("expected item v2e::cve::2, got %s", item2.URN)
	}
}

func TestBFSStrategy_ThreadSafe(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
	}

	strategy := NewBFSStrategy(items)
	done := make(chan bool)

	// Concurrent GetViewedCount
	for i := 0; i < 10; i++ {
		go func() {
			strategy.GetViewedCount()
			done <- true
		}()
	}

	// Concurrent GetProgress
	for i := 0; i < 10; i++ {
		go func() {
			strategy.GetProgress()
			done <- true
		}()
	}

	// Concurrent OnView
	for i := 0; i < 10; i++ {
		go func(idx int) {
			urn := fmt.Sprintf("v2e::cve::%d", (idx%2)+1)
			strategy.OnView(urn)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 30; i++ {
		<-done
	}
}

func TestBFSStrategy_SingleItem(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)
	ctx := context.Background()
	current := &LearningContext{}

	// Get item
	item, err := strategy.NextItem(ctx, current)
	if err != nil {
		t.Fatal(err)
	}

	if item.URN != "v2e::cve::1" {
		t.Errorf("expected item v2e::cve::1, got %s", item.URN)
	}

	// Mark as viewed
	strategy.OnView(item.URN)

	// Progress should be 100%
	progress := strategy.GetProgress()
	if progress != 100.0 {
		t.Errorf("expected progress 100%%, got %.2f%%", progress)
	}

	// No more items
	_, err = strategy.NextItem(ctx, current)
	if err != ErrNoMoreItems {
		t.Errorf("expected ErrNoMoreItems, got %v", err)
	}
}

func TestBFSStrategy_ContextPassThrough(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)
	current := &LearningContext{
		ViewedItems:    []string{"v2e::cve::0"},
		CompletedItems: []string{},
		AvailableItems: items,
	}

	item, err := strategy.NextItem(context.Background(), current)
	if err != nil {
		t.Fatal(err)
	}

	if item.URN != "v2e::cve::1" {
		t.Errorf("expected item v2e::cve::1, got %s", item.URN)
	}
}

func TestBFSStrategy_Reset_ClearsViewed(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
	}

	strategy := NewBFSStrategy(items)

	// View all items
	strategy.OnView("v2e::cve::1")
	strategy.OnView("v2e::cve::2")

	if strategy.GetViewedCount() != 2 {
		t.Errorf("expected viewed count 2 before reset, got %d", strategy.GetViewedCount())
	}

	// Reset
	strategy.Reset()

	// All items should be available again
	item1, err := strategy.NextItem(context.Background(), &LearningContext{})
	if err != nil {
		t.Fatal(err)
	}

	if item1.URN != "v2e::cve::1" {
		t.Errorf("expected first item after reset, got %s", item1.URN)
	}
}

func TestBFSStrategy_MultipleResets(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)

	// Multiple resets
	for i := 0; i < 5; i++ {
		err := strategy.Reset()
		if err != nil {
			t.Errorf("reset %d: unexpected error: %v", i, err)
		}
	}
}

func TestBFSStrategy_Name(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)

	if strategy.Name() != "bfs" {
		t.Errorf("expected name 'bfs', got %s", strategy.Name())
	}
}

func TestBFSStrategy_GetTotalCount(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "CVE 2"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "CVE 3"},
	}

	strategy := NewBFSStrategy(items)

	if strategy.GetTotalCount() != 3 {
		t.Errorf("expected total count 3, got %d", strategy.GetTotalCount())
	}
}

func TestBFSStrategy_GetTotalCount_Empty(t *testing.T) {
	items := []SecurityItem{}

	strategy := NewBFSStrategy(items)

	if strategy.GetTotalCount() != 0 {
		t.Errorf("expected total count 0, got %d", strategy.GetTotalCount())
	}
}

func TestBFSStrategy_GetViewedCount_Empty(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "CVE 1"},
	}

	strategy := NewBFSStrategy(items)

	if strategy.GetViewedCount() != 0 {
		t.Errorf("expected viewed count 0, got %d", strategy.GetViewedCount())
	}
}
