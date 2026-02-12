package transport

import (
	"testing"
)

func BenchmarkSpinLock(b *testing.B) {
	sl := &SpinLock{}
	counter := 0

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sl.Lock()
		counter++
		sl.Unlock()
	}
}

func BenchmarkSpinLockParallel(b *testing.B) {
	sl := &SpinLock{}
	var counter int64

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		localCounter := int64(0)
		for pb.Next() {
			sl.Lock()
			localCounter++
			sl.Unlock()
		}
		for {
			if sl.TryLock() {
				counter += localCounter
				sl.Unlock()
				break
			}
		}
	})
}

func BenchmarkSeqLockRead(b *testing.B) {
	sl := &SeqLock{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		seq := sl.ReadLock()
		_ = sl.ReadUnlock(seq)
	}
}

func BenchmarkSeqLockWrite(b *testing.B) {
	sl := &SeqLock{}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sl.WriteLock()
		sl.WriteUnlock()
	}
}

func BenchmarkShardedMutex(b *testing.B) {
	sm := NewShardedMutex(16)
	counter := 0

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		key := uintptr(i)
		sm.Lock(key)
		counter++
		sm.Unlock(key)
	}
}

func BenchmarkShardedMutexParallel(b *testing.B) {
	sm := NewShardedMutex(16)
	var counter int64

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		key := uintptr(0)
		localCounter := int64(0)
		for pb.Next() {
			sm.Lock(key)
			localCounter++
			sm.Unlock(key)
		}
		for {
			sm.Lock(key)
			if counter == 0 {
				counter = localCounter
				sm.Unlock(key)
				break
			}
			sm.Unlock(key)
		}
	})
}

func BenchmarkSemaphoreAcquireRelease(b *testing.B) {
	s := NewSemaphore(100)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.Acquire()
		s.Release()
	}
}

func BenchmarkSemaphoreTryAcquire(b *testing.B) {
	s := NewSemaphore(100)

	for i := 0; i < 100; i++ {
		s.Acquire()
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.TryAcquire()
		s.Release()
	}
}

func BenchmarkAtomicFlag(b *testing.B) {
	af := NewAtomicFlag(false)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		af.Set()
		af.Get()
		af.Clear()
		af.Swap(false)
	}
}

func BenchmarkAtomicFlagParallel(b *testing.B) {
	af := NewAtomicFlag(false)

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			af.Set()
			af.Get()
			af.Clear()
			af.Swap(false)
		}
	})
}

func BenchmarkReadWriteCounter(b *testing.B) {
	rwc := NewReadWriteCounter()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rwc.RLock()
		rwc.RUnlock()
	}
}

func BenchmarkReadWriteCounterExclusive(b *testing.B) {
	rwc := NewReadWriteCounter()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		rwc.WLock()
		rwc.WUnlock()
	}
}

func BenchmarkReadWriteCounterMixed(b *testing.B) {
	rwc := NewReadWriteCounter()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if i%4 == 0 {
			rwc.WLock()
			rwc.WUnlock()
		} else {
			rwc.RLock()
			rwc.RUnlock()
		}
	}
}
