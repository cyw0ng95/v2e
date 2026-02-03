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
	TotalGuides       int    `json:"total_guides"`
	ProcessedGuides   int    `json:"processed_guides"`
	FailedGuides      int    `json:"failed_guides"`
	TotalTables       int    `json:"total_tables"`
	ProcessedTables   int    `json:"processed_tables"`
	FailedTables      int    `json:"failed_tables"`
	TotalManifests    int    `json:"total_manifests"`
	ProcessedManifests int   `json:"processed_manifests"`
	FailedManifests   int    `json:"failed_manifests"`
	TotalDataStreams    int    `json:"total_data_streams"`
	ProcessedDataStreams int   `json:"processed_data_streams"`
	FailedDataStreams   int    `json:"failed_data_streams"`
	CurrentFile       string `json:"current_file,omitempty"`
	CurrentPhase      string `json:"current_phase,omitempty"` // "tables", "guides", "manifests", or "datastreams"
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

// executeImport runs the SSG import workflow with tick-tock pattern
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

	imp.logger.Info("SSG import workflow started for run: %s (tick-tock-tock-tock mode)", runID)

	// Step 1: Pull latest changes from SSG repository
	imp.logger.Info("[Step 1/7] Pulling SSG repository...")
	_, err := imp.rpcInvoker.InvokeRPC(ctx, "remote", "RPCSSGPullRepo", nil)
	if err != nil {
		imp.setFailed(runID, fmt.Sprintf("failed to pull repository: %v", err))
		imp.logger.Error("[Step 1/7] Failed to pull SSG repository: %v", err)
		return
	}
	imp.logger.Info("[Step 1/7] SSG repository pull completed successfully")

	// Step 2: List table files
	imp.logger.Info("[Step 2/7] Listing table files...")
	tableFiles, err := imp.listFiles(ctx, runID, "RPCSSGListTableFiles", "tables")
	if err != nil {
		return // Error already logged and state set
	}
	imp.logger.Info("[Step 2/7] Found %d table files to import", len(tableFiles))

	// Step 3: List guide files
	imp.logger.Info("[Step 3/7] Listing guide files...")
	guideFiles, err := imp.listFiles(ctx, runID, "RPCSSGListGuideFiles", "guides")
	if err != nil {
		return // Error already logged and state set
	}
	imp.logger.Info("[Step 3/7] Found %d guide files to import", len(guideFiles))

	// Step 4: List manifest files
	imp.logger.Info("[Step 4/7] Listing manifest files...")
	manifestFiles, err := imp.listFiles(ctx, runID, "RPCSSGListManifestFiles", "manifests")
	if err != nil {
		return // Error already logged and state set
	}
	imp.logger.Info("[Step 4/7] Found %d manifest files to import", len(manifestFiles))

	// Step 5: List data stream files
	imp.logger.Info("[Step 5/7] Listing data stream files...")
	dataStreamFiles, err := imp.listFiles(ctx, runID, "RPCSSGListDataStreamFiles", "data streams")
	if err != nil {
		return // Error already logged and state set
	}
	imp.logger.Info("[Step 5/7] Found %d data stream files to import", len(dataStreamFiles))

	// Update progress with totals
	imp.mu.Lock()
	if imp.activeRun != nil && imp.activeRun.ID == runID {
		imp.activeRun.Progress.TotalTables = len(tableFiles)
		imp.activeRun.Progress.TotalGuides = len(guideFiles)
		imp.activeRun.Progress.TotalManifests = len(manifestFiles)
		imp.activeRun.Progress.TotalDataStreams = len(dataStreamFiles)
	}
	imp.mu.Unlock()

	// Step 6: Import in tick-tock-tock-tock fashion (alternate between tables, guides, manifests, and data streams)
	imp.logger.Info("[Step 6/7] Starting tick-tock-tock-tock import: %d tables, %d guides, %d manifests, %d data streams", 
		len(tableFiles), len(guideFiles), len(manifestFiles), len(dataStreamFiles))
	maxLen := max(len(tableFiles), len(guideFiles), len(manifestFiles), len(dataStreamFiles))

	for i := 0; i < maxLen; i++ {
		// Check for cancellation/pause before each iteration
		if !imp.checkRunning(runID) {
			return
		}

		// Tick: Import table (if available)
		if i < len(tableFiles) {
			imp.mu.Lock()
			if imp.activeRun != nil && imp.activeRun.ID == runID {
				imp.activeRun.Progress.CurrentPhase = "tables"
			}
			imp.mu.Unlock()

			if !imp.importFile(ctx, runID, tableFiles[i], "table", "RPCSSGImportTable") {
				// Import failed, but continue with other files
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.FailedTables++
				}
				imp.mu.Unlock()
			} else {
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.ProcessedTables++
				}
				imp.mu.Unlock()
			}
		}

		// Check for cancellation/pause
		if !imp.checkRunning(runID) {
			return
		}

		// Tock: Import guide (if available)
		if i < len(guideFiles) {
			imp.mu.Lock()
			if imp.activeRun != nil && imp.activeRun.ID == runID {
				imp.activeRun.Progress.CurrentPhase = "guides"
			}
			imp.mu.Unlock()

			if !imp.importFile(ctx, runID, guideFiles[i], "guide", "RPCSSGImportGuide") {
				// Import failed, but continue with other files
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.FailedGuides++
				}
				imp.mu.Unlock()
			} else {
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.ProcessedGuides++
				}
				imp.mu.Unlock()
			}
		}

		// Check for cancellation/pause
		if !imp.checkRunning(runID) {
			return
		}

		// Tock-tock: Import manifest (if available)
		if i < len(manifestFiles) {
			imp.mu.Lock()
			if imp.activeRun != nil && imp.activeRun.ID == runID {
				imp.activeRun.Progress.CurrentPhase = "manifests"
			}
			imp.mu.Unlock()

			if !imp.importFile(ctx, runID, manifestFiles[i], "manifest", "RPCSSGImportManifest") {
				// Import failed, but continue with other files
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.FailedManifests++
				}
				imp.mu.Unlock()
			} else {
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.ProcessedManifests++
				}
				imp.mu.Unlock()
			}
		}

		// Check for cancellation/pause
		if !imp.checkRunning(runID) {
			return
		}

		// Tock-tock-tock: Import data stream (if available)
		if i < len(dataStreamFiles) {
			imp.mu.Lock()
			if imp.activeRun != nil && imp.activeRun.ID == runID {
				imp.activeRun.Progress.CurrentPhase = "datastreams"
			}
			imp.mu.Unlock()

			if !imp.importFile(ctx, runID, dataStreamFiles[i], "data stream", "RPCSSGImportDataStream") {
				// Import failed, but continue with other files
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.FailedDataStreams++
				}
				imp.mu.Unlock()
			} else {
				imp.mu.Lock()
				if imp.activeRun != nil && imp.activeRun.ID == runID {
					imp.activeRun.Progress.ProcessedDataStreams++
				}
				imp.mu.Unlock()
			}
		}
	}

	// Mark job as completed
	imp.mu.Lock()
	if imp.activeRun != nil && imp.activeRun.ID == runID {
		imp.activeRun.State = StateCompleted
		now := time.Now()
		imp.activeRun.CompletedAt = &now
		imp.activeRun.Progress.CurrentFile = ""
		imp.activeRun.Progress.CurrentPhase = ""
		imp.logger.Info("[Step 7/7] SSG import job completed: %s (tables: %d/%d, guides: %d/%d, manifests: %d/%d, data streams: %d/%d, failed: %d tables, %d guides, %d manifests, %d data streams)",
			runID,
			imp.activeRun.Progress.ProcessedTables, imp.activeRun.Progress.TotalTables,
			imp.activeRun.Progress.ProcessedGuides, imp.activeRun.Progress.TotalGuides,
			imp.activeRun.Progress.ProcessedManifests, imp.activeRun.Progress.TotalManifests,
			imp.activeRun.Progress.ProcessedDataStreams, imp.activeRun.Progress.TotalDataStreams,
			imp.activeRun.Progress.FailedTables, imp.activeRun.Progress.FailedGuides, 
			imp.activeRun.Progress.FailedManifests, imp.activeRun.Progress.FailedDataStreams)
	}
	imp.mu.Unlock()
}

// setFailed sets the job state to failed with an error message
func (imp *Importer) setFailed(runID, errorMsg string) {
	imp.mu.Lock()
	defer imp.mu.Unlock()
	if imp.activeRun != nil && imp.activeRun.ID == runID {
		imp.activeRun.State = StateFailed
		imp.activeRun.Error = errorMsg
		now := time.Now()
		imp.activeRun.CompletedAt = &now
	}
}

// checkRunning checks if the job should continue running (handles pause/stop)
func (imp *Importer) checkRunning(runID string) bool {
	for {
		imp.mu.RLock()
		if imp.activeRun == nil || imp.activeRun.ID != runID {
			imp.mu.RUnlock()
			return false
		}
		state := imp.activeRun.State
		imp.mu.RUnlock()

		if state == StateStopped {
			imp.logger.Info("Import job stopped, exiting")
			return false
		}

		if state == StatePaused {
			imp.logger.Debug("Import job paused, waiting...")
			time.Sleep(1 * time.Second)
			continue
		}

		return true
	}
}

// listFiles lists files from remote service
func (imp *Importer) listFiles(ctx context.Context, runID, rpcMethod, fileType string) ([]string, error) {
	resultMsg, err := imp.rpcInvoker.InvokeRPC(ctx, "remote", rpcMethod, nil)
	if err != nil {
		imp.setFailed(runID, fmt.Sprintf("failed to list %s: %v", fileType, err))
		imp.logger.Error("Failed to list %s: %v", fileType, err)
		return nil, err
	}

	msg, ok := resultMsg.(*subprocess.Message)
	if !ok || msg == nil {
		imp.setFailed(runID, fmt.Sprintf("invalid response format from %s", rpcMethod))
		imp.logger.Error("Invalid response format from %s", rpcMethod)
		return nil, fmt.Errorf("invalid response format")
	}

	var result map[string]interface{}
	if msg.Payload != nil {
		if err := subprocess.UnmarshalPayload(msg, &result); err != nil {
			imp.setFailed(runID, fmt.Sprintf("failed to unmarshal response: %v", err))
			imp.logger.Error("Failed to unmarshal response: %v", err)
			return nil, err
		}
	}

	filesInterface, ok := result["files"].([]interface{})
	if !ok {
		imp.setFailed(runID, fmt.Sprintf("invalid files format from %s", rpcMethod))
		imp.logger.Error("Invalid files format from %s", rpcMethod)
		return nil, fmt.Errorf("invalid files format")
	}

	// Convert to string slice
	files := make([]string, 0, len(filesInterface))
	for _, f := range filesInterface {
		if str, ok := f.(string); ok {
			files = append(files, str)
		}
	}

	return files, nil
}

// importFile imports a single file (table or guide)
func (imp *Importer) importFile(ctx context.Context, runID, filename, fileType, rpcMethod string) bool {
	imp.logger.Debug("Importing %s: %s", fileType, filename)

	// Update current file
	imp.mu.Lock()
	if imp.activeRun != nil && imp.activeRun.ID == runID {
		imp.activeRun.Progress.CurrentFile = filename
	}
	imp.mu.Unlock()

	// Get file path from remote service
	pathResultMsg, err := imp.rpcInvoker.InvokeRPC(ctx, "remote", "RPCSSGGetFilePath", map[string]interface{}{
		"filename": filename,
	})
	if err != nil {
		imp.logger.Warn("Failed to get path for %s: %v", filename, err)
		return false
	}

	pathMsg, ok := pathResultMsg.(*subprocess.Message)
	if !ok || pathMsg == nil {
		imp.logger.Warn("Invalid path response for %s", filename)
		return false
	}

	var pathResult map[string]interface{}
	if pathMsg.Payload != nil {
		if err := subprocess.UnmarshalPayload(pathMsg, &pathResult); err != nil {
			imp.logger.Warn("Failed to unmarshal path response for %s: %v", filename, err)
			return false
		}
	}

	path, ok := pathResult["path"].(string)
	if !ok {
		imp.logger.Warn("Invalid path response for %s", filename)
		return false
	}

	// Import the file with extended timeout
	importCtx, cancelImport := context.WithTimeout(ctx, 5*time.Minute)
	respMsg, err := imp.rpcInvoker.InvokeRPC(importCtx, "local", rpcMethod, map[string]interface{}{
		"path": path,
	})
	cancelImport()
	if err != nil {
		imp.logger.Warn("Failed to import %s %s: %v", fileType, filename, err)
		return false
	}

	msg, ok := respMsg.(*subprocess.Message)
	if !ok {
		imp.logger.Warn("Invalid response type for %s", filename)
		return false
	}
	if msg.Type == subprocess.MessageTypeError {
		imp.logger.Warn("Failed to import %s: %s", filename, msg.Error)
		return false
	}

	imp.logger.Info("Successfully imported %s: %s", fileType, filename)
	return true
}
