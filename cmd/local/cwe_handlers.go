package main

import (
	"context"
	"os"

	"github.com/bytedance/sonic"
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
			logger.Error("Failed to parse request: %v", err)
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
			logger.Error("Failed to get CWE: %v (cwe_id=%s)", err, req.CWEID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "CWE not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug("Found CWE: %+v", item)
		jsonData, err := sonic.Marshal(item)
		if err != nil {
			logger.Error("Failed to marshal CWE: %v (cwe_id=%s)", err, req.CWEID)
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
				logger.Error("Failed to parse request: %v", err)
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
			logger.Error("Failed to list CWEs: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list CWEs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		resp := map[string]interface{}{
			"cwes":   items,
			"offset": req.Offset,
			"limit":  req.Limit,
			"total":  total,
		}
		jsonData, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal CWEs: %v", err)
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
			logger.Error("Failed to parse request: %v", err)
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
			logger.Error("Failed to import CWE from raw JSON: %v (path: %s)", err, req.Path)
			if _, statErr := os.Stat(req.Path); statErr != nil {
				logger.Error("CWE import file stat error: %v (path: %s)", statErr, req.Path)
			}
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to import CWEs",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       []byte(`{"success":true}`),
		}, nil
	}
}
