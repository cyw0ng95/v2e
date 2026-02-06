/*
Package main implements the v2analysis RPC service for the UDA (Unified Data Analysis) framework.

Refer to service.md for the RPC API Specification and available methods.

Notes:
------
- Maintains an in-memory graph database of URN-based relationships
- Monitors UEE (Unified ETL Engine) status via RPC to the meta service
- Provides readonly access to data from local service for analysis
- Creates URN-based connection graphs between CVE, CWE, CAPEC, and ATT&CK objects
- All operations are readonly to ensure data integrity
- Service runs as a subprocess managed by the broker
- All communication is routed through the broker via RPC
*/
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/graph"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
	"github.com/cyw0ng95/v2e/pkg/urn"
	
	analysisfsm "github.com/cyw0ng95/v2e/pkg/analysis/fsm"
	analysisstorage "github.com/cyw0ng95/v2e/pkg/analysis/storage"
)

// AnalysisService manages the graph database and provides analysis capabilities
type AnalysisService struct {
	graph       *graph.Graph
	rpcClient   *rpc.Client
	logger      *common.Logger
	analyzeFSM  analysisfsm.AnalyzeFSM
	graphStore  *analysisstorage.GraphStore
	graphDBPath string
}

// NewAnalysisService creates a new analysis service
func NewAnalysisService(rpcClient *rpc.Client, logger *common.Logger, graphDBPath string) (*AnalysisService, error) {
	// Create FSM
	analyzeFSM := analysisfsm.NewAnalyzeFSM(logger)
	
	// Create graph storage
	graphStore, err := analysisstorage.NewGraphStore(graphDBPath, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create graph store: %w", err)
	}
	
	service := &AnalysisService{
		graph:       graph.New(),
		rpcClient:   rpcClient,
		logger:      logger,
		analyzeFSM:  analyzeFSM,
		graphStore:  graphStore,
		graphDBPath: graphDBPath,
	}
	
	// Try to load existing graph from storage
	if err := service.loadGraphFromStorage(); err != nil {
		logger.Warn("Failed to load graph from storage: %v", err)
		// Not a fatal error - continue with empty graph
	}
	
	// Start the FSM
	if err := analyzeFSM.Start(); err != nil {
		graphStore.Close()
		return nil, fmt.Errorf("failed to start FSM: %w", err)
	}
	
	return service, nil
}

// loadGraphFromStorage attempts to load the graph from storage
func (s *AnalysisService) loadGraphFromStorage() error {
	// Check if metadata exists
	metadata, err := s.graphStore.GetMetadata()
	if err != nil {
		// No existing graph
		return nil
	}
	
	s.logger.Info("Loading existing graph from storage (nodes: %d, edges: %d, last saved: %v)",
		metadata.NodeCount, metadata.EdgeCount, metadata.LastSaved)
	
	// Load the graph
	loadedGraph, err := s.graphStore.LoadGraph()
	if err != nil {
		return err
	}
	
	s.graph = loadedGraph
	return nil
}

// Close closes the analysis service and saves the graph
func (s *AnalysisService) Close() error {
	// Save graph before closing
	if s.graph.NodeCount() > 0 {
		s.logger.Info("Saving graph before shutdown...")
		if err := s.graphStore.SaveGraph(s.graph); err != nil {
			s.logger.Error("Failed to save graph: %v", err)
		}
	}
	
	// Stop FSM
	if err := s.analyzeFSM.Stop(); err != nil {
		s.logger.Warn("Error stopping FSM: %v", err)
	}
	
	// Close storage
	return s.graphStore.Close()
}

func main() {
	// Use standard startup pattern
	configStruct := subprocess.StandardStartupConfig{
		DefaultProcessID: "analysis",
		LogPrefix:        "[ANALYSIS] ",
	}
	sp, logger := subprocess.StandardStartup(configStruct)

	// Get graph database path from environment or use default
	graphDBPath := os.Getenv("GRAPH_DB_PATH")
	if graphDBPath == "" {
		graphDBPath = "analysis_graph.db"
	}

	// Create RPC client for communicating with other services
	rpcClient := rpc.NewClient(sp, logger, rpc.DefaultRPCTimeout)
	
	// Create analysis service with FSM and storage
	service, err := NewAnalysisService(rpcClient, logger, graphDBPath)
	if err != nil {
		logger.Error("Failed to create analysis service: %v", err)
		os.Exit(1)
	}
	defer service.Close()

	// Register RPC handlers
	sp.RegisterHandler("RPCGetGraphStats", createGetGraphStatsHandler(service))
	sp.RegisterHandler("RPCAddNode", createAddNodeHandler(service))
	sp.RegisterHandler("RPCAddEdge", createAddEdgeHandler(service))
	sp.RegisterHandler("RPCGetNode", createGetNodeHandler(service))
	sp.RegisterHandler("RPCGetNeighbors", createGetNeighborsHandler(service))
	sp.RegisterHandler("RPCFindPath", createFindPathHandler(service))
	sp.RegisterHandler("RPCGetNodesByType", createGetNodesByTypeHandler(service))
	sp.RegisterHandler("RPCGetUEEStatus", createGetUEEStatusHandler(service))
	sp.RegisterHandler("RPCBuildCVEGraph", createBuildCVEGraphHandler(service))
	sp.RegisterHandler("RPCClearGraph", createClearGraphHandler(service))
	
	// Register new FSM control handlers
	sp.RegisterHandler("RPCGetFSMState", createGetFSMStateHandler(service))
	sp.RegisterHandler("RPCPauseAnalysis", createPauseAnalysisHandler(service))
	sp.RegisterHandler("RPCResumeAnalysis", createResumeAnalysisHandler(service))
	sp.RegisterHandler("RPCSaveGraph", createSaveGraphHandler(service))
	sp.RegisterHandler("RPCLoadGraph", createLoadGraphHandler(service))

	logger.Info("UDA Analysis service started with FSM and persistence")
	logger.Info("Graph database: %s", graphDBPath)
	logger.Info("Graph database initialized")

	subprocess.RunWithDefaults(sp, logger)
	logger.Info("UDA Analysis service shutting down")
}

// createGetGraphStatsHandler returns statistics about the graph
func createGetGraphStatsHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		stats := map[string]interface{}{
			"node_count": service.graph.NodeCount(),
			"edge_count": service.graph.EdgeCount(),
		}
		return subprocess.NewSuccessResponse(msg, stats)
	}
}

// createAddNodeHandler adds a node to the graph
func createAddNodeHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			URN        string                 `json:"urn"`
			Properties map[string]interface{} `json:"properties"`
		}

		if err := subprocess.UnmarshalFast(msg.Payload, &params); err != nil {
			return subprocess.NewErrorResponse(msg, "invalid parameters: "+err.Error()), nil
		}

		u, err := urn.Parse(params.URN)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "invalid URN: "+err.Error()), nil
		}

		service.graph.AddNode(u, params.Properties)
		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"urn": u.String(),
		})
	}
}

// createAddEdgeHandler adds an edge between two nodes
func createAddEdgeHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			From       string                 `json:"from"`
			To         string                 `json:"to"`
			Type       string                 `json:"type"`
			Properties map[string]interface{} `json:"properties"`
		}

		if err := subprocess.UnmarshalFast(msg.Payload, &params); err != nil {
			return subprocess.NewErrorResponse(msg, "invalid parameters: "+err.Error()), nil
		}

		from, err := urn.Parse(params.From)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "invalid from URN: "+err.Error()), nil
		}

		to, err := urn.Parse(params.To)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "invalid to URN: "+err.Error()), nil
		}

		edgeType := graph.EdgeType(params.Type)
		if err := service.graph.AddEdge(from, to, edgeType, params.Properties); err != nil {
			return subprocess.NewErrorResponse(msg, "failed to add edge: "+err.Error()), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"from": from.String(),
			"to":   to.String(),
			"type": params.Type,
		})
	}
}

// createGetNodeHandler retrieves a node from the graph
func createGetNodeHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			URN string `json:"urn"`
		}

		if err := subprocess.UnmarshalFast(msg.Payload, &params); err != nil {
			return subprocess.NewErrorResponse(msg, "invalid parameters: "+err.Error()), nil
		}

		u, err := urn.Parse(params.URN)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "invalid URN: "+err.Error()), nil
		}

		node, exists := service.graph.GetNode(u)
		if !exists {
			return subprocess.NewErrorResponse(msg, "node not found"), nil
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"urn":        node.URN.String(),
			"properties": node.Properties,
		})
	}
}

// createGetNeighborsHandler gets all neighbors of a node
func createGetNeighborsHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			URN string `json:"urn"`
		}

		if err := subprocess.UnmarshalFast(msg.Payload, &params); err != nil {
			return subprocess.NewErrorResponse(msg, "invalid parameters: "+err.Error()), nil
		}

		u, err := urn.Parse(params.URN)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "invalid URN: "+err.Error()), nil
		}

		neighbors := service.graph.GetNeighbors(u)
		neighborStrings := make([]string, len(neighbors))
		for i, n := range neighbors {
			neighborStrings[i] = n.String()
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"neighbors": neighborStrings,
		})
	}
}

// createFindPathHandler finds a path between two nodes
func createFindPathHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			From string `json:"from"`
			To   string `json:"to"`
		}

		if err := subprocess.UnmarshalFast(msg.Payload, &params); err != nil {
			return subprocess.NewErrorResponse(msg, "invalid parameters: "+err.Error()), nil
		}

		from, err := urn.Parse(params.From)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "invalid from URN: "+err.Error()), nil
		}

		to, err := urn.Parse(params.To)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "invalid to URN: "+err.Error()), nil
		}

		path, found := service.graph.FindPath(from, to)
		if !found {
			return subprocess.NewErrorResponse(msg, "no path found"), nil
		}

		pathStrings := make([]string, len(path))
		for i, u := range path {
			pathStrings[i] = u.String()
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"path":   pathStrings,
			"length": len(path),
		})
	}
}

// createGetNodesByTypeHandler gets all nodes of a specific type
func createGetNodesByTypeHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			Type string `json:"type"`
		}

		if err := subprocess.UnmarshalFast(msg.Payload, &params); err != nil {
			return subprocess.NewErrorResponse(msg, "invalid parameters: "+err.Error()), nil
		}

		resourceType := urn.ResourceType(params.Type)
		nodes := service.graph.GetNodesByType(resourceType)

		nodeData := make([]map[string]interface{}, len(nodes))
		for i, node := range nodes {
			nodeData[i] = map[string]interface{}{
				"urn":        node.URN.String(),
				"properties": node.Properties,
			}
		}

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"nodes": nodeData,
			"count": len(nodes),
		})
	}
}

// createGetUEEStatusHandler queries the meta service for UEE status
func createGetUEEStatusHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		service.logger.Info("Querying UEE status from meta service")

		// Query meta service for active sessions
		resp, err := service.rpcClient.InvokeRPC(ctx, "meta", "RPCListSessions", nil)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "failed to query UEE status: "+err.Error()), nil
		}

		var sessions map[string]interface{}
		if err := subprocess.UnmarshalFast(resp.Payload, &sessions); err != nil {
			return subprocess.NewErrorResponse(msg, "failed to parse UEE response: "+err.Error()), nil
		}

		return subprocess.NewSuccessResponse(msg, sessions)
	}
}

// createBuildCVEGraphHandler builds a graph from CVE data in the local service
func createBuildCVEGraphHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			Limit int `json:"limit"`
		}

		if err := subprocess.UnmarshalFast(msg.Payload, &params); err != nil {
			// Use default limit
			params.Limit = 100
		}

		service.logger.Info("Building CVE graph with limit: %d", params.Limit)

		// Query local service for CVE data
		listParams := map[string]interface{}{
			"offset": 0,
			"limit":  params.Limit,
		}

		resp, err := service.rpcClient.InvokeRPC(ctx, "local", "RPCListCVEs", listParams)
		if err != nil {
			return subprocess.NewErrorResponse(msg, "failed to query CVE data: "+err.Error()), nil
		}

		var cveData struct {
			CVEs []map[string]interface{} `json:"cves"`
		}
		if err := subprocess.UnmarshalFast(resp.Payload, &cveData); err != nil {
			return subprocess.NewErrorResponse(msg, "failed to parse CVE response: "+err.Error()), nil
		}

		// Build graph from CVE data
		nodesAdded := 0
		edgesAdded := 0

		for _, cveMap := range cveData.CVEs {
			cveID, ok := cveMap["id"].(string)
			if !ok {
				continue
			}

			// Create CVE node
			cveURN, err := urn.New(urn.ProviderNVD, urn.TypeCVE, cveID)
			if err != nil {
				service.logger.Warn("Invalid CVE ID: %s", cveID)
				continue
			}

			service.graph.AddNode(cveURN, cveMap)
			nodesAdded++

			// Extract CWE references if available
			if cwes, ok := cveMap["cwe_ids"].([]interface{}); ok {
				for _, cweID := range cwes {
					cweIDStr, ok := cweID.(string)
					if !ok {
						continue
					}

					cweURN, err := urn.New(urn.ProviderMITRE, urn.TypeCWE, cweIDStr)
					if err != nil {
						continue
					}

					// Add CWE node if not exists
					if _, exists := service.graph.GetNode(cweURN); !exists {
						service.graph.AddNode(cweURN, map[string]interface{}{"id": cweIDStr})
						nodesAdded++
					}

					// Add edge from CVE to CWE
					if err := service.graph.AddEdge(cveURN, cweURN, graph.EdgeTypeReferences, nil); err == nil {
						edgesAdded++
					}
				}
			}
		}

		service.logger.Info("Graph build complete: %d nodes, %d edges added", nodesAdded, edgesAdded)

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"nodes_added": nodesAdded,
			"edges_added": edgesAdded,
			"total_nodes": service.graph.NodeCount(),
			"total_edges": service.graph.EdgeCount(),
		})
	}
}

// createClearGraphHandler clears all nodes and edges from the graph
func createClearGraphHandler(service *AnalysisService) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		service.graph.Clear()
		service.logger.Info("Graph cleared")

		return subprocess.NewSuccessResponse(msg, map[string]interface{}{
			"status": "cleared",
		})
	}
}

// createGetFSMStateHandler returns the current FSM state
func createGetFSMStateHandler(service *AnalysisService) subprocess.Handler {
return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
analyzeState := service.analyzeFSM.GetState()
graphFSM := service.analyzeFSM.GetGraphFSM()
graphState := graphFSM.GetState()

return subprocess.NewSuccessResponse(msg, map[string]interface{}{
"analyze_state": string(analyzeState),
"graph_state":   string(graphState),
})
}
}

// createPauseAnalysisHandler pauses the analysis service
func createPauseAnalysisHandler(service *AnalysisService) subprocess.Handler {
return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
if err := service.analyzeFSM.Pause(); err != nil {
return subprocess.NewErrorResponse(msg, "failed to pause analysis: "+err.Error()), nil
}

service.logger.Info("Analysis service paused")
return subprocess.NewSuccessResponse(msg, map[string]interface{}{
"status": "paused",
})
}
}

// createResumeAnalysisHandler resumes the analysis service
func createResumeAnalysisHandler(service *AnalysisService) subprocess.Handler {
return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
if err := service.analyzeFSM.Resume(); err != nil {
return subprocess.NewErrorResponse(msg, "failed to resume analysis: "+err.Error()), nil
}

service.logger.Info("Analysis service resumed")
return subprocess.NewSuccessResponse(msg, map[string]interface{}{
"status": "resumed",
})
}
}

// createSaveGraphHandler saves the graph to disk
func createSaveGraphHandler(service *AnalysisService) subprocess.Handler {
return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
graphFSM := service.analyzeFSM.GetGraphFSM()

// Start persistence
if err := graphFSM.StartPersist(); err != nil {
return subprocess.NewErrorResponse(msg, "failed to start persistence: "+err.Error()), nil
}

// Save graph
if err := service.graphStore.SaveGraph(service.graph); err != nil {
graphFSM.FailPersist(err)
return subprocess.NewErrorResponse(msg, "failed to save graph: "+err.Error()), nil
}

// Complete persistence
if err := graphFSM.CompletePersist(); err != nil {
return subprocess.NewErrorResponse(msg, "failed to complete persistence: "+err.Error()), nil
}

metadata, _ := service.graphStore.GetMetadata()
service.logger.Info("Graph saved to disk")

return subprocess.NewSuccessResponse(msg, map[string]interface{}{
"status":      "saved",
"node_count":  metadata.NodeCount,
"edge_count":  metadata.EdgeCount,
"last_saved":  metadata.LastSaved,
})
}
}

// createLoadGraphHandler loads the graph from disk
func createLoadGraphHandler(service *AnalysisService) subprocess.Handler {
return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
// Load graph
loadedGraph, err := service.graphStore.LoadGraph()
if err != nil {
return subprocess.NewErrorResponse(msg, "failed to load graph: "+err.Error()), nil
}

// Replace current graph
service.graph = loadedGraph
service.logger.Info("Graph loaded from disk")

return subprocess.NewSuccessResponse(msg, map[string]interface{}{
"status":      "loaded",
"node_count":  service.graph.NodeCount(),
"edge_count":  service.graph.EdgeCount(),
})
}
}
