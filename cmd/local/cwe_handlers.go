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
		item, err := store.GetByID(ctx, req.CWEID)
		if err != nil {
			logger.Error("Failed to get CWE: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "CWE not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		jsonData, err := sonic.Marshal(item)
		if err != nil {
			logger.Error("Failed to marshal CWE: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal CWE",
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

// createListCWEsHandler creates a handler for RPCListCWEs
func createListCWEsHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		items, err := store.ListAll(ctx)
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
		jsonData, err := sonic.Marshal(items)
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
			logger.Error("Failed to import CWEs: %v", err)
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

// createGetImportCWEStatusHandler creates a handler for RPCGetImportCWEStatus
func createGetImportCWEStatusHandler(store *cwe.LocalCWEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		cweImportPath := os.Getenv("CWE_IMPORT_PATH")
		if cweImportPath == "" {
			cweImportPath = "assets/cwe-raw.json"
		}
		status := map[string]interface{}{
			"importPath": cweImportPath,
			"imported":   false,
			"count":      0,
			"error":      "",
		}
		if _, err := os.Stat(cweImportPath); err == nil {
			items, err := store.ListAll(ctx)
			if err == nil {
				status["imported"] = true
				status["count"] = len(items)
			} else {
				status["error"] = err.Error()
			}
		} else {
			status["error"] = "import file not found"
		}
		jsonData, err := sonic.Marshal(status)
		if err != nil {
			logger.Error("Failed to marshal import status: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal import status",
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

// importCWEsAtStartup imports CWEs from a JSON file at startup if the file exists
func importCWEsAtStartup(store *cwe.LocalCWEStore, logger *common.Logger) {
	cweImportPath := os.Getenv("CWE_IMPORT_PATH")
	if cweImportPath == "" {
		cweImportPath = "assets/cwe-raw.json"
	}
	if _, err := os.Stat(cweImportPath); err == nil {
		if err := store.ImportFromJSON(cweImportPath); err != nil {
			logger.Error("Failed to import CWEs from %s: %v", cweImportPath, err)
		} else {
			logger.Info("Imported CWEs from %s", cweImportPath)
		}
	} else {
		logger.Warn("CWE import file not found: %s", cweImportPath)
	}
}
