package local

import (
	"context"
	"testing"
	"time"
)

func BenchmarkSmartConnectionPool_Query(b *testing.B) {
	db, err := NewOptimizedDB(":memory:")
	if err != nil {
		b.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	scp, err := NewSmartConnectionPool(db.db, 10, 5)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}
	defer scp.Close()

	ctx := context.Background()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scp.Query(ctx, "SELECT 1")
	}
}

func BenchmarkSmartConnectionPool_ConcurrentQuery(b *testing.B) {
	db, err := NewOptimizedDB(":memory:")
	if err != nil {
		b.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	scp, err := NewSmartConnectionPool(db.db, 10, 5)
	if err != nil {
		b.Fatalf("Failed to create pool: %v", err)
	}
	defer scp.Close()

	ctx := context.Background()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			scp.Query(ctx, "SELECT 1")
		}
	})
}

func BenchmarkQueryPatternAnalyzer_Record(b *testing.B) {
	qpa := NewQueryPatternAnalyzer(100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		qpa.RecordQuery("SELECT * FROM cve WHERE id = ?", time.Microsecond)
	}
}

func BenchmarkIntelligentBatcher_Flush(b *testing.B) {
	db, err := NewOptimizedDB(":memory:")
	if err != nil {
		b.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	ib := NewIntelligentBatcher(db.db, 2, 50, 100, 10*time.Millisecond)
	defer ib.Close()

	taskHandler := func(data interface{}) error {
		return nil
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			task := &BatchTask{
				ID:        "test",
				Operation: "test",
				Data:      pb,
				Handler:   taskHandler,
			}
			ib.AddTask(task)
		}
	})

	ib.Flush()
}

func BenchmarkPreparedStatementCache_Get(b *testing.B) {
	psc := NewPreparedStatementCache(100)
	psc.Set("key", "value")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		psc.Get("key")
	}
}

func BenchmarkPreparedStatementCache_Set(b *testing.B) {
	psc := NewPreparedStatementCache(100)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		key := string([]byte{byte(i % 100)})
		psc.Set(key, i)
	}
}

func BenchmarkBatchOperations(b *testing.B) {
	db, err := NewOptimizedDB(":memory:")
	if err != nil {
		b.Fatalf("Failed to create DB: %v", err)
	}
	defer db.Close()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			db.db.Exec("SELECT 1")
		}
	})
}
