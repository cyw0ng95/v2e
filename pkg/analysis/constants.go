package analysis

const (
	// GraphFSM Log Messages
	LogMsgGraphStateTransition   = "GraphFSM state transition: %s -> %s"
	LogMsgGraphBuildStarted      = "Graph build operation started"
	LogMsgGraphBuildCompleted    = "Graph build operation completed"
	LogMsgGraphBuildFailed       = "Graph build operation failed: %v"
	LogMsgGraphAnalysisStarted   = "Graph analysis operation started"
	LogMsgGraphAnalysisCompleted = "Graph analysis operation completed"
	LogMsgGraphPersistStarted    = "Graph persistence operation started"
	LogMsgGraphPersistCompleted  = "Graph persistence operation completed"
	LogMsgGraphPersistFailed     = "Graph persistence operation failed: %v"
	LogMsgGraphCleared           = "Graph cleared"

	// AnalyzeFSM Log Messages
	LogMsgAnalyzeStateTransition     = "AnalyzeFSM state transition: %s -> %s"
	LogMsgAnalyzeServiceStarted      = "Analysis service started"
	LogMsgAnalyzeServicePaused       = "Analysis service paused"
	LogMsgAnalyzeServiceResumed      = "Analysis service resumed"
	LogMsgAnalyzeServiceStopped      = "Analysis service stopped"
	LogMsgAnalyzeResourceConstrained = "Resource constraint detected: %s"

	// GraphStore Log Messages
	LogMsgGraphStoreOpened   = "Graph store opened: %s"
	LogMsgGraphStoreClosed   = "Graph store closed"
	LogMsgGraphSaveStarted   = "Saving graph data"
	LogMsgGraphSaveCompleted = "Graph data saved: %d bytes"
	LogMsgGraphSaveFailed    = "Failed to save graph data: %v"
	LogMsgGraphLoadStarted   = "Loading graph data"
	LogMsgGraphLoadCompleted = "Graph data loaded: %d bytes"
	LogMsgGraphLoadFailed    = "Failed to load graph data: %v"
	LogMsgCheckpointSaved    = "Graph checkpoint saved: %s"
	LogMsgCheckpointLoaded   = "Graph checkpoint loaded: %s"
	LogMsgCheckpointFailed   = "Failed to save/load checkpoint: %v"
)
