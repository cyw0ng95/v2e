package main

import (
	"context"
	"os"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/cwe"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createGetCWEByIDHandler creates a handler for RPCGetCWEByID
func createGetCWEByIDHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CWEID string `json:"cwe_id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing GetCWEByID request failed due to malformed payload: %s", string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.CWEID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "cwe_id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("GetCWEByID request: cwe_id=%s", req.CWEID)
		item, err := store.GetByID(ctx, req.CWEID)
		if err != nil {
			logger.Warn("Failed to get CWE: %v (cwe_id=%s)", err, req.CWEID)
			logger.Debug("Processing GetCWEByID request failed for CWE ID %s: %v", req.CWEID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "CWE not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("Found CWE: %+v", item)
		logger.Debug("Processing GetCWEByID request completed successfully for CWE ID %s", req.CWEID)
		jsonData, err := subprocess.MarshalFast(item)
		if err != nil {
			logger.Warn("Failed to marshal CWE: %v (cwe_id=%s)", err, req.CWEID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal CWE",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("Marshalled CWE JSON: %s", string(jsonData))
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}, nil
	}
}

// createListCWEsHandler creates a handler for RPCListCWEs
func createListCWEsHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info("RPCListCWEs handler invoked with message ID: %s", msg.ID)
		var req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Search string `json:"search"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse request: %v", err)
				logger.Debug("Processing ListCWEs request failed due to malformed payload: %s", string(msg.Payload))
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         "failed to parse request",
					CorrelationID: msg.CorrelationID,
					Target:        msg.Source,
				}, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		// Currently, search is ignored. Add search logic here if needed.
		common.Info("Listing CWEs with offset=%d, limit=%d", req.Offset, req.Limit)
		items, total, err := store.ListCWEsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn("Failed to list CWEs: %v", err)
			logger.Debug("Processing ListCWEs request failed: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list CWEs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("Processing ListCWEs request completed successfully: returned %d CWEs, total %d", len(items), total)
		resp := map[string]interface{}{
			"cwes":   items,
			"offset": req.Offset,
			"limit":  req.Limit,
			"total":  total,
		}
		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Warn("Failed to marshal CWEs: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal CWEs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}, nil
	}
}

// createImportCWEsHandler creates a handler for RPCImportCWEs
func createImportCWEsHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCImportCWEs handler invoked")
		var req struct {
			Path string `json:"path"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			logger.Debug("Processing ImportCWEs request failed due to malformed payload: %s", string(msg.Payload))
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("RPCImportCWEs received path: %s", req.Path)
		if req.Path == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "path is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		err := store.ImportFromJSON(req.Path)
		if err != nil {
			logger.Warn("Failed to import CWE from raw JSON: %v (path: %s)", err, req.Path)
			logger.Debug("Processing ImportCWEs request failed for path %s: %v", req.Path, err)
			if _, statErr := os.Stat(req.Path); statErr != nil {
				logger.Warn("CWE import file stat error: %v (path: %s)", statErr, req.Path)
			}
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to import CWEs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("Processing ImportCWEs request completed successfully for path %s", req.Path)
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       []byte(`{"success":true}`),
		}, nil
	}
}
