package local

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"gorm.io/gorm"
)

// IntelligentBatcher provides intelligent batch operations
type IntelligentBatcher struct {
	db            *gorm.DB
	workers       []*BatchWorker
	workerPool    chan *BatchTask
	batchSize     int
	maxBatchSize  int
	minBatchSize  int
	flushInterval time.Duration
	lastFlushTime time.Time
	pendingTasks  []*BatchTask
	pendingMu     sync.Mutex
	metrics       *BatchMetrics
	enabled       atomic.Bool
}

// BatchWorker represents a parallel batch worker
type BatchWorker struct {
	id         int
	taskChan   chan *BatchTask
	resultChan chan *BatchResult
	metrics    *BatchMetrics
}

// BatchTask represents a batch operation task
type BatchTask struct {
	ID        string
	Operation string
	Data      interface{}
	Handler   func(interface{}) error
	Context   context.Context
}

// BatchResult represents a result of a batch operation
type BatchResult struct {
	TaskID   string
	Success  bool
	Error    error
	Duration time.Duration
}

// BatchMetrics tracks batch operation metrics
type BatchMetrics struct {
	TotalBatches     int64
	TotalTasks       int64
	MergedBatches    int64
	AvgBatchSize     float64
	AvgFlushInterval int64
	SuccessfulTasks  int64
	FailedTasks      int64
	AvgTaskLatency   int64
}

// NewIntelligentBatcher creates a new intelligent batcher
func NewIntelligentBatcher(db *gorm.DB, workerCount, minBatchSize, maxBatchSize int, flushInterval time.Duration) *IntelligentBatcher {
	ib := &IntelligentBatcher{
		db:            db,
		workers:       make([]*BatchWorker, 0, workerCount),
		workerPool:    make(chan *BatchTask, workerCount*10),
		batchSize:     minBatchSize,
		maxBatchSize:  maxBatchSize,
		minBatchSize:  minBatchSize,
		flushInterval: flushInterval,
		lastFlushTime: time.Now(),
		pendingTasks:  make([]*BatchTask, 0, maxBatchSize),
		metrics:       &BatchMetrics{},
	}

	for i := 0; i < workerCount; i++ {
		worker := &BatchWorker{
			id:         i,
			taskChan:   make(chan *BatchTask, 100),
			resultChan: make(chan *BatchResult, 100),
			metrics:    ib.metrics,
		}
		ib.workers = append(ib.workers, worker)
		go worker.Run()
	}

	ib.enabled.Store(true)
	go ib.mergingLoop()
	go ib.flushLoop()

	return ib
}

// AddTask adds a task to the batch
func (ib *IntelligentBatcher) AddTask(task *BatchTask) error {
	if !ib.enabled.Load() {
		return gorm.ErrInvalidDB
	}

	ib.pendingMu.Lock()
	ib.pendingTasks = append(ib.pendingTasks, task)
	atomic.AddInt64(&ib.metrics.TotalTasks, 1)

	// Check if batch is ready to flush
	shouldFlush := len(ib.pendingTasks) >= ib.batchSize || time.Since(ib.lastFlushTime) >= ib.flushInterval
	ib.pendingMu.Unlock()

	if shouldFlush {
		go ib.flushBatch()
	}

	return nil
}

// Flush flushes all pending tasks
func (ib *IntelligentBatcher) Flush() error {
	ib.pendingMu.Lock()
	tasks := ib.pendingTasks
	ib.pendingTasks = make([]*BatchTask, 0, ib.maxBatchSize)
	ib.pendingMu.Unlock()

	if len(tasks) == 0 {
		return nil
	}

	atomic.AddInt64(&ib.metrics.TotalBatches, 1)
	ib.updateBatchSizeMetrics(len(tasks))

	tx := ib.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, task := range tasks {
		if task.Handler != nil {
			if err := task.Handler(task.Data); err != nil {
				tx.Rollback()
				atomic.AddInt64(&ib.metrics.FailedTasks, 1)
				return err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		atomic.AddInt64(&ib.metrics.FailedTasks, 1)
		return err
	}

	atomic.AddInt64(&ib.metrics.SuccessfulTasks, int64(len(tasks)))
	ib.lastFlushTime = time.Now()

	return nil
}

// flushBatch flushes pending tasks
func (ib *IntelligentBatcher) flushBatch() {
	ib.Flush()
}

// mergeBatches merges small batches for efficiency
func (ib *IntelligentBatcher) mergeBatches(currentSize int, recentSizes []int) int {
	if len(recentSizes) < 3 {
		return currentSize
	}

	// Calculate average of recent batch sizes
	avgSize := 0
	for _, size := range recentSizes {
		avgSize += size
	}
	avgSize /= len(recentSizes)

	// If current batch is much smaller than average, increase size
	if currentSize < avgSize/2 {
		atomic.AddInt64(&ib.metrics.MergedBatches, 1)
		newSize := int(float64(avgSize) * 1.2)
		if newSize > ib.maxBatchSize {
			return ib.maxBatchSize
		}
		return newSize
	}

	return currentSize
}

// mergingLoop handles batch merging logic
func (ib *IntelligentBatcher) mergingLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	recentSizes := make([]int, 0, 10)

	for range ticker.C {
		if !ib.enabled.Load() {
			break
		}

		ib.pendingMu.Lock()
		currentSize := len(ib.pendingTasks)
		ib.pendingMu.Unlock()

		newSize := ib.mergeBatches(currentSize, recentSizes)
		ib.batchSize = newSize

		if currentSize > 0 {
			recentSizes = append(recentSizes, currentSize)
			if len(recentSizes) > 10 {
				recentSizes = recentSizes[1:]
			}
		}
	}
}

// flushLoop periodically flushes pending tasks
func (ib *IntelligentBatcher) flushLoop() {
	ticker := time.NewTicker(ib.flushInterval)
	defer ticker.Stop()

	for range ticker.C {
		if !ib.enabled.Load() {
			break
		}

		ib.flushBatch()
	}
}

// updateBatchSizeMetrics updates batch size metrics
func (ib *IntelligentBatcher) updateBatchSizeMetrics(size int) {
	totalTasks := atomic.LoadInt64(&ib.metrics.TotalTasks)
	totalBatches := atomic.LoadInt64(&ib.metrics.TotalBatches)

	if totalBatches > 0 {
		ib.metrics.AvgBatchSize = float64(totalTasks) / float64(totalBatches)
	}
}

// GetMetrics returns batch operation metrics
func (ib *IntelligentBatcher) GetMetrics() BatchMetrics {
	return *ib.metrics
}

// Enable enables the batcher
func (ib *IntelligentBatcher) Enable() {
	ib.enabled.Store(true)
}

// Disable disables the batcher
func (ib *IntelligentBatcher) Disable() {
	ib.enabled.Store(false)
}

// Close closes the batcher and flushes remaining tasks
func (ib *IntelligentBatcher) Close() error {
	ib.Disable()
	return ib.Flush()
}

// Run runs a batch worker
func (bw *BatchWorker) Run() {
	for task := range bw.taskChan {
		startTime := time.Now()

		var result *BatchResult
		if task.Handler != nil {
			err := task.Handler(task.Data)
			result = &BatchResult{
				TaskID:   task.ID,
				Success:  err == nil,
				Error:    err,
				Duration: time.Since(startTime),
			}
		}

		bw.resultChan <- result
	}
}
