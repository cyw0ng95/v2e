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
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/taskflow"
	cwejob "github.com/cyw0ng95/v2e/pkg/cwe/job"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

const (
	// DefaultRPCTimeout is the default timeout for RPC requests to other services
	DefaultRPCTimeout = 30 * time.Second
	// DefaultSessionDBPath is the default path for the session database
	DefaultSessionDBPath = "session.db"
)

// RPCClient handles RPC communication with other services through the broker
type RPCClient struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]*requestEntry
	mu              sync.RWMutex
	correlationSeq  uint64
	logger          *common.Logger
}

type requestEntry struct {
	resp chan *subprocess.Message
	once sync.Once
}

func (e *requestEntry) signal(m *subprocess.Message) {
	e.once.Do(func() {
		e.resp <- m
		close(e.resp)
	})
}

func (e *requestEntry) close() {
	e.once.Do(func() {
		close(e.resp)
	})
}

// NewRPCClient creates a new RPC client for inter-service communication
func NewRPCClient(sp *subprocess.Subprocess, logger *common.Logger) *RPCClient {
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]*requestEntry),
		logger:          logger,
	}

	// Register handlers for response and error messages
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)

	return client
}

// handleResponse handles response messages from other services
func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Look up the pending request entry and remove it while holding the lock
	c.mu.Lock()
	entry := c.pendingRequests[msg.CorrelationID]
	if entry != nil {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()

	if entry != nil {
		entry.signal(msg)
	} else {
		c.logger.Warn("Received response for unknown correlation ID: %s", msg.CorrelationID)
	}

	return nil, nil // Don't send another response
}

// handleError handles error messages from other services (treat them as responses)
func (c *RPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Error messages are also valid responses
	return c.handleResponse(ctx, msg)
}

// InvokeRPC invokes an RPC method on another service through the broker
func (c *RPCClient) InvokeRPC(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	// Generate correlation ID
	c.mu.Lock()
	c.correlationSeq++
	correlationID := fmt.Sprintf("meta-rpc-%d-%d", time.Now().UnixNano(), c.correlationSeq)
	c.mu.Unlock()

	// Create response channel and entry
	resp := make(chan *subprocess.Message, 1)
	entry := &requestEntry{resp: resp}

	// Register pending request
	c.mu.Lock()
	c.pendingRequests[correlationID] = entry
	c.mu.Unlock()

	// Clean up on exit: remove from map and close entry
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, correlationID)
		c.mu.Unlock()
		entry.close()
	}()

	// Create request message
	var payload []byte
	if params != nil {
		data, err := subprocess.MarshalFast(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		payload = data
	}

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            method,
		Payload:       payload,
		Target:        target,
		CorrelationID: correlationID,
		Source:        c.sp.ID,
	}

	c.logger.Debug("Sending RPC request: method=%s, target=%s, correlationID=%s", method, target, correlationID)

	// Send request to broker (which will route to target)
	if err := c.sp.SendMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send RPC request: %w", err)
	}

	// Wait for response with timeout
	select {
	case response := <-resp:
		c.logger.Debug("Received RPC response: correlationID=%s, type=%s", correlationID, response.Type)
		return response, nil
	case <-time.After(DefaultRPCTimeout):
		c.logger.Warn("RPC timeout waiting for response: method=%s, target=%s, correlationID=%s", method, target, correlationID)
		c.logger.Debug("RPC invocation timed out: method=%s, target=%s, correlationID=%s", method, target, correlationID)
		return nil, fmt.Errorf("RPC timeout waiting for response from %s", target)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// RPCClientAdapter adapts RPCClient to job.RPCInvoker interface
type RPCClientAdapter struct {
	client *RPCClient
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

	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "meta"
	}

	// Log minimal startup info only (avoid dumping all environment variables)
	fmt.Fprintf(os.Stderr, "[meta] STARTUP: PROCESS_ID=%s SESSION_DB_PATH=%s\n", os.Getenv("PROCESS_ID"), os.Getenv("SESSION_DB_PATH"))

	// Set up logging using common subprocess framework
	logger, err := subprocess.SetupLogging(processID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[meta] Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	// Get run database path from environment or use default
	runDBPath := os.Getenv("SESSION_DB_PATH")
	if runDBPath == "" {
		runDBPath = DefaultSessionDBPath
	}
	logger.Info("[meta] Using run DB path: %s", runDBPath)
	if _, err := os.Stat(runDBPath); err == nil {
		logger.Info("[meta] Run DB file exists: %s", runDBPath)
	} else {
		logger.Warn("[meta] Run DB file does not exist or cannot stat: %s (err=%v)", runDBPath, err)
	}

	// Create run store
	logger.Info("[meta] Creating run store...")
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		logger.Error("[meta] Failed to create run store: %v", err)
		os.Exit(1)
	}
	logger.Info("[meta] Run store created successfully")
	defer runStore.Close()

	// Create subprocess instance
	sp := subprocess.New(processID)

	// Create RPC client for inter-service communication
	rpcClient := NewRPCClient(sp, logger)
	rpcAdapter := &RPCClientAdapter{client: rpcClient}

	// Create job executor with Taskflow (100 concurrent goroutines)
	jobExecutor := taskflow.NewJobExecutor(rpcAdapter, runStore, logger, 100)

	// Create CWE job controller (separate controller for view jobs)
	cweJobController := cwejob.NewController(rpcAdapter, logger)

	// Recover runs if needed after restart
	// This ensures job consistency when the service restarts
	recoverRuns(jobExecutor, logger)

	// Register RPC handlers for CRUD operations
	sp.RegisterHandler("RPCGetCVE", createGetCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCCreateCVE", createCreateCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCUpdateCVE", createUpdateCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCDeleteCVE", createDeleteCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCListCVEs", createListCVEsHandler(rpcClient, logger))
	sp.RegisterHandler("RPCCountCVEs", createCountCVEsHandler(rpcClient, logger))

	// Register job control RPC handlers
	sp.RegisterHandler("RPCStartSession", createStartSessionHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCStopSession", createStopSessionHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCGetSessionStatus", createGetSessionStatusHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCPauseJob", createPauseJobHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCResumeJob", createResumeJobHandler(jobExecutor, logger))

	// Register CWE view job RPC handlers
	sp.RegisterHandler("RPCStartCWEViewJob", createStartCWEViewJobHandler(cweJobController, logger))
	sp.RegisterHandler("RPCStopCWEViewJob", createStopCWEViewJobHandler(cweJobController, logger))

	logger.Info("[meta] CVE meta service started - orchestrates local and remote")

	// --- CWE Import Control ---
	go func() {
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		params := map[string]interface{}{"path": "assets/cwe-raw.json"}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCWEs", params)
		if err != nil {
			logger.Warn("Failed to import CWE on local: %v", err)
			logger.Debug("CWE import process failed: %v", err)
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Warn("CWE import error: %s", resp.Error)
			logger.Debug("CWE import process returned error: %s", resp.Error)
		} else {
			logger.Info("CWE import triggered on local")
			logger.Debug("CWE import process started successfully with path: %s", params["path"])
		}
	}()

	// --- CAPEC Import Control ---
	go func() {
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		// First check whether local already has CAPEC catalog metadata
		metaResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCGetCAPECCatalogMeta", nil)
		if err != nil {
			logger.Error("Failed to query CAPEC catalog meta on local: %v", err)
			// fall back to attempting import
		} else if metaResp.Type == subprocess.MessageTypeResponse {
			logger.Info("CAPEC catalog already present on local; skipping automatic import")
			return
		}
		// If meta not present or query failed, attempt import
		params := map[string]interface{}{"path": "assets/capec_contents_latest.xml", "xsd": "assets/capec_schema_latest.xsd"}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCAPECs", params)
		if err != nil {
			logger.Warn("Failed to import CAPEC on local: %v", err)
			logger.Debug("CAPEC import process failed: %v", err)
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Warn("CAPEC import error: %s", resp.Error)
			logger.Debug("CAPEC import process returned error: %s", resp.Error)
		} else {
			logger.Info("CAPEC import triggered on local")
			logger.Debug("CAPEC import process started successfully with path: %s", params["path"])
		}
	}()

	// Run with default lifecycle management
	subprocess.RunWithDefaults(sp, logger)
}

// createErrorResponse creates a properly formatted error response message
func createErrorResponse(msg *subprocess.Message, errorMsg string) *subprocess.Message {
	return &subprocess.Message{
		Type:          subprocess.MessageTypeError,
		ID:            msg.ID,
		Error:         errorMsg,
		CorrelationID: msg.CorrelationID,
		Target:        msg.Source,
	}
}

// isErrorResponse checks if an RPC response is an error and returns the error if so
func isErrorResponse(response *subprocess.Message) (bool, string) {
	if response.Type == subprocess.MessageTypeError {
		return true, response.Error
	}
	return false, ""
}

// createGetCVEHandler creates a handler that retrieves CVE data
// Flow: Check local storage first, if not found fetch from remote and save locally
func createGetCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing GetCVE request failed due to malformed payload: %s", string(msg.Payload))
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}

		logger.Info("RPCGetCVE: Processing request for CVE %s", req.CVEID)

		// Step 1: Check if CVE exists locally
		logger.Info("RPCGetCVE: Checking if CVE %s exists in local storage", req.CVEID)
		checkResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCIsCVEStoredByID", &rpc.CVEIDParams{CVEID: req.CVEID})
		if err != nil {
			logger.Warn("Failed to check local storage: %v", err)
			logger.Debug("GetCVE local storage check failed for CVE ID %s: %v", req.CVEID, err)
			return createErrorResponse(msg, fmt.Sprintf("failed to check local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(checkResp); isErr {
			logger.Warn("Error checking local storage: %s", errMsg)
			logger.Debug("GetCVE local storage check returned error for CVE ID %s: %s", req.CVEID, errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to check local storage: %s", errMsg)), nil
		}

		// Parse the check response
		var checkResult struct {
			Stored bool   `json:"stored"`
			CVEID  string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(checkResp, &checkResult); err != nil {
			logger.Error("Failed to parse check response: %v", err)
			logger.Debug("GetCVE failed to parse local storage check response for CVE ID %s: %v", req.CVEID, err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse check response: %v", err)), nil
		}

		var cveData *cve.CVEItem

		if checkResult.Stored {
			// Step 2a: CVE is stored locally, retrieve it
			logger.Info("RPCGetCVE: CVE %s found locally, retrieving from local storage", req.CVEID)
			getResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCGetCVEByID", &rpc.CVEIDParams{CVEID: req.CVEID})
			if err != nil {
				logger.Warn("Failed to get CVE from local storage: %v", err)
				logger.Debug("GetCVE failed to retrieve CVE from local storage for CVE ID %s: %v", req.CVEID, err)
				return createErrorResponse(msg, fmt.Sprintf("failed to get CVE from local storage: %v", err)), nil
			}

			// Check if the response is an error
			if isErr, errMsg := isErrorResponse(getResp); isErr {
				logger.Warn("Error getting CVE from local storage: %s", errMsg)
				logger.Debug("GetCVE local storage retrieval returned error for CVE ID %s: %s", req.CVEID, errMsg)
				return createErrorResponse(msg, fmt.Sprintf("failed to get CVE from local storage: %s", errMsg)), nil
			}

			if err := subprocess.UnmarshalPayload(getResp, &cveData); err != nil {
				logger.Error("Failed to parse local CVE data: %v", err)
				logger.Debug("GetCVE failed to parse local CVE data for CVE ID %s: %v", req.CVEID, err)
				return createErrorResponse(msg, fmt.Sprintf("failed to parse local CVE data: %v", err)), nil
			}
		} else {
			// Step 2b: CVE not found locally, fetch from remote
			logger.Info("RPCGetCVE: CVE %s not found locally, fetching from remote NVD API", req.CVEID)
			remoteResp, err := rpcClient.InvokeRPC(ctx, "remote", "RPCGetCVEByID", &rpc.CVEIDParams{CVEID: req.CVEID})
			if err != nil {
				logger.Warn("Failed to fetch CVE from remote: %v", err)
				logger.Debug("GetCVE failed to fetch CVE from remote for CVE ID %s: %v", req.CVEID, err)
				return createErrorResponse(msg, fmt.Sprintf("failed to fetch CVE from remote: %v", err)), nil
			}

			// Check if the response is an error
			if isErr, errMsg := isErrorResponse(remoteResp); isErr {
				logger.Warn("Error fetching CVE from remote: %s", errMsg)
				logger.Debug("GetCVE remote fetch returned error for CVE ID %s: %s", req.CVEID, errMsg)
				return createErrorResponse(msg, fmt.Sprintf("failed to fetch CVE from remote: %s", errMsg)), nil
			}

			// Parse remote response (NVD API format)
			var remoteResult cve.CVEResponse
			if err := subprocess.UnmarshalPayload(remoteResp, &remoteResult); err != nil {
				logger.Error("Failed to parse remote CVE response: %v", err)
				logger.Debug("GetCVE failed to parse remote CVE response for CVE ID %s: %v", req.CVEID, err)
				return createErrorResponse(msg, fmt.Sprintf("failed to parse remote CVE response: %v", err)), nil
			}

			// Extract CVE data from response
			if len(remoteResult.Vulnerabilities) == 0 {
				logger.Warn("CVE %s not found in NVD", req.CVEID)
				logger.Debug("GetCVE remote fetch found no vulnerabilities for CVE ID %s", req.CVEID)
				return createErrorResponse(msg, fmt.Sprintf("CVE %s not found", req.CVEID)), nil
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

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Marshal the result
		jsonData, err := subprocess.MarshalFast(cveData)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			logger.Debug("GetCVE failed to marshal response for CVE ID %s: %v", req.CVEID, err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCGetCVE: Successfully retrieved CVE %s", req.CVEID)
		logger.Debug("GetCVE request completed successfully for CVE ID %s", req.CVEID)
		return respMsg, nil
	}
}

// createCreateCVEHandler creates a handler that fetches CVE from NVD and saves locally
// This is essentially the same as Get, but enforces fetching from remote
func createCreateCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}

		logger.Info("RPCCreateCVE: Fetching CVE %s from NVD", req.CVEID)

		// Fetch from remote (NVD)
		remoteResp, err := rpcClient.InvokeRPC(ctx, "remote", "RPCGetCVEByID", map[string]interface{}{
			"cve_id": req.CVEID,
		})
		if err != nil {
			logger.Error("Failed to fetch CVE from remote: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to fetch CVE from remote: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(remoteResp); isErr {
			logger.Error("Error fetching CVE from remote: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to fetch CVE from remote: %s", errMsg)), nil
		}

		// Parse remote response (NVD API format)
		var remoteResult cve.CVEResponse
		if err := subprocess.UnmarshalPayload(remoteResp, &remoteResult); err != nil {
			logger.Error("Failed to parse remote CVE response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse remote CVE response: %v", err)), nil
		}

		// Extract CVE data from response
		if len(remoteResult.Vulnerabilities) == 0 {
			logger.Error("CVE %s not found in NVD", req.CVEID)
			return createErrorResponse(msg, fmt.Sprintf("CVE %s not found", req.CVEID)), nil
		}

		cveData := &remoteResult.Vulnerabilities[0].CVE

		// Save to local storage
		logger.Info("RPCCreateCVE: Saving CVE %s to local storage", req.CVEID)
		saveResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCSaveCVEByID", map[string]interface{}{
			"cve": cveData,
		})
		if err != nil {
			logger.Error("Failed to save CVE to local storage: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to save CVE to local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(saveResp); isErr {
			logger.Error("Error saving CVE to local storage: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to save CVE to local storage: %s", errMsg)), nil
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Return success response with CVE data
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.CVEID,
			"cve":     cveData,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCCreateCVE: Successfully created/fetched CVE %s", req.CVEID)
		return respMsg, nil
	}
}

// createUpdateCVEHandler creates a handler that refetches CVE from NVD and updates local storage
func createUpdateCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}

		logger.Info("RPCUpdateCVE: Refetching CVE %s from NVD to update local copy", req.CVEID)

		// Fetch latest data from remote (NVD)
		remoteResp, err := rpcClient.InvokeRPC(ctx, "remote", "RPCGetCVEByID", map[string]interface{}{
			"cve_id": req.CVEID,
		})
		if err != nil {
			logger.Error("Failed to fetch CVE from remote: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to fetch CVE from remote: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(remoteResp); isErr {
			logger.Error("Error fetching CVE from remote: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to fetch CVE from remote: %s", errMsg)), nil
		}

		// Parse remote response (NVD API format)
		var remoteResult cve.CVEResponse
		if err := subprocess.UnmarshalPayload(remoteResp, &remoteResult); err != nil {
			logger.Error("Failed to parse remote CVE response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse remote CVE response: %v", err)), nil
		}

		// Extract CVE data from response
		if len(remoteResult.Vulnerabilities) == 0 {
			logger.Error("CVE %s not found in NVD", req.CVEID)
			return createErrorResponse(msg, fmt.Sprintf("CVE %s not found", req.CVEID)), nil
		}

		cveData := &remoteResult.Vulnerabilities[0].CVE

		// Update local storage (save will update if exists, create if not)
		logger.Info("RPCUpdateCVE: Updating CVE %s in local storage", req.CVEID)
		saveResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCSaveCVEByID", map[string]interface{}{
			"cve": cveData,
		})
		if err != nil {
			logger.Error("Failed to update CVE in local storage: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to update CVE in local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(saveResp); isErr {
			logger.Error("Error updating CVE in local storage: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to update CVE in local storage: %s", errMsg)), nil
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Return success response with updated CVE data
		result := map[string]interface{}{
			"success": true,
			"cve_id":  req.CVEID,
			"cve":     cveData,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCUpdateCVE: Successfully updated CVE %s", req.CVEID)
		return respMsg, nil
	}
}

// createDeleteCVEHandler creates a handler that deletes CVE from local storage
func createDeleteCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}

		logger.Info("RPCDeleteCVE: Deleting CVE %s from local storage", req.CVEID)

		// Delete from local storage
		deleteResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCDeleteCVEByID", map[string]interface{}{
			"cve_id": req.CVEID,
		})
		if err != nil {
			logger.Error("Failed to delete CVE from local storage: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to delete CVE from local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(deleteResp); isErr {
			logger.Error("Error deleting CVE from local storage: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to delete CVE from local storage: %s", errMsg)), nil
		}

		// Parse delete response
		var deleteResult struct {
			Success bool   `json:"success"`
			CVEID   string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(deleteResp, &deleteResult); err != nil {
			logger.Error("Failed to parse delete response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse delete response: %v", err)), nil
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Return success response
		result := map[string]interface{}{
			"success": deleteResult.Success,
			"cve_id":  req.CVEID,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCDeleteCVE: Successfully deleted CVE %s", req.CVEID)
		return respMsg, nil
	}
}

// createListCVEsHandler creates a handler that lists CVEs from local storage
func createListCVEsHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		// Set defaults
		req.Offset = 0
		req.Limit = 10

		// Try to parse payload, but use defaults if parsing fails
		if msg.Payload != nil {
			_ = subprocess.UnmarshalPayload(msg, &req)
		}

		logger.Info("RPCListCVEs: Listing CVEs with offset=%d, limit=%d", req.Offset, req.Limit)

		// List from local storage
		listResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCListCVEs", map[string]interface{}{
			"offset": req.Offset,
			"limit":  req.Limit,
		})
		if err != nil {
			logger.Error("Failed to list CVEs from local storage: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to list CVEs from local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(listResp); isErr {
			logger.Error("Error listing CVEs from local storage: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to list CVEs from local storage: %s", errMsg)), nil
		}

		// Parse list response
		var listResult struct {
			CVEs  []cve.CVEItem `json:"cves"`
			Total int64         `json:"total"`
		}
		if err := subprocess.UnmarshalPayload(listResp, &listResult); err != nil {
			logger.Error("Failed to parse list response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse list response: %v", err)), nil
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Return list result
		result := map[string]interface{}{
			"cves":   listResult.CVEs,
			"total":  listResult.Total,
			"offset": req.Offset,
			"limit":  req.Limit,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCListCVEs: Successfully listed %d CVEs (total: %d)", len(listResult.CVEs), listResult.Total)
		return respMsg, nil
	}
}

// createCountCVEsHandler creates a handler that counts CVEs in local storage
func createCountCVEsHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCCountCVEs: Counting CVEs")

		// Count from local storage
		countResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCCountCVEs", map[string]interface{}{})
		if err != nil {
			logger.Error("Failed to count CVEs from local storage: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to count CVEs from local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(countResp); isErr {
			logger.Error("Error counting CVEs from local storage: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to count CVEs from local storage: %s", errMsg)), nil
		}

		// Parse count response
		var countResult struct {
			Count int64 `json:"count"`
		}
		if err := subprocess.UnmarshalPayload(countResp, &countResult); err != nil {
			logger.Error("Failed to parse count response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse count response: %v", err)), nil
		}

		// Create response message
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		// Return count result
		result := map[string]interface{}{
			"count": countResult.Count,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCCountCVEs: Successfully counted %d CVEs", countResult.Count)
		return respMsg, nil
	}
}

// createStartSessionHandler creates a handler that starts a new job run
func createStartSessionHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			SessionID       string `json:"session_id"`
			StartIndex      int    `json:"start_index"`
			ResultsPerBatch int    `json:"results_per_batch"`
		}

		// Set defaults
		req.StartIndex = 0
		req.ResultsPerBatch = 100

		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.SessionID == "" {
			logger.Error("session_id is required")
			return createErrorResponse(msg, "session_id is required"), nil
		}

		logger.Info("RPCStartSession: Starting job run %s (start_index=%d, batch_size=%d)",
			req.SessionID, req.StartIndex, req.ResultsPerBatch)

		// Start the job
		err := jobExecutor.Start(ctx, req.SessionID, req.StartIndex, req.ResultsPerBatch)
		if err != nil {
			logger.Error("Failed to start job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to start job: %v", err)), nil
		}

		// Get updated run state
		run, err := jobExecutor.GetStatus(req.SessionID)
		if err != nil {
			logger.Error("Failed to get run status: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get run status: %v", err)), nil
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		result := map[string]interface{}{
			"success":    true,
			"session_id": run.ID,
			"state":      run.State,
			"created_at": run.CreatedAt,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStartSession: Successfully started job run %s", run.ID)
		return respMsg, nil
	}
}

// createStopSessionHandler creates a handler that stops the current run
func createStopSessionHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCStopSession: Stopping current run")

		// Try to get active run first
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Error("Failed to get active run: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		// If no active run, try the latest persisted run as a fallback
		if run == nil {
			run, err = jobExecutor.GetLatestRun()
			if err != nil {
				logger.Error("Failed to get latest run: %v", err)
				return createErrorResponse(msg, fmt.Sprintf("failed to get latest run: %v", err)), nil
			}
			if run == nil {
				logger.Error("No active run to stop")
				return createErrorResponse(msg, "no active run"), nil
			}
		}

		// If run is actively running or paused, attempt to stop it
		if run.State == taskflow.StateRunning || run.State == taskflow.StatePaused {
			err = jobExecutor.Stop(run.ID)
			if err != nil {
				logger.Error("Failed to stop job: %v", err)
				return createErrorResponse(msg, fmt.Sprintf("failed to stop job: %v", err)), nil
			}
			// Refresh final run info
			run, _ = jobExecutor.GetStatus(run.ID)
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		result := map[string]interface{}{
			"success":       true,
			"session_id":    run.ID,
			"fetched_count": run.FetchedCount,
			"stored_count":  run.StoredCount,
			"error_count":   run.ErrorCount,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStopSession: Successfully stopped run %s", run.ID)
		return respMsg, nil
	}
}

// createGetSessionStatusHandler creates a handler that returns the current run status
func createGetSessionStatusHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetSessionStatus: Getting run status")

		// Get active run
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Error("Failed to get active run: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			// No active run - return empty status
			respMsg := &subprocess.Message{
				Type:          subprocess.MessageTypeResponse,
				ID:            msg.ID,
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}

			result := map[string]interface{}{
				"has_session": false,
			}

			jsonData, _ := subprocess.MarshalFast(result)
			respMsg.Payload = jsonData

			return respMsg, nil
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		result := map[string]interface{}{
			"has_session":       true,
			"session_id":        run.ID,
			"state":             run.State,
			"start_index":       run.StartIndex,
			"results_per_batch": run.ResultsPerBatch,
			"created_at":        run.CreatedAt,
			"updated_at":        run.UpdatedAt,
			"fetched_count":     run.FetchedCount,
			"stored_count":      run.StoredCount,
			"error_count":       run.ErrorCount,
			"error_message":     run.ErrorMessage,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData
		logger.Debug("RPCGetSessionStatus: Successfully retrieved run status")
		return respMsg, nil
	}
}

// createPauseJobHandler creates a handler that pauses the running job
func createPauseJobHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCPauseJob: Pausing job")

		// Get active run first
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Error("Failed to get active run: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Error("No active run to pause")
			return createErrorResponse(msg, "no active run"), nil
		}

		err = jobExecutor.Pause(run.ID)
		if err != nil {
			logger.Error("Failed to pause job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to pause job: %v", err)), nil
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		result := map[string]interface{}{
			"success": true,
			"state":   "paused",
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCPauseJob: Successfully paused job")
		return respMsg, nil
	}
}

// createStartCWEViewJobHandler starts a background job that fetches and saves CWE views
func createStartCWEViewJobHandler(jobController *cwejob.Controller, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req map[string]interface{}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Error("RPCStartCWEViewJob: failed to parse request: %v", err)
				return createErrorResponse(msg, "failed to parse request"), nil
			}
		}

		sessionID, err := jobController.Start(ctx, req)
		if err != nil {
			logger.Error("RPCStartCWEViewJob: failed to start job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to start job: %v", err)), nil
		}

		result := map[string]interface{}{"success": true, "session_id": sessionID}
		data, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("RPCStartCWEViewJob: failed to marshal response: %v", err)
			return createErrorResponse(msg, "failed to marshal response"), nil
		}
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, ID: msg.ID, CorrelationID: msg.CorrelationID, Target: msg.Source, Payload: data}, nil
	}
}

// createStopCWEViewJobHandler stops a running CWE view job
func createStopCWEViewJobHandler(jobController *cwejob.Controller, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			SessionID string `json:"session_id"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Error("RPCStopCWEViewJob: failed to parse request: %v", err)
				return createErrorResponse(msg, "failed to parse request"), nil
			}
		}

		err := jobController.Stop(ctx, req.SessionID)
		if err != nil && err != cwejob.ErrJobNotRunning {
			logger.Error("RPCStopCWEViewJob: failed to stop job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to stop job: %v", err)), nil
		}

		result := map[string]interface{}{"success": true, "session_id": req.SessionID}
		data, _ := subprocess.MarshalFast(result)
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, ID: msg.ID, CorrelationID: msg.CorrelationID, Target: msg.Source, Payload: data}, nil
	}
}

// createResumeJobHandler creates a handler that resumes a paused job
func createResumeJobHandler(jobExecutor *taskflow.JobExecutor, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCResumeJob: Resuming job")

		// Get active run (which should be paused)
		run, err := jobExecutor.GetActiveRun()
		if err != nil {
			logger.Error("Failed to get active run: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Error("No run to resume")
			return createErrorResponse(msg, "no paused run"), nil
		}

		err = jobExecutor.Resume(ctx, run.ID)
		if err != nil {
			logger.Error("Failed to resume job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to resume job: %v", err)), nil
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		result := map[string]interface{}{
			"success": true,
			"state":   "running",
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCResumeJob: Successfully resumed job")
		return respMsg, nil
	}
}
