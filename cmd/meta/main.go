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

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cve"
	"github.com/cyw0ng95/v2e/pkg/cve/job"
	"github.com/cyw0ng95/v2e/pkg/cve/session"
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
	pendingRequests map[string]chan *subprocess.Message
	mu              sync.RWMutex
	correlationSeq  uint64
	logger          *common.Logger
}

// NewRPCClient creates a new RPC client for inter-service communication
func NewRPCClient(sp *subprocess.Subprocess, logger *common.Logger) *RPCClient {
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]chan *subprocess.Message),
		logger:          logger,
	}

	// Register handlers for response and error messages
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)

	return client
}

// handleResponse handles response messages from other services
func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	// Look up the pending request
	c.mu.Lock()
	respChan, exists := c.pendingRequests[msg.CorrelationID]
	if exists {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()

	if exists {
		select {
		case respChan <- msg:
		case <-time.After(1 * time.Second):
			c.logger.Warn("Timeout sending response to channel for correlation ID: %s", msg.CorrelationID)
		}
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

	// Create response channel
	respChan := make(chan *subprocess.Message, 1)

	// Register pending request
	c.mu.Lock()
	c.pendingRequests[correlationID] = respChan
	c.mu.Unlock()

	// Clean up on exit
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, correlationID)
		c.mu.Unlock()
		close(respChan)
	}()

	// Create request message
	var payload []byte
	if params != nil {
		data, err := sonic.Marshal(params)
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
	case response := <-respChan:
		c.logger.Debug("Received RPC response: correlationID=%s, type=%s", correlationID, response.Type)
		return response, nil
	case <-time.After(DefaultRPCTimeout):
		c.logger.Error("RPC timeout waiting for response: method=%s, target=%s, correlationID=%s", method, target, correlationID)
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

// recoverSession checks for existing sessions after restart and recovers running jobs
// This ensures consistency when the service restarts.
//
// Recovery logic:
// - "running" sessions: Auto-resume (service crashed or was restarted while job was running)
// - "paused" sessions: Keep paused (user explicitly paused, don't auto-resume)
// - Other states: No action needed
func recoverSession(jobController *job.Controller, sessionManager *session.Manager, logger *common.Logger) {
	// Check if there's an existing session
	sess, err := sessionManager.GetSession()
	if err != nil {
		if err == session.ErrNoSession {
			logger.Info("No existing session to recover")
			return
		}
		logger.Error("Failed to check for existing session: %v", err)
		return
	}

	logger.Info("Found existing session: id=%s, state=%s, fetched=%d, stored=%d",
		sess.ID, sess.State, sess.FetchedCount, sess.StoredCount)

	// Only auto-recover if the session was in "running" state
	// Paused sessions should stay paused until explicitly resumed via RPCResumeJob
	if sess.State == session.StateRunning {
		logger.Info("Recovering running session: id=%s", sess.ID)

		// Use Start() for sessions that are already in "running" state
		// Start() doesn't check the state precondition like Resume() does
		err := jobController.Start(context.Background())
		if err != nil {
			// If start fails (e.g., job already running somehow), try to handle gracefully
			if err == job.ErrJobRunning {
				logger.Info("Job is already running - recovery not needed")
				return
			}
			logger.Error("Failed to recover running session: %v", err)
			// Don't change state on recovery failure - keep it as "running"
			// This way, next restart will try again
			logger.Warn("Session will remain in 'running' state for next recovery attempt")
			return
		}

		logger.Info("Successfully recovered running session: id=%s", sess.ID)
	} else if sess.State == session.StatePaused {
		logger.Info("Session is paused - not auto-recovering. Use RPCResumeJob to manually resume.")
	} else {
		logger.Info("Session state is '%s' - no recovery needed", sess.State)
	}
}

func main() {

	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "meta"
	}

	// Log all environment variables for debug
	fmt.Fprintf(os.Stderr, "[meta] ENV: PROCESS_ID=%s SESSION_DB_PATH=%s PWD=%s\n", os.Getenv("PROCESS_ID"), os.Getenv("SESSION_DB_PATH"), os.Getenv("PWD"))
	for _, e := range os.Environ() {
		fmt.Fprintf(os.Stderr, "[meta] ENV: %s\n", e)
	}

	// Set up logging using common subprocess framework
	logger, err := subprocess.SetupLogging(processID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[meta] Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	// Get session database path from environment or use default
	sessionDBPath := os.Getenv("SESSION_DB_PATH")
	if sessionDBPath == "" {
		sessionDBPath = DefaultSessionDBPath
	}
	logger.Info("[meta] Using session DB path: %s", sessionDBPath)
	if _, err := os.Stat(sessionDBPath); err == nil {
		logger.Info("[meta] Session DB file exists: %s", sessionDBPath)
	} else {
		logger.Warn("[meta] Session DB file does not exist or cannot stat: %s (err=%v)", sessionDBPath, err)
	}

	// Create session manager
	logger.Info("[meta] Creating session manager...")
	sessionManager, err := session.NewManager(sessionDBPath, logger)
	if err != nil {
		logger.Error("[meta] Failed to create session manager: %v", err)
		os.Exit(1)
	}
	logger.Info("[meta] Session manager created successfully")
	defer sessionManager.Close()

	// Create subprocess instance
	sp := subprocess.New(processID)

	// Create RPC client for inter-service communication
	rpcClient := NewRPCClient(sp, logger)
	rpcAdapter := &RPCClientAdapter{client: rpcClient}

	// Create job controller
	jobController := job.NewController(rpcAdapter, sessionManager, logger)

	// Recover session if needed after restart
	// This ensures job consistency when the service restarts
	recoverSession(jobController, sessionManager, logger)

	// Register RPC handlers for CRUD operations
	sp.RegisterHandler("RPCGetCVE", createGetCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCCreateCVE", createCreateCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCUpdateCVE", createUpdateCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCDeleteCVE", createDeleteCVEHandler(rpcClient, logger))
	sp.RegisterHandler("RPCListCVEs", createListCVEsHandler(rpcClient, logger))
	sp.RegisterHandler("RPCCountCVEs", createCountCVEsHandler(rpcClient, logger))

	// Register job control RPC handlers
	sp.RegisterHandler("RPCStartSession", createStartSessionHandler(sessionManager, jobController, logger))
	sp.RegisterHandler("RPCStopSession", createStopSessionHandler(sessionManager, jobController, logger))
	sp.RegisterHandler("RPCGetSessionStatus", createGetSessionStatusHandler(sessionManager, logger))
	sp.RegisterHandler("RPCPauseJob", createPauseJobHandler(jobController, logger))
	sp.RegisterHandler("RPCResumeJob", createResumeJobHandler(jobController, logger))

	logger.Info("[meta] CVE meta service started - orchestrates local and remote")

	// --- CWE Import Control ---
	go func() {
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		params := map[string]interface{}{"path": "assets/cwe-raw.json"}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCWEs", params)
		if err != nil {
			logger.Error("Failed to import CWE on local: %v", err)
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Error("CWE import error: %s", resp.Error)
		} else {
			logger.Info("CWE import triggered on local")
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
			logger.Error("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}

		logger.Info("RPCGetCVE: Processing request for CVE %s", req.CVEID)

		// Step 1: Check if CVE exists locally
		logger.Info("RPCGetCVE: Checking if CVE %s exists in local storage", req.CVEID)
		checkResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCIsCVEStoredByID", map[string]interface{}{
			"cve_id": req.CVEID,
		})
		if err != nil {
			logger.Error("Failed to check local storage: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to check local storage: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(checkResp); isErr {
			logger.Error("Error checking local storage: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to check local storage: %s", errMsg)), nil
		}

		// Parse the check response
		var checkResult struct {
			Stored bool   `json:"stored"`
			CVEID  string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(checkResp, &checkResult); err != nil {
			logger.Error("Failed to parse check response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse check response: %v", err)), nil
		}

		var cveData *cve.CVEItem

		if checkResult.Stored {
			// Step 2a: CVE is stored locally, retrieve it
			logger.Info("RPCGetCVE: CVE %s found locally, retrieving from local storage", req.CVEID)
			getResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCGetCVEByID", map[string]interface{}{
				"cve_id": req.CVEID,
			})
			if err != nil {
				logger.Error("Failed to get CVE from local storage: %v", err)
				return createErrorResponse(msg, fmt.Sprintf("failed to get CVE from local storage: %v", err)), nil
			}

			// Check if the response is an error
			if isErr, errMsg := isErrorResponse(getResp); isErr {
				logger.Error("Error getting CVE from local storage: %s", errMsg)
				return createErrorResponse(msg, fmt.Sprintf("failed to get CVE from local storage: %s", errMsg)), nil
			}

			if err := subprocess.UnmarshalPayload(getResp, &cveData); err != nil {
				logger.Error("Failed to parse local CVE data: %v", err)
				return createErrorResponse(msg, fmt.Sprintf("failed to parse local CVE data: %v", err)), nil
			}
		} else {
			// Step 2b: CVE not found locally, fetch from remote
			logger.Info("RPCGetCVE: CVE %s not found locally, fetching from remote NVD API", req.CVEID)
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

			cveData = &remoteResult.Vulnerabilities[0].CVE

			// Step 3: Save fetched CVE to local storage
			logger.Info("RPCGetCVE: Saving CVE %s to local storage", req.CVEID)
			_, err = rpcClient.InvokeRPC(ctx, "local", "RPCSaveCVEByID", map[string]interface{}{
				"cve": cveData,
			})
			if err != nil {
				logger.Warn("Failed to save CVE to local storage (continuing anyway): %v", err)
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
		jsonData, err := sonic.Marshal(cveData)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCGetCVE: Successfully retrieved CVE %s", req.CVEID)
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

		jsonData, err := sonic.Marshal(result)
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

		jsonData, err := sonic.Marshal(result)
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

		jsonData, err := sonic.Marshal(result)
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

		jsonData, err := sonic.Marshal(result)
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

		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCCountCVEs: Successfully counted %d CVEs", countResult.Count)
		return respMsg, nil
	}
}

// createStartSessionHandler creates a handler that starts a new job session
func createStartSessionHandler(sessionManager *session.Manager, jobController *job.Controller, logger *common.Logger) subprocess.Handler {
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

		logger.Info("RPCStartSession: Creating session %s (start_index=%d, batch_size=%d)",
			req.SessionID, req.StartIndex, req.ResultsPerBatch)

		// Create new session
		sess, err := sessionManager.CreateSession(req.SessionID, req.StartIndex, req.ResultsPerBatch)
		if err != nil {
			logger.Error("Failed to create session: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to create session: %v", err)), nil
		}

		// Start the job
		err = jobController.Start(ctx)
		if err != nil {
			logger.Error("Failed to start job: %v", err)
			// Clean up session if job failed to start
			sessionManager.DeleteSession()
			return createErrorResponse(msg, fmt.Sprintf("failed to start job: %v", err)), nil
		}

		// Get updated session state after starting the job
		sess, err = sessionManager.GetSession()
		if err != nil {
			logger.Error("Failed to get updated session: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get updated session: %v", err)), nil
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
			"session_id": sess.ID,
			"state":      sess.State,
			"created_at": sess.CreatedAt,
		}

		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStartSession: Successfully started session %s", sess.ID)
		return respMsg, nil
	}
}

// createStopSessionHandler creates a handler that stops the current session
func createStopSessionHandler(sessionManager *session.Manager, jobController *job.Controller, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCStopSession: Stopping current session")

		// Stop the job
		err := jobController.Stop()
		if err != nil && err != job.ErrJobNotRunning {
			logger.Error("Failed to stop job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to stop job: %v", err)), nil
		}

		// Get session info before deleting
		sess, err := sessionManager.GetSession()
		if err != nil {
			logger.Error("Failed to get session: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get session: %v", err)), nil
		}

		// Delete the session
		err = sessionManager.DeleteSession()
		if err != nil {
			logger.Error("Failed to delete session: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to delete session: %v", err)), nil
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
			"session_id":    sess.ID,
			"fetched_count": sess.FetchedCount,
			"stored_count":  sess.StoredCount,
			"error_count":   sess.ErrorCount,
		}

		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStopSession: Successfully stopped session %s", sess.ID)
		return respMsg, nil
	}
}

// createGetSessionStatusHandler creates a handler that returns the current session status
func createGetSessionStatusHandler(sessionManager *session.Manager, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCGetSessionStatus: Getting session status")

		// Get current session
		sess, err := sessionManager.GetSession()
		if err != nil {
			if err == session.ErrNoSession {
				// No session exists - return empty status
				respMsg := &subprocess.Message{
					Type:          subprocess.MessageTypeResponse,
					ID:            msg.ID,
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}

				result := map[string]interface{}{
					"has_session": false,
				}

				jsonData, _ := sonic.Marshal(result)
				respMsg.Payload = jsonData

				return respMsg, nil
			}

			logger.Error("Failed to get session: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get session: %v", err)), nil
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
			"session_id":        sess.ID,
			"state":             sess.State,
			"start_index":       sess.StartIndex,
			"results_per_batch": sess.ResultsPerBatch,
			"created_at":        sess.CreatedAt,
			"updated_at":        sess.UpdatedAt,
			"fetched_count":     sess.FetchedCount,
			"stored_count":      sess.StoredCount,
			"error_count":       sess.ErrorCount,
		}

		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Debug("RPCGetSessionStatus: Successfully retrieved session status")
		return respMsg, nil
	}
}

// createPauseJobHandler creates a handler that pauses the running job
func createPauseJobHandler(jobController *job.Controller, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCPauseJob: Pausing job")

		err := jobController.Pause()
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

		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCPauseJob: Successfully paused job")
		return respMsg, nil
	}
}

// createResumeJobHandler creates a handler that resumes a paused job
func createResumeJobHandler(jobController *job.Controller, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("RPCResumeJob: Resuming job")

		err := jobController.Resume(ctx)
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

		jsonData, err := sonic.Marshal(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCResumeJob: Successfully resumed job")
		return respMsg, nil
	}
}
