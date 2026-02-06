package transport

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type SpinLock struct {
	flag atomic.Int32
}

func (sl *SpinLock) Lock() {
	for !sl.flag.CompareAndSwap(0, 1) {
	}
}

func (sl *SpinLock) Unlock() {
	sl.flag.Store(0)
}

func (sl *SpinLock) TryLock() bool {
	return sl.flag.CompareAndSwap(0, 1)
}

type SeqLock struct {
	sequence atomic.Uint64
}

func (sl *SeqLock) ReadLock() uint64 {
	for {
		seq := sl.sequence.Load()
		if seq&1 == 0 {
			return seq
		}
	}
}

func (sl *SeqLock) ReadUnlock(seq uint64) bool {
	return sl.sequence.Load() == seq
}

func (sl *SeqLock) WriteLock() {
	for {
		seq := sl.sequence.Load()
		if seq&1 == 0 {
			if sl.sequence.CompareAndSwap(seq, seq+1) {
				return
			}
		}
	}
}

func (sl *SeqLock) WriteUnlock() {
	sl.sequence.Add(2)
}

type ShardedMutex struct {
	shards []sync.RWMutex
	mask   uintptr
}

func NewShardedMutex(numShards int) *ShardedMutex {
	shards := 1
	for shards < numShards {
		shards <<= 1
	}

	return &ShardedMutex{
		shards: make([]sync.RWMutex, shards),
		mask:   uintptr(shards - 1),
	}
}

func (sm *ShardedMutex) Lock(key uintptr) {
	index := key & sm.mask
	sm.shards[index].Lock()
}

func (sm *ShardedMutex) Unlock(key uintptr) {
	index := key & sm.mask
	sm.shards[index].Unlock()
}

func (sm *ShardedMutex) RLock(key uintptr) {
	index := key & sm.mask
	sm.shards[index].RLock()
}

func (sm *ShardedMutex) RUnlock(key uintptr) {
	index := key & sm.mask
	sm.shards[index].RUnlock()
}

type Semaphore struct {
	count   atomic.Int32
	waiters atomic.Int32
	cond    *sync.Cond
	mu      sync.Mutex
}

func NewSemaphore(initial int32) *Semaphore {
	s := &Semaphore{
		cond: sync.NewCond(&sync.Mutex{}),
	}
	s.count.Store(initial)
	return s
}

func (s *Semaphore) Acquire() {
	if s.count.Add(-1) >= 0 {
		return
	}

	s.waiters.Add(1)
	s.mu.Lock()
	for s.count.Load() < 0 {
		s.cond.Wait()
	}
	s.mu.Unlock()
	s.waiters.Add(-1)
}

func (s *Semaphore) TryAcquire() bool {
	for {
		count := s.count.Load()
		if count <= 0 {
			return false
		}
		if s.count.CompareAndSwap(count, count-1) {
			return true
		}
	}
}

func (s *Semaphore) Release() {
	if s.count.Add(1) > 0 {
		if s.waiters.Load() > 0 {
			s.cond.Signal()
		}
	}
}

func (s *Semaphore) Count() int32 {
	return s.count.Load()
}

type ReadWriteCounter struct {
	readers   atomic.Int32
	writers   atomic.Int32
	readWait  atomic.Int32
	writeWait atomic.Int32
	readCond  *sync.Cond
	writeCond *sync.Cond
	mu        sync.Mutex
}

func NewReadWriteCounter() *ReadWriteCounter {
	return &ReadWriteCounter{
		readCond:  sync.NewCond(&sync.Mutex{}),
		writeCond: sync.NewCond(&sync.Mutex{}),
	}
}

func (rwc *ReadWriteCounter) RLock() {
	rwc.writeWait.Add(1)
	for {
		if rwc.writers.Load() == 0 {
			break
		}
		rwc.mu.Lock()
		rwc.writeCond.Wait()
		rwc.mu.Unlock()
	}
	rwc.writeWait.Add(-1)

	rwc.readers.Add(1)
}

func (rwc *ReadWriteCounter) RUnlock() {
	rwc.readers.Add(-1)
	if rwc.readers.Load() == 0 && rwc.writeWait.Load() > 0 {
		rwc.mu.Lock()
		rwc.writeCond.Signal()
		rwc.mu.Unlock()
	}
}

func (rwc *ReadWriteCounter) WLock() {
	rwc.writers.Add(1)

	for {
		if rwc.readers.Load() == 0 && rwc.writers.Load() == 1 {
			break
		}
		if rwc.readers.Load() > 0 {
			rwc.readWait.Add(1)
			rwc.mu.Lock()
			rwc.readCond.Wait()
			rwc.mu.Unlock()
			rwc.readWait.Add(-1)
		} else {
			rwc.mu.Lock()
			rwc.writeCond.Wait()
			rwc.mu.Unlock()
		}
	}
}

func (rwc *ReadWriteCounter) WUnlock() {
	rwc.writers.Add(-1)

	if rwc.readWait.Load() > 0 {
		rwc.mu.Lock()
		for i := int32(0); i < rwc.readWait.Load(); i++ {
			rwc.readCond.Signal()
		}
		rwc.mu.Unlock()
	} else if rwc.writeWait.Load() > 0 {
		rwc.mu.Lock()
		rwc.writeCond.Signal()
		rwc.mu.Unlock()
	}
}

func (rwc *ReadWriteCounter) ReaderCount() int32 {
	return rwc.readers.Load()
}

func (rwc *ReadWriteCounter) WriterCount() int32 {
	return rwc.writers.Load()
}

type AtomicFlag struct {
	flag atomic.Int32
}

func NewAtomicFlag(initial bool) *AtomicFlag {
	af := &AtomicFlag{}
	if initial {
		af.flag.Store(1)
	}
	return af
}

func (af *AtomicFlag) Set() {
	af.flag.Store(1)
}

func (af *AtomicFlag) Clear() {
	af.flag.Store(0)
}

func (af *AtomicFlag) Get() bool {
	return af.flag.Load() != 0
}

func (af *AtomicFlag) Swap(new bool) bool {
	newVal := int32(0)
	if new {
		newVal = 1
	}
	old := af.flag.Swap(newVal)
	return old != 0
}

func (af *AtomicFlag) CompareAndSwap(old, new bool) bool {
	oldVal := int32(0)
	if old {
		oldVal = 1
	}
	newVal := int32(0)
	if new {
		newVal = 1
	}
	return af.flag.CompareAndSwap(oldVal, newVal)
}

type UnsafePointer[T any] struct {
	ptr atomic.Pointer[T]
}

func NewUnsafePointer[T any](initial *T) *UnsafePointer[T] {
	up := &UnsafePointer[T]{}
	up.ptr.Store(initial)
	return up
}

func (up *UnsafePointer[T]) Load() *T {
	return up.ptr.Load()
}

func (up *UnsafePointer[T]) Store(val *T) {
	up.ptr.Store(val)
}

func (up *UnsafePointer[T]) CompareAndSwap(old, new *T) bool {
	return up.ptr.CompareAndSwap(old, new)
}

func (up *UnsafePointer[T]) Swap(new *T) *T {
	return up.ptr.Swap(new)
}

type MemoryBarrier struct{}

func (mb *MemoryBarrier) Load() {
	runtime_KeepAlive := func(x any) {}
	_ = runtime_KeepAlive
}

func (mb *MemoryBarrier) Store() {
	_ = unsafe.Pointer(nil)
}

func (mb *MemoryBarrier) Full() {
	mb.Load()
	mb.Store()
}
