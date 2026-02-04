package workerpool

import (
	"context"
	"sync"
)

// worker represents a goroutine that processes tasks from the pool
type worker struct {
	id       int
	pool     *WorkerPool
	tasks    chan Task
	wg       *sync.WaitGroup
	stopChan chan struct{}
}

// newWorker creates a new worker
func newWorker(id int, pool *WorkerPool, wg *sync.WaitGroup) *worker {
	return &worker{
		id:       id,
		pool:     pool,
		tasks:    make(chan Task, 1), // Small buffer for task handoff
		wg:       wg,
		stopChan: make(chan struct{}),
	}
}

// start begins processing tasks in a goroutine
func (w *worker) start(ctx context.Context) {
	w.wg.Add(1)
	go w.run(ctx)
}

// run is the main worker loop
func (w *worker) run(ctx context.Context) {
	defer w.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case task := <-w.tasks:
			if task != nil {
				_ = task.Execute(ctx)
			}
		}
	}
}

// stop signals the worker to stop
func (w *worker) stop() {
	close(w.stopChan)
}

// submit sends a task to this worker
// Returns true if task was accepted, false if worker is stopping
func (w *worker) submit(task Task) bool {
	select {
	case <-w.stopChan:
		return false
	case w.tasks <- task:
		return true
	default:
		return false
	}
}
