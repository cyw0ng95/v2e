/*
Package main implements the meta RPC service.

RPC API Specification:

CVE Meta Service
====================

Service Type: RPC (stdin/stdout message passing)
Description: Orchestrates CVE fetching and storage operations by coordinating between local and remote services.

	Provides high-level CVE management and job control for continuous data synchronization.

Available RPC Methods:
---------------------

CVE Data Operations:

 1. RPCGetCVE
    Description: Retrieves CVE data, checking local storage first, then fetching from remote if not found
    Request Parameters:
    - cve_id (string, required): CVE identifier to retrieve
    Response:
    - cve (object): CVE object with all fields
    - source (string): "local" or "remote" indicating data source
    Errors:
    - Missing CVE ID: cve_id parameter is required
    - Not found: CVE not found in local or remote sources
    - RPC error: Failed to communicate with backend services
    Example:
    Request:  {"cve_id": "CVE-2021-44228"}
    Response: {"cve": {"id": "CVE-2021-44228", ...}, "source": "local"}

 2. RPCCreateCVE
    Description: Creates a new CVE record in local storage
    Request Parameters:
    - cve (object, required): CVE object to create
    Response:
    - success (bool): true if created successfully
    - cve_id (string): ID of the created CVE
    Errors:
    - Missing CVE data: cve parameter is required
    - Already exists: CVE already exists in database
    - Database error: Failed to create record
    Example:
    Request:  {"cve": {"id": "CVE-2024-12345", ...}}
    Response: {"success": true, "cve_id": "CVE-2024-12345"}

 3. RPCUpdateCVE
    Description: Updates an existing CVE record in local storage
    Request Parameters:
    - cve_id (string, required): CVE identifier to update
    - cve (object, required): Updated CVE object
    Response:
    - success (bool): true if updated successfully
    - cve_id (string): ID of the updated CVE
    Errors:
    - Missing parameters: cve_id and cve are required
    - Not found: CVE not found in database
    - Database error: Failed to update record
    Example:
    Request:  {"cve_id": "CVE-2021-44228", "cve": {...}}
    Response: {"success": true, "cve_id": "CVE-2021-44228"}

 4. RPCDeleteCVE
    Description: Deletes a CVE record from local storage
    Request Parameters:
    - cve_id (string, required): CVE identifier to delete
    Response:
    - success (bool): true if deleted successfully
    - cve_id (string): ID of the deleted CVE
    Errors:
    - Missing CVE ID: cve_id parameter is required
    - Not found: CVE not found in database
    - Database error: Failed to delete record
    Example:
    Request:  {"cve_id": "CVE-2021-44228"}
    Response: {"success": true, "cve_id": "CVE-2021-44228"}

 5. RPCListCVEs
    Description: Lists CVE records with pagination
    Request Parameters:
    - offset (int, optional): Starting offset (default: 0)
    - limit (int, optional): Maximum records to return (default: 10)
    Response:
    - cves ([]object): Array of CVE objects
    - offset (int): Starting offset used
    - limit (int): Limit used
    - total (int): Total CVEs in database
    Errors:
    - Database error: Failed to query database
    Example:
    Request:  {"offset": 0, "limit": 10}
    Response: {"cves": [...], "offset": 0, "limit": 10, "total": 150}

 6. RPCCountCVEs
    Description: Gets the total count of CVEs in local storage
    Request Parameters: None
    Response:
    - count (int): Total number of CVE records
    Errors:
    - Database error: Failed to query database
    Example:
    Request:  {}
    Response: {"count": 150}

Job Control Operations:

 7. RPCStartSession
    Description: Starts a new job session for continuous CVE fetching and storing
    Request Parameters:
    - session_id (string, required): Unique identifier for the session
    - start_index (int, optional): Starting index for NVD API pagination (default: 0)
    - results_per_batch (int, optional): CVEs per batch (default: 100)
    Response:
    - success (bool): true if session started
    - session_id (string): Session identifier
    - state (string): Session state ("running")
    - created_at (string): Session creation timestamp
    Errors:
    - Missing session ID: session_id is required
    - Already running: A job session is already active
    - Session error: Failed to create or start session
    Example:
    Request:  {"session_id": "bulk-fetch-2026", "start_index": 0, "results_per_batch": 100}
    Response: {"success": true, "session_id": "bulk-fetch-2026", "state": "running", "created_at": "2026-01-13T02:00:00Z"}

 8. RPCStopSession
    Description: Stops the current job session and cleans up resources
    Request Parameters: None
    Response:
    - success (bool): true if session stopped
    - session_id (string): Session identifier
    - fetched_count (int): Total CVEs fetched
    - stored_count (int): Total CVEs stored
    - error_count (int): Total errors encountered
    Errors:
    - No active session: No job session is running
    Example:
    Request:  {}
    Response: {"success": true, "session_id": "bulk-fetch-2026", "fetched_count": 150, "stored_count": 145, "error_count": 5}

 9. RPCGetSessionStatus
    Description: Gets the status and progress of the current job session
    Request Parameters: None
    Response:
    - has_session (bool): true if a session exists
    - session_id (string): Session identifier
    - state (string): Session state ("idle", "running", "paused")
    - start_index (int): Starting index
    - results_per_batch (int): Batch size
    - created_at (string): Creation timestamp
    - updated_at (string): Last update timestamp
    - fetched_count (int): CVEs fetched so far
    - stored_count (int): CVEs stored so far
    - error_count (int): Errors encountered
    Errors: None (returns has_session=false if no session)
    Example:
    Request:  {}
    Response: {"has_session": true, "session_id": "bulk-fetch-2026", "state": "running", "fetched_count": 150, ...}

 10. RPCPauseJob
    Description: Pauses the running job without deleting the session
    Request Parameters: None
    Response:
    - success (bool): true if job paused
    - state (string): New state ("paused")
    Errors:
    - No active session: No job session is running
    - Not running: Job is not in running state
    Example:
    Request:  {}
    Response: {"success": true, "state": "paused"}

 11. RPCResumeJob
    Description: Resumes a paused job from where it left off
    Request Parameters: None
    Response:
    - success (bool): true if job resumed
    - state (string): New state ("running")
    Errors:
    - No active session: No job session exists
    - Not paused: Job is not in paused state
    Example:
    Request:  {}
    Response: {"success": true, "state": "running"}

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

	// Set up logging using common subprocess framework
	logger, err := subprocess.SetupLogging(processID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	// Get session database path from environment or use default
	sessionDBPath := os.Getenv("SESSION_DB_PATH")
	if sessionDBPath == "" {
		sessionDBPath = DefaultSessionDBPath
	}

	// Create session manager
	sessionManager, err := session.NewManager(sessionDBPath)
	if err != nil {
		logger.Error("Failed to create session manager: %v", err)
		os.Exit(1)
	}
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

	logger.Info("CVE meta service started - orchestrates local and remote")

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

// Handler function implementations moved to cve_handlers.go for maintainability
