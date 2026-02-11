package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/cyw0ng95/v2e/pkg/cce"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createGetCCEByIDHandler creates a handler for RPCGetCCEByID
func createGetCCEByIDHandler(store *cce.LocalCCEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CCEID string `json:"cce_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			logger.Debug("Processing GetCCEByID request failed due to malformed payload: %s", string(msg.Payload))
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.CCEID, "cce_id"); errResp != nil {
			return errResp, nil
		}
		logger.Debug("GetCCEByID request: cce_id=%s", req.CCEID)
		item, err := store.GetCCEByID(ctx, req.CCEID)
		if err != nil {
			logger.Warn("Failed to get CCE: %v (cce_id=%s)", err, req.CCEID)
			logger.Debug("Processing GetCCEByID request failed for CCE ID %s: %v", req.CCEID, err)
			return subprocess.NewErrorResponse(msg, "CCE not found"), nil
		}
		logger.Debug("Found CCE: %+v", item)
		logger.Debug("Processing GetCCEByID request completed successfully for CCE ID %s", req.CCEID)
		resp, err := subprocess.NewSuccessResponse(msg, item)
		if err != nil {
			logger.Warn("Failed to marshal CCE: %v (cce_id=%s)", err, req.CCEID)
			return subprocess.NewErrorResponse(msg, "failed to marshal CCE"), nil
		}
		logger.Debug("Marshalled CCE JSON: %s", string(resp.Payload))
		return resp, nil
	}
}

// createListCCEsHandler creates a handler for RPCListCCEs
func createListCCEsHandler(store *cce.LocalCCEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info("RPCListCCEs handler invoked with message ID: %s", msg.ID)
		var req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Search string `json:"search"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn("Failed to parse request: %v", errResp.Error)
				logger.Debug("Processing ListCCEs request failed due to malformed payload: %s", string(msg.Payload))
				return errResp, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		var items []cce.CCE
		var total int64
		var err error
		if req.Search != "" {
			items, total, err = store.SearchCCEs(ctx, req.Search, req.Offset, req.Limit)
		} else {
			items, total, err = store.ListCCEs(ctx, req.Offset, req.Limit)
		}
		if err != nil {
			logger.Warn("Failed to list CCEs: %v", err)
			logger.Debug("Processing ListCCEs request failed: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to list CCEs"), nil
		}
		logger.Debug("Processing ListCCEs request completed successfully: returned %d CCEs, total %d", len(items), total)
		resp := map[string]interface{}{
			"cces":   items,
			"offset": req.Offset,
			"limit":  req.Limit,
			"total":  total,
		}
		msgResp, err := subprocess.NewSuccessResponse(msg, resp)
		if err != nil {
			logger.Warn("Failed to marshal CCEs: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to marshal CCEs"), nil
		}
		return msgResp, nil
	}
}

// createImportCCEsHandler creates a handler for RPCImportCCEs
func createImportCCEsHandler(store *cce.LocalCCEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgStartingImportCCE, msg.CorrelationID)
		logger.Debug("RPCImportCCEs handler invoked. msg.ID=%s, correlation_id=%s", msg.ID, msg.CorrelationID)
		var req struct {
			Path string `json:"path"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.Path, "path"); errResp != nil {
			return errResp, nil
		}
		if _, err := os.Stat(req.Path); os.IsNotExist(err) {
			logger.Warn("CCE Excel file does not exist: %s", req.Path)
			return subprocess.NewErrorResponse(msg, "CCE file not found"), nil
		}
		count, err := store.ImportCCEsFromExcel(ctx, req.Path)
		if err != nil {
			logger.Error("Failed to import CCEs: %v", err)
			logger.Debug("Processing ImportCCEs request failed for path %s: %v", req.Path, err)
			return subprocess.NewErrorResponse(msg, "failed to import CCEs"), nil
		}
		logger.Info("Successfully imported %d CCE entries from %s", count, req.Path)
		logger.Debug("Processing ImportCCEs request completed successfully for path %s. correlation_id=%s", req.Path, msg.CorrelationID)
		return subprocess.NewSuccessResponse(msg, map[string]bool{"success": true})
	}
}

// createImportCCEHandler creates a handler for RPCImportCCE (single CCE entry)
func createImportCCEHandler(store *cce.LocalCCEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CCEData []byte `json:"cceData"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if len(req.CCEData) == 0 {
			return subprocess.NewErrorResponse(msg, "cceData is required"), nil
		}
		var entry cce.CCE
		if err := json.Unmarshal(req.CCEData, &entry); err != nil {
			logger.Warn("Failed to unmarshal CCE data: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to parse CCE data"), nil
		}
		if err := store.CreateCCE(ctx, entry); err != nil {
			logger.Warn("Failed to create CCE: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to create CCE"), nil
		}
		return subprocess.NewSuccessResponse(msg, map[string]bool{"success": true})
	}
}

// createCountCCEsHandler creates a handler for RPCCountCCEs
func createCountCCEsHandler(store *cce.LocalCCEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		count, err := store.CountCCEs(ctx)
		if err != nil {
			logger.Warn("Failed to count CCEs: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to count CCEs"), nil
		}
		resp, err := subprocess.NewSuccessResponse(msg, map[string]interface{}{"count": count})
		if err != nil {
			logger.Warn("Failed to marshal count response: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to marshal count"), nil
		}
		return resp, nil
	}
}

// createDeleteCCEHandler creates a handler for RPCDeleteCCE
func createDeleteCCEHandler(store *cce.LocalCCEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CCEID string `json:"cce_id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.CCEID, "cce_id"); errResp != nil {
			return errResp, nil
		}
		if err := store.DeleteCCE(ctx, req.CCEID); err != nil {
			logger.Warn("Failed to delete CCE: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to delete CCE"), nil
		}
		return subprocess.NewSuccessResponse(msg, map[string]bool{"success": true})
	}
}

// createUpdateCCEHandler creates a handler for RPCUpdateCCE
func createUpdateCCEHandler(store *cce.LocalCCEStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			CCEData []byte `json:"cceData"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if len(req.CCEData) == 0 {
			return subprocess.NewErrorResponse(msg, "cceData is required"), nil
		}
		var entry cce.CCE
		if err := json.Unmarshal(req.CCEData, &entry); err != nil {
			logger.Warn("Failed to unmarshal CCE data: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to parse CCE data"), nil
		}
		if entry.ID == "" {
			return subprocess.NewErrorResponse(msg, "CCE ID is required"), nil
		}
		if err := store.UpdateCCE(ctx, entry); err != nil {
			logger.Warn("Failed to update CCE: %v", err)
			return subprocess.NewErrorResponse(msg, "failed to update CCE"), nil
		}
		return subprocess.NewSuccessResponse(msg, map[string]bool{"success": true})
	}
}
