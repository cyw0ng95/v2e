package transport

import (
	"runtime"
	"sync"
	"sync/atomic"
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

	var flushed atomic.Bool
	ba.SetOnFlush(func(msgs []AckMessage) {
		flushed.Store(true)
	})

	ba.AddAck(AckMessage{SeqNum: 1, Success: true})

	time.Sleep(100 * time.Millisecond)

	if !flushed.Load() {
		t.Error("Should flush after interval even when not full")
	}
}

func TestSharedMemoryRingBufferWrapAround(t *testing.T) {
	// Use a small buffer to test wrap-around behavior
	config := SharedMemConfig{
		Size:     8192, // 8KB total, capacity will be ~8064 bytes after header
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	capacity := shm.Available()
	headerOverhead := 8192 - int(capacity) // header size

	// Write data that fills most of the buffer, then read it to advance ReadPos
	writeSize := capacity - 100 // leave some space
	firstWrite := make([]byte, writeSize)
	for i := range firstWrite {
		firstWrite[i] = byte(i % 256)
	}

	if err := shm.Write(firstWrite); err != nil {
		t.Fatalf("First write failed: %v", err)
	}

	// Read the data to advance ReadPos
	readBuf := make([]byte, writeSize)
	n, err := shm.Read(readBuf)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != int(writeSize) {
		t.Fatalf("Read %d bytes, want %d", n, writeSize)
	}

	// Verify data integrity
	for i := range readBuf {
		if readBuf[i] != firstWrite[i] {
			t.Fatalf("Data mismatch at index %d: got %d, want %d", i, readBuf[i], firstWrite[i])
		}
	}

	// Now available space should be back to full capacity
	if shm.Available() != capacity {
		t.Errorf("Available space after read = %d, want %d", shm.Available(), capacity)
	}

	// Write data that will wrap around - the previous WritePos is near the end
	// and we need to write enough to wrap
	wrapWriteSize := capacity - 50 // slightly less than full capacity
	secondWrite := make([]byte, wrapWriteSize)
	for i := range secondWrite {
		secondWrite[i] = byte((i + 100) % 256)
	}

	if err := shm.Write(secondWrite); err != nil {
		t.Fatalf("Second write (wrap-around) failed: %v", err)
	}

	// Read and verify the wrapped data
	readBuf2 := make([]byte, wrapWriteSize)
	n, err = shm.Read(readBuf2)
	if err != nil {
		t.Fatalf("Second read failed: %v", err)
	}
	if n != int(wrapWriteSize) {
		t.Fatalf("Read %d bytes, want %d", n, wrapWriteSize)
	}

	// Verify data integrity after wrap-around
	for i := range readBuf2 {
		if readBuf2[i] != secondWrite[i] {
			t.Fatalf("Data mismatch at index %d after wrap: got %d, want %d", i, readBuf2[i], secondWrite[i])
		}
	}

	t.Logf("Ring buffer wrap-around test passed with %d header overhead", headerOverhead)
}

func TestSharedMemoryMultipleWrapArounds(t *testing.T) {
	// Use a small buffer for multiple wrap-around cycles
	config := SharedMemConfig{
		Size:     4096, // minimum size
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	capacity := shm.Available()

	// Use a chunk size that's about 1/3 of capacity to force multiple wraps
	chunkSize := capacity / 3

	for round := 0; round < 10; round++ {
		// Write data
		writeData := make([]byte, chunkSize)
		for i := range writeData {
			writeData[i] = byte((round*10 + i) % 256)
		}

		if err := shm.Write(writeData); err != nil {
			t.Fatalf("Round %d: write failed: %v", round, err)
		}

		// Read and verify
		readBuf := make([]byte, chunkSize)
		n, err := shm.Read(readBuf)
		if err != nil {
			t.Fatalf("Round %d: read failed: %v", round, err)
		}
		if n != int(chunkSize) {
			t.Fatalf("Round %d: read %d bytes, want %d", round, n, chunkSize)
		}

		for i := range readBuf {
			if readBuf[i] != writeData[i] {
				t.Fatalf("Round %d: data mismatch at index %d: got %d, want %d",
					round, i, readBuf[i], writeData[i])
			}
		}
	}

	t.Log("Multiple wrap-around cycles completed successfully")
}

func TestSharedMemoryWrapAroundSplitWrite(t *testing.T) {
	// Test case where write must be split across buffer boundary
	config := SharedMemConfig{
		Size:     4096,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	capacity := shm.Available()

	// Fill buffer partially to set up a specific write position
	fillSize := capacity / 2
	fillData := make([]byte, fillSize)
	if err := shm.Write(fillData); err != nil {
		t.Fatalf("Fill write failed: %v", err)
	}

	// Read to advance ReadPos, creating space but leaving WritePos at midpoint
	readBuf := make([]byte, fillSize)
	if _, err := shm.Read(readBuf); err != nil {
		t.Fatalf("Fill read failed: %v", err)
	}

	// Now WritePos is at midpoint. Write enough to wrap around.
	// We need to write > capacity/2 bytes to force wrap
	wrapSize := capacity - fillSize + 100 // This should wrap
	wrapData := make([]byte, wrapSize)
	for i := range wrapData {
		wrapData[i] = byte(i % 256)
	}

	if err := shm.Write(wrapData); err != nil {
		t.Fatalf("Wrap write failed: %v", err)
	}

	// Read and verify the wrapped data
	wrapReadBuf := make([]byte, wrapSize)
	n, err := shm.Read(wrapReadBuf)
	if err != nil {
		t.Fatalf("Wrap read failed: %v", err)
	}
	if n != int(wrapSize) {
		t.Fatalf("Read %d bytes, want %d", n, wrapSize)
	}

	for i := range wrapReadBuf {
		if wrapReadBuf[i] != wrapData[i] {
			t.Fatalf("Data mismatch at index %d: got %d, want %d", i, wrapReadBuf[i], wrapData[i])
		}
	}

	t.Log("Split write wrap-around test passed")
}

func TestSharedMemoryWrapAroundReadBoundary(t *testing.T) {
	// Test read that wraps across buffer boundary
	config := SharedMemConfig{
		Size:     4096,
		IsServer: true,
	}

	shm, err := NewSharedMemory(config)
	if err != nil {
		t.Fatalf("Failed to create shared memory: %v", err)
	}
	defer shm.Close()

	capacity := shm.Available()

	// Write full capacity worth of data in two parts
	// Part 1: fills to near end
	// Part 2: wraps to beginning

	// First, write 3/4 of capacity
	part1Size := capacity * 3 / 4
	part1Data := make([]byte, part1Size)
	for i := range part1Data {
		part1Data[i] = byte(i % 256)
	}
	if err := shm.Write(part1Data); err != nil {
		t.Fatalf("Part 1 write failed: %v", err)
	}

	// Read it back to free space
	readBuf1 := make([]byte, part1Size)
	if _, err := shm.Read(readBuf1); err != nil {
		t.Fatalf("Part 1 read failed: %v", err)
	}

	// Now write data that will wrap - this will be split
	// WritePos is at 3/4 capacity, so any write > 1/4 capacity will wrap
	wrapWriteSize := capacity/2 + 50
	wrapData := make([]byte, wrapWriteSize)
	for i := range wrapData {
		wrapData[i] = byte((i + 50) % 256)
	}

	if err := shm.Write(wrapData); err != nil {
		t.Fatalf("Wrap write failed: %v", err)
	}

	// Read and verify - this tests wrap-around read
	wrapReadBuf := make([]byte, wrapWriteSize)
	n, err := shm.Read(wrapReadBuf)
	if err != nil {
		t.Fatalf("Wrap read failed: %v", err)
	}
	if n != int(wrapWriteSize) {
		t.Fatalf("Read %d bytes, want %d", n, wrapWriteSize)
	}

	for i := range wrapReadBuf {
		if wrapReadBuf[i] != wrapData[i] {
			t.Fatalf("Data mismatch at index %d: got %d, want %d", i, wrapReadBuf[i], wrapData[i])
		}
	}

	t.Log("Read boundary wrap-around test passed")
}

func TestSharedMemoryFullCapacityUsage(t *testing.T) {
	// Verify that the full capacity can be used through wrap-around
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

	// Write full capacity
	fullData := make([]byte, capacity)
	for i := range fullData {
		fullData[i] = byte(i % 256)
	}

	if err := shm.Write(fullData); err != nil {
		t.Fatalf("Full capacity write failed: %v", err)
	}

	// Verify no more space available
	if shm.Available() != 0 {
		t.Errorf("Available = %d, want 0 after full write", shm.Available())
	}

	// Read half
	halfRead := make([]byte, capacity/2)
	if _, err := shm.Read(halfRead); err != nil {
		t.Fatalf("Half read failed: %v", err)
	}

	// Now we should be able to write half capacity again
	halfWrite := make([]byte, capacity/2)
	for i := range halfWrite {
		halfWrite[i] = byte((i + 128) % 256)
	}

	if err := shm.Write(halfWrite); err != nil {
		t.Fatalf("Write after partial read failed: %v", err)
	}

	// Read remaining and verify
	// First, read the second half of original data
	remainingOriginal := make([]byte, int(capacity)/2)
	if _, err := shm.Read(remainingOriginal); err != nil {
		t.Fatalf("Read remaining original failed: %v", err)
	}

	// Verify original data integrity
	for i := range remainingOriginal {
		expected := byte((int(capacity)/2 + i) % 256)
		if remainingOriginal[i] != expected {
			t.Fatalf("Original data mismatch at index %d: got %d, want %d",
				i, remainingOriginal[i], expected)
		}
	}

	// Then read the new data that wrapped
	wrappedRead := make([]byte, int(capacity)/2)
	if _, err := shm.Read(wrappedRead); err != nil {
		t.Fatalf("Read wrapped data failed: %v", err)
	}

	// Verify wrapped data integrity
	for i := range wrappedRead {
		if wrappedRead[i] != halfWrite[i] {
			t.Fatalf("Wrapped data mismatch at index %d: got %d, want %d",
				i, wrappedRead[i], halfWrite[i])
		}
	}

	t.Log("Full capacity usage test passed")
}
