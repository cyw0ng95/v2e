package workerpool

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrPoolClosed is returned when trying to submit to a closed pool
	ErrPoolClosed = errors.New("worker pool is closed")
	// ErrInvalidSize is returned when trying to resize to invalid size
	ErrInvalidSize = errors.New("invalid pool size")
)

// Config holds worker pool configuration
type Config struct {
	// InitialSize is the starting number of workers
	// If 0, defaults to runtime.NumCPU()
	InitialSize int
	// MinSize is the minimum number of workers
	MinSize int
	// MaxSize is the maximum number of workers
	MaxSize int
	// QueueSize is the size of the task queue
	// If 0, defaults to InitialSize * 10
	QueueSize int
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	numCPU := runtime.NumCPU()
	return &Config{
		InitialSize: numCPU,
		MinSize:     2,
		MaxSize:     numCPU * 4,
		QueueSize:   numCPU * 10,
	}
}

// WorkerPool manages a pool of worker goroutines
type WorkerPool struct {
	config    *Config
	size      int32 // Current pool size (atomic)
	taskQueue chan Task
	workers   []*worker
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	closed    int32 // Atomic flag
}

// NewWorkerPool creates a new worker pool with the given configuration
func NewWorkerPool(config *Config) *WorkerPool {
	if config == nil {
		config = DefaultConfig()
	}

	// Apply defaults
	if config.InitialSize == 0 {
		config.InitialSize = runtime.NumCPU()
	}
	if config.MinSize == 0 {
		config.MinSize = 2
	}
	if config.MaxSize == 0 {
		config.MaxSize = config.InitialSize * 4
	}
	if config.QueueSize == 0 {
		config.QueueSize = config.InitialSize * 10
	}

	// Validate config
	if config.InitialSize < config.MinSize {
		config.InitialSize = config.MinSize
	}
	if config.InitialSize > config.MaxSize {
		config.InitialSize = config.MaxSize
	}

	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		config:    config,
		size:      int32(config.InitialSize),
		taskQueue: make(chan Task, config.QueueSize),
		workers:   make([]*worker, 0, config.MaxSize),
		ctx:       ctx,
		cancel:    cancel,
	}

	// Start initial workers
	pool.startWorkers(config.InitialSize)

	// Start dispatcher
	pool.wg.Add(1)
	go pool.dispatcher()

	return pool
}

// startWorkers starts n new workers
func (p *WorkerPool) startWorkers(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < n; i++ {
		workerID := len(p.workers)
		w := newWorker(workerID, p, &p.wg)
		w.start(p.ctx)
		p.workers = append(p.workers, w)
	}
}

// stopWorkers stops n workers from the end
func (p *WorkerPool) stopWorkers(n int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if n > len(p.workers) {
		n = len(p.workers)
	}

	for i := 0; i < n; i++ {
		workerIdx := len(p.workers) - 1
		if workerIdx >= 0 {
			p.workers[workerIdx].stop()
			p.workers = p.workers[:workerIdx]
		}
	}
}

// dispatcher distributes tasks from the queue to workers
func (p *WorkerPool) dispatcher() {
	defer p.wg.Done()

	workerIdx := 0 // Round-robin index

	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.taskQueue:
			if task == nil {
				continue
			}

			// Get current workers
			p.mu.RLock()
			numWorkers := len(p.workers)
			p.mu.RUnlock()

			if numWorkers == 0 {
				// No workers available, skip task
				continue
			}

			// Try to submit to workers starting from round-robin position
			submitted := false
			for attempts := 0; attempts < numWorkers && !submitted; attempts++ {
				p.mu.RLock()
				if workerIdx >= len(p.workers) {
					workerIdx = 0
				}
				if workerIdx < len(p.workers) {
					worker := p.workers[workerIdx]
					p.mu.RUnlock()
					
					// Blocking submit with timeout
					select {
					case worker.tasks <- task:
						submitted = true
						workerIdx = (workerIdx + 1) % numWorkers
					case <-time.After(10 * time.Millisecond):
						// Worker is busy, try next one
						workerIdx = (workerIdx + 1) % numWorkers
					case <-p.ctx.Done():
						return
					}
				} else {
					p.mu.RUnlock()
					break
				}
			}
		}
	}
}

// Submit adds a task to the pool
// Returns an error if the pool is closed
func (p *WorkerPool) Submit(task Task) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	select {
	case <-p.ctx.Done():
		return ErrPoolClosed
	case p.taskQueue <- task:
		return nil
	}
}

// Resize changes the number of workers in the pool
// Returns an error if the new size is invalid or pool is closed
func (p *WorkerPool) Resize(newSize int) error {
	if atomic.LoadInt32(&p.closed) == 1 {
		return ErrPoolClosed
	}

	if newSize < p.config.MinSize || newSize > p.config.MaxSize {
		return fmt.Errorf("%w: size %d not in range [%d, %d]",
			ErrInvalidSize, newSize, p.config.MinSize, p.config.MaxSize)
	}

	currentSize := int(atomic.LoadInt32(&p.size))

	if newSize > currentSize {
		// Scale up: add workers
		diff := newSize - currentSize
		p.startWorkers(diff)
		atomic.StoreInt32(&p.size, int32(newSize))
	} else if newSize < currentSize {
		// Scale down: remove workers
		diff := currentSize - newSize
		p.stopWorkers(diff)
		atomic.StoreInt32(&p.size, int32(newSize))
	}

	return nil
}

// Size returns the current number of workers
func (p *WorkerPool) Size() int {
	return int(atomic.LoadInt32(&p.size))
}

// QueueDepth returns the current number of pending tasks
func (p *WorkerPool) QueueDepth() int {
	return len(p.taskQueue)
}

// Close shuts down the worker pool gracefully
// Waits for all workers to finish their current tasks
func (p *WorkerPool) Close() error {
	if !atomic.CompareAndSwapInt32(&p.closed, 0, 1) {
		return ErrPoolClosed
	}

	// Stop accepting new tasks
	close(p.taskQueue)

	// Signal all goroutines to stop
	p.cancel()

	// Wait for all workers to finish
	p.wg.Wait()

	return nil
}
