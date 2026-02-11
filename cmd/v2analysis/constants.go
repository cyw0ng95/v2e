package main

// Log message constants for the analysis service
const (
	LogMsgServiceStarted       = "Analysis service started"
	LogMsgServiceReady         = "Analysis service ready for requests"
	LogMsgServiceShutdown      = "Analysis service shutting down"
	LogMsgGraphInitialized     = "Graph database initialized"
	LogMsgNodeAdded            = "Node added to graph: %s"
	LogMsgEdgeAdded            = "Edge added: %s -> %s (%s)"
	LogMsgGraphCleared         = "Graph cleared"
	LogMsgQueryingUEEStatus    = "Querying UEE status from meta service"
	LogMsgBuildingCVEGraph     = "Building CVE graph with limit: %d"
	LogMsgGraphBuildComplete   = "Graph build complete: %d nodes, %d edges added"
	LogMsgInvalidURN           = "Invalid URN: %s"
	LogMsgNodeNotFound         = "Node not found: %s"
	LogMsgPathNotFound         = "No path found between %s and %s"
	LogMsgPathFound            = "Path found between %s and %s (length: %d)"
	LogMsgFailedToQueryService = "Failed to query %s service: %v"
)
