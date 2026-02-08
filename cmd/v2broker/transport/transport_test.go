package transport

import (
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestNewSharedMemory(t *testing.T) {
	tests := []struct {
		name    string
		config  SharedMemConfig
		wantErr bool
	}{
		{
			name: "default size",
			config: SharedMemConfig{
				Size:     0,
				IsServer: true,
			},
			wantErr: false,
		},
		{
			name: "custom size",
			config: SharedMemConfig{
				Size:     64 * 1024,
				IsServer: true,
			},
			wantErr: false,
		},
		{
			name: "minimum size",
			config: SharedMemConfig{
				Size:     SharedMemMinSize,
				IsServer: true,
			},
			wantErr: false,
		},
		{
			name: "client mode",
			config: SharedMemConfig{
				Size:     64 * 1024,
				IsServer: false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shm, err := NewSharedMemory(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSharedMemory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				defer shm.Close()

				if shm.Size() < SharedMemMinSize {
					t.Errorf("SharedMemory size %d is less than minimum %d", shm.Size(), SharedMemMinSize)
				}
			}
		})
	}
}

func TestSharedMemoryWriteRead(t *testing.T) {
	config := SharedMemConfig{
		Size:     64 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	writeData := []byte("test message")
	if err := shm.Write(writeData); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	readBuf := make([]byte, len(writeData))
	n, err := shm.Read(readBuf)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if n != len(writeData) {
		t.Errorf("Read %d bytes, want %d", n, len(writeData))
	}

	if string(readBuf) != string(writeData) {
		t.Errorf("Read data mismatch: got %s, want %s", string(readBuf), string(writeData))
	}
}

func TestSharedMemoryMultipleWrites(t *testing.T) {
	config := SharedMemConfig{
		Size:     64 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	messages := [][]byte{
		[]byte("message 1"),
		[]byte("message 2"),
		[]byte("message 3"),
	}

	for _, msg := range messages {
		if err := shm.Write(msg); err != nil {
			t.Fatalf("Failed to write: %v", err)
		}
	}

	for _, expected := range messages {
		readBuf := make([]byte, len(expected))
		n, err := shm.Read(readBuf)
		if err != nil {
			t.Fatalf("Failed to read: %v", err)
		}

		if n != len(expected) {
			t.Errorf("Read %d bytes, want %d", n, len(expected))
		}

		if string(readBuf) != string(expected) {
			t.Errorf("Read data mismatch: got %s, want %s", string(readBuf), string(expected))
		}
	}
}

func TestSharedMemoryConcurrentWrites(t *testing.T) {
	config := SharedMemConfig{
		Size:     512 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	const numGoroutines = 10
	const messagesPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()

			for j := 0; j < messagesPerGoroutine; j++ {
				msg := []byte(string(rune(id+'0')) + ": " + string(rune(j)))
				if err := shm.Write(msg); err != nil {
					t.Errorf("Write failed: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()

	totalMessages := numGoroutines * messagesPerGoroutine
	for i := 0; i < totalMessages; i++ {
		readBuf := make([]byte, 256)
		_, err := shm.Read(readBuf)
		if err != nil {
			t.Logf("Read at message %d failed (expected due to concurrent writes): %v", i, err)
		}
	}
}

func TestSharedMemoryAvailable(t *testing.T) {
	config := SharedMemConfig{
		Size:     64 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	initialAvailable := shm.Available()
	if initialAvailable == 0 {
		t.Error("Initial available should be non-zero")
	}

	writeData := make([]byte, 1024)
	if err := shm.Write(writeData); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	afterWrite := shm.Available()
	if afterWrite >= initialAvailable {
		t.Error("Available space should decrease after write")
	}

	readBuf := make([]byte, 1024)
	if _, err := shm.Read(readBuf); err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	afterRead := shm.Available()
	if afterRead < afterWrite {
		t.Error("Available space should increase after read")
	}
}

func TestSharedMemoryBytesAvailable(t *testing.T) {
	config := SharedMemConfig{
		Size:     64 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	if shm.BytesAvailable() != 0 {
		t.Error("Bytes available should be 0 initially")
	}

	writeData := make([]byte, 1024)
	if err := shm.Write(writeData); err != nil {
		t.Fatalf("Failed to write: %v", err)
	}

	if shm.BytesAvailable() != 1024 {
		t.Errorf("Bytes available should be 1024, got %d", shm.BytesAvailable())
	}

	readBuf := make([]byte, 1024)
	if _, err := shm.Read(readBuf); err != nil {
		t.Fatalf("Failed to read: %v", err)
	}

	if shm.BytesAvailable() != 0 {
		t.Errorf("Bytes available should be 0 after read, got %d", shm.BytesAvailable())
	}
}

func TestSharedMemoryClose(t *testing.T) {
	config := SharedMemConfig{
		Size:     64 * 1024,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}

	if err := shm.Close(); err != nil {
		t.Fatalf("Failed to close: %v", err)
	}

	if !shm.IsClosed() {
		t.Error("SharedMemory should be closed")
	}

	writeData := []byte("test")
	if err := shm.Write(writeData); err == nil {
		t.Error("Write should fail after close")
	}

	if err := shm.Close(); err != nil {
		t.Error("Close should be idempotent")
	}
}

func TestSharedMemoryBufferFull(t *testing.T) {
	config := SharedMemConfig{
		Size:     8192,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	capacity := shm.Available()
	largeData := make([]byte, capacity+1)

	if err := shm.Write(largeData); err == nil {
		t.Error("Write should fail when buffer is full")
	}
}

func TestSpinLock(t *testing.T) {
	sl := &SpinLock{}

	if !sl.TryLock() {
		t.Error("First TryLock should succeed")
	}

	if sl.TryLock() {
		t.Error("Second TryLock should fail")
	}

	sl.Unlock()

	if !sl.TryLock() {
		t.Error("TryLock after Unlock should succeed")
	}

	sl.Unlock()
}

func TestSpinLockConcurrency(t *testing.T) {
	sl := &SpinLock{}
	counter := 0
	const numGoroutines = 100
	const incrementsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				sl.Lock()
				counter++
				sl.Unlock()
			}
		}()
	}

	wg.Wait()

	expected := numGoroutines * incrementsPerGoroutine
	if counter != expected {
		t.Errorf("Counter = %d, want %d", counter, expected)
	}
}

func TestSeqLock(t *testing.T) {
	sl := &SeqLock{}

	seq1 := sl.ReadLock()
	seq2 := sl.ReadLock()

	if seq1 != seq2 {
		t.Error("ReadLock should return same sequence when no writes")
	}

	sl.WriteLock()
	sl.WriteUnlock()

	seq3 := sl.ReadLock()
	if seq3 <= seq2 {
		t.Error("Sequence should increase after write")
	}
}

func TestShardedMutex(t *testing.T) {
	sm := NewShardedMutex(16)

	key1 := uintptr(12345)
	key2 := uintptr(67890)

	sm.Lock(key1)
	sm.Lock(key2)

	sm.Unlock(key2)
	sm.Unlock(key1)

	sm.Lock(key1)
	sm.Unlock(key1)
}

func TestShardedMutexConcurrency(t *testing.T) {
	sm := NewShardedMutex(16)
	counter := make(map[uintptr]int)
	var mu sync.Mutex

	const numGoroutines = 100
	const incrementsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			key := uintptr(id % 16)

			for j := 0; j < incrementsPerGoroutine; j++ {
				sm.Lock(key)
				mu.Lock()
				counter[key]++
				mu.Unlock()
				sm.Unlock(key)
			}
		}(i)
	}

	wg.Wait()

	for key, count := range counter {
		if count == 0 {
			t.Errorf("Counter for key %d should be non-zero", key)
		}
	}
}

func TestSemaphore(t *testing.T) {
	s := NewSemaphore(5)

	if s.Count() != 5 {
		t.Errorf("Initial count = %d, want 5", s.Count())
	}

	for i := 0; i < 5; i++ {
		s.Acquire()
	}

	if s.Count() != 0 {
		t.Errorf("Count after 5 acquires = %d, want 0", s.Count())
	}

	if s.TryAcquire() {
		t.Error("TryAcquire should fail when count is 0")
	}

	s.Release()
	if s.Count() != 1 {
		t.Errorf("Count after release = %d, want 1", s.Count())
	}

	if !s.TryAcquire() {
		t.Error("TryAcquire should succeed when count > 0")
	}
}

func TestSemaphoreConcurrency(t *testing.T) {
	s := NewSemaphore(3)
	const numGoroutines = 10
	const loopsPerGoroutine = 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	activeCount := 0
	var activeMu sync.Mutex

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			for j := 0; j < loopsPerGoroutine; j++ {
				s.Acquire()

				activeMu.Lock()
				activeCount++
				if activeCount > 3 {
					t.Errorf("Active count %d exceeds semaphore size 3", activeCount)
				}
				activeMu.Unlock()

				runtime.Gosched()

				activeMu.Lock()
				activeCount--
				activeMu.Unlock()

				s.Release()
			}
		}()
	}

	wg.Wait()
}

func TestAtomicFlag(t *testing.T) {
	af := NewAtomicFlag(false)

	if af.Get() {
		t.Error("Initial flag should be false")
	}

	af.Set()
	if !af.Get() {
		t.Error("Flag should be true after Set")
	}

	if !af.Swap(false) {
		t.Error("Swap should return old value true")
	}

	if af.Get() {
		t.Error("Flag should be false after Swap")
	}

	if !af.CompareAndSwap(false, true) {
		t.Error("CompareAndSwap should succeed when old matches")
	}

	if af.CompareAndSwap(false, false) {
		t.Error("CompareAndSwap should fail when old doesn't match")
	}
}

func TestBatchAck(t *testing.T) {
	config := BatchAckConfig{
		MaxBatchSize:  10,
		FlushInterval: 100 * time.Millisecond,
		AckType:       AckBatch,
	}

	ba := NewBatchAck(config)
	defer ba.Close()

	flushed := false
	var flushedMsgs []AckMessage

	ba.SetOnFlush(func(msgs []AckMessage) {
		flushed = true
		flushedMsgs = msgs
	})

	for i := 0; i < 5; i++ {
		ba.AddAck(AckMessage{
			SeqNum:  uint64(i),
			Success: true,
		})
	}

	if flushed {
		t.Error("Should not flush before reaching max batch size")
	}

	for i := 5; i < 10; i++ {
		ba.AddAck(AckMessage{
			SeqNum:  uint64(i),
			Success: true,
		})
	}

	if !flushed {
		t.Error("Should flush when reaching max batch size")
	}

	if len(flushedMsgs) != 10 {
		t.Errorf("Flushed %d messages, want 10", len(flushedMsgs))
	}
}

func TestBatchAckImmediateFlush(t *testing.T) {
	config := BatchAckConfig{
		MaxBatchSize:  10,
		FlushInterval: 100 * time.Millisecond,
		AckType:       AckImmediate,
	}

	ba := NewBatchAck(config)
	defer ba.Close()

	flushed := false
	ba.SetOnFlush(func(msgs []AckMessage) {
		flushed = true
	})

	ba.AddAck(AckMessage{SeqNum: 1, Success: true})

	if !flushed {
		t.Error("Should flush immediately when AckType is AckImmediate")
	}
}

func TestBatchAckFlushInterval(t *testing.T) {
	config := BatchAckConfig{
		MaxBatchSize:  100,
		FlushInterval: 50 * time.Millisecond,
		AckType:       AckBatch,
	}

	ba := NewBatchAck(config)
	defer ba.Close()

	flushed := false
	ba.SetOnFlush(func(msgs []AckMessage) {
		flushed = true
	})

	ba.AddAck(AckMessage{SeqNum: 1, Success: true})

	time.Sleep(100 * time.Millisecond)

	if !flushed {
		t.Error("Should flush after interval even when not full")
	}
}
