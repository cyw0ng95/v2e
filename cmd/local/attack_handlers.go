package main

import (
	"context"
	"os"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/attack"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// createImportATTACKsHandler handles importing ATT&CK data from XLSX file
func createImportATTACKsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("RPCImportATTACKs handler invoked")
		var req struct {
			Path  string `json:"path"`
			Force bool   `json:"force,omitempty"`
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
		logger.Debug("RPCImportATTACKs received path: %s", req.Path)
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
			logger.Warn("Failed to import ATT&CK from XLSX: %v (path: %s)", err, req.Path)
			if _, statErr := os.Stat(req.Path); statErr != nil {
				logger.Warn("ATT&CK import file stat error: %v (path: %s)", statErr, req.Path)
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
		logger.Debug("GetAttackTechniqueByID request: id=%s", req.ID)
		item, err := store.GetTechniqueByID(ctx, req.ID)
		if err != nil {
			logger.Warn("Failed to get ATT&CK technique: %v (id=%s)", err, req.ID)
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
		jsonData, err := sonic.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK technique: %v (id=%s)", err, req.ID)
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
		logger.Debug("GetAttackTacticByID request: id=%s", req.ID)
		item, err := store.GetTacticByID(ctx, req.ID)
		if err != nil {
			logger.Warn("Failed to get ATT&CK tactic: %v (id=%s)", err, req.ID)
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
		jsonData, err := sonic.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK tactic: %v (id=%s)", err, req.ID)
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
		logger.Debug("GetAttackMitigationByID request: id=%s", req.ID)
		item, err := store.GetMitigationByID(ctx, req.ID)
		if err != nil {
			logger.Warn("Failed to get ATT&CK mitigation: %v (id=%s)", err, req.ID)
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
		jsonData, err := sonic.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK mitigation: %v (id=%s)", err, req.ID)
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
		logger.Debug("GetAttackSoftwareByID request: id=%s", req.ID)
		item, err := store.GetSoftwareByID(ctx, req.ID)
		if err != nil {
			logger.Warn("Failed to get ATT&CK software: %v (id=%s)", err, req.ID)
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
		jsonData, err := sonic.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK software: %v (id=%s)", err, req.ID)
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
		logger.Debug("GetAttackGroupByID request: id=%s", req.ID)
		item, err := store.GetGroupByID(ctx, req.ID)
		if err != nil {
			logger.Warn("Failed to get ATT&CK group: %v (id=%s)", err, req.ID)
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
		jsonData, err := sonic.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK group: %v (id=%s)", err, req.ID)
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

// createListAttackTechniquesHandler handles listing ATT&CK techniques with pagination and search
func createListAttackTechniquesHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Debug("Processing ListAttackTechniques request - Message ID: %s, Correlation ID: %s", msg.ID, msg.CorrelationID)
		var req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Search string `json:"search"`
		}
		if msg.Payload != nil {
			if err := subprocess.UnmarshalPayload(msg, &req); err != nil {
				logger.Warn("Failed to parse ListAttackTechniques request - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
				logger.Debug("Processing ListAttackTechniques request failed due to malformed payload - Message ID: %s, Payload: %s", msg.ID, string(msg.Payload))
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
		logger.Info("Processing ListAttackTechniques request - Message ID: %s, Correlation ID: %s, Offset: %d, Limit: %d, Search: %s", msg.ID, msg.CorrelationID, req.Offset, req.Limit, req.Search)

		var items []attack.AttackTechnique
		var total int64
		var err error

		if req.Search != "" {
			// Use search functionality
			items, total, err = store.SearchTechniques(ctx, req.Search, req.Offset, req.Limit)
		} else {
			// Use pagination functionality
			items, total, err = store.ListTechniquesPaginated(ctx, req.Offset, req.Limit)
		}

		if err != nil {
			logger.Warn("Failed to list ATT&CK techniques from store - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			logger.Debug("Processing ListAttackTechniques request failed - Message ID: %s, Error details: %v", msg.ID, err)
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
			logger.Debug("Mapping ATT&CK technique - Message ID: %s, Technique ID: %s", msg.ID, it.ID)
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
			"search":     req.Search, // Include search term in response
		}

		jsonData, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal ListAttackTechniques response - Message ID: %s, Correlation ID: %s, Error: %v", msg.ID, msg.CorrelationID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "failed to marshal ATT&CK techniques list: " + err.Error(),
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}

		logger.Info("Successfully processed ListAttackTechniques request - Message ID: %s, Correlation ID: %s, Returned: %d, Total: %d", msg.ID, msg.CorrelationID, len(items), total)
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
			Payload:       jsonData,
		}, nil
	}
}

// createListAttackTacticsHandler handles listing ATT&CK tactics with pagination and search
func createListAttackTacticsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info("RPCListAttackTactics handler invoked with message ID: %s", msg.ID)
		var req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Search string `json:"search"`
		}
		if msg.Payload != nil {
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
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info("Listing ATT&CK tactics with offset=%d, limit=%d, search=%s", req.Offset, req.Limit, req.Search)

		var items []attack.AttackTactic
		var total int64
		var err error

		if req.Search != "" {
			// Use search functionality
			items, total, err = store.SearchTactics(ctx, req.Search, req.Offset, req.Limit)
		} else {
			// Use pagination functionality
			items, total, err = store.ListTacticsPaginated(ctx, req.Offset, req.Limit)
		}

		if err != nil {
			logger.Warn("Failed to list ATT&CK tactics: %v", err)
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
			"search":  req.Search, // Include search term in response
		}

		jsonData, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK tactics list: %v", err)
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

// createListAttackMitigationsHandler handles listing ATT&CK mitigations with pagination and search
func createListAttackMitigationsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info("RPCListAttackMitigations handler invoked with message ID: %s", msg.ID)
		var req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Search string `json:"search"`
		}
		if msg.Payload != nil {
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
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info("Listing ATT&CK mitigations with offset=%d, limit=%d, search=%s", req.Offset, req.Limit, req.Search)

		var items []attack.AttackMitigation
		var total int64
		var err error

		if req.Search != "" {
			// Use search functionality
			items, total, err = store.SearchMitigations(ctx, req.Search, req.Offset, req.Limit)
		} else {
			// Use pagination functionality
			items, total, err = store.ListMitigationsPaginated(ctx, req.Offset, req.Limit)
		}

		if err != nil {
			logger.Warn("Failed to list ATT&CK mitigations: %v", err)
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
			"search":      req.Search, // Include search term in response
		}

		jsonData, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK mitigations list: %v", err)
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

// createListAttackSoftwareHandler handles listing ATT&CK software with pagination and search
func createListAttackSoftwareHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info("RPCListAttackSoftware handler invoked with message ID: %s", msg.ID)
		var req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Search string `json:"search"`
		}
		if msg.Payload != nil {
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
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info("Listing ATT&CK software with offset=%d, limit=%d, search=%s", req.Offset, req.Limit, req.Search)
		
		var items []attack.AttackSoftware
		var total int64
		var err error
		
		if req.Search != "" {
			// Use search functionality
			items, total, err = store.SearchSoftware(ctx, req.Search, req.Offset, req.Limit)
		} else {
			// Use pagination functionality
			items, total, err = store.ListSoftwarePaginated(ctx, req.Offset, req.Limit)
		}
		
		if err != nil {
			logger.Warn("Failed to list ATT&CK software: %v", err)
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
			"search":   req.Search, // Include search term in response
		}

		jsonData, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK software list: %v", err)
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

// createListAttackGroupsHandler handles listing ATT&CK groups with pagination and search
func createListAttackGroupsHandler(store *attack.LocalAttackStore, logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		common.Info("RPCListAttackGroups handler invoked with message ID: %s", msg.ID)
		var req struct {
			Offset int    `json:"offset"`
			Limit  int    `json:"limit"`
			Search string `json:"search"`
		}
		if msg.Payload != nil {
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
		}
		if req.Limit <= 0 || req.Limit > 1000 {
			req.Limit = 100
		}
		if req.Offset < 0 {
			req.Offset = 0
		}
		common.Info("Listing ATT&CK groups with offset=%d, limit=%d, search=%s", req.Offset, req.Limit, req.Search)
		
		var items []attack.AttackGroup
		var total int64
		var err error
		
		if req.Search != "" {
			// Use search functionality
			items, total, err = store.SearchGroups(ctx, req.Search, req.Offset, req.Limit)
		} else {
			// Use pagination functionality
			items, total, err = store.ListGroupsPaginated(ctx, req.Offset, req.Limit)
		}
		
		if err != nil {
			logger.Warn("Failed to list ATT&CK groups: %v", err)
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
			"search": req.Search, // Include search term in response
		}

		jsonData, err := sonic.Marshal(resp)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK groups list: %v", err)
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
		logger.Debug("RPCGetAttackImportMetadata handler invoked")

		meta, err := store.GetImportMetadata(ctx)
		if err != nil {
			logger.Warn("Failed to get ATT&CK import metadata: %v", err)
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
		jsonData, err := sonic.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal ATT&CK import metadata: %v", err)
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
