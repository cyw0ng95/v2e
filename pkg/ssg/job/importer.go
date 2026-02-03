// Package job provides SSG import job orchestration.
package job

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// DataType represents the type of job data
type DataType string

const (
	DataTypeSSG DataType = "ssg"
)

// JobState represents the current state of a job
type JobState string

const (
	StateQueued    JobState = "queued"
	StateRunning   JobState = "running"
	StatePaused    JobState = "paused"
	StateCompleted JobState = "completed"
	StateFailed    JobState = "failed"
	StateStopped   JobState = "stopped"
)

// String returns the string representation of the state
func (s JobState) String() string {
	return string(s)
}

// JobRun represents a single job run
type JobRun struct {
	ID          string            `json:"id"`
	DataType    DataType          `json:"data_type"`
	State       JobState          `json:"state"`
	StartedAt   time.Time         `json:"started_at"`
	CompletedAt *time.Time        `json:"completed_at,omitempty"`
	Error       string            `json:"error,omitempty"`
	Progress    JobProgress       `json:"progress"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// JobProgress tracks the progress of an import job
type JobProgress struct {
	TotalGuides    int    `json:"total_guides"`
	ProcessedGuides int    `json:"processed_guides"`
	FailedGuides   int    `json:"failed_guides"`
	CurrentFile    string `json:"current_file,omitempty"`
}

// RPCInvoker is an interface for making RPC calls to other services
type RPCInvoker interface {
	InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error)
}

// Importer orchestrates SSG import jobs
type Importer struct {
	rpcInvoker RPCInvoker
	logger     *common.Logger

	mu        sync.RWMutex
	activeRun *JobRun
	cancelCtx context.CancelFunc
}

// NewImporter creates a new SSG import job orchestrator
func NewImporter(rpcInvoker RPCInvoker, logger *common.Logger) *Importer {
	return &Importer{
		rpcInvoker: rpcInvoker,
		logger:     logger,
	}
}

// StartImport starts a new SSG import job
func (imp *Importer) StartImport(ctx context.Context, runID string) error {
	imp.mu.Lock()
	defer imp.mu.Unlock()

	// Check if a job is already running (or paused)
	// Allow starting new job if previous job is in terminal state (failed, completed, stopped)
	if imp.activeRun != nil {
		if imp.activeRun.State == StateRunning || imp.activeRun.State == StateQueued || imp.activeRun.State == StatePaused {
			return fmt.Errorf("import job already running: %s (state: %s)", imp.activeRun.ID, imp.activeRun.State)
		}
		// Previous job is in terminal state, clear it
		imp.activeRun = nil
		imp.cancelCtx = nil
	}

	// Create new run
	run := &JobRun{
		ID:        runID,
		DataType:  DataTypeSSG,
		State:     StateQueued,
		StartedAt: time.Now(),
		Progress:  JobProgress{},
	}

	// Transition to running
	run.State = StateRunning
	imp.activeRun = run

	// Create cancellable context
	jobCtx, cancel := context.WithCancel(ctx)
	imp.cancelCtx = cancel

	// Start job in background
	go imp.executeImport(jobCtx, runID)

	imp.logger.Info("SSG import job started: %s", runID)
	return nil
}

// StopImport stops the currently running import job
func (imp *Importer) StopImport(ctx context.Context) error {
	imp.mu.Lock()
	defer imp.mu.Unlock()

	if imp.activeRun == nil {
		return fmt.Errorf("no import job is currently running")
	}

	if imp.cancelCtx != nil {
		imp.cancelCtx()
	}

	imp.activeRun.State = StateStopped
	now := time.Now()
	imp.activeRun.CompletedAt = &now

	imp.logger.Info("SSG import job stopped: %s", imp.activeRun.ID)
	return nil
}

// PauseImport pauses the currently running import job
func (imp *Importer) PauseImport(ctx context.Context) error {
	imp.mu.Lock()
	defer imp.mu.Unlock()

	if imp.activeRun == nil {
		return fmt.Errorf("no import job is currently running")
	}

	if imp.activeRun.State != StateRunning {
		return fmt.Errorf("job is not running (current state: %s)", imp.activeRun.State)
	}

	imp.activeRun.State = StatePaused
	imp.logger.Info("SSG import job paused: %s", imp.activeRun.ID)
	return nil
}

// ResumeImport resumes a paused import job
func (imp *Importer) ResumeImport(ctx context.Context, runID string) error {
	imp.mu.Lock()
	defer imp.mu.Unlock()

	if imp.activeRun == nil {
		return fmt.Errorf("no import job to resume")
	}

	if imp.activeRun.State != StatePaused {
		return fmt.Errorf("job is not paused (current state: %s)", imp.activeRun.State)
	}

	imp.activeRun.State = StateRunning

	// Resume execution
	jobCtx, cancel := context.WithCancel(ctx)
	imp.cancelCtx = cancel

	go imp.executeImport(jobCtx, runID)

	imp.logger.Info("SSG import job resumed: %s", runID)
	return nil
}

// GetStatus returns the current status of the active job
func (imp *Importer) GetStatus(ctx context.Context) (*JobRun, error) {
	imp.mu.RLock()
	defer imp.mu.RUnlock()

	if imp.activeRun == nil {
		return nil, fmt.Errorf("no active import job")
	}

	// Return a copy to avoid external modification
	runCopy := *imp.activeRun
	return &runCopy, nil
}

// executeImport runs the SSG import workflow
func (imp *Importer) executeImport(ctx context.Context, runID string) {
	defer func() {
		if r := recover(); r != nil {
			imp.mu.Lock()
			if imp.activeRun != nil && imp.activeRun.ID == runID {
				imp.activeRun.State = StateFailed
				imp.activeRun.Error = fmt.Sprintf("panic: %v", r)
				now := time.Now()
				imp.activeRun.CompletedAt = &now
				imp.logger.Error("SSG import job panicked: %v", r)
			}
			imp.mu.Unlock()
		}
	}()

	imp.logger.Info("SSG import workflow started for run: %s", runID)

	// Step 1: Pull latest changes from SSG repository
	imp.logger.Info("[Step 1/4] Pulling SSG repository...")
	_, err := imp.rpcInvoker.InvokeRPC(ctx, "remote", "RPCSSGPullRepo", nil)
	if err != nil {
		imp.mu.Lock()
		if imp.activeRun.ID == runID {
			imp.activeRun.State = StateFailed
			imp.activeRun.Error = fmt.Sprintf("failed to pull repository: %v", err)
			now := time.Now()
			imp.activeRun.CompletedAt = &now
		}
		imp.mu.Unlock()
		imp.logger.Error("[Step 1/4] Failed to pull SSG repository: %v", err)
		return
	}
	imp.logger.Info("[Step 1/4] SSG repository pull completed successfully")

	// Step 2: List guide files
	imp.logger.Info("[Step 2/4] Listing guide files...")
	resultMsg, err := imp.rpcInvoker.InvokeRPC(ctx, "remote", "RPCSSGListGuideFiles", nil)
	if err != nil {
		imp.mu.Lock()
		if imp.activeRun.ID == runID {
			imp.activeRun.State = StateFailed
			imp.activeRun.Error = fmt.Sprintf("failed to list guide files: %v", err)
			now := time.Now()
			imp.activeRun.CompletedAt = &now
		}
		imp.mu.Unlock()
		imp.logger.Error("Failed to list guide files: %v", err)
		return
	}

	// Extract payload from subprocess.Message
	msg, ok := resultMsg.(*subprocess.Message)
	if !ok || msg == nil {
		imp.mu.Lock()
		if imp.activeRun.ID == runID {
			imp.activeRun.State = StateFailed
			imp.activeRun.Error = "invalid response format from RPCSSGListGuideFiles"
			now := time.Now()
			imp.activeRun.CompletedAt = &now
		}
		imp.mu.Unlock()
		imp.logger.Error("Invalid response format from RPCSSGListGuideFiles")
		return
	}

	// Unmarshal payload
	var result map[string]interface{}
	if msg.Payload != nil {
		if err := subprocess.UnmarshalPayload(msg, &result); err != nil {
			imp.mu.Lock()
			if imp.activeRun.ID == runID {
				imp.activeRun.State = StateFailed
				imp.activeRun.Error = fmt.Sprintf("failed to unmarshal response: %v", err)
				now := time.Now()
				imp.activeRun.CompletedAt = &now
			}
			imp.mu.Unlock()
			imp.logger.Error("Failed to unmarshal response: %v", err)
			return
		}
	}

	// Extract files from response
	files, ok := result["files"].([]interface{})
	if !ok {
		imp.mu.Lock()
		if imp.activeRun.ID == runID {
			imp.activeRun.State = StateFailed
			imp.activeRun.Error = "invalid response format from RPCSSGListGuideFiles"
			now := time.Now()
			imp.activeRun.CompletedAt = &now
		}
		imp.mu.Unlock()
		imp.logger.Error("Invalid response format from RPCSSGListGuideFiles")
		return
	}

	imp.logger.Info("[Step 2/4] Found %d guide files to import", len(files))

	// Update progress
	imp.mu.Lock()
	if imp.activeRun.ID == runID {
		imp.activeRun.Progress.TotalGuides = len(files)
	}
	imp.mu.Unlock()

	// Step 3: Import each guide file
	imp.logger.Info("[Step 3/4] Starting guide import for %d files...", len(files))
	for i, file := range files {
		// Check for cancellation/pause
		imp.mu.RLock()
		if imp.activeRun == nil || imp.activeRun.ID != runID {
			imp.mu.RUnlock()
			return
		}
		state := imp.activeRun.State
		imp.mu.RUnlock()

		if state == StateStopped {
			imp.logger.Info("Import job stopped, exiting")
			return
		}

		if state == StatePaused {
			imp.logger.Info("Import job paused, waiting...")
			for {
				time.Sleep(1 * time.Second)
				imp.mu.RLock()
				if imp.activeRun == nil || imp.activeRun.ID != runID {
					imp.mu.RUnlock()
					return
				}
				if imp.activeRun.State != StatePaused {
					state = imp.activeRun.State
					imp.mu.RUnlock()
					break
				}
				imp.mu.RUnlock()
			}
			if state == StateStopped {
				return
			}
		}

		filename, ok := file.(string)
		if !ok {
			continue
		}

		imp.logger.Debug("[%d/%d] Importing guide: %s", i+1, len(files), filename)

		// Update current file
		imp.mu.Lock()
		if imp.activeRun.ID == runID {
			imp.activeRun.Progress.CurrentFile = filename
		}
		imp.mu.Unlock()

		// Get file path from remote service
		pathResultMsg, err := imp.rpcInvoker.InvokeRPC(ctx, "remote", "RPCSSGGetFilePath", map[string]interface{}{
			"filename": filename,
		})
		if err != nil {
			imp.logger.Warn("Failed to get path for %s: %v", filename, err)
			imp.mu.Lock()
			if imp.activeRun.ID == runID {
				imp.activeRun.Progress.FailedGuides++
			}
			imp.mu.Unlock()
			continue
		}

		// Extract payload from subprocess.Message
		pathMsg, ok := pathResultMsg.(*subprocess.Message)
		if !ok || pathMsg == nil {
			imp.logger.Warn("Invalid path response for %s", filename)
			continue
		}

		// Unmarshal payload
		var pathResult map[string]interface{}
		if pathMsg.Payload != nil {
			if err := subprocess.UnmarshalPayload(pathMsg, &pathResult); err != nil {
				imp.logger.Warn("Failed to unmarshal path response for %s: %v", filename, err)
				continue
			}
		}

		path, ok := pathResult["path"].(string)
		if !ok {
			imp.logger.Warn("Invalid path response for %s", filename)
			continue
		}

		// Import the guide with extended timeout (5 minutes per guide)
		// Large guide files can take a long time to parse and save
		importCtx, cancelImport := context.WithTimeout(ctx, 5*time.Minute)
		respMsg, err := imp.rpcInvoker.InvokeRPC(importCtx, "local", "RPCSSGImportGuide", map[string]interface{}{
			"path": path,
		})
		cancelImport() // Clean up the context
		if err != nil {
			imp.logger.Warn("Failed to import %s: %v", filename, err)
			imp.mu.Lock()
			if imp.activeRun.ID == runID {
				imp.activeRun.Progress.FailedGuides++
			}
			imp.mu.Unlock()
			continue
		}

		// Check if the response is an error message
		msg, ok := respMsg.(*subprocess.Message)
		if !ok {
			imp.logger.Warn("Invalid response type for %s", filename)
			imp.mu.Lock()
			if imp.activeRun.ID == runID {
				imp.activeRun.Progress.FailedGuides++
			}
			imp.mu.Unlock()
			continue
		}
		if msg.Type == subprocess.MessageTypeError {
			imp.logger.Warn("Failed to import %s: %s", filename, msg.Error)
			imp.mu.Lock()
			if imp.activeRun.ID == runID {
				imp.activeRun.Progress.FailedGuides++
			}
			imp.mu.Unlock()
			continue
		}

		// Update processed count
		imp.mu.Lock()
		if imp.activeRun.ID == runID {
			imp.activeRun.Progress.ProcessedGuides++
		}
		imp.mu.Unlock()

		imp.logger.Info("Successfully imported %s", filename)
	}

	// Mark job as completed
	imp.mu.Lock()
	if imp.activeRun != nil && imp.activeRun.ID == runID {
		imp.activeRun.State = StateCompleted
		now := time.Now()
		imp.activeRun.CompletedAt = &now
		imp.activeRun.Progress.CurrentFile = ""
		imp.logger.Info("[Step 4/4] SSG import job completed: %s (processed: %d, failed: %d, total: %d)",
			runID, imp.activeRun.Progress.ProcessedGuides, imp.activeRun.Progress.FailedGuides, len(files))
	}
	imp.mu.Unlock()
}
