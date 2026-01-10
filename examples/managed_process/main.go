package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// WorkerProcess is an example of a managed process that handles tasks
// Users only need to implement OnMessage - lifecycle is handled automatically
type WorkerProcess struct {
	*proc.BaseProcess
	taskCount int
}

// NewWorkerProcess creates a new worker process
func NewWorkerProcess(id string) *WorkerProcess {
	return &WorkerProcess{
		BaseProcess: proc.NewBaseProcess(id),
	}
}

// OnMessage handles incoming messages
func (w *WorkerProcess) OnMessage(msg *proc.Message) error {
	switch msg.Type {
	case proc.MessageTypeRequest:
		return w.handleRequest(msg)
	case proc.MessageTypeEvent:
		return w.handleEvent(msg)
	default:
		common.Debug("Worker %s ignoring message type: %s", w.ID(), msg.Type)
	}
	return nil
}

func (w *WorkerProcess) handleRequest(msg *proc.Message) error {
	var payload map[string]interface{}
	if err := msg.UnmarshalPayload(&payload); err != nil {
		return w.SendError(msg.ID, err)
	}

	action, ok := payload["action"].(string)
	if !ok {
		return w.SendError(msg.ID, fmt.Errorf("missing action field"))
	}

	common.Info("Worker %s processing action: %s", w.ID(), action)
	
	switch action {
	case "process_task":
		w.taskCount++
		time.Sleep(100 * time.Millisecond) // Simulate work
		return w.SendResponse(msg.ID, map[string]interface{}{
			"status":     "completed",
			"task_count": w.taskCount,
			"worker_id":  w.ID(),
		})
	case "get_status":
		return w.SendResponse(msg.ID, map[string]interface{}{
			"status":     "running",
			"task_count": w.taskCount,
			"worker_id":  w.ID(),
		})
	default:
		return w.SendError(msg.ID, fmt.Errorf("unknown action: %s", action))
	}
}

func (w *WorkerProcess) handleEvent(msg *proc.Message) error {
	var payload map[string]interface{}
	if err := msg.UnmarshalPayload(&payload); err != nil {
		return nil // Ignore malformed events
	}

	event, ok := payload["event"].(string)
	if !ok {
		return nil
	}

	common.Debug("Worker %s received event: %s", w.ID(), event)
	return nil
}

func main() {
	// Set up logger
	common.SetLevel(common.InfoLevel)
	
	common.Info("Starting managed process example...")
	
	// Create broker
	broker := proc.NewBroker()
	broker.SetLogger(common.NewLogger(os.Stdout, "", common.InfoLevel))
	defer broker.Shutdown()

	// Create and register workers
	numWorkers := 3
	for i := 0; i < numWorkers; i++ {
		worker := NewWorkerProcess(fmt.Sprintf("worker-%d", i))
		if err := broker.RegisterManagedProcess(worker); err != nil {
			log.Fatalf("Failed to register worker: %v", err)
		}
		common.Info("Worker registered: worker-%d", i)
	}

	common.Info("All workers registered. Sending tasks...")

	// Send tasks to workers
	for i := 0; i < 10; i++ {
		workerID := fmt.Sprintf("worker-%d", i%numWorkers)
		
		msg, _ := proc.NewRequestMessage(fmt.Sprintf("task-%d", i), map[string]interface{}{
			"action": "process_task",
			"data":   fmt.Sprintf("task %d", i),
		})
		
		if err := broker.DispatchMessage(workerID, msg); err != nil {
			common.Error("Failed to dispatch to %s: %v", workerID, err)
		}
	}

	// Wait a bit for tasks to complete
	time.Sleep(1 * time.Second)

	// Get status from all workers
	common.Info("Requesting status from workers...")
	for i := 0; i < numWorkers; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		
		msg, _ := proc.NewRequestMessage(fmt.Sprintf("status-%d", i), map[string]interface{}{
			"action": "get_status",
		})
		
		if err := broker.DispatchMessage(workerID, msg); err != nil {
			common.Error("Failed to get status from %s: %v", workerID, err)
		}
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Wait for signal or timeout
	select {
	case <-sigChan:
		common.Info("Received shutdown signal")
	case <-time.After(3 * time.Second):
		common.Info("Example completed")
	}

	common.Info("Shutting down...")
}
