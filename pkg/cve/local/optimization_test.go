package local

import (
	"testing"
	"time"
)

func TestQueryPatternAnalyzer_Basic(t *testing.T) {
	qpa := NewQueryPatternAnalyzer(100)

	qpa.RecordQuery("SELECT * FROM cve WHERE id = ?", time.Millisecond)
	qpa.RecordQuery("SELECT * FROM cve WHERE id = ?", 2*time.Millisecond)

	stats := qpa.GetStats()
	if stats["total_patterns"] != 1 {
		t.Errorf("Expected 1 pattern, got %v", stats["total_patterns"])
	}
}

func TestQueryPatternAnalyzer_FrequentQueries(t *testing.T) {
	qpa := NewQueryPatternAnalyzer(100)

	sql1 := "SELECT * FROM cve WHERE id = ?"
	sql2 := "SELECT * FROM cwe WHERE id = ?"

	for i := 0; i < 10; i++ {
		qpa.RecordQuery(sql1, time.Millisecond)
	}

	for i := 0; i < 5; i++ {
		qpa.RecordQuery(sql2, time.Millisecond)
	}

	frequent := qpa.GetFrequentQueries(2)
	if len(frequent) != 2 {
		t.Errorf("Expected 2 frequent queries, got %d", len(frequent))
	}

	if frequent[0].ExecuteCount != 10 {
		t.Errorf("Expected first query executed 10 times, got %d", frequent[0].ExecuteCount)
	}
}

func TestIntelligentBatcher_BasicFlush(t *testing.T) {
	db, err := NewOptimizedDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	ib := NewIntelligentBatcher(db.db, 2, 2, 10, 1*time.Second)
	defer ib.Close()

	executedCount := 0
	taskHandler := func(data interface{}) error {
		executedCount++
		return nil
	}

	for i := 0; i < 5; i++ {
		task := &BatchTask{
			ID:        string([]byte{byte(i)}),
			Operation: "test",
			Data:      i,
			Handler:   taskHandler,
		}
		ib.AddTask(task)
	}

	ib.Flush()

	if executedCount != 5 {
		t.Errorf("Expected 5 tasks executed, got %d", executedCount)
	}
}

func TestIntelligentBatcher_Merging(t *testing.T) {
	db, err := NewOptimizedDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	ib := NewIntelligentBatcher(db.db, 2, 5, 20, 1*time.Second)
	defer ib.Close()

	executedCount := 0
	taskHandler := func(data interface{}) error {
		executedCount++
		return nil
	}

	for i := 0; i < 3; i++ {
		task := &BatchTask{
			ID:        string([]byte{byte(i)}),
			Operation: "test",
			Data:      i,
			Handler:   taskHandler,
		}
		ib.AddTask(task)
	}

	ib.Flush()

	if executedCount != 3 {
		t.Errorf("Expected 3 tasks executed, got %d", executedCount)
	}
}

func TestIntelligentBatcher_Metrics(t *testing.T) {
	db, err := NewOptimizedDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test DB: %v", err)
	}
	defer db.Close()

	// Use minBatchSize larger than test size to avoid async flushes during AddTask
	ib := NewIntelligentBatcher(db.db, 2, 20, 50, 10*time.Second)
	defer ib.Close()

	taskHandler := func(data interface{}) error {
		return nil
	}

	for i := 0; i < 10; i++ {
		task := &BatchTask{
			ID:        string([]byte{byte(i)}),
			Operation: "test",
			Data:      i,
			Handler:   taskHandler,
		}
		ib.AddTask(task)
	}

	ib.Flush()

	metrics := ib.GetMetrics()
	if metrics.TotalTasks != 10 {
		t.Errorf("Expected 10 total tasks, got %d", metrics.TotalTasks)
	}

	if metrics.SuccessfulTasks != 10 {
		t.Errorf("Expected 10 successful tasks, got %d", metrics.SuccessfulTasks)
	}
}

func TestPreparedStatementCache_Basic(t *testing.T) {
	psc := NewPreparedStatementCache(10)

	stmt := "test-stmt"
	psc.Set("key1", stmt)

	cached := psc.Get("key1")
	if cached != stmt {
		t.Errorf("Expected cached value, got %v", cached)
	}

	missed := psc.Get("nonexistent")
	if missed != nil {
		t.Errorf("Expected nil for missing key, got %v", missed)
	}
}

func TestPreparedStatementCache_Eviction(t *testing.T) {
	psc := NewPreparedStatementCache(3)

	for i := 0; i < 5; i++ {
		key := string([]byte{byte('a' + i)})
		psc.Set(key, i)
	}

	if len(psc.cache) > 3 {
		t.Errorf("Expected max 3 items in cache, got %d", len(psc.cache))
	}
}

func TestPreparedStatementCache_Metrics(t *testing.T) {
	psc := NewPreparedStatementCache(10)

	psc.Set("key1", 1)
	psc.Set("key2", 2)

	psc.Get("key1")
	psc.Get("key2")
	psc.Get("nonexistent")

	if psc.metrics.CacheHits != 2 {
		t.Errorf("Expected 2 cache hits, got %d", psc.metrics.CacheHits)
	}

	if psc.metrics.CacheMisses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", psc.metrics.CacheMisses)
	}
}
