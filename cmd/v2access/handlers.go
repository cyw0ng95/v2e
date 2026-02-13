package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/gin-gonic/gin"
)

// contextPool pools unused context objects for reuse in RPC handlers.
// This reduces allocations for the frequently created timeout contexts.
// Each pooled context includes its cancel function for proper cleanup.
var contextPool = sync.Pool{
	New: func() interface{} {
		// Return a new pooledContext struct - will be initialized on Get()
		return &pooledContext{}
	},
}

// pooledContext holds a context with its cancel function for pooling
type pooledContext struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// getPooledContext retrieves or creates a new timeout context from pool. Returns the context, its cancel function, and the pooled wrapper for cleanup.
func getPooledContext(timeout time.Duration) (ctx context.Context, cancel context.CancelFunc, pc *pooledContext) {
	// Get a pooled context object
	pc = contextPool.Get().(*pooledContext)

	// Create a new context with timeout
	// Note: We cannot reuse the context itself because canceled contexts
	// are not safe to reuse. However, we reuse the pooledContext struct
	// to reduce allocations of the wrapper struct.
	pc.ctx, pc.cancel = context.WithTimeout(context.Background(), timeout)
	return pc.ctx, pc.cancel, pc
}

// putPooledContext returns the pooled wrapper to the pool for reuse
func putPooledContext(pc *pooledContext) {
	if pc != nil {
		pc.ctx = nil
		pc.cancel = nil
		contextPool.Put(pc)
	}
}

// createPathRPCHandler creates a Gin handler for path-based RPC endpoints
func createPathRPCHandler(rpcClient *RPCClient, method string, target string) gin.HandlerFunc {
	return func(c *gin.Context) {
		common.Debug("Path-based RPC: %s %s", c.Request.Method, c.Request.URL.Path)

		var params map[string]interface{}
		if err := c.ShouldBindJSON(&params); err != nil {
			if err.Error() != "EOF" {
				common.Warn("Failed to parse params for %s: %v", method, err)
			}
			params = make(map[string]interface{})
		}

		rpcCtx, cancel, pc := getPooledContext(rpcClient.rpcTimeout)
		defer cancel()
		defer putPooledContext(pc)

		response, err := rpcClient.InvokeRPCWithTarget(rpcCtx, target, method, params)
		if err != nil {
			common.Error("Path RPC error for %s: %v", method, err)
			httpErrorResponse(c, http.StatusOK, fmt.Sprintf("RPC error: %v", err))
			return
		}

		if isError, errMsg := subprocess.IsErrorResponse(response); isError {
			common.Warn("Path RPC response is error for %s: %s", method, errMsg)
			httpErrorResponse(c, http.StatusOK, errMsg)
			return
		}

		var payload interface{}
		if response.Payload != nil {
			if err := subprocess.UnmarshalFast(response.Payload, &payload); err != nil {
				common.Error("Failed to parse response for %s: %v", method, err)
				httpErrorResponse(c, http.StatusOK, fmt.Sprintf("Failed to parse response: %v", err))
				return
			}
		}

		httpSuccessResponse(c, payload)
	}
}

// HTTP response helpers for reducing boilerplate in handlers

// httpErrorResponse sends an error response with given code and message.
func httpErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, gin.H{
		"retcode": code,
		"message": message,
		"payload": nil,
	})
}

// httpSuccessResponse sends a success response with given payload.
func httpSuccessResponse(c *gin.Context, payload interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"retcode": 0,
		"message": "success",
		"payload": payload,
	})
}

// registerHandlers registers REST endpoints on provided router group
func registerHandlers(restful *gin.RouterGroup, rpcClient *RPCClient) {
	// Health check endpoint
	restful.GET("/health", func(c *gin.Context) {
		common.Debug(LogMsgHealthCheckReceived)
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// Path-based RPC endpoints: /restful/rpc/{resource}/{action}
	// Example: POST /restful/rpc/cve/get
	rpcGroup := restful.Group("/rpc")
	{
		// CVE endpoints
		rpcGroup.POST("/cve/get", createPathRPCHandler(rpcClient, "RPCGetCVE", "local"))
		rpcGroup.POST("/cve/create", createPathRPCHandler(rpcClient, "RPCCreateCVE", "local"))
		rpcGroup.POST("/cve/update", createPathRPCHandler(rpcClient, "RPCUpdateCVE", "local"))
		rpcGroup.POST("/cve/delete", createPathRPCHandler(rpcClient, "RPCDeleteCVE", "local"))
		rpcGroup.POST("/cve/list", createPathRPCHandler(rpcClient, "RPCListCVEs", "local"))
		rpcGroup.POST("/cve/count", createPathRPCHandler(rpcClient, "RPCCountCVEs", "local"))

		// CWE endpoints
		rpcGroup.POST("/cwe/get", createPathRPCHandler(rpcClient, "RPCGetCWEByID", "local"))
		rpcGroup.POST("/cwe/list", createPathRPCHandler(rpcClient, "RPCListCWEs", "local"))
		rpcGroup.POST("/cwe/import", createPathRPCHandler(rpcClient, "RPCImportCWEs", "local"))

		// CWE View endpoints
		rpcGroup.POST("/cwe-view/save", createPathRPCHandler(rpcClient, "RPCSaveCWEView", "local"))
		rpcGroup.POST("/cwe-view/get", createPathRPCHandler(rpcClient, "RPCGetCWEViewByID", "local"))
		rpcGroup.POST("/cwe-view/list", createPathRPCHandler(rpcClient, "RPCListCWEViews", "local"))
		rpcGroup.POST("/cwe-view/delete", createPathRPCHandler(rpcClient, "RPCDeleteCWEView", "local"))
		rpcGroup.POST("/cwe-view/start-job", createPathRPCHandler(rpcClient, "RPCStartCWEViewJob", "meta"))
		rpcGroup.POST("/cwe-view/stop-job", createPathRPCHandler(rpcClient, "RPCStopCWEViewJob", "meta"))

		// CAPEC endpoints
		rpcGroup.POST("/capec/get", createPathRPCHandler(rpcClient, "RPCGetCAPECByID", "local"))
		rpcGroup.POST("/capec/list", createPathRPCHandler(rpcClient, "RPCListCAPECs", "local"))
		rpcGroup.POST("/capec/import", createPathRPCHandler(rpcClient, "RPCImportCAPECs", "local"))
		rpcGroup.POST("/capec/force-import", createPathRPCHandler(rpcClient, "RPCForceImportCAPECs", "local"))
		rpcGroup.POST("/capec/metadata", createPathRPCHandler(rpcClient, "RPCGetCAPECCatalogMeta", "local"))

		// ATT&CK endpoints
		rpcGroup.POST("/attack/technique", createPathRPCHandler(rpcClient, "RPCGetAttackTechnique", "local"))
		rpcGroup.POST("/attack/tactic", createPathRPCHandler(rpcClient, "RPCGetAttackTactic", "local"))
		rpcGroup.POST("/attack/mitigation", createPathRPCHandler(rpcClient, "RPCGetAttackMitigation", "local"))
		rpcGroup.POST("/attack/software", createPathRPCHandler(rpcClient, "RPCGetAttackSoftware", "local"))
		rpcGroup.POST("/attack/group", createPathRPCHandler(rpcClient, "RPCGetAttackGroup", "local"))
		rpcGroup.POST("/attack/technique-by-id", createPathRPCHandler(rpcClient, "RPCGetAttackTechniqueByID", "local"))
		rpcGroup.POST("/attack/tactic-by-id", createPathRPCHandler(rpcClient, "RPCGetAttackTacticByID", "local"))
		rpcGroup.POST("/attack/mitigation-by-id", createPathRPCHandler(rpcClient, "RPCGetAttackMitigationByID", "local"))
		rpcGroup.POST("/attack/software-by-id", createPathRPCHandler(rpcClient, "RPCGetAttackSoftwareByID", "local"))
		rpcGroup.POST("/attack/group-by-id", createPathRPCHandler(rpcClient, "RPCGetAttackGroupByID", "local"))
		rpcGroup.POST("/attack/techniques", createPathRPCHandler(rpcClient, "RPCListAttackTechniques", "local"))
		rpcGroup.POST("/attack/tactics", createPathRPCHandler(rpcClient, "RPCListAttackTactics", "local"))
		rpcGroup.POST("/attack/mitigations", createPathRPCHandler(rpcClient, "RPCListAttackMitigations", "local"))
		rpcGroup.POST("/attack/softwares", createPathRPCHandler(rpcClient, "RPCListAttackSoftware", "local"))
		rpcGroup.POST("/attack/groups", createPathRPCHandler(rpcClient, "RPCListAttackGroups", "local"))
		rpcGroup.POST("/attack/import", createPathRPCHandler(rpcClient, "RPCImportATTACKs", "local"))
		rpcGroup.POST("/attack/import-metadata", createPathRPCHandler(rpcClient, "RPCGetAttackImportMetadata", "local"))

		// ASVS endpoints
		rpcGroup.POST("/asvs/list", createPathRPCHandler(rpcClient, "RPCListASVS", "local"))
		rpcGroup.POST("/asvs/get", createPathRPCHandler(rpcClient, "RPCGetASVSByID", "local"))
		rpcGroup.POST("/asvs/import", createPathRPCHandler(rpcClient, "RPCImportASVS", "local"))

		// CCE endpoints
		rpcGroup.POST("/cce/get", createPathRPCHandler(rpcClient, "RPCGetCCEByID", "local"))
		rpcGroup.POST("/cce/list", createPathRPCHandler(rpcClient, "RPCListCCEs", "local"))
		rpcGroup.POST("/cce/import", createPathRPCHandler(rpcClient, "RPCImportCCEs", "local"))
		rpcGroup.POST("/cce/import-one", createPathRPCHandler(rpcClient, "RPCImportCCE", "local"))
		rpcGroup.POST("/cce/count", createPathRPCHandler(rpcClient, "RPCCountCCEs", "local"))
		rpcGroup.POST("/cce/delete", createPathRPCHandler(rpcClient, "RPCDeleteCCE", "local"))
		rpcGroup.POST("/cce/update", createPathRPCHandler(rpcClient, "RPCUpdateCCE", "local"))

		// Session/Job endpoints
		rpcGroup.POST("/session/start", createPathRPCHandler(rpcClient, "RPCStartSession", "meta"))
		rpcGroup.POST("/session/start-typed", createPathRPCHandler(rpcClient, "RPCStartTypedSession", "meta"))
		rpcGroup.POST("/session/stop", createPathRPCHandler(rpcClient, "RPCStopSession", "meta"))
		rpcGroup.POST("/session/status", createPathRPCHandler(rpcClient, "RPCGetSessionStatus", "meta"))
		rpcGroup.POST("/job/pause", createPathRPCHandler(rpcClient, "RPCPauseJob", "meta"))
		rpcGroup.POST("/job/resume", createPathRPCHandler(rpcClient, "RPCResumeJob", "meta"))

		// Bookmark endpoints
		rpcGroup.POST("/bookmark/create", createPathRPCHandler(rpcClient, "RPCCreateBookmark", "local"))
		rpcGroup.POST("/bookmark/get", createPathRPCHandler(rpcClient, "RPCGetBookmark", "local"))
		rpcGroup.POST("/bookmark/update", createPathRPCHandler(rpcClient, "RPCUpdateBookmark", "local"))
		rpcGroup.POST("/bookmark/delete", createPathRPCHandler(rpcClient, "RPCDeleteBookmark", "local"))
		rpcGroup.POST("/bookmark/list", createPathRPCHandler(rpcClient, "RPCListBookmarks", "local"))

		// Note endpoints
		rpcGroup.POST("/note/add", createPathRPCHandler(rpcClient, "RPCAddNote", "local"))
		rpcGroup.POST("/note/get", createPathRPCHandler(rpcClient, "RPCGetNote", "local"))
		rpcGroup.POST("/note/update", createPathRPCHandler(rpcClient, "RPCUpdateNote", "local"))
		rpcGroup.POST("/note/delete", createPathRPCHandler(rpcClient, "RPCDeleteNote", "local"))
		rpcGroup.POST("/note/by-bookmark", createPathRPCHandler(rpcClient, "RPCGetNotesByBookmark", "local"))

		// Memory Card endpoints
		rpcGroup.POST("/memory-card/create", createPathRPCHandler(rpcClient, "RPCCreateMemoryCard", "local"))
		rpcGroup.POST("/memory-card/get", createPathRPCHandler(rpcClient, "RPCGetMemoryCard", "local"))
		rpcGroup.POST("/memory-card/update", createPathRPCHandler(rpcClient, "RPCUpdateMemoryCard", "local"))
		rpcGroup.POST("/memory-card/delete", createPathRPCHandler(rpcClient, "RPCDeleteMemoryCard", "local"))
		rpcGroup.POST("/memory-card/list", createPathRPCHandler(rpcClient, "RPCListMemoryCards", "local"))
		rpcGroup.POST("/memory-card/rate", createPathRPCHandler(rpcClient, "RPCRateMemoryCard", "local"))

		// GLC endpoints
		rpcGroup.POST("/glc/graph/create", createPathRPCHandler(rpcClient, "RPCGLCGraphCreate", "local"))
		rpcGroup.POST("/glc/graph/get", createPathRPCHandler(rpcClient, "RPCGLCGraphGet", "local"))
		rpcGroup.POST("/glc/graph/update", createPathRPCHandler(rpcClient, "RPCGLCGraphUpdate", "local"))
		rpcGroup.POST("/glc/graph/delete", createPathRPCHandler(rpcClient, "RPCGLCGraphDelete", "local"))
		rpcGroup.POST("/glc/graph/list", createPathRPCHandler(rpcClient, "RPCGLCGraphList", "local"))
		rpcGroup.POST("/glc/graph/list-recent", createPathRPCHandler(rpcClient, "RPCGLCGraphListRecent", "local"))
		rpcGroup.POST("/glc/version/get", createPathRPCHandler(rpcClient, "RPCGLCVersionGet", "local"))
		rpcGroup.POST("/glc/version/list", createPathRPCHandler(rpcClient, "RPCGLCVersionList", "local"))
		rpcGroup.POST("/glc/version/restore", createPathRPCHandler(rpcClient, "RPCGLCVersionRestore", "local"))
		rpcGroup.POST("/glc/preset/create", createPathRPCHandler(rpcClient, "RPCGLCPresetCreate", "local"))
		rpcGroup.POST("/glc/preset/get", createPathRPCHandler(rpcClient, "RPCGLCPresetGet", "local"))
		rpcGroup.POST("/glc/preset/update", createPathRPCHandler(rpcClient, "RPCGLCPresetUpdate", "local"))
		rpcGroup.POST("/glc/preset/delete", createPathRPCHandler(rpcClient, "RPCGLCPresetDelete", "local"))
		rpcGroup.POST("/glc/preset/list", createPathRPCHandler(rpcClient, "RPCGLCPresetList", "local"))
		rpcGroup.POST("/glc/share/create", createPathRPCHandler(rpcClient, "RPCGLCShareCreateLink", "local"))
		rpcGroup.POST("/glc/share/get", createPathRPCHandler(rpcClient, "RPCGLCShareGetShared", "local"))
		rpcGroup.POST("/glc/share/embed", createPathRPCHandler(rpcClient, "RPCGLCShareGetEmbedData", "local"))

		// Analysis endpoints
		rpcGroup.POST("/analysis/stats", createPathRPCHandler(rpcClient, "RPCGetGraphStats", "analysis"))
		rpcGroup.POST("/analysis/node/add", createPathRPCHandler(rpcClient, "RPCAddNode", "analysis"))
		rpcGroup.POST("/analysis/edge/add", createPathRPCHandler(rpcClient, "RPCAddEdge", "analysis"))
		rpcGroup.POST("/analysis/node/get", createPathRPCHandler(rpcClient, "RPCGetNode", "analysis"))
		rpcGroup.POST("/analysis/neighbors", createPathRPCHandler(rpcClient, "RPCGetNeighbors", "analysis"))
		rpcGroup.POST("/analysis/path/find", createPathRPCHandler(rpcClient, "RPCFindPath", "analysis"))
		rpcGroup.POST("/analysis/nodes/by-type", createPathRPCHandler(rpcClient, "RPCGetNodesByType", "analysis"))
		rpcGroup.POST("/analysis/status", createPathRPCHandler(rpcClient, "RPCGetUEEStatus", "analysis"))
		rpcGroup.POST("/analysis/graph/build", createPathRPCHandler(rpcClient, "RPCBuildCVEGraph", "analysis"))
		rpcGroup.POST("/analysis/graph/clear", createPathRPCHandler(rpcClient, "RPCClearGraph", "analysis"))
		rpcGroup.POST("/analysis/fsm/state", createPathRPCHandler(rpcClient, "RPCGetFSMState", "analysis"))
		rpcGroup.POST("/analysis/fsm/pause", createPathRPCHandler(rpcClient, "RPCPauseAnalysis", "analysis"))
		rpcGroup.POST("/analysis/fsm/resume", createPathRPCHandler(rpcClient, "RPCResumeAnalysis", "analysis"))
		rpcGroup.POST("/analysis/graph/save", createPathRPCHandler(rpcClient, "RPCSaveGraph", "analysis"))
		rpcGroup.POST("/analysis/graph/load", createPathRPCHandler(rpcClient, "RPCLoadGraph", "analysis"))

		// System endpoints
		rpcGroup.POST("/system/metrics", createPathRPCHandler(rpcClient, "RPCGetSysMetrics", "sysmon"))

		// ETL endpoints
		rpcGroup.POST("/etl/tree", createPathRPCHandler(rpcClient, "RPCGetEtlTree", "meta"))
		rpcGroup.POST("/etl/provider/start", createPathRPCHandler(rpcClient, "RPCStartProvider", "meta"))
		rpcGroup.POST("/etl/provider/pause", createPathRPCHandler(rpcClient, "RPCPauseProvider", "meta"))
		rpcGroup.POST("/etl/provider/stop", createPathRPCHandler(rpcClient, "RPCStopProvider", "meta"))
		rpcGroup.POST("/etl/performance-policy", createPathRPCHandler(rpcClient, "RPCUpdatePerformancePolicy", "meta"))
		rpcGroup.POST("/etl/kernel-metrics", createPathRPCHandler(rpcClient, "RPCGetKernelMetrics", "meta"))

		// SSG endpoints
		rpcGroup.POST("/ssg/import-guide", createPathRPCHandler(rpcClient, "RPCSSGImportGuide", "local"))
		rpcGroup.POST("/ssg/import-table", createPathRPCHandler(rpcClient, "RPCSSGImportTable", "local"))
		rpcGroup.POST("/ssg/guide", createPathRPCHandler(rpcClient, "RPCSSGGetGuide", "local"))
		rpcGroup.POST("/ssg/guides", createPathRPCHandler(rpcClient, "RPCSSGListGuides", "local"))
		rpcGroup.POST("/ssg/tables", createPathRPCHandler(rpcClient, "RPCSSGListTables", "local"))
		rpcGroup.POST("/ssg/table", createPathRPCHandler(rpcClient, "RPCSSGGetTable", "local"))
		rpcGroup.POST("/ssg/table-entries", createPathRPCHandler(rpcClient, "RPCSSGGetTableEntries", "local"))
		rpcGroup.POST("/ssg/tree", createPathRPCHandler(rpcClient, "RPCSSGGetTree", "local"))
		rpcGroup.POST("/ssg/tree-node", createPathRPCHandler(rpcClient, "RPCSSGGetTreeNode", "local"))
		rpcGroup.POST("/ssg/group", createPathRPCHandler(rpcClient, "RPCSSGGetGroup", "local"))
		rpcGroup.POST("/ssg/child-groups", createPathRPCHandler(rpcClient, "RPCSSGGetChildGroups", "local"))
		rpcGroup.POST("/ssg/rule", createPathRPCHandler(rpcClient, "RPCSSGGetRule", "local"))
		rpcGroup.POST("/ssg/rules", createPathRPCHandler(rpcClient, "RPCSSGListRules", "local"))
		rpcGroup.POST("/ssg/child-rules", createPathRPCHandler(rpcClient, "RPCSSGGetChildRules", "local"))
		rpcGroup.POST("/ssg/import-manifest", createPathRPCHandler(rpcClient, "RPCSSGImportManifest", "local"))
		rpcGroup.POST("/ssg/manifests", createPathRPCHandler(rpcClient, "RPCSSGListManifests", "local"))
		rpcGroup.POST("/ssg/manifest", createPathRPCHandler(rpcClient, "RPCSSGGetManifest", "local"))
		rpcGroup.POST("/ssg/profiles", createPathRPCHandler(rpcClient, "RPCSSGListProfiles", "local"))
		rpcGroup.POST("/ssg/profile", createPathRPCHandler(rpcClient, "RPCSSGGetProfile", "local"))
		rpcGroup.POST("/ssg/profile-rules", createPathRPCHandler(rpcClient, "RPCSSGGetProfileRules", "local"))
		rpcGroup.POST("/ssg/import-datastream", createPathRPCHandler(rpcClient, "RPCSSGImportDataStream", "local"))
		rpcGroup.POST("/ssg/datastreams", createPathRPCHandler(rpcClient, "RPCSSGListDataStreams", "local"))
		rpcGroup.POST("/ssg/datastream", createPathRPCHandler(rpcClient, "RPCSSGGetDataStream", "local"))
		rpcGroup.POST("/ssg/ds-profiles", createPathRPCHandler(rpcClient, "RPCSSGListDSProfiles", "local"))
		rpcGroup.POST("/ssg/ds-profile", createPathRPCHandler(rpcClient, "RPCSSGGetDSProfile", "local"))
		rpcGroup.POST("/ssg/ds-profile-rules", createPathRPCHandler(rpcClient, "RPCSSGGetDSProfileRules", "local"))
		rpcGroup.POST("/ssg/ds-groups", createPathRPCHandler(rpcClient, "RPCSSGListDSGroups", "local"))
		rpcGroup.POST("/ssg/ds-rules", createPathRPCHandler(rpcClient, "RPCSSGListDSRules", "local"))
		rpcGroup.POST("/ssg/ds-rule", createPathRPCHandler(rpcClient, "RPCSSGGetDSRule", "local"))
		rpcGroup.POST("/ssg/cross-references", createPathRPCHandler(rpcClient, "RPCSSGGetCrossReferences", "local"))
		rpcGroup.POST("/ssg/find-related", createPathRPCHandler(rpcClient, "RPCSSGFindRelatedObjects", "local"))
		rpcGroup.POST("/ssg/job/start", createPathRPCHandler(rpcClient, "RPCSSGStartImportJob", "meta"))
		rpcGroup.POST("/ssg/job/stop", createPathRPCHandler(rpcClient, "RPCSSGStopImportJob", "meta"))
		rpcGroup.POST("/ssg/job/pause", createPathRPCHandler(rpcClient, "RPCSSGPauseImportJob", "meta"))
		rpcGroup.POST("/ssg/job/resume", createPathRPCHandler(rpcClient, "RPCSSGResumeImportJob", "meta"))
		rpcGroup.POST("/ssg/job/status", createPathRPCHandler(rpcClient, "RPCSSGGetImportStatus", "meta"))
	}

	// Generic RPC forwarding endpoint (backward compatible)
	restful.POST("/rpc", func(c *gin.Context) {
		common.Debug(LogMsgHTTPRequestReceived, c.Request.Method, c.Request.URL.Path)

		// Parse request body
		var request struct {
			Method string                 `json:"method" binding:"required"`
			Params map[string]interface{} `json:"params"`
			Target string                 `json:"target"` // Optional target process (defaults to "broker")
		}

		if err := c.ShouldBindJSON(&request); err != nil {
			common.Warn(LogMsgRequestParsingError, err)
			httpErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
			common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, http.StatusBadRequest)
			return
		}

		// Default target to broker if not specified
		target := request.Target
		if target == "" {
			target = "broker"
		}

		common.Info(LogMsgRPCForwardingStarted, request.Method, target)
		if request.Params != nil {
			common.Debug(LogMsgRPCForwardingParams, request.Params)
		}

		// Forward RPC request to target process (use configured rpc timeout)
		requestCtx := c.Request.Context()
		common.Info(LogMsgRPCInvokeStarted, target, request.Method)
		common.Debug("RPC request context value: %v", requestCtx)

		// Check if context is already done before making RPC call
		select {
		case <-requestCtx.Done():
			err := requestCtx.Err()
			common.Error("HTTP request context already canceled before RPC call: %v", err)
			// Use appropriate status code for context cancellation
			statusCode := http.StatusRequestTimeout
			if err == context.Canceled {
				statusCode = http.StatusServiceUnavailable
			}
			httpErrorResponse(c, statusCode, fmt.Sprintf("Request context canceled: %v", err))
			return
		default:
			// Context is not done, proceed with RPC
		}

		// Create a separate context for RPC call to avoid cancellation from HTTP context
		// This prevents RPC call from being canceled when the HTTP client disconnects
		// Use sync.Pool to reduce allocations for frequently created timeout contexts
		rpcCtx, cancel, pc := getPooledContext(rpcClient.rpcTimeout)
		defer cancel()
		defer putPooledContext(pc)

		response, err := rpcClient.InvokeRPCWithTarget(rpcCtx, target, request.Method, request.Params)
		common.Debug(LogMsgRPCInvokeCompleted, target, request.Method)

		// Log context state after RPC call completes
		select {
		case <-requestCtx.Done():
			err := requestCtx.Err()
			common.Warn("HTTP request context canceled after RPC call: %v", err)
		default:
			// Context is still active
		}

		if err != nil {
			common.Error(LogMsgRPCForwardingError, err)
			httpErrorResponse(c, http.StatusOK, fmt.Sprintf("RPC error: %v", err))
			common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, 200)
			return
		}

		// Check response type using subprocess helper
		if isError, errMsg := subprocess.IsErrorResponse(response); isError {
			common.Warn("RPC response is an error: %s", errMsg)
			httpErrorResponse(c, http.StatusOK, errMsg)
			common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, 200)
			return
		}

		// Parse payload
		var payload interface{}
		if response.Payload != nil {
			common.Debug(LogMsgRPCResponseParsing)
			if err := subprocess.UnmarshalFast(response.Payload, &payload); err != nil {
				common.Error(LogMsgRPCResponseParseError, err)
				httpErrorResponse(c, http.StatusOK, fmt.Sprintf("Failed to parse response: %v", err))
				common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, 200)
				return
			}
			common.Debug(LogMsgRPCResponseParsed)
		}

		// Return success response
		httpSuccessResponse(c, payload)
		common.Info(LogMsgRPCForwardingComplete, request.Method, target)
		common.Debug(LogMsgHTTPRequestProcessed, c.Request.Method, c.Request.URL.Path, http.StatusOK)
	})
}
