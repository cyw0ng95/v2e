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
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

const (
	// DefaultRPCTimeout is the default timeout for RPC requests to other services
	DefaultRPCTimeout = 30 * time.Second
	// DefaultSessionDBPath is the default path for the session database
	DefaultSessionDBPath = "session.db"
)

// DataType represents the type of data being populated
// Using the same DataType from taskflow package

const (
	DataTypeCVE    = taskflow.DataTypeCVE
	DataTypeCWE    = taskflow.DataTypeCWE
	DataTypeCAPEC  = taskflow.DataTypeCAPEC
	DataTypeATTACK = taskflow.DataTypeATTACK
)

// Alias the DataType from the taskflow package so local code can use it directly
type DataType = taskflow.DataType

// RPCClient handles RPC communication with other services through the broker
type RPCClient struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]*requestEntry
	mu              sync.RWMutex
	correlationSeq  uint64
	logger          *common.Logger
}

// DataPopulationController manages data population for different data types
type DataPopulationController struct {
	rpcClient *RPCClient
	logger    *common.Logger
}

// NewDataPopulationController creates a new controller for data population
func NewDataPopulationController(rpcClient *RPCClient, logger *common.Logger) *DataPopulationController {
	return &DataPopulationController{
		rpcClient: rpcClient,
		logger:    logger,
	}
}

// StartDataPopulation starts a data population job for a specific data type
func (c *DataPopulationController) StartDataPopulation(ctx context.Context, dataType DataType, params map[string]interface{}) (string, error) {
	sessionID := fmt.Sprintf("%s-%d", string(dataType), time.Now().Unix())

	switch dataType {
	case DataTypeCWE:
		return c.startCWEImport(ctx, sessionID, params)
	case DataTypeCAPEC:
		return c.startCAPECImport(ctx, sessionID, params)
	case DataTypeATTACK:
		return c.startATTACKImport(ctx, sessionID, params)
	default:
		return "", fmt.Errorf("unsupported data type: %s", dataType)
	}
}

// startCWEImport starts a CWE import job
func (c *DataPopulationController) startCWEImport(ctx context.Context, sessionID string, params map[string]interface{}) (string, error) {
	path, ok := params["path"].(string)
	if !ok {
		path = "assets/cwe-raw.json" // default path
	}

	c.logger.Info("Starting CWE import: session_id=%s, path=%s", sessionID, path)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	paramsObj := &rpc.ImportParams{Path: path}
	resp, err := c.rpcClient.InvokeRPC(ctx, "local", "RPCImportCWEs", paramsObj)
	if err != nil {
		c.logger.Warn("Failed to start CWE import: %v", err)
		return "", fmt.Errorf("failed to start CWE import: %w", err)
	}

	if isErr, errMsg := isErrorResponse(resp); isErr {
		c.logger.Warn("CWE import returned error: %s", errMsg)
		return "", fmt.Errorf("CWE import failed: %s", errMsg)
	}

	c.logger.Info("CWE import started successfully: session_id=%s", sessionID)
	return sessionID, nil
}

// startCAPECImport starts a CAPEC import job
func (c *DataPopulationController) startCAPECImport(ctx context.Context, sessionID string, params map[string]interface{}) (string, error) {
	path, ok := params["path"].(string)
	if !ok {
		path = "assets/capec_contents_latest.xml" // default path
	}

	xsd, ok := params["xsd"].(string)
	if !ok {
		xsd = "assets/capec_schema_latest.xsd" // default xsd
	}

	force, _ := params["force"].(bool)

	c.logger.Info("Starting CAPEC import: session_id=%s, path=%s, xsd=%s, force=%t", sessionID, path, xsd, force)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	paramsObj := &rpc.ImportParams{Path: path, XSD: xsd, Force: force}
	resp, err := c.rpcClient.InvokeRPC(ctx, "local", "RPCImportCAPECs", paramsObj)
	if err != nil {
		c.logger.Warn("Failed to start CAPEC import: %v", err)
		return "", fmt.Errorf("failed to start CAPEC import: %w", err)
	}

	if isErr, errMsg := isErrorResponse(resp); isErr {
		c.logger.Warn("CAPEC import returned error: %s", errMsg)
		return "", fmt.Errorf("CAPEC import failed: %s", errMsg)
	}

	c.logger.Info("CAPEC import started successfully: session_id=%s", sessionID)
	return sessionID, nil
}

// startATTACKImport starts an ATT&CK import job
func (c *DataPopulationController) startATTACKImport(ctx context.Context, sessionID string, params map[string]interface{}) (string, error) {
	path, ok := params["path"].(string)
	if !ok {
		path = "assets/enterprise-attack.xlsx" // default path
	}

	force, _ := params["force"].(bool)

	c.logger.Info("Starting ATT&CK import: session_id=%s, path=%s, force=%t", sessionID, path, force)

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	paramsObj := &rpc.ImportParams{Path: path, Force: force}
	resp, err := c.rpcClient.InvokeRPC(ctx, "local", "RPCImportATTACKs", paramsObj)
	if err != nil {
		c.logger.Warn("Failed to start ATT&CK import: %v", err)
		return "", fmt.Errorf("failed to start ATT&CK import: %w", err)
	}

	if isErr, errMsg := isErrorResponse(resp); isErr {
		c.logger.Warn("ATT&CK import returned error: %s", errMsg)
		return "", fmt.Errorf("ATT&CK import failed: %s", errMsg)
	}

	c.logger.Info("ATT&CK import started successfully: session_id=%s", sessionID)
	return sessionID, nil
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
	// Use a bootstrap logger for initial messages before the full logging system is ready
	bootstrapLogger := common.NewLogger(os.Stderr, "", common.InfoLevel)
	bootstrapLogger.Info(LogMsgStartup, os.Getenv("PROCESS_ID"), os.Getenv("SESSION_DB_PATH"))

	// Set up logging using common subprocess framework
	logger, err := subprocess.SetupLogging(processID)
	if err != nil {
		bootstrapLogger.Error(LogMsgFailedToSetupLogging, err)
		os.Exit(1)
	}

	// Get run database path from environment or use default
	runDBPath := os.Getenv("SESSION_DB_PATH")
	if runDBPath == "" {
		runDBPath = DefaultSessionDBPath
	}
	logger.Info(LogMsgUsingRunDBPath, runDBPath)
	if _, err := os.Stat(runDBPath); err == nil {
		logger.Info(LogMsgRunDBFileExists, runDBPath)
	} else {
		logger.Warn(LogMsgRunDBFileDoesNotExist, runDBPath, err)
	}

	// Create run store
	logger.Info(LogMsgCreatingRunStore)
	runStore, err := taskflow.NewRunStore(runDBPath, logger)
	if err != nil {
		logger.Error(LogMsgFailedToCreateRunStore, err)
		os.Exit(1)
	}
	logger.Info(LogMsgRunStoreCreated)
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

	// Create data population controller for all data types
	dataPopController := NewDataPopulationController(rpcClient, logger)

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
	sp.RegisterHandler("RPCStartTypedSession", createStartTypedSessionHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCStopSession", createStopSessionHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCGetSessionStatus", createGetSessionStatusHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCPauseJob", createPauseJobHandler(jobExecutor, logger))
	sp.RegisterHandler("RPCResumeJob", createResumeJobHandler(jobExecutor, logger))

	// Register CWE view job RPC handlers
	sp.RegisterHandler("RPCStartCWEViewJob", createStartCWEViewJobHandler(cweJobController, logger))
	sp.RegisterHandler("RPCStopCWEViewJob", createStopCWEViewJobHandler(cweJobController, logger))

	// Register data population RPC handlers
	sp.RegisterHandler("RPCStartCWEImport", createStartCWEImportHandler(dataPopController, logger))
	sp.RegisterHandler("RPCStartCAPECImport", createStartCAPECImportHandler(dataPopController, logger))
	sp.RegisterHandler("RPCStartATTACKImport", createStartATTACKImportHandler(dataPopController, logger))

	logger.Info("[meta] CVE meta service started - orchestrates local and remote")

	// --- CWE Import Control ---
	go func() {
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		params := &rpc.ImportParams{Path: "assets/cwe-raw.json"}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCWEs", params)
		if err != nil {
			logger.Warn("Failed to import CWE on local: %v", err)
			logger.Debug("CWE import process failed: %v", err)
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Warn("CWE import error: %s", resp.Error)
			logger.Debug("CWE import process returned error: %s", resp.Error)
		} else {
			logger.Info("CWE import triggered on local")
			logger.Debug("CWE import process started successfully with path: %s", params.Path)
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
			logger.Warn("Failed to query CAPEC catalog meta on local: %v", err)
			// fall back to attempting import
		} else if metaResp.Type == subprocess.MessageTypeResponse {
			logger.Info("CAPEC catalog already present on local; skipping automatic import")
			return
		}
		// If meta not present or query failed, attempt import
		params := &rpc.ImportParams{Path: "assets/capec_contents_latest.xml", XSD: "assets/capec_schema_latest.xsd"}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCAPECs", params)
		if err != nil {
			logger.Warn("Failed to import CAPEC on local: %v", err)
			logger.Debug("CAPEC import process failed: %v", err)
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Warn("CAPEC import error: %s", resp.Error)
			logger.Debug("CAPEC import process returned error: %s", resp.Error)
		} else {
			logger.Info("CAPEC import triggered on local")
			logger.Debug("CAPEC import process started successfully with path: %s", params.Path)
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

// createStartCWEImportHandler creates a handler for starting CWE import
func createStartCWEImportHandler(controller *DataPopulationController, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgRPCStartCWEImport)

		var req map[string]interface{}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedToParseRequest, err)
				return createErrorResponse(msg, "failed to parse request"), nil
			}
		}

		sessionID, err := controller.StartDataPopulation(ctx, DataTypeCWE, req)
		if err != nil {
			logger.Error(LogMsgFailedToStartCWEImport, err)
			return createErrorResponse(msg, fmt.Sprintf("failed to start CWE import: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
			"data_type":  string(DataTypeCWE),
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error(LogMsgFailedToMarshalResponse, err)
			return createErrorResponse(msg, "failed to marshal response"), nil
		}

		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}

		logger.Info(LogMsgSuccessStartCWEImport, sessionID)
		return respMsg, nil
	}
}

// createStartCAPECImportHandler creates a handler for starting CAPEC import
func createStartCAPECImportHandler(controller *DataPopulationController, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgRPCStartCAPECImport)

		var req map[string]interface{}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedToParseRequest, err)
				return createErrorResponse(msg, "failed to parse request"), nil
			}
		}

		sessionID, err := controller.StartDataPopulation(ctx, DataTypeCAPEC, req)
		if err != nil {
			logger.Warn(LogMsgFailedToStartCAPECImport, err)
			return createErrorResponse(msg, fmt.Sprintf("failed to start CAPEC import: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
			"data_type":  string(DataTypeCAPEC),
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Warn(LogMsgFailedToMarshalResponse, err)
			return createErrorResponse(msg, "failed to marshal response"), nil
		}

		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}

		logger.Info(LogMsgSuccessStartCAPECImport, sessionID)
		return respMsg, nil
	}
}

// createStartATTACKImportHandler creates a handler for starting ATT&CK import
func createStartATTACKImportHandler(controller *DataPopulationController, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgRPCStartATTACKImport)

		var req map[string]interface{}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedToParseRequest, err)
				return createErrorResponse(msg, "failed to parse request"), nil
			}
		}

		sessionID, err := controller.StartDataPopulation(ctx, DataTypeATTACK, req)
		if err != nil {
			logger.Warn(LogMsgFailedToStartATTACKImport, err)
			return createErrorResponse(msg, fmt.Sprintf("failed to start ATT&CK import: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
			"data_type":  string(DataTypeATTACK),
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Warn(LogMsgFailedToMarshalResponse, err)
			return createErrorResponse(msg, "failed to marshal response"), nil
		}

		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}

		logger.Info(LogMsgSuccessStartATTACKImport, sessionID)
		return respMsg, nil
	}
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
			logger.Warn("cve_id is required but was empty or missing")
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
			logger.Warn("Failed to parse check response: %v", err)
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
				logger.Warn("Failed to parse local CVE data: %v", err)
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
				logger.Warn("Failed to parse remote CVE response: %v", err)
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
			logger.Warn("Failed to marshal response: %v", err)
			logger.Debug("GetCVE failed to marshal response for CVE ID %s: %v", req.CVEID, err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCGetCVE: Successfully retrieved CVE %s", req.CVEID)
		logger.Debug("GetCVE request completed successfully for CVE ID %s", req.CVEID)
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
			logger.Warn("Failed to get active run: %v", err)
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

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Warn("Failed to marshal response: %v", err)
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
			logger.Warn("Failed to get active run: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Warn("No active run to pause")
			return createErrorResponse(msg, "no active run"), nil
		}

		err = jobExecutor.Pause(run.ID)
		if err != nil {
			logger.Warn("Failed to pause job: %v", err)
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
			logger.Warn("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCPauseJob: Successfully paused job")
		return respMsg, nil
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
			return createErrorResponse(msg, fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Warn("No active run to resume")
			return createErrorResponse(msg, "no active run"), nil
		}

		err = jobExecutor.Resume(ctx, run.ID)
		if err != nil {
			logger.Warn("Failed to resume job: %v", err)
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
			logger.Warn("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCResumeJob: Successfully resumed job")
		return respMsg, nil
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
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		logger.Info("RPCStartSession: Starting new job session with data type %s", req.DataType)

		if req.DataType == "" {
			logger.Warn("data_type is required but was empty or missing")
			return createErrorResponse(msg, "data_type is required"), nil
		}

		if req.StartIndex < 0 {
			logger.Warn("start_index must be non-negative")
			return createErrorResponse(msg, "start_index must be non-negative"), nil
		}

		if req.ResultsPerBatch <= 0 {
			logger.Warn("results_per_batch must be positive")
			return createErrorResponse(msg, "results_per_batch must be positive"), nil
		}

		sessionID := fmt.Sprintf("%s-%d", req.DataType, time.Now().Unix())

		err := jobExecutor.StartTyped(ctx, sessionID, req.StartIndex, req.ResultsPerBatch, req.DataType)
		if err != nil {
			logger.Warn("Failed to start job session: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to start job session: %v", err)), nil
		}

		// Get updated run state
		run, err := jobExecutor.GetStatus(sessionID)
		if err != nil {
			logger.Warn("Failed to get run status: %v", err)
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
			"success":     true,
			"session_id":  run.ID,
			"data_type":   string(run.DataType),
			"state":       run.State,
			"created_at":  run.CreatedAt,
			"start_index": run.StartIndex,
			"batch_size":  run.ResultsPerBatch,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Warn("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStartSession: Successfully started job session: %s", run.ID)
		return respMsg, nil
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
			return createErrorResponse(msg, fmt.Sprintf("failed to get active run: %v", err)), nil
		}

		if run == nil {
			logger.Warn("No active run to stop")
			return createErrorResponse(msg, "no active run"), nil
		}

		err = jobExecutor.Stop(run.ID)
		if err != nil {
			logger.Warn("Failed to stop job session: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to stop job session: %v", err)), nil
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
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Warn("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStopSession: Successfully stopped job session")
		return respMsg, nil
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
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		// Start the CWE view job
		sessionID, err := controller.Start(ctx, req.Params)
		if err != nil {
			logger.Warn("Failed to start CWE view job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to start CWE view job: %v", err)), nil
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
			"session_id": sessionID,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Warn("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStartCWEViewJob: Successfully started CWE view job: %s", sessionID)
		return respMsg, nil
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
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.SessionID == "" {
			logger.Error("session_id is required but was empty or missing")
			return createErrorResponse(msg, "session_id is required"), nil
		}

		// Stop the CWE view job
		err := controller.Stop(ctx, req.SessionID)
		if err != nil {
			logger.Error("Failed to stop CWE view job: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to stop CWE view job: %v", err)), nil
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
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStopCWEViewJob: Successfully stopped CWE view job: %s", req.SessionID)
		return respMsg, nil
	}
}

// createCreateCVEHandler creates a handler that creates a new CVE
func createCreateCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req cve.CVEItem
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		// Create the CVE
		// Validate required fields before making RPC calls
		if req.ID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCCreateCVE", &req)
		if err != nil {
			logger.Warn("Failed to create CVE: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to create CVE: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(resp); isErr {
			logger.Warn("Error creating CVE: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to create CVE: %s", errMsg)), nil
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
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCCreateCVE: Successfully created CVE")
		return respMsg, nil
	}
}

// createUpdateCVEHandler creates a handler that updates an existing CVE
func createUpdateCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req cve.CVEItem
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		// Update the CVE
		// Validate required fields before making RPC calls
		if req.ID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCUpdateCVE", &req)
		if err != nil {
			logger.Warn("Failed to update CVE: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to update CVE: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(resp); isErr {
			logger.Warn("Error updating CVE: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to update CVE: %s", errMsg)), nil
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
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCUpdateCVE: Successfully updated CVE")
		return respMsg, nil
	}
}

// createDeleteCVEHandler creates a handler that deletes an existing CVE
func createDeleteCVEHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			CVEID string `json:"cve_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.CVEID == "" {
			logger.Error("cve_id is required but was empty or missing")
			return createErrorResponse(msg, "cve_id is required"), nil
		}

		// Delete the CVE
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCDeleteCVE", &rpc.CVEIDParams{CVEID: req.CVEID})
		if err != nil {
			logger.Warn("Failed to delete CVE: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to delete CVE: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(resp); isErr {
			logger.Warn("Error deleting CVE: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to delete CVE: %s", errMsg)), nil
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
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCDeleteCVE: Successfully deleted CVE")
		return respMsg, nil
	}
}

// createListCVEsHandler creates a handler that lists CVEs
func createListCVEsHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Parse the request payload
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.Offset < 0 {
			logger.Error("offset must be non-negative")
			return createErrorResponse(msg, "offset must be non-negative"), nil
		}

		if req.Limit <= 0 {
			logger.Error("limit must be positive")
			return createErrorResponse(msg, "limit must be positive"), nil
		}

		// List CVEs
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCListCVEs", &req)
		if err != nil {
			logger.Warn("Failed to list CVEs: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to list CVEs: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(resp); isErr {
			logger.Warn("Error listing CVEs: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to list CVEs: %s", errMsg)), nil
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		jsonData, err := subprocess.MarshalFast(resp.Payload)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCListCVEs: Successfully listed CVEs")
		return respMsg, nil
	}
}

// createCountCVEsHandler creates a handler that counts CVEs
func createCountCVEsHandler(rpcClient *RPCClient, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		// Count CVEs
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCCountCVEs", nil)
		if err != nil {
			logger.Warn("Failed to count CVEs: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to count CVEs: %v", err)), nil
		}

		// Check if the response is an error
		if isErr, errMsg := isErrorResponse(resp); isErr {
			logger.Warn("Error counting CVEs: %s", errMsg)
			return createErrorResponse(msg, fmt.Sprintf("failed to count CVEs: %s", errMsg)), nil
		}

		// Create response
		respMsg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}

		jsonData, err := subprocess.MarshalFast(resp.Payload)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCCountCVEs: Successfully counted CVEs")
		return respMsg, nil
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

		// Set defaults
		req.StartIndex = 0
		req.ResultsPerBatch = 100
		req.DataType = taskflow.DataTypeCVE // default to CVE

		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Error("Failed to parse request: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to parse request: %v", err)), nil
		}

		if req.SessionID == "" {
			logger.Error("session_id is required")
			return createErrorResponse(msg, "session_id is required"), nil
		}

		logger.Info("RPCStartTypedSession: Starting job run %s (data_type=%s, start_index=%d, batch_size=%d)",
			req.SessionID, req.DataType, req.StartIndex, req.ResultsPerBatch)

		// Start the job with the specified data type
		err := jobExecutor.StartTyped(ctx, req.SessionID, req.StartIndex, req.ResultsPerBatch, req.DataType)
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
			"success":     true,
			"session_id":  run.ID,
			"state":       run.State,
			"data_type":   run.DataType,
			"created_at":  run.CreatedAt,
			"start_index": run.StartIndex,
			"batch_size":  run.ResultsPerBatch,
			"params":      req.Params,
		}

		jsonData, err := subprocess.MarshalFast(result)
		if err != nil {
			logger.Error("Failed to marshal response: %v", err)
			return createErrorResponse(msg, fmt.Sprintf("failed to marshal response: %v", err)), nil
		}
		respMsg.Payload = jsonData

		logger.Info("RPCStartTypedSession: Successfully started job run %s (type: %s)", run.ID, run.DataType)
		return respMsg, nil
	}
}
