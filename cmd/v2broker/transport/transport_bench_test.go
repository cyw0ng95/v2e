package transport

import (
	"testing"
)

func BenchmarkSharedMemoryWrite(b *testing.B) {
	config := SharedMemConfig{
		Size:     1024 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		b.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	data := make([]byte, 256)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := shm.Write(data); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
	}
}

func BenchmarkSharedMemoryRead(b *testing.B) {
	config := SharedMemConfig{
		Size:     1024 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		b.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	data := make([]byte, 256)
	for i := 0; i < 1000; i++ {
		_ = shm.Write(data)
	}

	buf := make([]byte, 256)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := shm.Read(buf)
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

func BenchmarkSharedMemoryWriteRead(b *testing.B) {
	config := SharedMemConfig{
		Size:     1024 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		b.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	data := make([]byte, 256)
	buf := make([]byte, 256)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		if err := shm.Write(data); err != nil {
			b.Fatalf("Write failed: %v", err)
		}
		_, err := shm.Read(buf)
		if err != nil {
			b.Fatalf("Read failed: %v", err)
		}
	}
}

func BenchmarkSharedMemoryConcurrentWrites(b *testing.B) {
	config := SharedMemConfig{
		Size:     4 * 1024 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		b.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	data := make([]byte, 256)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := shm.Write(data); err != nil {
				b.Fatalf("Write failed: %v", err)
			}
		}
	})
}

func BenchmarkSharedMemoryConcurrentReads(b *testing.B) {
	config := SharedMemConfig{
		Size:     4 * 1024 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		b.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	data := make([]byte, 256)
	for i := 0; i < 10000; i++ {
		_ = shm.Write(data)
	}

	buf := make([]byte, 256)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := shm.Read(buf)
			if err != nil {
				b.Fatalf("Read failed: %v", err)
			}
		}
	})
}

func BenchmarkSharedMemoryConcurrentWriteRead(b *testing.B) {
	config := SharedMemConfig{
		Size:     8 * 1024 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		b.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	data := make([]byte, 256)
	buf := make([]byte, 256)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := shm.Write(data); err != nil {
				b.Fatalf("Write failed: %v", err)
			}
			_, err := shm.Read(buf)
			if err != nil {
				b.Fatalf("Read failed: %v", err)
			}
		}
	})
}

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

func BenchmarkBatchAckImmediate(b *testing.B) {
	config := BatchAckConfig{
		MaxBatchSize:  100,
		FlushInterval: 100,
		AckType:       AckImmediate,
	}

	ba := NewBatchAck(config)
	defer ba.Close()

	msg := AckMessage{SeqNum: 1, Success: true}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ba.AddAck(msg)
	}
}

func BenchmarkBatchAckBatch(b *testing.B) {
	config := BatchAckConfig{
		MaxBatchSize:  100,
		FlushInterval: 100,
		AckType:       AckBatch,
	}

	ba := NewBatchAck(config)
	defer ba.Close()

	msg := AckMessage{SeqNum: 1, Success: true}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ba.AddAck(msg)
		if ba.PendingCount() >= 100 {
			_ = ba.Flush()
		}
	}
}

func BenchmarkBatchAckBatchAdd(b *testing.B) {
	config := BatchAckConfig{
		MaxBatchSize:  1000,
		FlushInterval: 1000,
		AckType:       AckBatch,
	}

	ba := NewBatchAck(config)
	defer ba.Close()

	msgs := make([]AckMessage, 10)
	for i := range msgs {
		msgs[i] = AckMessage{SeqNum: uint64(i), Success: true}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		ba.AddAckBatch(msgs)
	}
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

func BenchmarkSharedMemoryVariousSizes(b *testing.B) {
	sizes := []int{64, 256, 1024, 4096, 16384}

	for _, size := range sizes {
		b.Run(string(rune('A'+size/1024)), func(b *testing.B) {
			config := SharedMemConfig{
				Size:     2 * 1024 * 1024,
				IsServer: true,
			}

			shm, err := NewSharedMemory(config)
			if err != nil {
				b.Fatalf("Failed to create shared memory: %v", err)
			}
			defer shm.Close()

			data := make([]byte, size)
			buf := make([]byte, size)

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_ = shm.Write(data)
				_, _ = shm.Read(buf)
			}
		})
	}
}
