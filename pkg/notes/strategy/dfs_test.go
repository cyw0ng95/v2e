package strategy

import (
	"context"
	"fmt"
	"testing"
)

func TestNewItemGraph(t *testing.T) {
	graph := NewItemGraph()

	if graph == nil {
		t.Fatal("NewItemGraph returned nil")
	}

	if graph.links == nil {
		t.Error("Expected links map to be initialized")
	}

	if len(graph.links) != 0 {
		t.Error("Expected empty links map")
	}
}

func TestItemGraph_AddLink(t *testing.T) {
	graph := NewItemGraph()

	graph.AddLink("urn:1", "urn:2")
	graph.AddLink("urn:1", "urn:3")
	graph.AddLink("urn:2", "urn:4")

	links1 := graph.GetLinks("urn:1")
	if len(links1) != 2 {
		t.Errorf("Expected 2 links for urn:1, got %d", len(links1))
	}

	if len(graph.GetLinks("urn:2")) != 1 {
		t.Error("Expected 1 link for urn:2")
	}

	if graph.GetLinks("urn:nonexistent") != nil {
		t.Error("Expected nil for nonexistent URN")
	}
}

func TestItemGraph_GetLinks(t *testing.T) {
	graph := NewItemGraph()
	graph.AddLink("urn:1", "urn:2")
	graph.AddLink("urn:1", "urn:3")

	links := graph.GetLinks("urn:1")
	if len(links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(links))
	}

	links2 := graph.GetLinks("urn:1")
	if links2[0] != links[0] || links2[1] != links[1] {
		t.Error("GetLinks should return a copy, not the original slice")
	}
}

func TestNewDFSStrategy(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	if strategy == nil {
		t.Fatal("NewDFSStrategy returned nil")
	}

	if strategy.itemGraph != graph {
		t.Error("Expected itemGraph to be set")
	}

	if len(strategy.pathStack) != 0 {
		t.Error("Expected empty pathStack")
	}

	if len(strategy.viewed) != 0 {
		t.Error("Expected empty viewed map")
	}
}

func TestDFSStrategy_Name(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	name := strategy.Name()
	if name != "dfs" {
		t.Errorf("Expected name 'dfs', got '%s'", name)
	}
}

func TestDFSStrategy_NextItem(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)
	ctx := context.Background()

	current := &LearningContext{
		AvailableItems: []SecurityItem{
			{URN: "urn:1", Type: "cve", Title: "CVE-2024-001"},
			{URN: "urn:2", Type: "cwe", Title: "CWE-79"},
		},
	}

	item, err := strategy.NextItem(ctx, current)
	if err != ErrSwitchStrategy {
		t.Errorf("Expected ErrSwitchStrategy when stack is empty, got %v", err)
	}

	if item != nil {
		t.Error("Expected nil item when stack is empty")
	}

	strategy.PushStack("urn:1")
	item, err = strategy.NextItem(ctx, current)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if item == nil {
		t.Fatal("Expected item, got nil")
	}

	if item.URN != "urn:1" {
		t.Errorf("Expected URN 'urn:1', got '%s'", item.URN)
	}

	if item.Context != "deep_dive" {
		t.Errorf("Expected context 'deep_dive', got '%s'", item.Context)
	}
}

func TestDFSStrategy_NextItem_NotFound(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)
	ctx := context.Background()

	current := &LearningContext{
		AvailableItems: []SecurityItem{
			{URN: "urn:1", Type: "cve", Title: "CVE-2024-001"},
		},
	}

	strategy.PushStack("urn:nonexistent")
	item, err := strategy.NextItem(ctx, current)

	if err == nil {
		t.Error("Expected error for nonexistent item")
	}

	if item != nil {
		t.Error("Expected nil item for nonexistent URN")
	}
}

func TestDFSStrategy_OnView(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	err := strategy.OnView("urn:1")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	count := strategy.GetViewedCount()
	if count != 1 {
		t.Errorf("Expected viewed count 1, got %d", count)
	}

	strategy.OnView("urn:2")
	count = strategy.GetViewedCount()
	if count != 2 {
		t.Errorf("Expected viewed count 2, got %d", count)
	}

	strategy.OnView("urn:1")
	count = strategy.GetViewedCount()
	if count != 2 {
		t.Error("Viewed count should not increase for duplicate views")
	}
}

func TestDFSStrategy_OnFollowLink(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	err := strategy.OnFollowLink("urn:1", "urn:2")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if strategy.GetStackDepth() != 1 {
		t.Errorf("Expected stack depth 1, got %d", strategy.GetStackDepth())
	}

	if strategy.GetViewedCount() != 1 {
		t.Error("Expected viewed count 1")
	}

	err = strategy.OnFollowLink("", "urn:3")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if strategy.GetStackDepth() != 1 {
		t.Error("Stack depth should not increase when fromURN is empty")
	}
}

func TestDFSStrategy_OnGoBack(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)
	ctx := context.Background()

	strategy.PushStack("urn:1")
	strategy.PushStack("urn:2")

	item, err := strategy.OnGoBack(ctx)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if item == nil {
		t.Fatal("Expected item, got nil")
	}

	if item.URN != "urn:2" {
		t.Errorf("Expected URN 'urn:2', got '%s'", item.URN)
	}

	if strategy.GetStackDepth() != 1 {
		t.Errorf("Expected stack depth 1 after pop, got %d", strategy.GetStackDepth())
	}

	item, _ = strategy.OnGoBack(ctx)
	if item.URN != "urn:1" {
		t.Errorf("Expected URN 'urn:1', got '%s'", item.URN)
	}
}

func TestDFSStrategy_OnGoBack_EmptyStack(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)
	ctx := context.Background()

	item, err := strategy.OnGoBack(ctx)

	if err != ErrSwitchStrategy {
		t.Errorf("Expected ErrSwitchStrategy when stack is empty, got %v", err)
	}

	if item != nil {
		t.Error("Expected nil item when stack is empty")
	}
}

func TestDFSStrategy_Reset(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	strategy.PushStack("urn:1")
	strategy.PushStack("urn:2")
	strategy.OnView("urn:1")
	strategy.OnView("urn:3")

	err := strategy.Reset()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if strategy.GetStackDepth() != 0 {
		t.Errorf("Expected stack depth 0 after reset, got %d", strategy.GetStackDepth())
	}

	if strategy.GetViewedCount() != 0 {
		t.Error("Expected viewed count 0 after reset")
	}
}

func TestDFSStrategy_GetLinkedItems(t *testing.T) {
	graph := NewItemGraph()
	graph.AddLink("urn:1", "urn:2")
	graph.AddLink("urn:1", "urn:3")

	strategy := NewDFSStrategy(graph)
	links := strategy.GetLinkedItems("urn:1")

	if len(links) != 2 {
		t.Errorf("Expected 2 linked items, got %d", len(links))
	}

	links = strategy.GetLinkedItems("urn:nonexistent")
	if links != nil {
		t.Error("Expected nil for nonexistent URN")
	}
}

func TestDFSStrategy_GetLinkedItems_NilGraph(t *testing.T) {
	strategy := NewDFSStrategy(nil)
	links := strategy.GetLinkedItems("urn:1")

	if links != nil {
		t.Error("Expected nil when itemGraph is nil")
	}
}

func TestDFSStrategy_PushStack(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	strategy.PushStack("urn:1")
	if strategy.GetStackDepth() != 1 {
		t.Errorf("Expected stack depth 1, got %d", strategy.GetStackDepth())
	}

	strategy.PushStack("urn:2")
	strategy.PushStack("urn:3")
	if strategy.GetStackDepth() != 3 {
		t.Errorf("Expected stack depth 3, got %d", strategy.GetStackDepth())
	}
}

func TestDFSStrategy_GetStackDepth(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	if strategy.GetStackDepth() != 0 {
		t.Errorf("Expected initial stack depth 0, got %d", strategy.GetStackDepth())
	}

	strategy.PushStack("urn:1")
	if strategy.GetStackDepth() != 1 {
		t.Errorf("Expected stack depth 1, got %d", strategy.GetStackDepth())
	}

	for i := 0; i < 10; i++ {
		strategy.PushStack("urn:1")
	}

	if strategy.GetStackDepth() != 11 {
		t.Errorf("Expected stack depth 11, got %d", strategy.GetStackDepth())
	}
}

func TestDFSStrategy_GetViewedCount(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	if strategy.GetViewedCount() != 0 {
		t.Errorf("Expected initial viewed count 0, got %d", strategy.GetViewedCount())
	}

	for i := 0; i < 5; i++ {
		urn := fmt.Sprintf("urn:%d", i)
		strategy.OnView(urn)
	}

	if strategy.GetViewedCount() != 5 {
		t.Errorf("Expected viewed count 5, got %d", strategy.GetViewedCount())
	}

	// Add one more distinct URN
	strategy.OnView("urn:5")
}

func TestDFSStrategy_NextItem_FullCycle(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)
	ctx := context.Background()

	current := &LearningContext{
		AvailableItems: []SecurityItem{
			{URN: "urn:1", Type: "cve", Title: "Item 1"},
			{URN: "urn:2", Type: "cwe", Title: "Item 2"},
			{URN: "urn:3", Type: "capec", Title: "Item 3"},
		},
	}

	strategy.PushStack("urn:1")
	strategy.PushStack("urn:2")
	strategy.PushStack("urn:3")

	items := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		item, err := strategy.NextItem(ctx, current)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		items = append(items, item.URN)
	}

	if items[0] != "urn:3" || items[1] != "urn:2" || items[2] != "urn:1" {
		t.Errorf("Expected items in LIFO order, got %v", items)
	}

	_, err := strategy.NextItem(ctx, current)
	if err != ErrSwitchStrategy {
		t.Errorf("Expected ErrSwitchStrategy when stack is empty, got %v", err)
	}
}

func TestDFSStrategy_ThreadSafety(t *testing.T) {
	graph := NewItemGraph()
	strategy := NewDFSStrategy(graph)

	done := make(chan bool)

	for i := 0; i < 100; i++ {
		go func(id int) {
			defer func() { done <- true }()
			urn := "urn:" + string(rune('1'+id%5))
			strategy.PushStack(urn)
			strategy.OnView(urn)
			strategy.GetLinkedItems(urn)
			strategy.GetStackDepth()
			strategy.GetViewedCount()
		}(i)
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	strategy.Reset()
}

func TestItemGraph_Bidirectional(t *testing.T) {
	graph := NewItemGraph()

	graph.AddLink("urn:1", "urn:2")
	graph.AddLink("urn:2", "urn:1")

	links1 := graph.GetLinks("urn:1")
	links2 := graph.GetLinks("urn:2")

	if len(links1) != 1 || links1[0] != "urn:2" {
		t.Error("Expected urn:1 to link to urn:2")
	}

	if len(links2) != 1 || links2[0] != "urn:1" {
		t.Error("Expected urn:2 to link to urn:1")
	}
}
