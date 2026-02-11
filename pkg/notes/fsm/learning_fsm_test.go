package fsm

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestNewLearningFSM(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE", Source: "cwe"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Check initial state
	if lfsm.state != LearningStateIdle {
		t.Errorf("expected initial state %s, got %s", LearningStateIdle, lfsm.state)
	}
	if lfsm.currentStrategy != "bfs" {
		t.Errorf("expected initial strategy bfs, got %s", lfsm.currentStrategy)
	}
	if len(lfsm.viewedItems) != 0 {
		t.Errorf("expected no viewed items, got %d", len(lfsm.viewedItems))
	}
}

func TestLearningFSM_LoadItem_BFS(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE 1", Source: "nvd"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "Test CVE 2", Source: "nvd"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "Test CVE 3", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Load first item
	item1, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if item1.URN != "v2e::cve::1" {
		t.Errorf("expected first item v2e::cve::1, got %s", item1.URN)
	}
	if item1.Context != "browsing" {
		t.Errorf("expected context browsing, got %s", item1.Context)
	}

	// Mark as viewed
	err = lfsm.MarkViewed("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	// Verify it was marked as viewed
	if len(lfsm.viewedItems) != 1 {
		t.Errorf("expected 1 viewed item, got %d", len(lfsm.viewedItems))
	}

	// Load second item by clearing current item first
	err = lfsm.MarkLearned(item1.URN)
	if err != nil {
		t.Fatal(err)
	}

	item2, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if item2.URN != "v2e::cve::2" {
		t.Errorf("expected second item v2e::cve::2, got %s", item2.URN)
	}
}

func TestLearningFSM_MarkViewed(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.MarkViewed("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	if len(lfsm.viewedItems) != 1 {
		t.Errorf("expected 1 viewed item, got %d", len(lfsm.viewedItems))
	}
	if lfsm.viewedItems[0] != "v2e::cve::1" {
		t.Errorf("expected viewed item v2e::cve::1, got %s", lfsm.viewedItems[0])
	}
}

func TestLearningFSM_MarkLearned(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.MarkLearned("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	if len(lfsm.completedItems) != 1 {
		t.Errorf("expected 1 completed item, got %d", len(lfsm.completedItems))
	}
	if lfsm.completedItems[0] != "v2e::cve::1" {
		t.Errorf("expected completed item v2e::cve::1, got %s", lfsm.completedItems[0])
	}
}

func TestLearningFSM_StatePersistence(t *testing.T) {
	dbPath := t.TempDir() + "/test.db"
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage1, err := NewBoltDBStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}

	lfsm1, err := NewLearningFSM(storage1, items)
	if err != nil {
		t.Fatal(err)
	}

	// Mark as viewed and completed
	err = lfsm1.MarkViewed("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm1.MarkLearned("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	// Save state
	err = lfsm1.SaveState()
	if err != nil {
		t.Fatal(err)
	}

	// Close and reopen
	storage1.Close()

	storage2, err := NewBoltDBStorage(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer storage2.Close()

	lfsm2, err := NewLearningFSM(storage2, items)
	if err != nil {
		t.Fatal(err)
	}

	// Check state was restored
	if len(lfsm2.viewedItems) != 1 {
		t.Errorf("expected 1 viewed item, got %d", len(lfsm2.viewedItems))
	}
	if len(lfsm2.completedItems) != 1 {
		t.Errorf("expected 1 completed item, got %d", len(lfsm2.completedItems))
	}
}

func TestLearningFSM_GetContext(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE 1", Source: "nvd"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "Test CVE 2", Source: "nvd"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "Test CVE 3", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Mark first two as learned
	err = lfsm.MarkLearned("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.MarkLearned("v2e::cve::2")
	if err != nil {
		t.Fatal(err)
	}

	// Get context
	ctx := lfsm.GetContext()

	if len(ctx.ViewedItems) != 0 {
		t.Errorf("expected 0 viewed items in context, got %d", len(ctx.ViewedItems))
	}
	if len(ctx.CompletedItems) != 2 {
		t.Errorf("expected 2 completed items in context, got %d", len(ctx.CompletedItems))
	}
	if len(ctx.AvailableItems) != 3 {
		t.Errorf("expected 3 available items in context, got %d", len(ctx.AvailableItems))
	}
}

func TestLearningFSM_GetCurrentItemViaLoadItem(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Load an item
	item, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if item.URN != "v2e::cve::1" {
		t.Errorf("expected item v2e::cve::1, got %s", item.URN)
	}
}

func TestLearningFSM_PauseAndResume(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Load an item to set state to browsing
	_, err = lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Pause
	err = lfsm.Pause()
	if err != nil {
		t.Fatal(err)
	}

	if lfsm.state != LearningStatePaused {
		t.Errorf("expected state %s, got %s", LearningStatePaused, lfsm.state)
	}

	// Resume
	err = lfsm.Resume()
	if err != nil {
		t.Fatal(err)
	}

	if lfsm.state != LearningStateBrowsing {
		t.Errorf("expected state %s, got %s", LearningStateBrowsing, lfsm.state)
	}
}

func TestItemGraph(t *testing.T) {
	graph := NewItemGraph()

	// Add links
	graph.AddLink("v2e::cve::1", "v2e::cwe::1")
	graph.AddLink("v2e::cve::1", "v2e::capec::1")

	// Get links
	links := graph.GetLinks("v2e::cve::1")
	if len(links) != 2 {
		t.Errorf("expected 2 links, got %d", len(links))
	}

	// Check specific links
	hasCWE := false
	hasCAPEC := false
	for _, link := range links {
		if link == "v2e::cwe::1" {
			hasCWE = true
		}
		if link == "v2e::capec::1" {
			hasCAPEC = true
		}
	}

	if !hasCWE {
		t.Error("expected link to v2e::cwe::1")
	}
	if !hasCAPEC {
		t.Error("expected link to v2e::capec::1")
	}
}

func TestLearningFSM_FollowLink_DFS(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE", Source: "cwe"},
		{URN: "v2e::capec::1", Type: "capec", ID: "1", Title: "Test CAPEC", Source: "capec"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Load initial item
	item1, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Follow link to related item
	err = lfsm.FollowLink(item1.URN, "v2e::cwe::1")
	if err != nil {
		t.Fatal(err)
	}

	// Check strategy changed to DFS
	if lfsm.currentStrategy != "dfs" {
		t.Errorf("expected strategy dfs, got %s", lfsm.currentStrategy)
	}

	// Check item was added to path stack
	if len(lfsm.pathStack) == 0 {
		t.Error("expected non-empty path stack")
	}

	// Check state changed to deep dive
	if lfsm.state != LearningStateDeepDive {
		t.Errorf("expected state deep_dive, got %s", lfsm.state)
	}
}

func TestLearningFSM_GoBack(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE", Source: "cwe"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Setup DFS navigation
	item1, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.FollowLink(item1.URN, "v2e::cwe::1")
	if err != nil {
		t.Fatal(err)
	}

	// Go back should return to previous item
	prevItem, err := lfsm.GoBack(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if prevItem.URN != "v2e::cve::1" {
		t.Errorf("expected previous item v2e::cve::1, got %s", prevItem.URN)
	}

	// Path stack should be empty
	if len(lfsm.pathStack) != 0 {
		t.Errorf("expected empty path stack, got %d items", len(lfsm.pathStack))
	}
}

func TestLearningFSM_GoBack_EmptyStack_SwitchesToBFS(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Try to go back with empty stack
	item, err := lfsm.GoBack(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Should return first BFS item
	if item.URN != "v2e::cve::1" {
		t.Errorf("expected first item v2e::cve::1, got %s", item.URN)
	}

	// Should switch to BFS strategy
	if lfsm.currentStrategy != "bfs" {
		t.Errorf("expected strategy bfs, got %s", lfsm.currentStrategy)
	}
}

func TestLearningFSM_LoadItem_DFS(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE", Source: "cwe"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Manually set DFS mode and add item to stack
	lfsm.mu.Lock()
	lfsm.currentStrategy = "dfs"
	lfsm.pathStack = append(lfsm.pathStack, "v2e::cve::1")
	lfsm.mu.Unlock()

	item, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if item.URN != "v2e::cve::1" {
		t.Errorf("expected item v2e::cve::1, got %s", item.URN)
	}

	if item.Context != "deep_dive" {
		t.Errorf("expected context deep_dive, got %s", item.Context)
	}
}

func TestLearningFSM_MarkViewed_Duplicate(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Mark as viewed twice
	err = lfsm.MarkViewed("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.MarkViewed("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	// Should only have one viewed item
	if len(lfsm.viewedItems) != 1 {
		t.Errorf("expected 1 viewed item, got %d", len(lfsm.viewedItems))
	}
}

func TestLearningFSM_MarkLearned_Duplicate(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Mark as learned twice
	err = lfsm.MarkLearned("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.MarkLearned("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	// Should only have one learned item
	if len(lfsm.completedItems) != 1 {
		t.Errorf("expected 1 completed item, got %d", len(lfsm.completedItems))
	}
}

func TestLearningFSM_MarkLearned_ClearsCurrentItem(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Load item
	item, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if lfsm.currentItemURN != item.URN {
		t.Error("expected current item to be set")
	}

	// Mark as learned
	err = lfsm.MarkLearned(item.URN)
	if err != nil {
		t.Fatal(err)
	}

	// Current item should be cleared
	if lfsm.currentItemURN != "" {
		t.Errorf("expected current item to be cleared, got %s", lfsm.currentItemURN)
	}
}

func TestLearningFSM_Resume_FromNonPausedState(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Try to resume from idle state
	err = lfsm.Resume()
	if err == nil {
		t.Error("expected error when resuming from non-paused state")
	}
}

func TestLearningFSM_GetState_ThreadSafe(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Get state should be thread-safe
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			lfsm.GetState()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestLearningFSM_GetContext_Immutable(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	ctx := lfsm.GetContext()

	// Modify context should not affect original FSM
	ctx.ViewedItems = append(ctx.ViewedItems, "v2e::cve::2")
	ctx.CompletedItems = append(ctx.CompletedItems, "v2e::cve::3")

	if len(lfsm.viewedItems) != 0 {
		t.Error("modifying context should not affect original FSM")
	}

	if len(lfsm.completedItems) != 0 {
		t.Error("modifying context should not affect original FSM")
	}
}

func TestLearningFSM_LoadItem_NoMoreItems(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Mark all items as viewed
	for _, item := range items {
		err = lfsm.MarkViewed(item.URN)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Try to load next item
	_, err = lfsm.LoadItem(context.Background())
	if err == nil {
		t.Error("expected error when no more items")
	}
}

func TestLearningFSM_NilStorage(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	// Create FSM with nil storage
	lfsm, err := NewLearningFSM(nil, items)
	if err != nil {
		t.Fatal(err)
	}

	// Operations should still work without storage
	err = lfsm.MarkViewed("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.MarkLearned("v2e::cve::1")
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.SaveState()
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.LoadState()
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.Pause()
	if err != nil {
		t.Fatal(err)
	}

	err = lfsm.Resume()
	if err != nil {
		t.Fatal(err)
	}
}

func TestLearningFSM_GetItemByURN_NotFound(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Try to get non-existent item
	_, err = lfsm.getItemByURN("v2e::cve::999")
	if err == nil {
		t.Error("expected error for non-existent item")
	}
}

func TestLearningFSM_LastActivityUpdate(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	initialActivity := lfsm.lastActivity

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Perform operation that updates last activity
	_ = lfsm.MarkViewed("v2e::cve::1")

	if !lfsm.lastActivity.After(initialActivity) {
		t.Error("expected last activity to be updated")
	}
}

func TestLearningFSM_BuildItemGraph(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "Test CVE 2", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE", Source: "cwe"},
		{URN: "v2e::cwe::2", Type: "cwe", ID: "2", Title: "Test CWE 2", Source: "cwe"},
		{URN: "v2e::capec::1", Type: "capec", ID: "1", Title: "Test CAPEC", Source: "capec"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Check graph was built
	if lfsm.itemGraph == nil {
		t.Error("expected item graph to be built")
	}

	// Check links exist
	cveLinks := lfsm.itemGraph.GetLinks("v2e::cve::1")
	if len(cveLinks) == 0 {
		t.Error("expected CVE to have links to CWE")
	}
}

func TestLearningFSM_UnknownStrategy(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Set unknown strategy
	lfsm.mu.Lock()
	lfsm.currentStrategy = "unknown"
	lfsm.mu.Unlock()

	// Load item should fail with unknown strategy
	_, err = lfsm.LoadItem(context.Background())
	if err == nil {
		t.Error("expected error for unknown strategy")
	}
}

func TestLearningFSM_ContextTimeout_SaveState(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Create context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// Wait for context to expire
	time.Sleep(10 * time.Millisecond)

	// Save should fail due to timeout
	err = lfsm.SaveStateWithContext(ctx)
	if err == nil {
		t.Error("expected timeout error")
	}
}

func TestLearningFSM_ContextAlreadyCancelled_SaveState(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Create already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Save should fail immediately
	err = lfsm.SaveStateWithContext(ctx)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestLearningFSM_MultipleItems_BFS_Sequential(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE 1", Source: "nvd"},
		{URN: "v2e::cve::2", Type: "cve", ID: "2", Title: "Test CVE 2", Source: "nvd"},
		{URN: "v2e::cve::3", Type: "cve", ID: "3", Title: "Test CVE 3", Source: "nvd"},
		{URN: "v2e::cve::4", Type: "cve", ID: "4", Title: "Test CVE 4", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Load all items sequentially
	for i := 0; i < 4; i++ {
		lfsm.mu.Lock()
		lfsm.currentItemURN = ""
		lfsm.mu.Unlock()

		item, err := lfsm.LoadItem(context.Background())
		if err != nil {
			t.Fatal(err)
		}

		expectedURN := fmt.Sprintf("v2e::cve::%d", i+1)
		if item.URN != expectedURN {
			t.Errorf("item %d: expected %s, got %s", i, expectedURN, item.URN)
		}

		// Mark as viewed (not learned) so it doesn't affect currentItemURN
		err = lfsm.MarkViewed(item.URN)
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestLearningFSM_DFS_DeepNavigation(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE 1", Source: "cwe"},
		{URN: "v2e::cwe::2", Type: "cwe", ID: "2", Title: "Test CWE 2", Source: "cwe"},
		{URN: "v2e::capec::1", Type: "capec", ID: "1", Title: "Test CAPEC 1", Source: "capec"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Start with CVE
	item1, err := lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Dive into CWE 1
	err = lfsm.FollowLink(item1.URN, "v2e::cwe::1")
	if err != nil {
		t.Fatal(err)
	}

	// Dive into CAPEC
	err = lfsm.FollowLink("v2e::cwe::1", "v2e::capec::1")
	if err != nil {
		t.Fatal(err)
	}

	// Should have 2 items in path stack
	if len(lfsm.pathStack) != 2 {
		t.Errorf("expected 2 items in path stack, got %d", len(lfsm.pathStack))
	}

	// Go back to CWE
	item2, err := lfsm.GoBack(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if item2.URN != "v2e::cwe::1" {
		t.Errorf("expected to go back to v2e::cwe::1, got %s", item2.URN)
	}

	// Stack should have 1 item now
	if len(lfsm.pathStack) != 1 {
		t.Errorf("expected 1 item in path stack, got %d", len(lfsm.pathStack))
	}
}

func TestItemGraph_EmptyLinks(t *testing.T) {
	graph := NewItemGraph()

	links := graph.GetLinks("nonexistent")
	if links != nil {
		t.Errorf("expected nil for non-existent URN, got %v", links)
	}
}

func TestItemGraph_ThreadSafe(t *testing.T) {
	graph := NewItemGraph()

	done := make(chan bool)

	// Concurrent AddLink
	for i := 0; i < 10; i++ {
		go func(idx int) {
			graph.AddLink(fmt.Sprintf("urn%d", idx), fmt.Sprintf("urn%d", idx+1))
			done <- true
		}(i)
	}

	// Concurrent GetLinks
	for i := 0; i < 10; i++ {
		go func(idx int) {
			graph.GetLinks(fmt.Sprintf("urn%d", idx))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestLearningFSM_PathStackLimit(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE", Source: "cwe"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Build up path stack
	for i := 0; i < 100; i++ {
		lfsm.pathStack = append(lfsm.pathStack, fmt.Sprintf("urn%d", i))
	}

	// Stack should grow without limit
	if len(lfsm.pathStack) != 100 {
		t.Errorf("expected 100 items in stack, got %d", len(lfsm.pathStack))
	}
}

func TestLearningFSM_StateTransition_BFS_To_DFS(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
		{URN: "v2e::cwe::1", Type: "cwe", ID: "1", Title: "Test CWE", Source: "cwe"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Start in BFS mode
	if lfsm.state != LearningStateIdle {
		t.Errorf("expected initial state idle, got %s", lfsm.state)
	}

	// Load item to switch to browsing
	_, err = lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if lfsm.state != LearningStateBrowsing {
		t.Errorf("expected state browsing, got %s", lfsm.state)
	}

	// Follow link to switch to DFS
	err = lfsm.FollowLink("v2e::cve::1", "v2e::cwe::1")
	if err != nil {
		t.Fatal(err)
	}

	if lfsm.state != LearningStateDeepDive {
		t.Errorf("expected state deep_dive, got %s", lfsm.state)
	}
}

func TestLearningFSM_StateTransition_Paused(t *testing.T) {
	items := []SecurityItem{
		{URN: "v2e::cve::1", Type: "cve", ID: "1", Title: "Test CVE", Source: "nvd"},
	}

	storage, err := NewBoltDBStorage(t.TempDir() + "/test.db")
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()

	lfsm, err := NewLearningFSM(storage, items)
	if err != nil {
		t.Fatal(err)
	}

	// Load item to set state
	_, err = lfsm.LoadItem(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Pause
	err = lfsm.Pause()
	if err != nil {
		t.Fatal(err)
	}

	if lfsm.state != LearningStatePaused {
		t.Errorf("expected state paused, got %s", lfsm.state)
	}

	// Resume
	err = lfsm.Resume()
	if err != nil {
		t.Fatal(err)
	}

	if lfsm.state != LearningStateBrowsing {
		t.Errorf("expected state browsing after resume, got %s", lfsm.state)
	}
}
