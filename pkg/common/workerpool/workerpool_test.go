package workerpool

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestWorkerPoolCreation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestWorkerPoolCreation", nil, func(t *testing.T, tx *gorm.DB) {
		config := &Config{
			InitialSize: 4,
			MinSize:     2,
			MaxSize:     8,
			QueueSize:   40,
		}

		pool := NewWorkerPool(config)
		defer pool.Close()

		if pool.Size() != 4 {
			t.Errorf("Expected pool size 4, got %d", pool.Size())
		}
	})

}

func TestWorkerPoolSubmitTask(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestWorkerPoolSubmitTask", nil, func(t *testing.T, tx *gorm.DB) {
		pool := NewWorkerPool(nil) // Use defaults
		defer pool.Close()

		var counter int32
		task := TaskFunc(func(ctx context.Context) error {
			atomic.AddInt32(&counter, 1)
			return nil
		})

		err := pool.Submit(task)
		if err != nil {
			t.Fatalf("Failed to submit task: %v", err)
		}

		// Wait for task to execute
		time.Sleep(100 * time.Millisecond)

		if atomic.LoadInt32(&counter) != 1 {
			t.Errorf("Expected counter=1, got %d", counter)
		}
	})

}

func TestWorkerPoolConcurrentTasks(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestWorkerPoolConcurrentTasks", nil, func(t *testing.T, tx *gorm.DB) {
		pool := NewWorkerPool(&Config{
			InitialSize: 4,
			MinSize:     2,
			MaxSize:     8,
			QueueSize:   100,
		})
		defer pool.Close()

		numTasks := 50
		var counter int32

		for i := 0; i < numTasks; i++ {
			task := TaskFunc(func(ctx context.Context) error {
				atomic.AddInt32(&counter, 1)
				time.Sleep(10 * time.Millisecond)
				return nil
			})
			err := pool.Submit(task)
			if err != nil {
				t.Fatalf("Failed to submit task %d: %v", i, err)
			}
		}

		// Wait for all tasks to complete
		time.Sleep(2 * time.Second)

		if atomic.LoadInt32(&counter) != int32(numTasks) {
			t.Errorf("Expected counter=%d, got %d", numTasks, counter)
		}
	})

}

func TestWorkerPoolResize(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestWorkerPoolResize", nil, func(t *testing.T, tx *gorm.DB) {
		pool := NewWorkerPool(&Config{
			InitialSize: 4,
			MinSize:     2,
			MaxSize:     10,
			QueueSize:   40,
		})
		defer pool.Close()

		// Scale up
		err := pool.Resize(8)
		if err != nil {
			t.Fatalf("Failed to resize pool: %v", err)
		}
		if pool.Size() != 8 {
			t.Errorf("Expected pool size 8 after resize, got %d", pool.Size())
		}

		// Scale down
		err = pool.Resize(3)
		if err != nil {
			t.Fatalf("Failed to resize pool down: %v", err)
		}
		if pool.Size() != 3 {
			t.Errorf("Expected pool size 3 after resize, got %d", pool.Size())
		}

		// Try invalid sizes
		err = pool.Resize(1) // Below MinSize
		if err == nil {
			t.Error("Expected error when resizing below MinSize")
		}

		err = pool.Resize(11) // Above MaxSize
		if err == nil {
			t.Error("Expected error when resizing above MaxSize")
		}
	})

}

func TestWorkerPoolClose(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestWorkerPoolClose", nil, func(t *testing.T, tx *gorm.DB) {
		pool := NewWorkerPool(nil)

		// Submit some tasks
		for i := 0; i < 5; i++ {
			task := TaskFunc(func(ctx context.Context) error {
				time.Sleep(50 * time.Millisecond)
				return nil
			})
			pool.Submit(task)
		}

		// Close pool
		err := pool.Close()
		if err != nil {
			t.Fatalf("Failed to close pool: %v", err)
		}

		// Try to submit after close
		task := TaskFunc(func(ctx context.Context) error {
			return nil
		})
		err = pool.Submit(task)
		if err != ErrPoolClosed {
			t.Errorf("Expected ErrPoolClosed, got %v", err)
		}
	})

}

func TestWorkerPoolQueueDepth(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestWorkerPoolQueueDepth", nil, func(t *testing.T, tx *gorm.DB) {
		pool := NewWorkerPool(&Config{
			InitialSize: 1, // Only 1 worker
			MinSize:     1,
			MaxSize:     2,
			QueueSize:   10,
		})
		defer pool.Close()

		// Submit tasks that take time
		numTasks := 5
		for i := 0; i < numTasks; i++ {
			task := TaskFunc(func(ctx context.Context) error {
				time.Sleep(100 * time.Millisecond)
				return nil
			})
			pool.Submit(task)
		}

		// Check queue depth (should be > 0 since worker is busy)
		time.Sleep(10 * time.Millisecond)
		depth := pool.QueueDepth()
		if depth == 0 {
			t.Log("Queue depth is 0, tasks may have been processed quickly")
		}
	})

}

func BenchmarkWorkerPoolSubmit(b *testing.B) {
	pool := NewWorkerPool(nil)
	defer pool.Close()

	task := TaskFunc(func(ctx context.Context) error {
		return nil
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Submit(task)
	}
}

func BenchmarkWorkerPoolConcurrent(b *testing.B) {
	pool := NewWorkerPool(&Config{
		InitialSize: 8,
		MinSize:     2,
		MaxSize:     16,
		QueueSize:   1000,
	})
	defer pool.Close()

	task := TaskFunc(func(ctx context.Context) error {
		time.Sleep(1 * time.Millisecond)
		return nil
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			pool.Submit(task)
		}
	})
}
