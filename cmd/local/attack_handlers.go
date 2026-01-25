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
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn(LogMsgFailedParseReq, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug(LogMsgImportATTACKReceivedPath, req.Path)
		if req.Path == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "path is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if err := store.ImportFromXLSX(req.Path, req.Force); err != nil {
			logger.Warn(LogMsgFailedImportATTACKXLSX, err, req.Path)
			if _, statErr := os.Stat(req.Path); statErr != nil {
				logger.Warn(LogMsgATTACKImportStatError, statErr, req.Path)
			}
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to import ATT&CKs",
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

// createGetAttackTechniqueByIDHandler handles getting an ATT&CK technique by ID
func createGetAttackTechniqueByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.ID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug(LogMsgGetAttackTechniqueByIDReq, req.ID)
		item, err := store.GetTechniqueByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackTechnique, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "ATT&CK technique not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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
		jsonData, err := subprocess.MarshalFast(payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackTechnique, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK technique",
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

// createGetAttackTacticByIDHandler handles getting an ATT&CK tactic by ID
func createGetAttackTacticByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.ID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug(LogMsgGetAttackTacticByIDReq, req.ID)
		item, err := store.GetTacticByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackTactic, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "ATT&CK tactic not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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
		jsonData, err := subprocess.MarshalFast(payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackTactic, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK tactic",
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

// createGetAttackMitigationByIDHandler handles getting an ATT&CK mitigation by ID
func createGetAttackMitigationByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.ID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug(LogMsgGetAttackMitigationByIDReq, req.ID)
		item, err := store.GetMitigationByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackMitigation, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "ATT&CK mitigation not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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
		jsonData, err := subprocess.MarshalFast(payload)
		if err != nil {
			logger.Error(LogMsgFailedMarshalAttackMitigation, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK mitigation",
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

// createGetAttackSoftwareByIDHandler handles getting an ATT&CK software by ID
func createGetAttackSoftwareByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.ID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug(LogMsgGetAttackSoftwareByIDReq, req.ID)
		item, err := store.GetSoftwareByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackSoftware, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "ATT&CK software not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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
		jsonData, err := subprocess.MarshalFast(payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackSoftware, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK software",
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

// createGetAttackGroupByIDHandler handles getting an ATT&CK group by ID
func createGetAttackGroupByIDHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		var req struct {
			ID string `json:"id"`
		}
		if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
			logger.Warn("Failed to parse request: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to parse request",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		if req.ID == "" {
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "id is required",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Debug(LogMsgGetAttackGroupByIDReq, req.ID)
		item, err := store.GetGroupByID(ctx, req.ID)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackGroup, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "ATT&CK group not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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
		jsonData, err := subprocess.MarshalFast(payload)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalAttackGroup, err, req.ID)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK group",
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

// createListAttackTechniquesHandler handles listing ATT&CK techniques with pagination
func createListAttackTechniquesHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgProcessingListAttackTechniques, msg.ID, msg.CorrelationID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedParseListAttackTechniques, msg.ID, msg.CorrelationID, err)
				logger.Debug(LogMsgProcessingListAttackTechniquesFailed, msg.ID, string(msg.Payload))
				return &subprocess.Message{
					Type:          subprocess.MessageTypeError,
					ID:            msg.ID,
					Error:         "failed to parse request: " + err.Error(),
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
		logger.Info(LogMsgListAttackTechniquesParams, msg.ID, msg.CorrelationID, req.Offset, req.Limit)
		items, total, err := store.ListTechniquesPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackTechniques, msg.ID, msg.CorrelationID, err)
			logger.Debug(LogMsgProcessingListAttackTechniquesError, msg.ID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list ATT&CK techniques: " + err.Error(),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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

		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackTechniques, msg.ID, msg.CorrelationID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK techniques list: " + err.Error(),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Info(LogMsgSuccessListAttackTechniques, msg.ID, msg.CorrelationID, len(items), total)
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}, nil
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
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedParseReq, err)
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
		common.Info(LogMsgListAttackTacticsParams, req.Offset, req.Limit)
		items, total, err := store.ListTacticsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackTactics, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list ATT&CK tactics",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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

		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackTactics, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK tactics list",
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

// createListAttackMitigationsHandler handles listing ATT&CK mitigations with pagination
func createListAttackMitigationsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info(LogMsgListAttackMitigationsInvoked, msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedParseReq, err)
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
		common.Info(LogMsgListAttackMitigationsParams, req.Offset, req.Limit)
		items, total, err := store.ListMitigationsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackMitigations, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list ATT&CK mitigations",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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

		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackMitigations, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK mitigations list",
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

// createListAttackSoftwareHandler handles listing ATT&CK software with pagination
func createListAttackSoftwareHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info(LogMsgListAttackSoftwareInvoked, msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedParseReq, err)
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
		common.Info(LogMsgListAttackSoftwareParams, req.Offset, req.Limit)
		items, total, err := store.ListSoftwarePaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackSoftware, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list ATT&CK software",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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

		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackSoftware, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK software list",
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

// createListAttackGroupsHandler handles listing ATT&CK groups with pagination
func createListAttackGroupsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info(LogMsgListAttackGroupsInvoked, msg.ID)
		var req struct {
			Offset int `json:"offset"`
			Limit  int `json:"limit"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn(LogMsgFailedParseReq, err)
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
		common.Info(LogMsgListAttackGroupsParams, req.Offset, req.Limit)
		items, total, err := store.ListGroupsPaginated(ctx, req.Offset, req.Limit)
		if err != nil {
			logger.Warn(LogMsgFailedListAttackGroups, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to list ATT&CK groups",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
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

		jsonData, err := subprocess.MarshalFast(resp)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalListAttackGroups, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK groups list",
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

// createGetAttackImportMetadataHandler handles getting ATT&CK import metadata
func createGetAttackImportMetadataHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug(LogMsgGetAttackImportMetadataInvoked)

		meta, err := store.GetImportMetadata(ctx)
		if err != nil {
			logger.Warn(LogMsgFailedGetAttackImportMetadata, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "ATT&CK import metadata not found",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		// Build a client-friendly payload
		payload := map[string]interface{}{
			"id":             meta.ID,
			"imported_at":    meta.ImportedAt,
			"source_file":    meta.SourceFile,
			"total_records":  meta.TotalRecords,
			"import_version": meta.ImportVersion,
		}
		jsonData, err := subprocess.MarshalFast(payload)
		if err != nil {
			logger.Error(LogMsgFailedMarshalAttackImportMetadata, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK import metadata",
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
