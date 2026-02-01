package main

import (
	"context"
	"os"

	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createImportATTACKsHandler handles importing ATT&CK data from XLSX file
func createImportATTACKsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgImportATTACKInvoked)
		var req struct {
			Path  string `json:"path"`
			Force bool   `json:"force,omitempty"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn(LogMsgFailedParseReq, errResp.Error)
			return errResp, nil
		}
		logger.Debug(LogMsgImportATTACKReceivedPath, req.Path)
		if errResp := subprocess.RequireField(msg, req.Path, "path"); errResp != nil {
			return errResp, nil
		}
		if err := store.ImportFromXLSX(req.Path, req.Force); err != nil {
			logger.Warn(LogMsgFailedImportATTACKXLSX, err, req.Path)
			if _, statErr := os.Stat(req.Path); statErr != nil {
				logger.Warn(LogMsgATTACKImportStatError, statErr, req.Path)
			}
			return subprocess.NewErrorResponse(msg, "failed to import ATT&CKs"), nil
		}
		return subprocess.NewSuccessResponse(msg, map[string]bool{"success": true})
	}
}

// createGetAttackTechniqueByIDHandler handles getting an ATT&CK technique by ID
func createGetAttackTechniqueByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			return errResp, nil
		}
		logger.Debug(LogMsgGetAttackTechniqueByIDReq, req.ID)
		item, err := store.GetTechniqueByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackTechnique, err, req.ID)
			return subprocess.NewErrorResponse(msg, "ATT&CK technique not found"), nil
		}

		// Build a client-friendly payload
		payload := map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"description": item.Description,
			"domain":      item.Domain,
			"platform":    item.Platform,
			"created":     item.Created,
			"modified":    item.Modified,
			"revoked":     item.Revoked,
			"deprecated":  item.Deprecated,
		}
		resp, err := subprocess.NewSuccessResponse(msg, payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackTechnique, err, req.ID)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK technique"), nil
		}
		return resp, nil
	}
}

// createGetAttackTacticByIDHandler handles getting an ATT&CK tactic by ID
func createGetAttackTacticByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			return errResp, nil
		}
		logger.Debug(LogMsgGetAttackTacticByIDReq, req.ID)
		item, err := store.GetTacticByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackTactic, err, req.ID)
			return subprocess.NewErrorResponse(msg, "ATT&CK tactic not found"), nil
		}

		// Build a client-friendly payload
		payload := map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"description": item.Description,
			"domain":      item.Domain,
			"created":     item.Created,
			"modified":    item.Modified,
		}
		resp, err := subprocess.NewSuccessResponse(msg, payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackTactic, err, req.ID)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK tactic"), nil
		}
		return resp, nil
	}
}

// createGetAttackMitigationByIDHandler handles getting an ATT&CK mitigation by ID
func createGetAttackMitigationByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			return errResp, nil
		}
		logger.Debug(LogMsgGetAttackMitigationByIDReq, req.ID)
		item, err := store.GetMitigationByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackMitigation, err, req.ID)
			return subprocess.NewErrorResponse(msg, "ATT&CK mitigation not found"), nil
		}

		// Build a client-friendly payload
		payload := map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"description": item.Description,
			"domain":      item.Domain,
			"created":     item.Created,
			"modified":    item.Modified,
		}
		resp, err := subprocess.NewSuccessResponse(msg, payload)
		if err != nil {
			logger.Error(LogMsgFailedMarshalAttackMitigation, err, req.ID)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK mitigation"), nil
		}
		return resp, nil
	}
}

// createGetAttackSoftwareByIDHandler handles getting an ATT&CK software by ID
func createGetAttackSoftwareByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			return errResp, nil
		}
		logger.Debug(LogMsgGetAttackSoftwareByIDReq, req.ID)
		item, err := store.GetSoftwareByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackSoftware, err, req.ID)
			return subprocess.NewErrorResponse(msg, "ATT&CK software not found"), nil
		}

		// Build a client-friendly payload
		payload := map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"description": item.Description,
			"type":        item.Type,
			"domain":      item.Domain,
			"created":     item.Created,
			"modified":    item.Modified,
		}
		resp, err := subprocess.NewSuccessResponse(msg, payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackSoftware, err, req.ID)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK software"), nil
		}
		return resp, nil
	}
}

// createGetAttackGroupByIDHandler handles getting an ATT&CK group by ID
func createGetAttackGroupByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
			logger.Warn("Failed to parse request: %v", errResp.Error)
			return errResp, nil
		}
		if errResp := subprocess.RequireField(msg, req.ID, "id"); errResp != nil {
			return errResp, nil
		}
		logger.Debug(LogMsgGetAttackGroupByIDReq, req.ID)
		item, err := store.GetGroupByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackGroup, err, req.ID)
			return subprocess.NewErrorResponse(msg, "ATT&CK group not found"), nil
		}

		// Build a client-friendly payload
		payload := map[string]interface{}{
			"id":          item.ID,
			"name":        item.Name,
			"description": item.Description,
			"domain":      item.Domain,
			"created":     item.Created,
			"modified":    item.Modified,
		}
		resp, err := subprocess.NewSuccessResponse(msg, payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackGroup, err, req.ID)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK group"), nil
		}
		return resp, nil
	}
}

// createListAttackTechniquesHandler handles listing ATT&CK techniques with pagination
func createListAttackTechniquesHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgProcessingListAttackTechniques, msg.ID, msg.CorrelationID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn(LogMsgFailedParseListAttackTechniques, msg.ID, msg.CorrelationID, errResp.Error)
				logger.Debug(LogMsgProcessingListAttackTechniquesFailed, msg.ID, string(msg.Payload))
				return errResp, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		logger.Info(LogMsgListAttackTechniquesParams, msg.ID, msg.CorrelationID, req.Offset, req.Limit)
		items, total, err := store.ListTechniquesPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackTechniques, msg.ID, msg.CorrelationID, err)
			logger.Debug(LogMsgProcessingListAttackTechniquesError, msg.ID, err)
			return subprocess.NewErrorResponse(msg, "failed to list ATT&CK techniques: "+err.Error()), nil
		}
		// Map DB models to client-friendly objects
		mapped := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			logger.Debug(LogMsgMappingAttackTechnique, msg.ID, it.ID)
			mapped = append(mapped, map[string]interface{}{
				"id":          it.ID,
				"name":        it.Name,
				"description": it.Description,
				"domain":      it.Domain,
				"platform":    it.Platform,
				"created":     it.Created,
				"modified":    it.Modified,
				"revoked":     it.Revoked,
				"deprecated":  it.Deprecated,
			})
		}

		resp := map[string]interface{}{
			"techniques": mapped,
			"offset":     req.Offset,
			"limit":      req.Limit,
			"total":      total,
		}

		msgResp, err := subprocess.NewSuccessResponse(msg, resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackTechniques, msg.ID, msg.CorrelationID, err)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK techniques list: "+err.Error()), nil
		}

		logger.Info(LogMsgSuccessListAttackTechniques, msg.ID, msg.CorrelationID, len(items), total)
		return msgResp, nil
	}
}

// createListAttackTacticsHandler handles listing ATT&CK tactics with pagination
func createListAttackTacticsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info(LogMsgListAttackTacticsInvoked, msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn(LogMsgFailedParseReq, errResp.Error)
				return errResp, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info(LogMsgListAttackTacticsParams, req.Offset, req.Limit)
		items, total, err := store.ListTacticsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackTactics, err)
			return subprocess.NewErrorResponse(msg, "failed to list ATT&CK tactics"), nil
		}
		// Map DB models to client-friendly objects
		mapped := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			mapped = append(mapped, map[string]interface{}{
				"id":          it.ID,
				"name":        it.Name,
				"description": it.Description,
				"domain":      it.Domain,
				"created":     it.Created,
				"modified":    it.Modified,
			})
		}

		resp := map[string]interface{}{
			"tactics": mapped,
			"offset":  req.Offset,
			"limit":   req.Limit,
			"total":   total,
		}

		msgResp, err := subprocess.NewSuccessResponse(msg, resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackTactics, err)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK tactics list"), nil
		}

		return msgResp, nil
	}
}

// createListAttackMitigationsHandler handles listing ATT&CK mitigations with pagination
func createListAttackMitigationsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info(LogMsgListAttackMitigationsInvoked, msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn(LogMsgFailedParseReq, errResp.Error)
				return errResp, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info(LogMsgListAttackMitigationsParams, req.Offset, req.Limit)
		items, total, err := store.ListMitigationsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackMitigations, err)
			return subprocess.NewErrorResponse(msg, "failed to list ATT&CK mitigations"), nil
		}
		// Map DB models to client-friendly objects
		mapped := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			mapped = append(mapped, map[string]interface{}{
				"id":          it.ID,
				"name":        it.Name,
				"description": it.Description,
				"domain":      it.Domain,
				"created":     it.Created,
				"modified":    it.Modified,
			})
		}

		resp := map[string]interface{}{
			"mitigations": mapped,
			"offset":      req.Offset,
			"limit":       req.Limit,
			"total":       total,
		}

		msgResp, err := subprocess.NewSuccessResponse(msg, resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackMitigations, err)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK mitigations list"), nil
		}

		return msgResp, nil
	}
}

// createListAttackSoftwareHandler handles listing ATT&CK software with pagination
func createListAttackSoftwareHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info(LogMsgListAttackSoftwareInvoked, msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn(LogMsgFailedParseReq, errResp.Error)
				return errResp, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info(LogMsgListAttackSoftwareParams, req.Offset, req.Limit)
		items, total, err := store.ListSoftwarePaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackSoftware, err)
			return subprocess.NewErrorResponse(msg, "failed to list ATT&CK software"), nil
		}
		// Map DB models to client-friendly objects
		mapped := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			mapped = append(mapped, map[string]interface{}{
				"id":          it.ID,
				"name":        it.Name,
				"description": it.Description,
				"type":        it.Type,
				"domain":      it.Domain,
				"created":     it.Created,
				"modified":    it.Modified,
			})
		}

		resp := map[string]interface{}{
			"software": mapped,
			"offset":   req.Offset,
			"limit":    req.Limit,
			"total":    total,
		}

		msgResp, err := subprocess.NewSuccessResponse(msg, resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackSoftware, err)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK software list"), nil
		}

		return msgResp, nil
	}
}

// createListAttackGroupsHandler handles listing ATT&CK groups with pagination
func createListAttackGroupsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info(LogMsgListAttackGroupsInvoked, msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if errResp := subprocess.ParseRequest(msg, &req); errResp != nil {
				logger.Warn(LogMsgFailedParseReq, errResp.Error)
				return errResp, nil
			}
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info(LogMsgListAttackGroupsParams, req.Offset, req.Limit)
		items, total, err := store.ListGroupsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackGroups, err)
			return subprocess.NewErrorResponse(msg, "failed to list ATT&CK groups"), nil
		}
		// Map DB models to client-friendly objects
		mapped := make([]map[string]interface{}, 0, len(items))
		for _, it := range items {
			mapped = append(mapped, map[string]interface{}{
				"id":          it.ID,
				"name":        it.Name,
				"description": it.Description,
				"domain":      it.Domain,
				"created":     it.Created,
				"modified":    it.Modified,
			})
		}

		resp := map[string]interface{}{
			"groups": mapped,
			"offset": req.Offset,
			"limit":  req.Limit,
			"total":  total,
		}

		msgResp, err := subprocess.NewSuccessResponse(msg, resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackGroups, err)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK groups list"), nil
		}

		return msgResp, nil
	}
}

// createGetAttackImportMetadataHandler handles getting ATT&CK import metadata
func createGetAttackImportMetadataHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgGetAttackImportMetadataInvoked)

		meta, err := store.GetImportMetadata(ctx)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackImportMetadata, err)
			return subprocess.NewErrorResponse(msg, "ATT&CK import metadata not found"), nil
		}

		// Build a client-friendly payload
		payload := map[string]interface{}{
			"id":             meta.ID,
			"imported_at":    meta.ImportedAt,
			"source_file":    meta.SourceFile,
			"total_records":  meta.TotalRecords,
			"import_version": meta.ImportVersion,
		}
		resp, err := subprocess.NewSuccessResponse(msg, payload)
		if err != nil {
			logger.Error(LogMsgFailedMarshalAttackImportMetadata, err)
			return subprocess.NewErrorResponse(msg, "failed to marshal ATT&CK import metadata"), nil
		}
		return resp, nil
	}
}

// createGetAttackTechniqueHandler handles getting an ATT&CK technique by ID (alias for GetAttackTechniqueByID)
func createGetAttackTechniqueHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return createGetAttackTechniqueByIDHandler(store, logger)
}

// createGetAttackTacticHandler handles getting an ATT&CK tactic by ID (alias for GetAttackTacticByID)
func createGetAttackTacticHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return createGetAttackTacticByIDHandler(store, logger)
}

// createGetAttackMitigationHandler handles getting an ATT&CK mitigation by ID (alias for GetAttackMitigationByID)
func createGetAttackMitigationHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return createGetAttackMitigationByIDHandler(store, logger)
}

// createGetAttackSoftwareHandler handles getting an ATT&CK software by ID (alias for GetAttackSoftwareByID)
func createGetAttackSoftwareHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return createGetAttackSoftwareByIDHandler(store, logger)
}

// createGetAttackGroupHandler handles getting an ATT&CK group by ID (alias for GetAttackGroupByID)
func createGetAttackGroupHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return createGetAttackGroupByIDHandler(store, logger)
}
