package taskflow

import (
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

func BenchmarkTieredPool_Get_Put(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := tp.Get(1000)
		tp.Put(buf)
	}
}

func BenchmarkTieredPool_Get_Put_Direct(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 1000)
		_ = buf
	}
}

func BenchmarkTieredPool_MixedSizes(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	sizes := []int{100, 500, 2000, 8000, 20000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		size := sizes[i%len(sizes)]
		buf := tp.Get(size)
		tp.Put(buf)
	}
}

func BenchmarkTieredPool_GetOnly(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := tp.Get(1000)
		_ = buf
	}
}

func BenchmarkTieredPool_PutOnly(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	// Warm up pool
	bufs := make([][]byte, 100)
	for i := range bufs {
		bufs[i] = tp.Get(1000)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bufs[i%len(bufs)]
		tp.Put(buf)
	}
}

func BenchmarkTieredPool_SmallBuffer(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := tp.Get(100)
		tp.Put(buf)
	}
}

func BenchmarkTieredPool_MediumBuffer(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := tp.Get(2000)
		tp.Put(buf)
	}
}

func BenchmarkTieredPool_LargeBuffer(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := tp.Get(15000)
		tp.Put(buf)
	}
}

func BenchmarkTieredPool_Concurrent(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := tp.Get(1000)
			tp.Put(buf)
		}
	})
}

func BenchmarkTieredPool_WithMetrics(b *testing.B) {
	tp := NewTieredPoolWithDefaults()
	metrics := NewPoolMetrics()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := tp.Get(1000)
		metrics.RecordAllocation(PoolSmall, 1000, 0)
		metrics.RecordHit(PoolSmall)
		tp.Put(buf)
		metrics.RecordRelease(PoolSmall)
	}
}

func BenchmarkTieredPool_NoReuse(b *testing.B) {
	tp := NewTieredPoolWithDefaults()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := tp.Get(1000)
		// Don't return to pool
		_ = buf
	}
}

func BenchmarkMessage_Pool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			msg := proc.GetMessage()
			msg.Type = proc.MessageTypeRequest
			msg.ID = "test"
			proc.PutMessage(msg)
		}
	})
}

func BenchmarkMessage_NoPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		msg := &proc.Message{
			Type: proc.MessageTypeRequest,
			ID:   "test",
		}
		_ = msg
	}
}
