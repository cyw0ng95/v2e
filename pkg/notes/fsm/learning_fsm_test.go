package fsm

import (
	"context"
	"testing"
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
