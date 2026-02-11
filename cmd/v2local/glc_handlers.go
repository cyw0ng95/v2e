package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/glc"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// ============================================================================
// Graph Handlers
// ============================================================================

// createGLCGraphHandler handles RPCGLCGraphCreate
func createGLCGraphHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			PresetID    string `json:"preset_id"`
			Nodes       string `json:"nodes"`
			Edges       string `json:"edges"`
			Viewport    string `json:"viewport"`
			Tags        string `json:"tags"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCGraphCreate request: %v", errResp.Error)
			return errResp, nil
		}

		if params.Name == "" {
			return subprocess.NewErrorResponse(msg, "name is required"), nil
		}
		if params.PresetID == "" {
			return subprocess.NewErrorResponse(msg, "preset_id is required"), nil
		}

		// Default empty arrays for nodes/edges
		if params.Nodes == "" {
			params.Nodes = "[]"
		}
		if params.Edges == "" {
			params.Edges = "[]"
		}

		graph, err := store.CreateGraph(ctx, params.Name, params.Description, params.PresetID, params.Nodes, params.Edges, params.Viewport)
		if err != nil {
			logger.Warn("Failed to create graph: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to create graph: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true, "graph": graph}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// getGLCGraphHandler handles RPCGLCGraphGet
func getGLCGraphHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			GraphID string `json:"graph_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCGraphGet request: %v", errResp.Error)
			return errResp, nil
		}

		if params.GraphID == "" {
			return subprocess.NewErrorResponse(msg, "graph_id is required"), nil
		}

		graph, err := store.GetGraph(ctx, params.GraphID)
		if err != nil {
			logger.Warn("Failed to get graph: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get graph: %v", err)), nil
		}

		resp := map[string]interface{}{"graph": graph}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// updateGLCGraphHandler handles RPCGLCGraphUpdate
func updateGLCGraphHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]interface{}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCGraphUpdate request: %v", errResp.Error)
			return errResp, nil
		}

		graphID, ok := params["graph_id"].(string)
		if !ok || graphID == "" {
			return subprocess.NewErrorResponse(msg, "graph_id is required"), nil
		}

		// Remove graph_id from updates
		delete(params, "graph_id")

		graph, err := store.UpdateGraph(ctx, graphID, params)
		if err != nil {
			logger.Warn("Failed to update graph: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to update graph: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true, "graph": graph}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// deleteGLCGraphHandler handles RPCGLCGraphDelete
func deleteGLCGraphHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			GraphID string `json:"graph_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCGraphDelete request: %v", errResp.Error)
			return errResp, nil
		}

		if params.GraphID == "" {
			return subprocess.NewErrorResponse(msg, "graph_id is required"), nil
		}

		if err := store.DeleteGraph(ctx, params.GraphID); err != nil {
			logger.Warn("Failed to delete graph: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete graph: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// listGLCGraphsHandler handles RPCGLCGraphList
func listGLCGraphsHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			PresetID string `json:"preset_id"`
			Offset   int    `json:"offset"`
			Limit    int    `json:"limit"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCGraphList request: %v", errResp.Error)
			return errResp, nil
		}

		if params.Limit <= 0 {
			params.Limit = 20
		}
		if params.Limit > 100 {
			params.Limit = 100
		}

		graphs, total, err := store.ListGraphs(ctx, params.PresetID, params.Offset, params.Limit)
		if err != nil {
			logger.Warn("Failed to list graphs: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list graphs: %v", err)), nil
		}

		resp := map[string]interface{}{
			"graphs": graphs,
			"total":  total,
			"offset": params.Offset,
			"limit":  params.Limit,
		}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// listRecentGLCGraphsHandler handles RPCGLCGraphListRecent
func listRecentGLCGraphsHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			Limit int `json:"limit"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCGraphListRecent request: %v", errResp.Error)
			return errResp, nil
		}

		if params.Limit <= 0 {
			params.Limit = 10
		}
		if params.Limit > 50 {
			params.Limit = 50
		}

		graphs, err := store.ListRecentGraphs(ctx, params.Limit)
		if err != nil {
			logger.Warn("Failed to list recent graphs: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list recent graphs: %v", err)), nil
		}

		resp := map[string]interface{}{"graphs": graphs}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// ============================================================================
// Version Handlers
// ============================================================================

// getGLCVersionHandler handles RPCGLCVersionGet
func getGLCVersionHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			GraphID string `json:"graph_id"`
			Version int    `json:"version"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCVersionGet request: %v", errResp.Error)
			return errResp, nil
		}

		if params.GraphID == "" {
			return subprocess.NewErrorResponse(msg, "graph_id is required"), nil
		}

		// Get graph to find DB ID
		graph, err := store.GetGraph(ctx, params.GraphID)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("graph not found: %v", err)), nil
		}

		version, err := store.GetVersion(ctx, graph.ID, params.Version)
		if err != nil {
			logger.Warn("Failed to get version: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get version: %v", err)), nil
		}

		resp := map[string]interface{}{"version": version}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// listGLCVersionsHandler handles RPCGLCVersionList
func listGLCVersionsHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			GraphID string `json:"graph_id"`
			Limit   int    `json:"limit"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCVersionList request: %v", errResp.Error)
			return errResp, nil
		}

		if params.GraphID == "" {
			return subprocess.NewErrorResponse(msg, "graph_id is required"), nil
		}

		// Get graph to find DB ID
		graph, err := store.GetGraph(ctx, params.GraphID)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("graph not found: %v", err)), nil
		}

		if params.Limit <= 0 {
			params.Limit = 20
		}

		versions, err := store.ListVersions(ctx, graph.ID, params.Limit)
		if err != nil {
			logger.Warn("Failed to list versions: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list versions: %v", err)), nil
		}

		resp := map[string]interface{}{"versions": versions}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// restoreGLCVersionHandler handles RPCGLCVersionRestore
func restoreGLCVersionHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			GraphID string `json:"graph_id"`
			Version int    `json:"version"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCVersionRestore request: %v", errResp.Error)
			return errResp, nil
		}

		if params.GraphID == "" {
			return subprocess.NewErrorResponse(msg, "graph_id is required"), nil
		}

		graph, err := store.RestoreVersion(ctx, params.GraphID, params.Version)
		if err != nil {
			logger.Warn("Failed to restore version: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to restore version: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true, "graph": graph}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// ============================================================================
// Preset Handlers
// ============================================================================

// createGLCPresetHandler handles RPCGLCPresetCreate
func createGLCPresetHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			Name        string          `json:"name"`
			Version     string          `json:"version"`
			Description string          `json:"description"`
			Author      string          `json:"author"`
			Theme       json.RawMessage `json:"theme"`
			Behavior    json.RawMessage `json:"behavior"`
			NodeTypes   json.RawMessage `json:"node_types"`
			Relations   json.RawMessage `json:"relations"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCPresetCreate request: %v", errResp.Error)
			return errResp, nil
		}

		if params.Name == "" {
			return subprocess.NewErrorResponse(msg, "name is required"), nil
		}

		preset, err := store.CreateUserPreset(ctx, params.Name, params.Version, params.Description, params.Author,
			string(params.Theme), string(params.Behavior), string(params.NodeTypes), string(params.Relations))
		if err != nil {
			logger.Warn("Failed to create preset: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to create preset: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true, "preset": preset}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// getGLCPresetHandler handles RPCGLCPresetGet
func getGLCPresetHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			PresetID string `json:"preset_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCPresetGet request: %v", errResp.Error)
			return errResp, nil
		}

		if params.PresetID == "" {
			return subprocess.NewErrorResponse(msg, "preset_id is required"), nil
		}

		preset, err := store.GetUserPreset(ctx, params.PresetID)
		if err != nil {
			logger.Warn("Failed to get preset: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get preset: %v", err)), nil
		}

		resp := map[string]interface{}{"preset": preset}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// updateGLCPresetHandler handles RPCGLCPresetUpdate
func updateGLCPresetHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params map[string]interface{}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCPresetUpdate request: %v", errResp.Error)
			return errResp, nil
		}

		presetID, ok := params["preset_id"].(string)
		if !ok || presetID == "" {
			return subprocess.NewErrorResponse(msg, "preset_id is required"), nil
		}

		delete(params, "preset_id")

		preset, err := store.UpdateUserPreset(ctx, presetID, params)
		if err != nil {
			logger.Warn("Failed to update preset: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to update preset: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true, "preset": preset}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// deleteGLCPresetHandler handles RPCGLCPresetDelete
func deleteGLCPresetHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			PresetID string `json:"preset_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCPresetDelete request: %v", errResp.Error)
			return errResp, nil
		}

		if params.PresetID == "" {
			return subprocess.NewErrorResponse(msg, "preset_id is required"), nil
		}

		if err := store.DeleteUserPreset(ctx, params.PresetID); err != nil {
			logger.Warn("Failed to delete preset: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to delete preset: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// listGLCPresetsHandler handles RPCGLCPresetList
func listGLCPresetsHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		presets, err := store.ListUserPresets(ctx)
		if err != nil {
			logger.Warn("Failed to list presets: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to list presets: %v", err)), nil
		}

		resp := map[string]interface{}{"presets": presets}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// ============================================================================
// Share Link Handlers
// ============================================================================

// createGLCShareLinkHandler handles RPCGLCShareCreateLink
func createGLCShareLinkHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			GraphID   string   `json:"graph_id"`
			Password  string   `json:"password"`
			ExpiresIn *float64 `json:"expires_in_hours"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCShareCreateLink request: %v", errResp.Error)
			return errResp, nil
		}

		if params.GraphID == "" {
			return subprocess.NewErrorResponse(msg, "graph_id is required"), nil
		}

		var expiresIn *time.Duration
		if params.ExpiresIn != nil && *params.ExpiresIn > 0 {
			d := time.Duration(*params.ExpiresIn * float64(time.Hour))
			expiresIn = &d
		}

		link, err := store.CreateShareLink(ctx, params.GraphID, params.Password, expiresIn)
		if err != nil {
			logger.Warn("Failed to create share link: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to create share link: %v", err)), nil
		}

		resp := map[string]interface{}{"success": true, "share_link": link}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// getGLCSharedGraphHandler handles RPCGLCShareGetShared
func getGLCSharedGraphHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			LinkID   string `json:"link_id"`
			Password string `json:"password"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCShareGetShared request: %v", errResp.Error)
			return errResp, nil
		}

		if params.LinkID == "" {
			return subprocess.NewErrorResponse(msg, "link_id is required"), nil
		}

		graph, err := store.GetGraphByShareLink(ctx, params.LinkID, params.Password)
		if err != nil {
			logger.Warn("Failed to get shared graph: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get shared graph: %v", err)), nil
		}

		resp := map[string]interface{}{"graph": graph}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// getGLCShareEmbedDataHandler handles RPCGLCShareGetEmbedData
func getGLCShareEmbedDataHandler(store *glc.Store, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var params struct {
			LinkID string `json:"link_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &params); errResp != nil {
			logger.Warn("Failed to parse RPCGLCShareGetEmbedData request: %v", errResp.Error)
			return errResp, nil
		}

		if params.LinkID == "" {
			return subprocess.NewErrorResponse(msg, "link_id is required"), nil
		}

		link, err := store.GetShareLink(ctx, params.LinkID)
		if err != nil {
			logger.Warn("Failed to get share link: %v", err)
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get share link: %v", err)), nil
		}

		graph, err := store.GetGraph(ctx, link.GraphID)
		if err != nil {
			return subprocess.NewErrorResponse(msg, fmt.Sprintf("failed to get graph: %v", err)), nil
		}

		resp := map[string]interface{}{
			"share_link": link,
			"graph":      graph,
		}
		return subprocess.NewSuccessResponse(msg, resp)
	}
}

// RegisterGLCHandlers registers all GLC RPC handlers
func RegisterGLCHandlers(sp *subprocess.Subprocess, store *glc.Store, logger *common.Logger) {
	// Graph handlers
	sp.RegisterHandler("RPCGLCGraphCreate", createGLCGraphHandler(store, logger))
	sp.RegisterHandler("RPCGLCGraphGet", getGLCGraphHandler(store, logger))
	sp.RegisterHandler("RPCGLCGraphUpdate", updateGLCGraphHandler(store, logger))
	sp.RegisterHandler("RPCGLCGraphDelete", deleteGLCGraphHandler(store, logger))
	sp.RegisterHandler("RPCGLCGraphList", listGLCGraphsHandler(store, logger))
	sp.RegisterHandler("RPCGLCGraphListRecent", listRecentGLCGraphsHandler(store, logger))

	// Version handlers
	sp.RegisterHandler("RPCGLCVersionGet", getGLCVersionHandler(store, logger))
	sp.RegisterHandler("RPCGLCVersionList", listGLCVersionsHandler(store, logger))
	sp.RegisterHandler("RPCGLCVersionRestore", restoreGLCVersionHandler(store, logger))

	// Preset handlers
	sp.RegisterHandler("RPCGLCPresetCreate", createGLCPresetHandler(store, logger))
	sp.RegisterHandler("RPCGLCPresetGet", getGLCPresetHandler(store, logger))
	sp.RegisterHandler("RPCGLCPresetUpdate", updateGLCPresetHandler(store, logger))
	sp.RegisterHandler("RPCGLCPresetDelete", deleteGLCPresetHandler(store, logger))
	sp.RegisterHandler("RPCGLCPresetList", listGLCPresetsHandler(store, logger))

	// Share handlers
	sp.RegisterHandler("RPCGLCShareCreateLink", createGLCShareLinkHandler(store, logger))
	sp.RegisterHandler("RPCGLCShareGetShared", getGLCSharedGraphHandler(store, logger))
	sp.RegisterHandler("RPCGLCShareGetEmbedData", getGLCShareEmbedDataHandler(store, logger))

	logger.Info("GLC handlers registered")
}
