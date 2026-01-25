package taskflow

import (
	"sync"
	"testing"
)

// Test concurrent UpdateProgress calls aggregate correctly.
func TestRunStore_UpdateProgress_Concurrent(t *testing.T) {
	rs := NewTempRunStore(t)
	defer rs.Close()

	runID := "concurrent-run"
	if _, err := rs.CreateRun(runID, 0, 10, DataTypeCVE); err != nil {
		t.Fatalf("CreateRun failed: %v", err)
	}

	const goroutines = 10
	const perG = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				if err := rs.UpdateProgress(runID, 1, 2, 0); err != nil {
					t.Fatalf("UpdateProgress failed: %v", err)
				}
			}
		}()
	}
	wg.Wait()

	r, err := rs.GetRun(runID)
	if err != nil {
		t.Fatalf("GetRun failed: %v", err)
	}

	expectedFetched := int64(goroutines * perG)
	expectedStored := int64(goroutines * perG * 2)

	if r.FetchedCount != expectedFetched {
		t.Fatalf("unexpected fetched count: got %d want %d", r.FetchedCount, expectedFetched)
	}
	if r.StoredCount != expectedStored {
		t.Fatalf("unexpected stored count: got %d want %d", r.StoredCount, expectedStored)
	}
	if r.State == StateFailed {
		t.Fatalf("run ended up failed unexpectedly: %+v", r)
	}
}
