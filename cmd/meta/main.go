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
)

const (
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

// DataPopulationController manages data population for different data types
type DataPopulationController struct {
	rpcClient *rpc.Client
	logger    *common.Logger
}

// NewDataPopulationController creates a new controller for data population
func NewDataPopulationController(rpcClient *rpc.Client, logger *common.Logger) *DataPopulationController {
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

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	paramsObj := &rpc.ImportParams{Path: path}
	c.logger.Debug("About to invoke RPCImportCWEs on local service")
	resp, err := c.rpcClient.InvokeRPC(ctx, "local", "RPCImportCWEs", paramsObj)
	if err != nil {
		c.logger.Error("Failed to start CWE import: %v", err)
		return "", fmt.Errorf("failed to start CWE import: %w", err)
	}

	if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
		c.logger.Error("CWE import returned error: %s", errMsg)
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

	force, _ := params["force"].(bool)

	c.logger.Info("Starting CAPEC import: session_id=%s, path=%s, force=%t", sessionID, path, force)

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	paramsObj := &rpc.ImportParams{Path: path, Force: force}
	c.logger.Debug("About to invoke RPCImportCAPECs on local service")
	resp, err := c.rpcClient.InvokeRPC(ctx, "local", "RPCImportCAPECs", paramsObj)
	if err != nil {
		c.logger.Error("Failed to start CAPEC import: %v", err)
		return "", fmt.Errorf("failed to start CAPEC import: %w", err)
	}

	if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
		c.logger.Error("CAPEC import returned error: %s", errMsg)
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

	ctx, cancel := context.WithTimeout(ctx, 120*time.Second)
	defer cancel()

	paramsObj := &rpc.ImportParams{Path: path, Force: force}
	c.logger.Debug("About to invoke RPCImportATTACKs on local service")
	resp, err := c.rpcClient.InvokeRPC(ctx, "local", "RPCImportATTACKs", paramsObj)
	if err != nil {
		c.logger.Error("Failed to start ATT&CK import: %v", err)
		return "", fmt.Errorf("failed to start ATT&CK import: %w", err)
	}

	if isErr, errMsg := subprocess.IsErrorResponse(resp); isErr {
		c.logger.Error("ATT&CK import returned error: %s", errMsg)
		return "", fmt.Errorf("ATT&CK import failed: %s", errMsg)
	}

	c.logger.Info("ATT&CK import started successfully: session_id=%s", sessionID)
	return sessionID, nil
}

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
		runDBPath = DefaultSessionDBPath
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
	rpcClient := rpc.NewClient(sp, logger, 60*time.Second)
	rpcAdapter := &RPCClientAdapter{client: rpcClient}
	logger.Info(LogMsgRPCAdapterCreated)

	// Create job executor with Taskflow (100 concurrent goroutines)
	logger.Info(LogMsgJobExecutorCreated, 100)
	jobExecutor := taskflow.NewJobExecutor(rpcAdapter, runStore, logger, 100)

	// Create CWE job controller (separate controller for view jobs)
	logger.Info(LogMsgCWEJobControllerCreated)
	cweJobController := cwejob.NewController(rpcAdapter, logger)

	// Create data population controller for all data types
	logger.Info(LogMsgDataPopControllerCreated)
	dataPopController := NewDataPopulationController(rpcClient, logger)

	// Recover runs if needed after restart
	// This ensures job consistency when the service restarts
	logger.Info(LogMsgRunRecoveryStarted)
	recoverRuns(jobExecutor, logger)
	logger.Info(LogMsgRunRecoveryCompleted)

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

	// Register data population RPC handlers
	sp.RegisterHandler("RPCStartCWEImport", createStartCWEImportHandler(dataPopController, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStartCWEImport")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStartCWEImport")
	sp.RegisterHandler("RPCStartCAPECImport", createStartCAPECImportHandler(dataPopController, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStartCAPECImport")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStartCAPECImport")
	sp.RegisterHandler("RPCStartATTACKImport", createStartATTACKImportHandler(dataPopController, logger))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCStartATTACKImport")
	logger.Debug(LogMsgRPCClientHandlerRegistered, "RPCStartATTACKImport")

	// Register Memory Card proxy handlers
	registerMemoryCardProxyHandlers(sp, rpcClient, logger)

	logger.Info(LogMsgServiceStarted)
	logger.Info(LogMsgServiceReady)

	// --- CWE Import Control ---
	go func() {
		logger.Info("Starting CWE import control routine...")
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		params := &rpc.ImportParams{Path: "assets/cwe-raw.json"}
		logger.Info(LogMsgCWEImportTriggered, params.Path)
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
		logger.Info("Starting CAPEC import control routine...")
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()
		// First check whether local already has CAPEC catalog metadata
		logger.Info("Checking for existing CAPEC catalog metadata...")
		metaResp, err := rpcClient.InvokeRPC(ctx, "local", "RPCGetCAPECCatalogMeta", nil)
		if err != nil {
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
	logger.Info("Starting subprocess with default lifecycle management")
	logger.Debug(LogMsgSubprocessRunStarted)
	subprocess.RunWithDefaults(sp, logger)
	logger.Debug(LogMsgSubprocessRunCompleted)
	logger.Info(LogMsgServiceShutdownStarting)
	logger.Info(LogMsgServiceShutdownComplete)
}

// createStartCWEImportHandler creates a handler for starting CWE import
func createStartCWEImportHandler(controller *DataPopulationController, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgRPCStartCWEImport)

		var req map[string]interface{}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedToParseRequest, err)
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", "failed to parse request"), nil
			}
		}

		sessionID, err := controller.StartDataPopulation(ctx, DataTypeCWE, req)
		if err != nil {
			logger.Error(LogMsgFailedToStartCWEImport, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to start CWE import: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
			"data_type":  string(DataTypeCWE),
		}

		respMsg, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Error(LogMsgFailedToMarshalResponse, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "failed to marshal response"), nil
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
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", "failed to parse request"), nil
			}
		}

		sessionID, err := controller.StartDataPopulation(ctx, DataTypeCAPEC, req)
		if err != nil {
			logger.Warn(LogMsgFailedToStartCAPECImport, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to start CAPEC import: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
			"data_type":  string(DataTypeCAPEC),
		}

		respMsg, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Warn(LogMsgFailedToMarshalResponse, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "failed to marshal response"), nil
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
				return subprocess.NewErrorResponseWithPrefix(msg, "meta", "failed to parse request"), nil
			}
		}

		sessionID, err := controller.StartDataPopulation(ctx, DataTypeATTACK, req)
		if err != nil {
			logger.Warn(LogMsgFailedToStartATTACKImport, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", fmt.Sprintf("failed to start ATT&CK import: %v", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"session_id": sessionID,
			"data_type":  string(DataTypeATTACK),
		}

		respMsg, err := subprocess.NewSuccessResponse(msg, result)
		if err != nil {
			logger.Warn(LogMsgFailedToMarshalResponse, err)
			return subprocess.NewErrorResponseWithPrefix(msg, "meta", "failed to marshal response"), nil
		}

		logger.Info(LogMsgSuccessStartATTACKImport, sessionID)
		return respMsg, nil
	}
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
