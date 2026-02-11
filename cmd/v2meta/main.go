/*
Package main implements the meta RPC service.

Refer to service.md for details about the RPC API Specification, including available methods, request/response formats, and error handling.

Notes:
------
- Orchestrates operations between local and remote services
- Job sessions are persistent (stored in bbolt K-V database)
- Only one job session can run at a time
- Session state survives service restarts
- Uses RPC to communicate with local and remote services
- All communication is routed through the broker
- Session database path: session.db (default)
*/
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/taskflow"
	cwejob "github.com/cyw0ng95/v2e/pkg/cwe/job"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	ssgjob "github.com/cyw0ng95/v2e/pkg/ssg/job"
)

// RPCClientAdapter adapts rpc.Client to job.RPCInvoker interface
type RPCClientAdapter struct {
	client *rpc.Client
}

// InvokeRPC implements job.RPCInvoker interface
func (a *RPCClientAdapter) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	return a.client.InvokeRPC(ctx, target, method, params)
}

// recoverRuns checks for existing runs after restart and recovers running jobs
// This ensures consistency when the service restarts.
//
// Recovery logic:
// - "running" runs: Auto-resume (service crashed or was restarted while job was running)
// - "paused" runs: Keep paused (user explicitly paused, don't auto-resume)
// - Terminal states: No action needed
func recoverRuns(jobExecutor *taskflow.JobExecutor, logger *common.Logger) {
	if err := jobExecutor.RecoverRuns(context.Background()); err != nil {
		logger.Warn("Failed to recover runs: %v", err)
		logger.Debug("Run recovery failed: %v", err)
	}
}

func main() {

	// Use common startup utility to standardize initialization
	configStruct := subprocess.StandardStartupConfig{
		DefaultProcessID: "meta",
		LogPrefix:        "[META] ",
	}
	sp, logger := subprocess.StandardStartup(configStruct)

	// Get run database path from environment or use default
	runDBPath := os.Getenv("SESSION_DB_PATH")
	if runDBPath == "" {
		runDBPath = common.DefaultSessionDBPath
		logger.Info(LogMsgRunDBPathDefaultUsed, runDBPath)
	} else {
		logger.Info(LogMsgRunDBPathConfigured, runDBPath)
	}
	logger.Info(LogMsgUsingRunDBPath, runDBPath)
	if fileInfo, err := os.Stat(runDBPath); err == nil {
		logger.Info(LogMsgRunDBFileExists, runDBPath)
		logger.Info(LogMsgRunDBStatInfo, fileInfo.Size(), fileInfo.ModTime())
	} else {
		logger.Warn(LogMsgRunDBFileDoesNotExist, runDBPath, err)
	}

	// Create run store
	logger.Info(LogMsgRunStoreOpening, runDBPath)
	logger.Info(LogMsgCreatingRunStore)
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		logger.Error(LogMsgFailedToCreateRunStore, err)
		os.Exit(1)
	}
	logger.Info(LogMsgRunStoreOpened)
	logger.Info(LogMsgRunStoreCreated)
	defer func() {
		logger.Info(LogMsgRunStoreClosing)
		runStore.Close()
	}()

	// Create RPC client for inter-service communication
	logger.Info(LogMsgRPCClientCreated)
	rpcClient := rpc.NewClient(sp, logger, common.DefaultRPCTimeout)
	rpcAdapter := &RPCClientAdapter{client: rpcClient}
	logger.Info(LogMsgRPCAdapterCreated)

	// Create job executor with Taskflow (100 concurrent goroutines)
	logger.Info(LogMsgJobExecutorCreated, 100)
	jobExecutor := taskflow.NewJobExecutor(rpcAdapter, runStore, logger, 100)

	// Create CWE job controller (separate controller for view jobs)
	logger.Info(LogMsgCWEJobControllerCreated)
	cweJobController := cwejob.NewController(rpcAdapter, logger)

	// Create SSG import job orchestrator
	logger.Info("SSG import job orchestrator created")
	ssgImporter := ssgjob.NewImporter(rpcAdapter, logger)

	// Recover runs if needed after restart
	// This ensures job consistency when the service restarts
	logger.Info(LogMsgRunRecoveryStarted)
	recoverRuns(jobExecutor, logger)
	logger.Info(LogMsgRunRecoveryCompleted)

	// Initialize UEE FSM infrastructure
	logger.Info("Initializing UEE FSM infrastructure...")
	if err := initFSMInfrastructure(logger, runDBPath, sp); err != nil {
		logger.Error("Failed to initialize UEE FSM infrastructure: %v", err)
		// Continue without FSM infrastructure for now
	} else {
		// Register FSM control RPC handlers
		fsmHandlers := CreateFSMRPCHandlers(logger)
		for name, handler := range fsmHandlers {
			sp.RegisterHandler(name, handler)
			logger.Info(LogMsgRPCHandlerRegistered, name)
			logger.Debug(LogMsgRPCClientHandlerRegistered, name)
		}
	}

	// Register RPC handlers for CRUD operations
	logger.Info("Registering RPC handlers...")
	sp.RegisterHandler("RPCGetCVE", createGetCVEHandler(rpcClient, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetCVE")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCGetCVE")
	sp.RegisterHandler("RPCCreateCVE", createCreateCVEHandler(rpcClient, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCCreateCVE")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCCreateCVE")
	sp.RegisterHandler("RPCUpdateCVE", createUpdateCVEHandler(rpcClient, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCUpdateCVE")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCUpdateCVE")
	sp.RegisterHandler("RPCDeleteCVE", createDeleteCVEHandler(rpcClient, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCDeleteCVE")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCDeleteCVE")
	sp.RegisterHandler("RPCListCVEs", createListCVEsHandler(rpcClient, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCListCVEs")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCListCVEs")
	sp.RegisterHandler("RPCCountCVEs", createCountCVEsHandler(rpcClient, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCCountCVEs")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCCountCVEs")

	// Register job control RPC handlers
	sp.RegisterHandler("RPCStartSession", createStartSessionHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStartSession")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStartSession")
	sp.RegisterHandler("RPCStartTypedSession", createStartTypedSessionHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStartTypedSession")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStartTypedSession")
	sp.RegisterHandler("RPCStopSession", createStopSessionHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStopSession")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStopSession")
	sp.RegisterHandler("RPCGetSessionStatus", createGetSessionStatusHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetSessionStatus")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCGetSessionStatus")
	sp.RegisterHandler("RPCPauseJob", createPauseJobHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCPauseJob")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCPauseJob")
	sp.RegisterHandler("RPCResumeJob", createResumeJobHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCResumeJob")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCResumeJob")

	// Register CWE view job RPC handlers
	sp.RegisterHandler("RPCStartCWEViewJob", createStartCWEViewJobHandler(cweJobController, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStartCWEViewJob")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStartCWEViewJob")
	sp.RegisterHandler("RPCStopCWEViewJob", createStopCWEViewJobHandler(cweJobController, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStopCWEViewJob")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStopCWEViewJob")

	// Register SSG import job RPC handlers
	RegisterSSGJobHandlers(sp, ssgImporter, logger)

	// Register Memory Card proxy handlers
	registerMemoryCardProxyHandlers(sp, rpcClient, logger)

	// Register ETL tree handler
	sp.RegisterHandler("RPCGetEtlTree", createGetEtlTreeHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetEtlTree")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCGetEtlTree")

	// Register kernel metrics handler
	sp.RegisterHandler("RPCGetKernelMetrics", createGetKernelMetricsHandler(rpcClient, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetKernelMetrics")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCGetKernelMetrics")

	// Register provider control handlers
	sp.RegisterHandler("RPCStartProvider", createStartProviderHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStartProvider")
	sp.RegisterHandler("RPCPauseProvider", createPauseProviderHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCPauseProvider")
	sp.RegisterHandler("RPCStopProvider", createStopProviderHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStopProvider")

	// Register performance policy handler
	sp.RegisterHandler("RPCUpdatePerformancePolicy", createUpdatePerformancePolicyHandler(jobExecutor, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCUpdatePerformancePolicy")

	logger.Info(LogMsgServiceStarted)
	logger.Info(LogMsgServiceReady)

	// Create a shared context for import control goroutines that gets cancelled on shutdown
	importCtx, importCancel := context.WithCancel(context.Background())
	defer importCancel()

	// --- CWE Import Control ---
	go func() {
		logger.Info("Starting CWE import control routine...")
		select {
		case <-time.After(2 * time.Second):
		case <-importCtx.Done():
			logger.Info("CWE import control cancelled during initial delay")
			return
		}
		ctx, cancel := context.WithTimeout(importCtx, 120*time.Second)
		defer cancel()
		params := &rpc.ImportParams{Path: "assets/cwe-raw.json"}
		logger.Info(LogMsgCWEImportTriggered, params.Path)
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCWEs", params)
		if err != nil {
			if ctx.Err() == context.Canceled {
				logger.Info("CWE import cancelled")
			} else {
				logger.Warn("Failed to import CWE on local: %v", err)
			}
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Warn("CWE import error: %s", resp.Error)
		} else {
			logger.Info("CWE import triggered on local")
		}
	}()

	// --- CAPEC Import Control ---
	go func() {
		logger.Info("Starting CAPEC import control routine...")
		select {
		case <-time.After(2 * time.Second):
		case <-importCtx.Done():
			logger.Info("CAPEC import control cancelled during initial delay")
			return
		}
		ctx, cancel := context.WithTimeout(importCtx, 120*time.Second)
		defer cancel()
		// First check whether local already has CAPEC catalog metadata
		logger.Info("Checking for existing CAPEC catalog metadata...")
		metaResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCGetCAPECCatalogMeta", nil)
		if err != nil {
			if ctx.Err() == context.Canceled {
				logger.Info("CAPEC metadata check cancelled")
				return
			}
			logger.Warn("Failed to query CAPEC catalog meta on local: %v", err)
			// fall back to attempting import
		} else if metaResp.Type == subprocess.MessageTypeResponse {
			logger.Info("CAPEC catalog already present on local; skipping automatic import")
			return
		}
		// If meta not present or query failed, attempt import
		params := &rpc.ImportParams{Path: "assets/capec_contents_latest.xml"}
		logger.Info(LogMsgCAPECImportTriggered, params.Path)
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCAPECs", params)
		if err != nil {
			if ctx.Err() == context.Canceled {
				logger.Info("CAPEC import cancelled")
			} else {
				logger.Warn("Failed to import CAPEC on local: %v", err)
			}
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Warn("CAPEC import error: %s", resp.Error)
		} else {
			logger.Info("CAPEC import triggered on local")
		}
	}()

	// Run with default lifecycle management
	logger.Info("Starting subprocess with default lifecycle management")
	logger.Debug(LogMsgSubprocessRunStarted)
	subprocess.RunWithDefaults(sp, logger)
	logger.Debug(LogMsgSubprocessRunCompleted)
	logger.Info(LogMsgServiceShutdownStarting)
	logger.Info(LogMsgServiceShutdownComplete)
}

// createGetCVEHandler creates a handler that retrieves CVE data
// Flow: Check local storage first, if not found fetch from remote and save locally
func createGetCVEHandler(rpcClient *rpc.Client, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgRPCHandlerCalled, "RPCGetCVE")
		logger.Debug(LogMsgRPCRequestReceived, msg.Type, msg.ID, msg.Source, msg.CorrelationID)

		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing GetCVE request failed due to malformed payload: %s", string(msg.Payload))
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Warn("cve_id is required but was empty or missing")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "cve_id is required"), nil
		}

		logger.Info("RPCGetCVE: Processing request for CVE %s", req.CVEID)

		// Step 1: Check if CVE exists locally
		logger.Info("RPCGetCVE: Checking if CVE %s exists in local storage", req.CVEID)
		checkResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCIsCVEStoredByID", &rpc.CVEIDParams{CVEID: req.CVEID})
		if err != nil {
			logger.Warn("Failed to check local storage: %v", err)
			logger.Debug("GetCVE local storage check failed for CVE ID %s: %v", req.CVEID, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to check local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := subprocess.IsErrorResponse(checkResp); isErr {
			logger.Warn("Error checking local storage: %s", errMsg)
			logger.Debug("GetCVE local storage check returned error for CVE ID %s: %s", req.CVEID, errMsg)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to check local storage: %s", errMsg)), nil
		}

		// Parse the check response
		var checkResult struct {
			Stored bool   `json:"stored"`
			CVEID  string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(checkResp, &checkResult); err != nil {
			logger.Warn("Failed to parse check response: %v", err)
			logger.Debug("GetCVE failed to parse local storage check response for CVE ID %s: %v", req.CVEID, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse check response: %v", err)), nil
		}

		var cveData *cve.CVEItem

		if checkResult.Stored {
			// Step 2a: CVE is stored locally, retrieve it
			logger.Info("RPCGetCVE: CVE %s found locally, retrieving from local storage", req.CVEID)
			getResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCGetCVEByID", &rpc.CVEIDParams{CVEID: req.CVEID})
			if err != nil {
				logger.Warn("Failed to get CVE from local storage: %v", err)
				logger.Debug("GetCVE failed to retrieve CVE from local storage for CVE ID %s: %v", req.CVEID, err)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get CVE from local storage: %v", err)), nil
			}

			// Check if the response is an error
			if isErr, errMsg := subprocess.IsErrorResponse(getResp); isErr {
				logger.Warn("Error getting CVE from local storage: %s", errMsg)
				logger.Debug("GetCVE local storage retrieval returned error for CVE ID %s: %s", req.CVEID, errMsg)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get CVE from local storage: %s", errMsg)), nil
			}

			if err := subprocess.UnmarshalPayload(getResp, &cveData); err != nil {
				logger.Warn("Failed to parse local CVE data: %v", err)
				logger.Debug("GetCVE failed to parse local CVE data for CVE ID %s: %v", req.CVEID, err)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse local CVE data: %v", err)), nil
			}
		} else {
			// Step 2b: CVE not found locally, fetch from remote
			logger.Info("RPCGetCVE: CVE %s not found locally, fetching from remote NVD API", req.CVEID)
			remoteResp, err := rpcClient.InvokeRPC(ctx, "remote", "RPCGetCVEByID", &rpc.CVEIDParams{CVEID: req.CVEID})
			if err != nil {
				logger.Warn("Failed to fetch CVE from remote: %v", err)
				logger.Debug("GetCVE failed to fetch CVE from remote for CVE ID %s: %v", req.CVEID, err)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to fetch CVE from remote: %v", err)), nil
			}

			// Check if the response is an error
			if isErr, errMsg := subprocess.IsErrorResponse(remoteResp); isErr {
				logger.Warn("Error fetching CVE from remote: %s", errMsg)
				logger.Debug("GetCVE remote fetch returned error for CVE ID %s: %s", req.CVEID, errMsg)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to fetch CVE from remote: %s", errMsg)), nil
			}

			// Parse remote response (NVD API format)
			var remoteResult cve.CVEResponse
			if err := subprocess.UnmarshalPayload(remoteResp, &remoteResult); err != nil {
				logger.Warn("Failed to parse remote CVE response: %v", err)
				logger.Debug("GetCVE failed to parse remote CVE response for CVE ID %s: %v", req.CVEID, err)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse remote CVE response: %v", err)), nil
			}

			// Extract CVE data from response
			if len(remoteResult.Vulnerabilities) == 0 {
				logger.Warn("CVE %s not found in NVD", req.CVEID)
				logger.Debug("GetCVE remote fetch found no vulnerabilities for CVE ID %s", req.CVEID)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("CVE %s not found", req.CVEID)), nil
			}

			cveData = &remoteResult.Vulnerabilities[0].CVE

			// Step 3: Save fetched CVE to local storage
			logger.Info("RPCGetCVE: Saving CVE %s to local storage", req.CVEID)
			_, err = rpcClient.InvokeRPC(ctx, "local", "RPCSaveCVEByID", &rpc.SaveCVEByIDParams{CVE: *cveData})
			if err != nil {
				logger.Warn("Failed to save CVE to local storage (continuing anyway): %v", err)
				logger.Debug("GetCVE save to local storage failed for CVE ID %s: %v", req.CVEID, err)
				// Continue even if save fails - we still have the data
			}
		}

		logger.Info("RPCGetCVE: Successfully retrieved CVE %s", req.CVEID)
		logger.Debug("GetCVE request completed successfully for CVE ID %s", req.CVEID)
		return subprocess.NewSuccessResponse(msg, cveData)
	}
}

// createGetSessionStatusHandler creates a handler that returns the current run status
func createGetSessionStatusHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetSessionStatus: Getting run status")

		// Get active run
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Warn("Failed to get active run: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			// No active run - return empty status
			return subprocess.NewSuccessResponse(msg, map[string]interface{}{
				"has_session": false,
			})
		}

		// Prepare the enhanced result with new data structures
		result := map[string]interface{}{
			"has_session":       true,
			"session_id":        run.ID,
			"state":             run.State,
			"data_type":         run.DataType, // New field
			"start_index":       run.StartIndex,
			"results_per_batch": run.ResultsPerBatch,
			"created_at":        run.CreatedAt,
			"updated_at":        run.UpdatedAt,
			"fetched_count":     run.FetchedCount,
			"stored_count":      run.StoredCount,
			"error_count":       run.ErrorCount,
			"error_message":     run.ErrorMessage,
			// New fields for enhanced progress tracking
			"progress": run.Progress,
			"params":   run.Params,
		}

		logger.Debug("RPCGetSessionStatus: Successfully retrieved run status")
		return subprocess.NewSuccessResponse(msg, result)
	}
}

// createPauseJobHandler creates a handler that pauses the running job
func createPauseJobHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCPauseJob: Pausing job")

		// Get active run first
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Warn("Failed to get active run: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Warn("No active run to pause")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "no active run"), nil
		}

		err = jobExecutor.Pause(run.ID)
		if err != nil {
			logger.Warn("Failed to pause job: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to pause job: %v", err)), nil
		}

		logger.Info("RPCPauseJob: Successfully paused job")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
			"state":   "paused",
		})
	}
}

// createResumeJobHandler creates a handler that resumes the paused job
func createResumeJobHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCResumeJob: Resuming job")

		// Get active run first
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Warn("Failed to get active run: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Warn("No active run to resume")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "no active run"), nil
		}

		err = jobExecutor.Resume(ctx, run.ID)
		if err != nil {
			logger.Warn("Failed to resume job: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to resume job: %v", err)), nil
		}

		logger.Info("RPCResumeJob: Successfully resumed job")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
			"state":   "running",
		})
	}
}

// createStartSessionHandler creates a handler that starts a new job session
func createStartSessionHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			DataType        taskflow.DataType `json:"data_type"`
			StartIndex      int               `json:"start_index"`
			ResultsPerBatch int               `json:"results_per_batch"`
			Params          map[string]interface{}
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		logger.Info("RPCStartSession: Starting new job session with data type %s", req.DataType)

		if req.DataType == "" {
			logger.Warn("data_type is required but was empty or missing")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "data_type is required"), nil
		}

		if req.StartIndex < 0 {
			logger.Warn("start_index must be non-negative")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "start_index must be non-negative"), nil
		}

		if req.ResultsPerBatch <= 0 {
			logger.Warn("results_per_batch must be positive")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "results_per_batch must be positive"), nil
		}

		sessionID := fmt.Sprintf("%s-%d", req.DataType, time.Now().Unix())

		err := jobExecutor.StartTyped(ctx, sessionID, req.StartIndex, req.ResultsPerBatch, req.DataType)
		if err != nil {
			logger.Warn("Failed to start job session: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to start job session: %v", err)), nil
		}

		// Get updated run state
		run, err := jobExecutor.GetStatus(sessionID)
		if err != nil {
			logger.Warn("Failed to get run status: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get run status: %v", err)), nil
		}

		logger.Info("RPCStartSession: Successfully started job session: %s", run.ID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":     true,
			"session_id":  run.ID,
			"data_type":   string(run.DataType),
			"state":       run.State,
			"created_at":  run.CreatedAt,
			"start_index": run.StartIndex,
			"batch_size":  run.ResultsPerBatch,
		})
	}
}

// createStopSessionHandler creates a handler that stops the running job session
func createStopSessionHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCStopSession: Stopping job session")

		// Get active run first
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Warn("Failed to get active run: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Warn("No active run to stop")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "no active run"), nil
		}

		err = jobExecutor.Stop(run.ID)
		if err != nil {
			logger.Warn("Failed to stop job session: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to stop job session: %v", err)), nil
		}

		logger.Info("RPCStopSession: Successfully stopped job session")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createStartCWEViewJobHandler creates a handler that starts a CWE view job
func createStartCWEViewJobHandler(controller *cwejob.Controller, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCStartCWEViewJob: Starting CWE view job")

		// Parse the request payload
		var req struct {
			Params map[string]interface{}
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		// Start the CWE view job
		sessionID, err := controller.Start(ctx, req.Params)
		if err != nil {
			logger.Warn("Failed to start CWE view job: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to start CWE view job: %v", err)), nil
		}

		logger.Info("RPCStartCWEViewJob: Successfully started CWE view job: %s", sessionID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
		})
	}
}

// createStopCWEViewJobHandler creates a handler that stops a CWE view job
func createStopCWEViewJobHandler(controller *cwejob.Controller, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCStopCWEViewJob: Stopping CWE view job")

		// Parse the request payload
		var req struct {
			SessionID string `json:"session_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.SessionID == "" {
			logger.Error("session_id is required but was empty or missing")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "session_id is required"), nil
		}

		// Stop the CWE view job
		err := controller.Stop(ctx, req.SessionID)
		if err != nil {
			logger.Error("Failed to stop CWE view job: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to stop CWE view job: %v", err)), nil
		}

		logger.Info("RPCStopCWEViewJob: Successfully stopped CWE view job: %s", req.SessionID)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createCreateCVEHandler creates a handler that creates a new CVE
func createCreateCVEHandler(rpcClient *rpc.Client, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgRPCHandlerCalled, "RPCCreateCVE")
		logger.Debug(LogMsgRPCRequestReceived, msg.Type, msg.ID, msg.Source, msg.CorrelationID)

		// Parse the request payload
		var req cve.CVEItem
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		// Create the CVE
		// Validate required fields before making RPC calls
		if req.ID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "cve_id is required"), nil
		}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCCreateCVE", &req)
		if err != nil {
			logger.Warn("Failed to create CVE: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to create CVE: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
			logger.Warn("Error creating CVE: %s", errMsg)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to create CVE: %s", errMsg)), nil
		}

		logger.Info("RPCCreateCVE: Successfully created CVE")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createUpdateCVEHandler creates a handler that updates an existing CVE
func createUpdateCVEHandler(rpcClient *rpc.Client, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgRPCHandlerCalled, "RPCUpdateCVE")
		logger.Debug(LogMsgRPCRequestReceived, msg.Type, msg.ID, msg.Source, msg.CorrelationID)

		// Parse the request payload
		var req cve.CVEItem
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		// Update the CVE
		// Validate required fields before making RPC calls
		if req.ID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "cve_id is required"), nil
		}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCUpdateCVE", &req)
		if err != nil {
			logger.Warn("Failed to update CVE: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to update CVE: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
			logger.Warn("Error updating CVE: %s", errMsg)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to update CVE: %s", errMsg)), nil
		}

		logger.Info("RPCUpdateCVE: Successfully updated CVE")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createDeleteCVEHandler creates a handler that deletes an existing CVE
func createDeleteCVEHandler(rpcClient *rpc.Client, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgRPCHandlerCalled, "RPCDeleteCVE")
		logger.Debug(LogMsgRPCRequestReceived, msg.Type, msg.ID, msg.Source, msg.CorrelationID)

		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "cve_id is required"), nil
		}

		// Delete the CVE
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCDeleteCVE", &rpc.CVEIDParams{CVEID: req.CVEID})
		if err != nil {
			logger.Warn("Failed to delete CVE: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to delete CVE: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
			logger.Warn("Error deleting CVE: %s", errMsg)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to delete CVE: %s", errMsg)), nil
		}

		logger.Info("RPCDeleteCVE: Successfully deleted CVE")
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success": true,
		})
	}
}

// createListCVEsHandler creates a handler that lists CVEs
func createListCVEsHandler(rpcClient *rpc.Client, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgRPCHandlerCalled, "RPCListCVEs")
		logger.Debug(LogMsgRPCRequestReceived, msg.Type, msg.ID, msg.Source, msg.CorrelationID)

		// Parse the request payload
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.Offset < 0 {
			logger.Error("offset must be non-negative")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "offset must be non-negative"), nil
		}

		if req.Limit <= 0 {
			logger.Error("limit must be positive")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "limit must be positive"), nil
		}

		// List CVEs
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCListCVEs", &req)
		if err != nil {
			logger.Warn("Failed to list CVEs: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to list CVEs: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
			logger.Warn("Error listing CVEs: %s", errMsg)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to list CVEs: %s", errMsg)), nil
		}

		logger.Info("RPCListCVEs: Successfully listed CVEs")
		// Forward the response directly (payload is already marshaled)
		return resp, nil
	}
}

// createCountCVEsHandler creates a handler that counts CVEs
func createCountCVEsHandler(rpcClient *rpc.Client, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgRPCHandlerCalled, "RPCCountCVEs")
		logger.Debug(LogMsgRPCRequestReceived, msg.Type, msg.ID, msg.Source, msg.CorrelationID)

		// Count CVEs
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCCountCVEs", nil)
		if err != nil {
			logger.Warn("Failed to count CVEs: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to count CVEs: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
			logger.Warn("Error counting CVEs: %s", errMsg)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to count CVEs: %s", errMsg)), nil
		}

		logger.Info("RPCCountCVEs: Successfully counted CVEs")
		// Forward the response directly (payload is already marshaled)
		return resp, nil
	}
}

// createStartTypedSessionHandler creates a handler that starts a new job run with a specific data type
func createStartTypedSessionHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			SessionID       string                 `json:"session_id"`
			StartIndex      int                    `json:"start_index"`
			ResultsPerBatch int                    `json:"results_per_batch"`
			DataType        taskflow.DataType      `json:"data_type"`
			Params          map[string]interface{} `json:"params,omitempty"`
		}

		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		// Set defaults only if not provided (after unmarshaling)
		if req.StartIndex == 0 {
			req.StartIndex = 0
		}
		if req.ResultsPerBatch == 0 {
			req.ResultsPerBatch = 100
		}
		if req.DataType == "" {
			req.DataType = taskflow.DataTypeCVE // default to CVE
		}

		if req.SessionID == "" {
			logger.Error("session_id is required")
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "session_id is required"), nil
		}

		logger.Info("RPCStartTypedSession: Starting job run %s (data_type=%s, start_index=%d, batch_size=%d)",
			req.SessionID, req.DataType, req.StartIndex, req.ResultsPerBatch)

		// Start the job with the specified data type
		err := jobExecutor.StartTyped(ctx, req.SessionID, req.StartIndex, req.ResultsPerBatch, req.DataType)
		if err != nil {
			logger.Error("Failed to start job: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to start job: %v", err)), nil
		}

		// Get updated run state
		run, err := jobExecutor.GetStatus(req.SessionID)
		if err != nil {
			logger.Error("Failed to get run status: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get run status: %v", err)), nil
		}

		logger.Info("RPCStartTypedSession: Successfully started job run %s (type: %s)", run.ID, run.DataType)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"success":     true,
			"session_id":  run.ID,
			"state":       run.State,
			"data_type":   run.DataType,
			"created_at":  run.CreatedAt,
			"start_index": run.StartIndex,
			"batch_size":  run.ResultsPerBatch,
			"params":      req.Params,
		})
	}
}

// createProxyHandler returns a handler that proxies the RPC call to the given target and method.
func createProxyHandler(rpcClient *rpc.Client, logger *common.Logger, target, method string) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Forward the payload as-is to the target service/method
		var params interface{}
		if msg.Payload != nil {
			params = msg.Payload
		}
		resp, err := rpcClient.InvokeRPC(ctx, target, method, params)
		if err != nil {
			logger.Warn("Proxy handler failed: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("proxy handler failed: %v", err)), nil
		}
		return resp, nil
	}
}

// Memory Card RPC handlers
func registerMemoryCardProxyHandlers(sp *subprocess.Subprocess, rpcClient *rpc.Client, logger *common.Logger) {
	sp.RegisterHandler("RPCCreateMemoryCard", createProxyHandler(rpcClient, logger, "local", "RPCCreateMemoryCard"))
	sp.RegisterHandler("RPCGetMemoryCard", createProxyHandler(rpcClient, logger, "local", "RPCGetMemoryCard"))
	sp.RegisterHandler("RPCUpdateMemoryCard", createProxyHandler(rpcClient, logger, "local", "RPCUpdateMemoryCard"))
	sp.RegisterHandler("RPCDeleteMemoryCard", createProxyHandler(rpcClient, logger, "local", "RPCDeleteMemoryCard"))
	sp.RegisterHandler("RPCListMemoryCards", createProxyHandler(rpcClient, logger, "local", "RPCListMemoryCards"))
	logger.Info("Memory Card proxy handlers registered")
}

// createGetEtlTreeHandler creates a handler that returns the ETL tree with macro FSM and provider states
func createGetEtlTreeHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetEtlTree: Getting ETL tree")

		// Get active run to determine macro state
		activeRun, err := jobExecutor.GetActiveRun()
		macroState := "IDLE"
		providers := make([]map[string]interface{}, 0)
		activeProviders := 0

		if err == nil && activeRun != nil {
			// Map run state to macro FSM state
			switch activeRun.State {
			case taskflow.StateQueued:
				macroState = "BOOTSTRAPPING"
			case taskflow.StateRunning:
				macroState = "ORCHESTRATING"
			case taskflow.StatePaused:
				macroState = "STABILIZING"
			case taskflow.StateCompleted, taskflow.StateStopped:
				macroState = "DRAINING"
			default:
				macroState = "IDLE"
			}

			// Build provider node from active run
			providerState := "IDLE"
			switch activeRun.State {
			case taskflow.StateQueued:
				providerState = "ACQUIRING"
			case taskflow.StateRunning:
				providerState = "RUNNING"
				activeProviders++
			case taskflow.StatePaused:
				providerState = "WAITING_QUOTA"
			case taskflow.StateCompleted:
				providerState = "TERMINATED"
			case taskflow.StateStopped:
				providerState = "PAUSED"
			}

			providerType := string(activeRun.DataType)
			if providerType == "" {
				providerType = "unknown"
			}

			providers = append(providers, map[string]interface{}{
				"id":             activeRun.ID,
				"providerType":   providerType,
				"state":          providerState,
				"processedCount": activeRun.FetchedCount,
				"errorCount":     activeRun.ErrorCount,
				"permitsHeld":    0, // No permit system after CCE removal
				"createdAt":      activeRun.CreatedAt.Format(time.RFC3339),
				"updatedAt":      activeRun.UpdatedAt.Format(time.RFC3339),
			})
		}

		// Build macro node
		now := time.Now()
		macro := map[string]interface{}{
			"id":        "main-orchestrator",
			"state":     macroState,
			"providers": providers,
			"createdAt": now.Add(-24 * time.Hour).Format(time.RFC3339), // Default creation time
			"updatedAt": now.Format(time.RFC3339),
		}

		result := map[string]interface{}{
			"tree": map[string]interface{}{
				"macro":           macro,
				"totalProviders":  len(providers),
				"activeProviders": activeProviders,
			},
		}

		logger.Debug("RPCGetEtlTree: Successfully retrieved ETL tree")
		return subprocess.NewSuccessResponse(msg, result)
	}
}

// createGetKernelMetricsHandler creates a handler that returns kernel performance metrics
func createGetKernelMetricsHandler(rpcClient *rpc.Client, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetKernelMetrics: Getting kernel metrics")

		// Query broker for message stats
		resp, err := rpcClient.InvokeRPC(ctx, "broker", "RPCGetMessageStats", nil)
		if err != nil {
			logger.Warn("Failed to get message stats from broker: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get metrics: %v", err)), nil
		}

		if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
			logger.Warn("Error getting message stats: %s", errMsg)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to get metrics: %s", errMsg)), nil
		}

		// Parse stats response
		var stats map[string]interface{}
		if err := subprocess.UnmarshalPayload(resp, &stats); err != nil {
			logger.Warn("Failed to parse stats response: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to parse metrics: %v", err)), nil
		}

		// Extract metrics from response
		total, hasTotal := stats["total"].(map[string]interface{})
		totalSent := 0
		totalReceived := 0
		if hasTotal {
			if ts, ok := total["total_sent"].(float64); ok {
				totalSent = int(ts)
			}
			if tr, ok := total["total_received"].(float64); ok {
				totalReceived = int(tr)
			}
		}

		// Build metrics response with calculated values
		now := time.Now()
		result := map[string]interface{}{
			"metrics": map[string]interface{}{
				"p99Latency":       10.0,                                    // Default P99 latency (ms)
				"bufferSaturation": 25.0,                                    // Default buffer saturation (%)
				"messageRate":      float64(totalSent+totalReceived) / 60.0, // Messages per second (approx)
				"errorRate":        0.0,                                     // Default error rate
				"timestamp":        now.Format(time.RFC3339),
			},
		}

		logger.Debug("RPCGetKernelMetrics: Successfully retrieved kernel metrics")
		return subprocess.NewSuccessResponse(msg, result)
	}
}

// createStartProviderHandler creates a handler to start a provider
func createStartProviderHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCStartProvider: Starting provider")

		var params struct {
			ProviderID string `json:"providerId"`
		}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			logger.Warn("Failed to parse start provider params: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("invalid params: %v", err)), nil
		}

		// Get active run for the provider
		activeRun, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Warn("No active run found: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "no active run found"), nil
		}

		if activeRun.ID != params.ProviderID {
			logger.Warn("Active run %s does not match requested provider %s", activeRun.ID, params.ProviderID)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "provider not active"), nil
		}

		// Resume the paused run
		if err := jobExecutor.Resume(ctx, activeRun.ID); err != nil {
			logger.Warn("Failed to resume run %s: %v", activeRun.ID, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to start provider: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"providerId": params.ProviderID,
		}

		logger.Debug("RPCStartProvider: Provider started successfully")
		return subprocess.NewSuccessResponse(msg, result)
	}
}

// createPauseProviderHandler creates a handler to pause a provider
func createPauseProviderHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCPauseProvider: Pausing provider")

		var params struct {
			ProviderID string `json:"providerId"`
		}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			logger.Warn("Failed to parse pause provider params: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("invalid params: %v", err)), nil
		}

		// Get active run for the provider
		activeRun, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Warn("No active run found: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "no active run found"), nil
		}

		if activeRun.ID != params.ProviderID {
			logger.Warn("Active run %s does not match requested provider %s", activeRun.ID, params.ProviderID)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "provider not active"), nil
		}

		// Pause the run
		if err := jobExecutor.Pause(activeRun.ID); err != nil {
			logger.Warn("Failed to pause run %s: %v", activeRun.ID, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to pause provider: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"providerId": params.ProviderID,
		}

		logger.Debug("RPCPauseProvider: Provider paused successfully")
		return subprocess.NewSuccessResponse(msg, result)
	}
}

// createStopProviderHandler creates a handler to stop a provider
func createStopProviderHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCStopProvider: Stopping provider")

		var params struct {
			ProviderID string `json:"providerId"`
		}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			logger.Warn("Failed to parse stop provider params: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("invalid params: %v", err)), nil
		}

		// Get active run for the provider
		activeRun, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Warn("No active run found: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "no active run found"), nil
		}

		if activeRun.ID != params.ProviderID {
			logger.Warn("Active run %s does not match requested provider %s", activeRun.ID, params.ProviderID)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "provider not active"), nil
		}

		// Stop the run
		if err := jobExecutor.Stop(activeRun.ID); err != nil {
			logger.Warn("Failed to stop run %s: %v", activeRun.ID, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to stop provider: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"providerId": params.ProviderID,
		}

		logger.Debug("RPCStopProvider: Provider stopped successfully")
		return subprocess.NewSuccessResponse(msg, result)
	}
}

// createUpdatePerformancePolicyHandler creates a handler to update performance policy
func createUpdatePerformancePolicyHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCUpdatePerformancePolicy: Updating performance policy")

		var params struct {
			ProviderID string                 `json:"providerId"`
			Policy     map[string]interface{} `json:"policy"`
		}
		if err := subprocess.UnmarshalPayload(msg, &params); err != nil {
			logger.Warn("Failed to parse update policy params: %v", err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("invalid params: %v", err)), nil
		}

		// In a real implementation, this would update the performance policy
		// for the provider. For now, we'll log it and return success.
		logger.Info("Performance policy update for provider %s: %+v", params.ProviderID, params.Policy)

		result := map[string]interface{}{
			"success":    true,
			"providerId": params.ProviderID,
		}

		logger.Debug("RPCUpdatePerformancePolicy: Policy updated successfully")
		return subprocess.NewSuccessResponse(msg, result)
	}
}
